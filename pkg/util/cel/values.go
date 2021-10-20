// Copyright 2021 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cel

import (
	"fmt"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"reflect"
	"sync"
)

// UnstructuredToVal converts a Kubernetes unstructured data element to a CEL Val.
func UnstructuredToVal(unstructured interface{}, schema *spec.Schema) ref.Val {
	if schema.Type.Contains("object") {
		m, ok := unstructured.(map[string]interface{})
		if !ok {
			types.NewErr("invalid data, expected map[string]interface{} to match the provided schema with type=object")
		}
		if schema.AdditionalProperties != nil && schema.AdditionalProperties.Allows == true && schema.AdditionalProperties.Schema != nil {
			return &unstructuredMap{
				value:      m,
				schema:     schema,
				propSchema: func(key string) *spec.Schema { return schema.AdditionalProperties.Schema },
			}
		} else if schema.Properties != nil {
			return &unstructuredMap{
				value:      m,
				schema:     schema,
				propSchema: func(key string) *spec.Schema { schema := schema.Properties[key]; return &schema },
			}
		} else {
			types.NewErr("invalid object type, expected either Properties or AdditionalProperties with Allows=true and non-empty Schema")
		}
	}
	if schema.Type.Contains("array") {
		if schema.Items == nil || schema.Items.Schema == nil {
			types.NewErr("invalid array type, expected Items with a non-empty Schema")
		}
		l, ok := unstructured.([]interface{})
		if ok {
			types.NewErr("invalid data, expected []interface{} to match the provided schema with type=array")
		}
		typedList := unstructuredList{elements: l, itemsSchema: schema.Items.Schema}
		if listType, ok := schema.Extensions.GetString("x-kubernetes-list-type"); ok {
			switch listType {
			case "map":
				mapKeys, ok := schema.Extensions.GetStringSlice("x-kubernetes-list-map-keys")
				if !ok {
					types.NewErr("invalid map list type, expected x-kubernetes-list-map-keys to be provided when x-kubernetes-list-type=map")
				}
				return &unstructuredMapList{unstructuredList: typedList, keyFields: mapKeys}
			case "set":
				return &unstructuredSetList{unstructuredList: typedList}
			}
		}
		return &typedList
	}
	return types.DefaultTypeAdapter.NativeToValue(unstructured)
}

// unstructuredMapList represents an unstructured data instance of an OpenAPI array with x-kubernetes-list-type=map.
type unstructuredMapList struct {
	unstructuredList
	keyFields []string

	sync.Once // for for lazy load of mapOfList since it is only needed if Equals is called
	mapOfList map[interface{}]interface{}
}

func (t *unstructuredMapList) getMap() map[interface{}]interface{} {
	t.Do(func() {
		t.mapOfList = make(map[interface{}]interface{}, len(t.elements))
		for _, e := range t.elements {
			t.mapOfList[t.toMapKey(e)] = e
		}
	})
	return t.mapOfList
}

// toMapKey returns a valid golang map key for the given element of the map list.
// element must be a valid map list entry where all map key fields are scalar types (which are comparable in go
// and valid for use in a golang map key).
func (t *unstructuredMapList) toMapKey(element interface{}) interface{} {
	eObj, ok := element.(map[string]interface{})
	if !ok {
		return types.NewErr("unexpected data format for element of array with x-kubernetes-list-type=map: %T", element)
	}
	// Arrays are comparable in go and may be used as map keys, but maps and slices are not.
	// So we can special case small numbers of key fields as arrays and fall back to serialization
	// for larger numbers of key fields
	if len(t.keyFields) == 1 {
		return eObj[t.keyFields[0]]
	}
	if len(t.keyFields) == 2 {
		return [2]interface{}{eObj[t.keyFields[0]], eObj[t.keyFields[1]]}
	}
	if len(t.keyFields) == 3 {
		return [3]interface{}{eObj[t.keyFields[0]], eObj[t.keyFields[1]], eObj[t.keyFields[2]]}
	}

	key := make([]interface{}, len(t.keyFields))
	for i, kf := range t.keyFields {
		key[i] = eObj[kf]
	}
	return fmt.Sprintf("%v", key)
}

func (t *unstructuredMapList) Equal(other ref.Val) ref.Val {
	oMapList, ok := other.(*unstructuredMapList)
	if !ok || t.itemsSchema != oMapList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}
	sz := types.Int(len(t.elements))
	if sz != oMapList.Size() {
		return types.False
	}
	oMap := oMapList.getMap()
	for _, v := range t.elements {
		k := t.toMapKey(v)
		oVal, ok := oMap[k]
		if !ok {
			return types.False
		}
		if UnstructuredToVal(v, t.itemsSchema).Equal(UnstructuredToVal(oVal, t.itemsSchema)) != types.True {
			return types.False
		}
	}
	return types.True
}

