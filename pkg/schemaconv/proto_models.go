/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schemaconv

import (
	"errors"
	"path"
	"strings"

	"k8s.io/kube-openapi/pkg/util/proto"
	"sigs.k8s.io/structured-merge-diff/v4/schema"
)

// ToSchema converts openapi definitions into a schema suitable for structured
// merge (i.e. kubectl apply v2).
func ToSchema(models proto.Models) (*schema.Schema, error) {
	return ToSchemaWithPreserveUnknownFields(models, false)
}

// ToSchemaWithPreserveUnknownFields converts openapi definitions into a schema suitable for structured
// merge (i.e. kubectl apply v2), it will preserve unknown fields if specified.
func ToSchemaWithPreserveUnknownFields(models proto.Models, preserveUnknownFields bool) (*schema.Schema, error) {
	c := convert{
		preserveUnknownFields: preserveUnknownFields,
		output:                &schema.Schema{},
	}
	for _, name := range models.ListModels() {
		model := models.LookupModel(name)

		var a schema.Atom
		c2 := c.push(name, &a)
		model.Accept(c2)
		c.pop(c2)

		c.insertTypeDef(name, a)
	}

	if len(c.errorMessages) > 0 {
		return nil, errors.New(strings.Join(c.errorMessages, "\n"))
	}

	c.addCommonTypes()
	return c.output, nil
}

func (c *convert) makeRef(model proto.Schema, preserveUnknownFields bool) schema.TypeRef {
	var tr schema.TypeRef
	if r, ok := model.(*proto.Ref); ok {
		if r.Reference() == "io.k8s.apimachinery.pkg.runtime.RawExtension" {
			return schema.TypeRef{
				NamedType: &untypedName,
			}
		}
		// reference a named type
		_, n := path.Split(r.Reference())
		tr.NamedType = &n

		ext := model.GetExtensions()
		if val, ok := ext["x-kubernetes-map-type"]; ok {
			switch val {
			case "atomic":
				relationship := schema.Atomic
				tr.ElementRelationship = &relationship
			case "granular":
				relationship := schema.Separable
				tr.ElementRelationship = &relationship
			default:
				c.reportError("unknown map type %v", val)
			}
		}
	} else {
		// compute the type inline
		c2 := c.push("inlined in "+c.currentName, &tr.Inlined)
		c2.preserveUnknownFields = preserveUnknownFields
		model.Accept(c2)
		c.pop(c2)

		if tr == (schema.TypeRef{}) {
			// emit warning?
			tr.NamedType = &untypedName
		}
	}
	return tr
}

func (c *convert) VisitKind(k *proto.Kind) {
	preserveUnknownFields := c.preserveUnknownFields
	if p, ok := k.GetExtensions()["x-kubernetes-preserve-unknown-fields"]; ok && p == true {
		preserveUnknownFields = true
	}

	a := c.top()
	a.Map = &schema.Map{}
	for _, name := range k.FieldOrder {
		member := k.Fields[name]
		tr := c.makeRef(member, preserveUnknownFields)
		a.Map.Fields = append(a.Map.Fields, schema.StructField{
			Name:    name,
			Type:    tr,
			Default: member.GetDefault(),
		})
	}

	unions, err := makeUnions(k.GetExtensions())
	if err != nil {
		c.reportError(err.Error())
		return
	}
	// TODO: We should check that the fields and discriminator
	// specified in the union are actual fields in the struct.
	a.Map.Unions = unions

	if preserveUnknownFields {
		a.Map.ElementType = schema.TypeRef{
			NamedType: &deducedName,
		}
	}

	ext := k.GetExtensions()
	if val, ok := ext["x-kubernetes-map-type"]; ok {
		switch val {
		case "atomic":
			a.Map.ElementRelationship = schema.Atomic
		case "granular":
			a.Map.ElementRelationship = schema.Separable
		default:
			c.reportError("unknown map type %v", val)
		}
	}
}

func (c *convert) VisitArray(a *proto.Array) {
	atom := c.top()
	atom.List = &schema.List{
		ElementRelationship: schema.Atomic,
	}
	l := atom.List
	l.ElementType = c.makeRef(a.SubType, c.preserveUnknownFields)

	ext := a.GetExtensions()

	if val, ok := ext["x-kubernetes-list-type"]; ok {
		if val == "atomic" {
			l.ElementRelationship = schema.Atomic
		} else if val == "set" {
			l.ElementRelationship = schema.Associative
		} else if val == "map" {
			l.ElementRelationship = schema.Associative
			if keys, ok := ext["x-kubernetes-list-map-keys"]; ok {
				if keyNames, ok := toStringSlice(keys); ok {
					l.Keys = keyNames
				} else {
					c.reportError("uninterpreted map keys: %#v", keys)
				}
			} else {
				c.reportError("missing map keys")
			}
		} else {
			c.reportError("unknown list type %v", val)
			l.ElementRelationship = schema.Atomic
		}
	} else if val, ok := ext["x-kubernetes-patch-strategy"]; ok {
		if val == "merge" || val == "merge,retainKeys" {
			l.ElementRelationship = schema.Associative
			if key, ok := ext["x-kubernetes-patch-merge-key"]; ok {
				if keyName, ok := key.(string); ok {
					l.Keys = []string{keyName}
				} else {
					c.reportError("uninterpreted merge key: %#v", key)
				}
			} else {
				// It's not an error for this to be absent, it
				// means it's a set.
			}
		} else if val == "retainKeys" {
		} else {
			c.reportError("unknown patch strategy %v", val)
			l.ElementRelationship = schema.Atomic
		}
	}
}

func (c *convert) VisitMap(m *proto.Map) {
	a := c.top()
	a.Map = &schema.Map{}
	a.Map.ElementType = c.makeRef(m.SubType, c.preserveUnknownFields)

	ext := m.GetExtensions()
	if val, ok := ext["x-kubernetes-map-type"]; ok {
		switch val {
		case "atomic":
			a.Map.ElementRelationship = schema.Atomic
		case "granular":
			a.Map.ElementRelationship = schema.Separable
		default:
			c.reportError("unknown map type %v", val)
		}
	}
}

func (c *convert) VisitPrimitive(p *proto.Primitive) {
	a := c.top()
	if c.currentName == quantityResource {
		a.Scalar = ptr(schema.Scalar("untyped"))
	} else {
		switch p.Type {
		case proto.Integer:
			a.Scalar = ptr(schema.Numeric)
		case proto.Number:
			a.Scalar = ptr(schema.Numeric)
		case proto.String:
			switch p.Format {
			case "":
				a.Scalar = ptr(schema.String)
			case "byte":
				// byte really means []byte and is encoded as a string.
				a.Scalar = ptr(schema.String)
			case "int-or-string":
				a.Scalar = ptr(schema.Scalar("untyped"))
			case "date-time":
				a.Scalar = ptr(schema.Scalar("untyped"))
			default:
				a.Scalar = ptr(schema.Scalar("untyped"))
			}
		case proto.Boolean:
			a.Scalar = ptr(schema.Boolean)
		default:
			a.Scalar = ptr(schema.Scalar("untyped"))
		}
	}
}

func (c *convert) VisitArbitrary(a *proto.Arbitrary) {
	*c.top() = deducedDef.Atom
}

func (c *convert) VisitReference(proto.Reference) {
	// Do nothing, we handle references specially
}
