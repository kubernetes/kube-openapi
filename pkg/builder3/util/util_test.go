/*
Copyright 2022 The Kubernetes Authors.

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

package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"k8s.io/kube-openapi/pkg/schemamutation"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/randfill"
)

// Compile-time assertions that fail if fields are added to, removed from, or
// reordered in spec.Schema, spec.VendorExtensible, spec.SchemaProps, or
// spec.SwaggerSchemaProps without updating hasSiblingElements.
var (
	_ = struct {
		spec.VendorExtensible
		spec.SchemaProps
		spec.SwaggerSchemaProps
		ExtraProps map[string]interface{}
	}(spec.Schema{})
	_ = struct {
		Extensions spec.Extensions
	}(spec.VendorExtensible{})
	_ = struct {
		ID                   string
		Ref                  spec.Ref
		Schema               spec.SchemaURL
		Description          string
		Type                 spec.StringOrArray
		Nullable             bool
		Format               string
		Title                string
		Default              interface{}
		Maximum              *float64
		ExclusiveMaximum     bool
		Minimum              *float64
		ExclusiveMinimum     bool
		MaxLength            *int64
		MinLength            *int64
		Pattern              string
		MaxItems             *int64
		MinItems             *int64
		UniqueItems          bool
		MultipleOf           *float64
		Enum                 []interface{}
		MaxProperties        *int64
		MinProperties        *int64
		Required             []string
		Items                *spec.SchemaOrArray
		AllOf                []spec.Schema
		OneOf                []spec.Schema
		AnyOf                []spec.Schema
		Not                  *spec.Schema
		Properties           map[string]spec.Schema
		AdditionalProperties *spec.SchemaOrBool
		PatternProperties    map[string]spec.Schema
		Dependencies         spec.Dependencies
		AdditionalItems      *spec.SchemaOrBool
		Definitions          spec.Definitions
	}(spec.SchemaProps{})
	_ = struct {
		Discriminator string
		ReadOnly      bool
		ExternalDocs  *spec.ExternalDocumentation
		Example       interface{}
	}(spec.SwaggerSchemaProps{})
)

func referenceWrapRefs(schema *spec.Schema) *spec.Schema {
	walker := schemamutation.Walker{
		SchemaCallback: func(schema *spec.Schema) *spec.Schema {
			orig := schema
			clone := func() {
				if orig == schema {
					schema = new(spec.Schema)
					*schema = *orig
				}
			}
			if schema.Ref.String() != "" && !reflect.DeepEqual(*schema, spec.Schema{SchemaProps: spec.SchemaProps{Ref: schema.Ref}}) {
				clone()
				refSchema := new(spec.Schema)
				refSchema.Ref = schema.Ref
				schema.Ref = spec.Ref{}
				schema.AllOf = []spec.Schema{*refSchema}
			}
			return schema
		},
		RefCallback: schemamutation.RefCallbackNoop,
	}
	return walker.WalkSchema(schema)
}

func makeBenchmarkSchema() *spec.Schema {
	props := make(map[string]spec.Schema, 24)
	for i := 0; i < 16; i++ {
		props[fmt.Sprintf("field_%d", i)] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: fmt.Sprintf("Primitive field %d", i),
				Type:        []string{"string"},
			},
		}
	}
	for i := 0; i < 4; i++ {
		props[fmt.Sprintf("pure_ref_%d", i)] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Ref: spec.MustCreateRef(fmt.Sprintf("#/components/schemas/io.k8s.api.core.v1.PureType%d", i)),
			},
		}
	}
	for i := 0; i < 4; i++ {
		props[fmt.Sprintf("sibling_ref_%d", i)] = spec.Schema{
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{"x-kubernetes-patch-strategy": "merge"},
			},
			SchemaProps: spec.SchemaProps{
				Description: fmt.Sprintf("Ref field with sibling description %d", i),
				Ref:         spec.MustCreateRef(fmt.Sprintf("#/components/schemas/io.k8s.api.core.v1.RefType%d", i)),
				Nullable:    i%2 == 0,
			},
		}
	}
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Benchmark object schema",
			Type:        []string{"object"},
			Properties:  props,
		},
	}
}

func TestWrapRefsEquivalence(t *testing.T) {
	s := makeBenchmarkSchema()
	origBytes, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	expected := referenceWrapRefs(s)
	actual := WrapRefs(s)

	afterOrigBytes, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterOrigBytes) != string(origBytes) {
		t.Fatalf("WrapRefs(s) mutated input schema:\ngot:  %s\nwant: %s", string(afterOrigBytes), string(origBytes))
	}

	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	actualBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	if string(actualBytes) != string(expectedBytes) {
		t.Fatalf("WrapRefs(s) mismatch:\ngot:  %s\nwant: %s", string(actualBytes), string(expectedBytes))
	}

	// Fuzz test 200 random schemas with refs and sibling fields
	for i := 0; i < 200; i++ {
		r := rand.New(rand.NewSource(int64(1000 + i)))
		f := randfill.New().RandSource(r).NilChance(0.65).NumElements(0, 2)
		f.Funcs(
			func(ref *spec.Ref, c randfill.Continue) {
				if c.Bool() {
					*ref = spec.MustCreateRef(fmt.Sprintf("#/components/schemas/Type%d", c.Intn(10)))
				} else {
					*ref = spec.Ref{}
				}
			},
			func(url *spec.SchemaURL, c randfill.Continue) {
				if c.Bool() {
					*url = spec.SchemaURL("http://example.com/schema")
				}
			},
			func(sa *spec.SchemaOrStringArray, c randfill.Continue) {
				*sa = spec.SchemaOrStringArray{}
			},
			func(d *spec.Dependencies, c randfill.Continue) {
				*d = nil
			},
			func(v *interface{}, c randfill.Continue) {
				if c.Bool() {
					*v = "default-val"
				}
			},
			func(sp *spec.SchemaProps, c randfill.Continue) {
				if c.Float64() < 0.7 {
					return
				}
				c.FillNoCustom(sp)
			},
		)
		fuzzSchema := &spec.Schema{}
		f.Fill(fuzzSchema)
		raw, err := json.Marshal(fuzzSchema)
		if err != nil {
			continue
		}
		var norm1, norm2 spec.Schema
		if err := json.Unmarshal(raw, &norm1); err != nil {
			continue
		}
		if err := json.Unmarshal(raw, &norm2); err != nil {
			continue
		}
		exp := referenceWrapRefs(&norm1)
		act := WrapRefs(&norm2)
		expJSON, err := json.Marshal(exp)
		if err != nil {
			t.Fatal(err)
		}
		actJSON, err := json.Marshal(act)
		if err != nil {
			t.Fatal(err)
		}
		if string(actJSON) != string(expJSON) {
			t.Fatalf("WrapRefs fuzz iteration %d mismatch:\ngot:  %s\nwant: %s", i, string(actJSON), string(expJSON))
		}
	}
}

func TestHasSiblingElementsAllFields(t *testing.T) {
	ref := spec.MustCreateRef("#/components/schemas/Target")
	pureRef := &spec.Schema{SchemaProps: spec.SchemaProps{Ref: ref}}
	if got := WrapRefs(pureRef); got != pureRef {
		t.Fatalf("WrapRefs(pureRef) = %v, want unmodified pointer %v", got, pureRef)
	}

	setNonZeroField := func(fv reflect.Value) {
		switch fv.Kind() {
		case reflect.String:
			fv.SetString("x")
		case reflect.Bool:
			fv.SetBool(true)
		case reflect.Slice:
			fv.Set(reflect.MakeSlice(fv.Type(), 0, 0))
		case reflect.Map:
			fv.Set(reflect.MakeMap(fv.Type()))
		case reflect.Ptr:
			fv.Set(reflect.New(fv.Type().Elem()))
		case reflect.Interface:
			fv.Set(reflect.ValueOf("x"))
		default:
			t.Fatalf("unhandled field kind %v for type %v", fv.Kind(), fv.Type())
		}
	}

	var checkStructFields func(parentPath string, rt reflect.Type, indexPath []int)
	checkStructFields = func(parentPath string, rt reflect.Type, indexPath []int) {
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			path := append(append([]int(nil), indexPath...), i)
			fieldName := sf.Name
			if parentPath != "" {
				fieldName = parentPath + "." + sf.Name
			}
			if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
				checkStructFields(fieldName, sf.Type, path)
				continue
			}
			if sf.Name == "Ref" {
				continue
			}
			s := spec.Schema{SchemaProps: spec.SchemaProps{Ref: ref}}
			fv := reflect.ValueOf(&s).Elem().FieldByIndex(path)
			setNonZeroField(fv)

			got := WrapRefs(&s)
			want := referenceWrapRefs(&s)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("WrapRefs with sibling field %s: got %+v, want %+v", fieldName, got, want)
			}
			if len(got.AllOf) != 1 || got.AllOf[0].Ref.String() != ref.String() || got.Ref.String() != "" {
				t.Errorf("WrapRefs with sibling field %s did not wrap $ref into AllOf: got %+v", fieldName, got)
			}
		}
	}

	checkStructFields("", reflect.TypeOf(spec.Schema{}), nil)
}

func BenchmarkWrapRefs(b *testing.B) {
	schema := makeBenchmarkSchema()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapRefs(schema)
	}
}