func (t *unstructuredMapList) Add(other ref.Val) ref.Val {
	// TODO: use same merge routine SSA uses
	oMapList, ok := other.(*unstructuredMapList)
	if !ok || t.itemsSchema != oMapList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}
	elements := make([]interface{}, 0, len(t.elements)+len(oMapList.elements))
	oMap := oMapList.getMap()
	overwrites := map[interface{}]struct{}{}
	for _, e := range t.elements {
		k := t.toMapKey(e)
		if oe, ok := oMap[k]; ok {
			elements = append(elements, oe)
			overwrites[k] = struct{}{}
		} else {
			elements = append(elements, e)
		}
	}
	for _, oe := range oMapList.elements {
		k := t.toMapKey(oe)
		if _, ok := overwrites[k]; ok {
			continue
		}
		elements = append(elements, oe)
	}
	return &unstructuredMapList{
		unstructuredList: unstructuredList{elements: elements, itemsSchema: t.itemsSchema},
		keyFields:        t.keyFields,
	}
}

// unstructuredSetList represents an unstructured data instance of an OpenAPI array with x-kubernetes-list-type=set.
type unstructuredSetList struct {
	unstructuredList
	keyFields []string

	sync.Once // for for lazy load of setOfList since it is only needed if Equals is called
	set       map[interface{}]struct{}
}

func (t *unstructuredSetList) getSet() map[interface{}]struct{} {
	// sets are only allowed to contain scalar elements, which are comparable in go, and can safely be used as
	// golang map keys
	t.Do(func() {
		t.set = make(map[interface{}]struct{}, len(t.elements))
		for _, e := range t.elements {
			t.set[e] = struct{}{}
		}
	})
	return t.set
}

func (t *unstructuredSetList) Equal(other ref.Val) ref.Val {
	oSetList, ok := other.(*unstructuredSetList)
	if !ok || t.itemsSchema != oSetList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}
	sz := types.Int(len(t.elements))
	if sz != oSetList.Size() {
		return types.False
	}
	oSet := oSetList.getSet()
	for _, v := range t.elements {
		_, ok := oSet[v]
		if !ok {
			return types.False
		}
	}
	return types.True
}

func (t *unstructuredSetList) Add(other ref.Val) ref.Val {
	oSetList, ok := other.(*unstructuredSetList)
	if !ok || t.itemsSchema != oSetList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}

	elements := make([]interface{}, 0, len(t.elements)+len(oSetList.elements))
	oSet := oSetList.getSet()
	intersection := map[interface{}]struct{}{}
	for _, e := range t.elements {
		if _, ok := oSet[e]; ok {
			intersection[e] = struct{}{}
		}
		elements = append(elements, e)
	}
	for _, oe := range oSetList.elements {
		if _, ok := intersection[oe]; ok {
			continue
		}
		elements = append(elements, oe)
	}
	return &unstructuredMapList{
		unstructuredList: unstructuredList{elements: elements, itemsSchema: t.itemsSchema},
		keyFields:        t.keyFields,
	}
}

// unstructuredList represents an unstructured data instance of an OpenAPI array with x-kubernetes-list-type=atomic (the default).
type unstructuredList struct {
	elements    []interface{}
	itemsSchema *spec.Schema
}

var _ = traits.Lister(&unstructuredList{})

func (t *unstructuredList) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	switch typeDesc.Kind() {
	case reflect.Slice:
		return t.elements, nil
	}
	return nil, fmt.Errorf("type conversion error from '%s' to '%s'", t.Type(), typeDesc)
}

func (t *unstructuredList) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.ListType:
		return t
	}
	return types.NewErr("type conversion error from '%s' to '%s'", t.Type(), typeValue.TypeName())
}

func (t *unstructuredList) Equal(other ref.Val) ref.Val {
	oList, ok := other.(*unstructuredList)
	if !ok || t.itemsSchema != oList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}
	sz := types.Int(len(t.elements))
	if sz != oList.Size() {
		return types.False
	}
	for i := types.Int(0); i < sz; i++ {
		if t.Get(i).Equal(oList.Get(i)) != types.True {
			return types.False
		}
	}
	return types.True
}

func (t *unstructuredList) Type() ref.Type {
	return types.ListType
}

func (t *unstructuredList) Value() interface{} {
	return t.elements
}

func (t *unstructuredList) Add(other ref.Val) ref.Val {
	oList, ok := other.(*unstructuredList)
	if !ok || t.itemsSchema != oList.itemsSchema {
		return types.MaybeNoSuchOverloadErr(other)
	}

	elements := append(t.elements, oList.elements...)
	return &unstructuredList{elements: elements, itemsSchema: t.itemsSchema}
}

func (t *unstructuredList) Contains(val ref.Val) ref.Val {
	if types.IsUnknownOrError(val) {
		return val
	}
	var err ref.Val
	sz := len(t.elements)
	for i := 0; i < sz; i++ {
		elem := UnstructuredToVal(t.elements[i], t.itemsSchema)
		cmp := elem.Equal(val)
		b, ok := cmp.(types.Bool)
		if !ok && err == nil {
			err = types.MaybeNoSuchOverloadErr(cmp)
		}
		if b == types.True {
			return types.True
		}
	}
	if err != nil {
		return err
	}
	return types.False
}

func (t *unstructuredList) Get(idx ref.Val) ref.Val {
	iv, isInt := idx.(types.Int)
	if !isInt {
		return types.ValOrErr(idx, "unsupported index: %v", idx)
	}
	i := int(iv)
	if i < 0 || i >= len(t.elements) {
		return types.NewErr("index out of bounds: %v", idx)
	}
	return UnstructuredToVal(t.elements[i], t.itemsSchema)
}

func (t *unstructuredList) Iterator() traits.Iterator {
	items := make([]ref.Val, len(t.elements))
	for i, item := range t.elements {
		itemCopy := item
		items[i] = UnstructuredToVal(itemCopy, t.itemsSchema)
	}
	return &listIterator{unstructuredList: t, items: items}
}

type listIterator struct {
	*unstructuredList
	items []ref.Val
	idx   int
}

func (it *listIterator) HasNext() ref.Val {
	return types.Bool(it.idx < len(it.items))
}

func (it *listIterator) Next() ref.Val {
	item := it.items[it.idx]
	it.idx++
	return item
}

func (t unstructuredList) Size() ref.Val {
	return types.Int(len(t.elements))
}

// unstructuredMap represented an unstructured data instance of an OpenAPI object.
type unstructuredMap struct {
	value  map[string]interface{}
	schema *spec.Schema
	// propSchema finds the schema to use for a particular map key.
	propSchema func(key string) *spec.Schema
}

var _ = traits.Mapper(&unstructuredMap{})

func (t unstructuredMap) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	switch typeDesc.Kind() {
	case reflect.Map:
		return t.value, nil
	}
	return nil, fmt.Errorf("type conversion error from '%s' to '%s'", t.Type(), typeDesc)
}

func (t unstructuredMap) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.MapType:
		return t
	}
	return types.NewErr("type conversion error from '%s' to '%s'", t.Type(), typeValue.TypeName())
}

func (t unstructuredMap) Equal(other ref.Val) ref.Val {
	oMap, isMap := other.(traits.Mapper)
	if !isMap {
		return types.MaybeNoSuchOverloadErr(other)
	}
	if t.Size() != oMap.Size() {
		return types.False
	}
	for key, value := range t.value {
		ov, found := oMap.Find(types.String(key))
		if !found {
			return types.False
		}
		v := UnstructuredToVal(value, t.propSchema(key))
		vEq := v.Equal(ov)
		if vEq != types.True {
			return vEq
		}
	}
	return types.True
}

func (t unstructuredMap) Type() ref.Type {
	return types.MapType
}

func (t unstructuredMap) Value() interface{} {
	return t.value
}

func (t unstructuredMap) Contains(key ref.Val) ref.Val {
	v, found := t.Find(key)
	if v != nil && types.IsUnknownOrError(v) {
		return v
	}

	return types.Bool(found)
}

func (t *unstructuredMap) Get(key ref.Val) ref.Val {
	v, found := t.Find(key)
	if found {
		return v
	}
	return types.ValOrErr(key, "no such key: %v", key)
}

func (t *unstructuredMap) Iterator() traits.Iterator {
	keys := make([]ref.Val, len(t.value))
	i := 0
	for k := range t.value {
		keys[i] = types.String(k)
		i++
	}
	return &mapIterator{unstructuredMap: t, keys: keys}
}

type mapIterator struct {
	*unstructuredMap
	keys []ref.Val
	idx  int
}

func (it *mapIterator) HasNext() ref.Val {
	return types.Bool(it.idx < len(it.keys))
}

func (it *mapIterator) Next() ref.Val {
	key := it.keys[it.idx]
	it.idx++
	return key
}

func (t *unstructuredMap) Size() ref.Val {
	return types.Int(len(t.value))
}

func (t *unstructuredMap) Find(key ref.Val) (ref.Val, bool) {
	keyStr, ok := key.(types.String)
	if !ok {
		return types.MaybeNoSuchOverloadErr(key), true
	}
	k := keyStr.Value().(string)
	if v, ok := t.value[k]; ok {
		return UnstructuredToVal(v, t.propSchema(k)), true
	}

	return nil, false
}
