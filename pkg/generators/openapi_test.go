/*
Copyright 2016 The Kubernetes Authors.

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

package generators

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go/format"

	"github.com/stretchr/testify/assert"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
)

func construct(t *testing.T, files map[string]string, testNamer namer.Namer) (*parser.Builder, types.Universe, []*types.Type) {
	b := parser.New()
	for name, src := range files {
		if err := b.AddFileForTest(filepath.Dir(name), name, []byte(src)); err != nil {
			t.Fatal(err)
		}
	}
	u, err := b.FindTypes()
	if err != nil {
		t.Fatal(err)
	}
	orderer := namer.Orderer{Namer: testNamer}
	o := orderer.OrderUniverse(u)
	return b, u, o
}

func testOpenAPITypeWriter(t *testing.T, code string) (error, error, *assert.Assertions, *bytes.Buffer, *bytes.Buffer, []string) {
	assert := assert.New(t)
	var testFiles = map[string]string{
		"base/foo/bar.go": code,
	}
	outputPackage := "base/output"
	imports := generator.NewImportTrackerForPackage(outputPackage)
	rawNamer := namer.NewRawNamer(outputPackage, imports)
	namers := namer.NameSystems{
		"raw": rawNamer,
		"private": &namer.NameStrategy{
			Join: func(pre string, in []string, post string) string {
				return strings.Join(in, "_")
			},
			PrependPackageNames: 4, // enough to fully qualify from k8s.io/api/...
		},
	}
	builder, universe, _ := construct(t, testFiles, rawNamer)
	context, err := generator.NewContext(builder, namers, "raw")
	if err != nil {
		t.Fatal(err)
	}
	blahT := universe.Type(types.Name{Package: "base/foo", Name: "Blah"})

	callBuffer := &bytes.Buffer{}
	callSW := generator.NewSnippetWriter(callBuffer, context, "$", "$")
	callError := newOpenAPITypeWriter(callSW, context).generateCall(blahT)

	funcBuffer := &bytes.Buffer{}
	funcSW := generator.NewSnippetWriter(funcBuffer, context, "$", "$")
	funcError := newOpenAPITypeWriter(funcSW, context).generate(blahT)

	return callError, funcError, assert, callBuffer, funcBuffer, imports.ImportLines()
}

func TestSimple(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
type Blah struct {
	// A simple string
	String string
	// A simple int
	Int int `+"`"+`json:",omitempty"`+"`"+`
	// An int considered string simple int
	IntString int `+"`"+`json:",string"`+"`"+`
	// A simple int64
	Int64 int64
	// A simple int32
	Int32 int32
	// A simple int16
	Int16 int16
	// A simple int8
	Int8 int8
	// A simple int
	Uint uint
	// A simple int64
	Uint64 uint64
	// A simple int32
	Uint32 uint32
	// A simple int16
	Uint16 uint16
	// A simple int8
	Uint8 uint8
	// A simple byte
	Byte byte
	// A simple boolean
	Bool bool
	// A simple float64
	Float64 float64
	// A simple float32
	Float32 float32
	// a base64 encoded characters
	ByteArray []byte
	// a member with an extension
	// +k8s:openapi-gen=x-kubernetes-member-tag:member_test
	WithExtension string
	// a member with struct tag as extension
	// +patchStrategy=merge
	// +patchMergeKey=pmk
	WithStructTagExtension string `+"`"+`patchStrategy:"merge" patchMergeKey:"pmk"`+"`"+`
	// a member with a list type
	// +listType=atomic
	// +default=["foo", "bar"]
	WithListType []string
	// a member with a map type
	// +listType=atomic
	// +default={"foo": "bar", "fizz": "buzz"}
	Map map[string]string
	// a member with a string pointer
	// +default="foo"
	StringPointer *string
	// an int member with a default
	// +default=1
	OmittedInt int `+"`"+`json:"omitted,omitempty"`+"`"+`
}
		`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"String": {
SchemaProps: spec.SchemaProps{
Description: "A simple string",
Default: "",
Type: []string{"string"},
Format: "",
},
},
"Int64": {
SchemaProps: spec.SchemaProps{
Description: "A simple int64",
Default: 0,
Type: []string{"integer"},
Format: "int64",
},
},
"Int32": {
SchemaProps: spec.SchemaProps{
Description: "A simple int32",
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
"Int16": {
SchemaProps: spec.SchemaProps{
Description: "A simple int16",
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
"Int8": {
SchemaProps: spec.SchemaProps{
Description: "A simple int8",
Default: 0,
Type: []string{"integer"},
Format: "byte",
},
},
"Uint": {
SchemaProps: spec.SchemaProps{
Description: "A simple int",
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
"Uint64": {
SchemaProps: spec.SchemaProps{
Description: "A simple int64",
Default: 0,
Type: []string{"integer"},
Format: "int64",
},
},
"Uint32": {
SchemaProps: spec.SchemaProps{
Description: "A simple int32",
Default: 0,
Type: []string{"integer"},
Format: "int64",
},
},
"Uint16": {
SchemaProps: spec.SchemaProps{
Description: "A simple int16",
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
"Uint8": {
SchemaProps: spec.SchemaProps{
Description: "A simple int8",
Default: 0,
Type: []string{"integer"},
Format: "byte",
},
},
"Byte": {
SchemaProps: spec.SchemaProps{
Description: "A simple byte",
Default: 0,
Type: []string{"integer"},
Format: "byte",
},
},
"Bool": {
SchemaProps: spec.SchemaProps{
Description: "A simple boolean",
Default: false,
Type: []string{"boolean"},
Format: "",
},
},
"Float64": {
SchemaProps: spec.SchemaProps{
Description: "A simple float64",
Default: 0,
Type: []string{"number"},
Format: "double",
},
},
"Float32": {
SchemaProps: spec.SchemaProps{
Description: "A simple float32",
Default: 0,
Type: []string{"number"},
Format: "float",
},
},
"ByteArray": {
SchemaProps: spec.SchemaProps{
Description: "a base64 encoded characters",
Type: []string{"string"},
Format: "byte",
},
},
"WithExtension": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-member-tag": "member_test",
},
},
SchemaProps: spec.SchemaProps{
Description: "a member with an extension",
Default: "",
Type: []string{"string"},
Format: "",
},
},
"WithStructTagExtension": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-patch-merge-key": "pmk",
"x-kubernetes-patch-strategy": "merge",
},
},
SchemaProps: spec.SchemaProps{
Description: "a member with struct tag as extension",
Default: "",
Type: []string{"string"},
Format: "",
},
},
"WithListType": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-list-type": "atomic",
},
},
SchemaProps: spec.SchemaProps{
Description: "a member with a list type",
Default: []interface {}{"foo", "bar"},
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
"Map": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-list-type": "atomic",
},
},
SchemaProps: spec.SchemaProps{
Description: "a member with a map type",
Default: map[string]interface {}{"fizz":"buzz", "foo":"bar"},
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
"StringPointer": {
SchemaProps: spec.SchemaProps{
Description: "a member with a string pointer",
Default: "foo",
Type: []string{"string"},
Format: "",
},
},
"omitted": {
SchemaProps: spec.SchemaProps{
Description: "an int member with a default",
Default: 1,
Type: []string{"integer"},
Format: "int32",
},
},
},
Required: []string{"String","Int64","Int32","Int16","Int8","Uint","Uint64","Uint32","Uint16","Uint8","Byte","Bool","Float64","Float32","ByteArray","WithExtension","WithStructTagExtension","WithListType","Map","StringPointer"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
},
},
},
}
}

`, funcBuffer.String())
}

func TestEmptyProperties(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah demonstrate a struct without fields.
type Blah struct {
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct without fields.",
Type: []string{"object"},
},
},
}
}

`, funcBuffer.String())
}

func TestNestedStruct(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Nested is used as struct field
type Nested struct {
  // A simple string
  String string
}

// Blah demonstrate a struct with struct field.
type Blah struct {
  // A struct field
  Field Nested
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with struct field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"Field": {
SchemaProps: spec.SchemaProps{
Description: "A struct field",
Default: map[string]interface {}{},
Ref: ref("base/foo.Nested"),
},
},
},
Required: []string{"Field"},
},
},
Dependencies: []string{
"base/foo.Nested",},
}
}

`, funcBuffer.String())
}

func TestNestedStructPointer(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Nested is used as struct pointer field
type Nested struct {
  // A simple string
  String string
}

// Blah demonstrate a struct with struct pointer field.
type Blah struct {
  // A struct pointer field
  Field *Nested
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with struct pointer field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"Field": {
SchemaProps: spec.SchemaProps{
Description: "A struct pointer field",
Ref: ref("base/foo.Nested"),
},
},
},
Required: []string{"Field"},
},
},
Dependencies: []string{
"base/foo.Nested",},
}
}

`, funcBuffer.String())
}

func TestEmbeddedStruct(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Nested is used as embedded struct field
type Nested struct {
  // A simple string
  String string
}

// Blah demonstrate a struct with embedded struct field.
type Blah struct {
  // An embedded struct field
  Nested
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with embedded struct field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"Nested": {
SchemaProps: spec.SchemaProps{
Description: "An embedded struct field",
Default: map[string]interface {}{},
Ref: ref("base/foo.Nested"),
},
},
},
Required: []string{"Nested"},
},
},
Dependencies: []string{
"base/foo.Nested",},
}
}

`, funcBuffer.String())
}

func TestSingleEmbeddedStruct(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import "time"

// Nested is used as embedded struct field
type Nested struct {
  // A simple string
  time.Duration
}

// Blah demonstrate a struct with embedded struct field.
type Blah struct {
  // An embedded struct field
  // +default="10ms"
  Nested `+"`"+`json:"nested,omitempty" protobuf:"bytes,5,opt,name=nested"`+"`"+`
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with embedded struct field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"nested": {
SchemaProps: spec.SchemaProps{
Description: "An embedded struct field",
Default: "10ms",
Ref: ref("base/foo.Nested"),
},
},
},
},
},
Dependencies: []string{
"base/foo.Nested",},
}
}

`, funcBuffer.String())
}

func TestEmbeddedInlineStruct(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Nested is used as embedded inline struct field
type Nested struct {
  // A simple string
  String string
}

// Blah demonstrate a struct with embedded inline struct field.
type Blah struct {
  // An embedded inline struct field
  Nested `+"`"+`json:",inline,omitempty"`+"`"+`
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with embedded inline struct field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"String": {
SchemaProps: spec.SchemaProps{
Description: "A simple string",
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
Required: []string{"String"},
},
},
}
}

`, funcBuffer.String())
}

func TestEmbeddedInlineStructPointer(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Nested is used as embedded inline struct pointer field.
type Nested struct {
  // A simple string
  String string
}

// Blah demonstrate a struct with embedded inline struct pointer field.
type Blah struct {
  // An embedded inline struct pointer field
  *Nested `+"`"+`json:",inline,omitempty"`+"`"+`
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah demonstrate a struct with embedded inline struct pointer field.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"String": {
SchemaProps: spec.SchemaProps{
Description: "A simple string",
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
Required: []string{"String"},
},
},
}
}

`, funcBuffer.String())
}

func TestNestedMapString(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Map sample tests openAPIGen.generateMapProperty method.
type Blah struct {
	// A sample String to String map
	StringToArray map[string]map[string]string
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Map sample tests openAPIGen.generateMapProperty method.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"StringToArray": {
SchemaProps: spec.SchemaProps{
Description: "A sample String to String map",
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
},
},
},
},
Required: []string{"StringToArray"},
},
},
}
}

`, funcBuffer.String())
}

func TestNestedMapInt(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Map sample tests openAPIGen.generateMapProperty method.
type Blah struct {
	// A sample String to String map
	StringToArray map[string]map[string]int
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Map sample tests openAPIGen.generateMapProperty method.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"StringToArray": {
SchemaProps: spec.SchemaProps{
Description: "A sample String to String map",
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
},
},
},
},
},
},
},
Required: []string{"StringToArray"},
},
},
}
}

`, funcBuffer.String())
}

func TestNestedMapBoolean(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Map sample tests openAPIGen.generateMapProperty method.
type Blah struct {
	// A sample String to String map
	StringToArray map[string]map[string]bool
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Map sample tests openAPIGen.generateMapProperty method.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"StringToArray": {
SchemaProps: spec.SchemaProps{
Description: "A sample String to String map",
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: false,
Type: []string{"boolean"},
Format: "",
},
},
},
},
},
},
},
},
},
Required: []string{"StringToArray"},
},
},
}
}

`, funcBuffer.String())
}

func TestFailingSample1(t *testing.T) {
	_, funcErr, assert, _, _, _ := testOpenAPITypeWriter(t, `
package foo

// Map sample tests openAPIGen.generateMapProperty method.
type Blah struct {
	// A sample String to String map
	StringToArray map[string]map[string]map[int]string
}
	`)
	if assert.Error(funcErr, "An error was expected") {
		assert.Equal(funcErr, fmt.Errorf("failed to generate map property in base/foo.Blah: StringToArray: map with non-string keys are not supported by OpenAPI in map[int]string"))
	}
}

func TestFailingSample2(t *testing.T) {
	_, funcErr, assert, _, _, _ := testOpenAPITypeWriter(t, `
package foo

// Map sample tests openAPIGen.generateMapProperty method.
type Blah struct {
	// A sample String to String map
	StringToArray map[int]string
}	`)
	if assert.Error(funcErr, "An error was expected") {
		assert.Equal(funcErr, fmt.Errorf("failed to generate map property in base/foo.Blah: StringToArray: map with non-string keys are not supported by OpenAPI in map[int]string"))
	}
}

func TestFailingDefaultEnforced(t *testing.T) {
	tests := []struct {
		definition    string
		expectedError error
	}{
		{
			definition: `
package foo

type Blah struct {
	// +default=5
	Int int
}	`,
			expectedError: fmt.Errorf("failed to generate default in base/foo.Blah: Int: invalid default value (5) for non-pointer/non-omitempty. If specified, must be: 0"),
		},
		{
			definition: `
package foo

type Blah struct {
	// +default={"foo": 5}
	Struct struct{
		foo int
	}
}	`,
			expectedError: fmt.Errorf(`failed to generate default in base/foo.Blah: Struct: invalid default value (map[string]interface {}{"foo":5}) for non-pointer/non-omitempty. If specified, must be: {}`),
		},
		{
			definition: `
package foo

type Blah struct {
	List []Item

}

// +default="foo"
type Item string	`,
			expectedError: fmt.Errorf(`failed to generate slice property in base/foo.Blah: List: invalid default value ("foo") for non-pointer/non-omitempty. If specified, must be: ""`),
		},
		{
			definition: `
package foo

type Blah struct {
	Map map[string]Item

}

// +default="foo"
type Item string	`,
			expectedError: fmt.Errorf(`failed to generate map property in base/foo.Blah: Map: invalid default value ("foo") for non-pointer/non-omitempty. If specified, must be: ""`),
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, funcErr, assert, _, _, _ := testOpenAPITypeWriter(t, test.definition)
			if assert.Error(funcErr, "An error was expected") {
				assert.Equal(funcErr, test.expectedError)
			}
		})
	}
}

func TestCustomDef(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import openapi "k8s.io/kube-openapi/pkg/common"

type Blah struct {
}

func (_ Blah) OpenAPIDefinition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		},
	}
}
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": foo.Blah{}.OpenAPIDefinition(),
`, callBuffer.String())
	assert.Equal(``, funcBuffer.String())
}

func TestCustomDefV3(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import openapi "k8s.io/kube-openapi/pkg/common"

type Blah struct {
}

func (_ Blah) OpenAPIV3Definition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		},
	}
}
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": foo.Blah{}.OpenAPIV3Definition(),
`, callBuffer.String())
	assert.Equal(``, funcBuffer.String())
}

func TestCustomDefV2AndV3(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import openapi "k8s.io/kube-openapi/pkg/common"

type Blah struct {
}

func (_ Blah) OpenAPIV3Definition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		},
	}
}

func (_ Blah) OpenAPIDefinition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		},
	}
}
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": common.EmbedOpenAPIDefinitionIntoV2Extension(foo.Blah{}.OpenAPIV3Definition(), foo.Blah{}.OpenAPIDefinition()),
`, callBuffer.String())
	assert.Equal(``, funcBuffer.String())
}

func TestCustomDefs(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a custom type
type Blah struct {
}

func (_ Blah) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Blah) OpenAPISchemaFormat() string { return "date-time" }
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a custom type",
Type:foo.Blah{}.OpenAPISchemaType(),
Format:foo.Blah{}.OpenAPISchemaFormat(),
},
},
}
}

`, funcBuffer.String())
}

func TestCustomDefsV3(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import openapi "k8s.io/kube-openapi/pkg/common"

// Blah is a custom type
type Blah struct {
}

func (_ Blah) OpenAPIV3Definition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		},
	}
}

func (_ Blah) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Blah) OpenAPISchemaFormat() string { return "date-time" }
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.EmbedOpenAPIDefinitionIntoV2Extension(foo.Blah{}.OpenAPIV3Definition(), common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a custom type",
Type:foo.Blah{}.OpenAPISchemaType(),
Format:foo.Blah{}.OpenAPISchemaFormat(),
},
},
})
}

`, funcBuffer.String())
}

func TestV3OneOfTypes(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a custom type
type Blah struct {
}

func (_ Blah) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Blah) OpenAPISchemaFormat() string { return "date-time" }
func (_ Blah) OpenAPIV3OneOfTypes() []string { return []string{"string", "number"} }

`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.EmbedOpenAPIDefinitionIntoV2Extension(common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a custom type",
OneOf:common.GenerateOpenAPIV3OneOfSchema(foo.Blah{}.OpenAPIV3OneOfTypes()),
Format:foo.Blah{}.OpenAPISchemaFormat(),
},
},
},common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a custom type",
Type:foo.Blah{}.OpenAPISchemaType(),
Format:foo.Blah{}.OpenAPISchemaFormat(),
},
},
})
}

`, funcBuffer.String())
}

func TestPointer(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// PointerSample demonstrate pointer's properties
type Blah struct {
	// A string pointer
	StringPointer *string
	// A struct pointer
	StructPointer *Blah
	// A slice pointer
	SlicePointer *[]string
	// A map pointer
	MapPointer *map[string]string
}
	`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "PointerSample demonstrate pointer's properties",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"StringPointer": {
SchemaProps: spec.SchemaProps{
Description: "A string pointer",
Type: []string{"string"},
Format: "",
},
},
"StructPointer": {
SchemaProps: spec.SchemaProps{
Description: "A struct pointer",
Ref: ref("base/foo.Blah"),
},
},
"SlicePointer": {
SchemaProps: spec.SchemaProps{
Description: "A slice pointer",
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
"MapPointer": {
SchemaProps: spec.SchemaProps{
Description: "A map pointer",
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
},
Required: []string{"StringPointer","StructPointer","SlicePointer","MapPointer"},
},
},
Dependencies: []string{
"base/foo.Blah",},
}
}

`, funcBuffer.String())
}

func TestNestedLists(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
type Blah struct {
	// Nested list
	NestedList [][]int64
}
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"NestedList": {
SchemaProps: spec.SchemaProps{
Description: "Nested list",
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: 0,
Type: []string{"integer"},
Format: "int64",
},
},
},
},
},
},
},
},
},
Required: []string{"NestedList"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
},
},
},
}
}

`, funcBuffer.String())
}

func TestNestListOfMaps(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
type Blah struct {
	// Nested list of maps
	NestedListOfMaps [][]map[string]string
}
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"NestedListOfMaps": {
SchemaProps: spec.SchemaProps{
Description: "Nested list of maps",
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Type: []string{"object"},
AdditionalProperties: &spec.SchemaOrBool{
Allows: true,
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
},
},
},
},
},
},
},
Required: []string{"NestedListOfMaps"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
},
},
},
}
}

`, funcBuffer.String())
}

func TestExtensions(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
type Blah struct {
	// a member with a list type with two map keys
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	WithListField []string

	// another member with a list type with one map key
	// +listType=map
	// +listMapKey=port
	WithListField2 []string
}
		`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"WithListField": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-list-map-keys": []interface{}{
"port",
"protocol",
},
"x-kubernetes-list-type": "map",
},
},
SchemaProps: spec.SchemaProps{
Description: "a member with a list type with two map keys",
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
"WithListField2": {
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-list-map-keys": []interface{}{
"port",
},
"x-kubernetes-list-type": "map",
},
},
SchemaProps: spec.SchemaProps{
Description: "another member with a list type with one map key",
Type: []string{"array"},
Items: &spec.SchemaOrArray{
Schema: &spec.Schema{
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
},
},
},
},
Required: []string{"WithListField","WithListField2"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
},
},
},
}
}

`, funcBuffer.String())
}

func TestUnion(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
// +union
type Blah struct {
	// +unionDiscriminator
	Discriminator *string `+"`"+`json:"discriminator"`+"`"+`
        // +optional
        Numeric int `+"`"+`json:"numeric"`+"`"+`
        // +optional
        String string `+"`"+`json:"string"`+"`"+`
        // +optional
        Float float64 `+"`"+`json:"float"`+"`"+`
}
		`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"discriminator": {
SchemaProps: spec.SchemaProps{
Type: []string{"string"},
Format: "",
},
},
"numeric": {
SchemaProps: spec.SchemaProps{
Default: 0,
Type: []string{"integer"},
Format: "int32",
},
},
"string": {
SchemaProps: spec.SchemaProps{
Default: "",
Type: []string{"string"},
Format: "",
},
},
"float": {
SchemaProps: spec.SchemaProps{
Default: 0,
Type: []string{"number"},
Format: "double",
},
},
},
Required: []string{"discriminator"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
"x-kubernetes-unions": []interface{}{
map[string]interface{}{
"discriminator": "discriminator",
"fields-to-discriminateBy": map[string]interface{}{
"float": "Float",
"numeric": "Numeric",
"string": "String",
},
},
},
},
},
},
}
}

`, funcBuffer.String())
}

func TestEnum(t *testing.T) {
	callErr, funcErr, assert, _, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// EnumType is the enumType.
// +enum
type EnumType string

// EnumA is a.
const EnumA EnumType = "a"
// EnumB is b.
const EnumB EnumType = "b"

// Blah is a test.
// +k8s:openapi-gen=true
// +k8s:openapi-gen=x-kubernetes-type-tag:type_test
type Blah struct {
  // Value is the value.
	Value EnumType
	NoCommentEnum EnumType
  // +optional
	OptionalEnum *EnumType
}`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	_ = assert
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Description: "Blah is a test.",
Type: []string{"object"},
Properties: map[string]spec.Schema{
"Value": {
SchemaProps: spec.SchemaProps{`+"\n"+
		"Description: \"Value is the value.\\n\\nPossible enum values:\\n - `\\\"a\\\"` is a.\\n - `\\\"b\\\"` is b.\","+`
Default: "",
Type: []string{"string"},
Format: "",
Enum: []interface{}{"a", "b"},
},
},
"NoCommentEnum": {
SchemaProps: spec.SchemaProps{`+"\n"+
		"Description: \"Possible enum values:\\n - `\\\"a\\\"` is a.\\n - `\\\"b\\\"` is b.\","+`
Default: "",
Type: []string{"string"},
Format: "",
Enum: []interface{}{"a", "b"},
},
},
"OptionalEnum": {
SchemaProps: spec.SchemaProps{`+"\n"+
		"Description: \"Possible enum values:\\n - `\\\"a\\\"` is a.\\n - `\\\"b\\\"` is b.\","+`
Type: []string{"string"},
Format: "",
Enum: []interface{}{"a", "b"},
},
},
},
Required: []string{"Value","NoCommentEnum"},
},
VendorExtensible: spec.VendorExtensible{
Extensions: spec.Extensions{
"x-kubernetes-type-tag": "type_test",
},
},
},
}
}

`, funcBuffer.String())
}

func TestSymbolReference(t *testing.T) {
	callErr, funcErr, assert, _, funcBuffer, imports := testOpenAPITypeWriter(t, `
package foo

// +k8s:openapi-gen=true
type Blah struct {
	// +default="A Default Value"
	// +optional
	Value *string

	// User constant local to the output package fully qualified
	// +default=ref(base/output.MyConst)
	// +optional
	FullyQualifiedOutputValue *string

	// Local to types but not to output
	// +default=ref(MyConst)
	// +optional
	LocalValue *string

	// +default=ref(base/foo.MyConst)
	// +optional
	FullyQualifiedLocalValue *string

	// +default=ref(k8s.io/api/v1.TerminationPathDefault)
	// +optional
	FullyQualifiedExternalValue *string
}
	`)
	assert.NoError(funcErr)
	assert.NoError(callErr)
	assert.ElementsMatch(imports, []string{`foo "base/foo"`, `v1 "k8s.io/api/v1"`, `common "k8s.io/kube-openapi/pkg/common"`, `spec "k8s.io/kube-openapi/pkg/validation/spec"`})

	if formatted, err := format.Source(funcBuffer.Bytes()); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(string(formatted), `func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
				Properties: map[string]spec.Schema{
					"Value": {
						SchemaProps: spec.SchemaProps{
							Default: "A Default Value",
							Type:    []string{"string"},
							Format:  "",
						},
					},
					"FullyQualifiedOutputValue": {
						SchemaProps: spec.SchemaProps{
							Description: "User constant local to the output package fully qualified",
							Default:     MyConst,
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"LocalValue": {
						SchemaProps: spec.SchemaProps{
							Description: "Local to types but not to output",
							Default:     foo.MyConst,
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"FullyQualifiedLocalValue": {
						SchemaProps: spec.SchemaProps{
							Default: foo.MyConst,
							Type:    []string{"string"},
							Format:  "",
						},
					},
					"FullyQualifiedExternalValue": {
						SchemaProps: spec.SchemaProps{
							Default: v1.TerminationPathDefault,
							Type:    []string{"string"},
							Format:  "",
						},
					},
				},
			},
		},
	}
}

`)
	}

}

// Show that types with unmarshalJSON in their hierarchy do not have struct
// defaults enforced, and that aliases and embededd types are respected
func TestMustEnforceDefaultStruct(t *testing.T) {
	callErr, funcErr, assert, _, funcBuffer, imports := testOpenAPITypeWriter(t, `
package foo

type Time struct {
	value interface{}
}


type TimeWithoutUnmarshal struct {
	value interface{}
}

func (_ TimeWithoutUnmarshal) OpenAPISchemaType() []string { return []string{"string"} }
func (_ TimeWithoutUnmarshal) OpenAPISchemaFormat() string { return "date-time" }

func (_ Time) UnmarshalJSON([]byte) error {
	return nil
}


func (_ Time) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Time) OpenAPISchemaFormat() string { return "date-time" }

// Time with UnmarshalJSON defined on pointer instead of struct
type MicroTime struct {
	value interface{}
}

func (t *MicroTime) UnmarshalJSON([]byte) error {
	return nil
}

func (_ MicroTime) OpenAPISchemaType() []string { return []string{"string"} }
func (_ MicroTime) OpenAPISchemaFormat() string { return "date-time" }

type Int64 int64

type Duration struct {
	Int64
}

func (_ Duration) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Duration) OpenAPISchemaFormat() string { return "" }

type NothingSpecial struct {
	Field string
}

// +k8s:openapi-gen=true
type Blah struct {
	Embedded Duration
	PointerUnmarshal MicroTime
	StructUnmarshal Time
	NoUnmarshal TimeWithoutUnmarshal
	Regular NothingSpecial
}
	`)
	assert.NoError(funcErr)
	assert.NoError(callErr)
	assert.ElementsMatch(imports, []string{`foo "base/foo"`, `common "k8s.io/kube-openapi/pkg/common"`, `spec "k8s.io/kube-openapi/pkg/validation/spec"`})

	if formatted, err := format.Source(funcBuffer.Bytes()); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(string(formatted), `func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
				Properties: map[string]spec.Schema{
					"Embedded": {
						SchemaProps: spec.SchemaProps{
							Default: 0,
							Ref:     ref("base/foo.Duration"),
						},
					},
					"PointerUnmarshal": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("base/foo.MicroTime"),
						},
					},
					"StructUnmarshal": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("base/foo.Time"),
						},
					},
					"NoUnmarshal": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("base/foo.TimeWithoutUnmarshal"),
						},
					},
					"Regular": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("base/foo.NothingSpecial"),
						},
					},
				},
				Required: []string{"Embedded", "PointerUnmarshal", "StructUnmarshal", "NoUnmarshal", "Regular"},
			},
		},
		Dependencies: []string{
			"base/foo.Duration", "base/foo.MicroTime", "base/foo.NothingSpecial", "base/foo.Time", "base/foo.TimeWithoutUnmarshal"},
	}
}

`)
	}

}

func TestMarkerComments(t *testing.T) {

	callErr, funcErr, assert, _, funcBuffer, imports := testOpenAPITypeWriter(t, `
package foo

// +k8s:openapi-gen=true
// +k8s:validation:maxProperties=10
// +k8s:validation:minProperties=1
// +k8s:validation:exclusiveMinimum
// +k8s:validation:exclusiveMaximum
type Blah struct {

	// Integer with min and max values
	// +k8s:validation:minimum=0
	// +k8s:validation:maximum=10
	// +k8s:validation:exclusiveMinimum
	// +k8s:validation:exclusiveMaximum
	IntValue int

	// String with min and max lengths
	// +k8s:validation:minLength=1
	// +k8s:validation:maxLength=10
	// +k8s:validation:pattern="^foo$[0-9]+"
	StringValue string

	// +k8s:validation:maxitems=10
	// +k8s:validation:minItems=1
	// +k8s:validation:uniqueItems
	ArrayValue []string

	// +k8s:validation:maxProperties=10
	// +k8s:validation:minProperties=1
	ObjValue map[string]interface{}
}
	`)
	assert.NoError(funcErr)
	assert.NoError(callErr)
	assert.ElementsMatch(imports, []string{`foo "base/foo"`, `common "k8s.io/kube-openapi/pkg/common"`, `spec "k8s.io/kube-openapi/pkg/validation/spec"`, `ptr "k8s.io/utils/ptr"`})

	if formatted, err := format.Source(funcBuffer.Bytes()); err != nil {
		t.Fatalf("%v\n%v", err, string(funcBuffer.Bytes()))
	} else {
		formatted_expected, ree := format.Source([]byte(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
		return common.OpenAPIDefinition{
			Schema: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: 			  []string{"object"},
					ExclusiveMinimum: true,
					ExclusiveMaximum: true,
					MinProperties:    ptr.To[int64](1),
					MaxProperties:    ptr.To[int64](10),
					Properties: map[string]spec.Schema{
						"IntValue": {
							SchemaProps: spec.SchemaProps{
								Description: "Integer with min and max values",
								Default: 	 0,
								Minimum:	 ptr.To[float64](0),
								Maximum:	 ptr.To[float64](10),
								ExclusiveMinimum: true,
								ExclusiveMaximum: true,
								Type:        []string{"integer"},
								Format:	  	 "int32",
							},
						},
						"StringValue": {
							SchemaProps: spec.SchemaProps{
								Description: "String with min and max lengths",
								Default:	 "",
								MinLength:	 ptr.To[int64](1),
								MaxLength:	 ptr.To[int64](10),
								Pattern:	 "^foo$[0-9]+",
								Type:        []string{"string"},
								Format:	  	 "",
							},
						},
						"ArrayValue": {
							SchemaProps: spec.SchemaProps{
								MinItems:	 ptr.To[int64](1),
								MaxItems:	 ptr.To[int64](10),
								UniqueItems: true,
								Type: []string{"array"},
								Items: &spec.SchemaOrArray{
									Schema: &spec.Schema{
										SchemaProps: spec.SchemaProps{
											Default: "",
											Type:    []string{"string"},
											Format:  "",
										},
									},
								},
							},
						},
						"ObjValue": {
							SchemaProps: spec.SchemaProps{
								MinProperties:	 ptr.To[int64](1),
								MaxProperties:	 ptr.To[int64](10),
								Type: []string{"object"},
									AdditionalProperties: &spec.SchemaOrBool{
										Allows: true,
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type:   []string{"object"},
												Format: "",
											},
										},
									},
							},
						},
					},
					Required: []string{"IntValue", "StringValue", "ArrayValue", "ObjValue"},
				},
			},
		}
	}

`))
		if ree != nil {
			t.Fatal(ree)
		}
		assert.Equal(string(formatted), string(formatted_expected))
	}
}

func TestCELMarkerComments(t *testing.T) {

	callErr, funcErr, assert, _, funcBuffer, imports := testOpenAPITypeWriter(t, `
		package foo

		// +k8s:openapi-gen=true
		// +k8s:validation:cel[0]:rule="self == oldSelf"
		// +k8s:validation:cel[0]:message="message1"
		type Blah struct {
			// +k8s:validation:cel[0]:rule="self.length() > 0"
			// +k8s:validation:cel[0]:message="string message"
			// +k8s:validation:cel[1]:rule="self.length() % 2 == 0"
			// +k8s:validation:cel[1]:messageExpression="self + ' hello'"
			// +k8s:validation:cel[1]:optionalOldSelf
			// +optional
			Field string
		}
	`)

	assert.NoError(funcErr)
	assert.NoError(callErr)
	assert.ElementsMatch(imports, []string{`foo "base/foo"`, `common "k8s.io/kube-openapi/pkg/common"`, `spec "k8s.io/kube-openapi/pkg/validation/spec"`})

	if formatted, err := format.Source(funcBuffer.Bytes()); err != nil {
		t.Fatalf("%v\n%v", err, string(funcBuffer.Bytes()))
	} else {
		formatted_expected, ree := format.Source([]byte(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
		return common.OpenAPIDefinition{
			Schema: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: 			  []string{"object"},
					Properties: map[string]spec.Schema{
						"Field": {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									"x-kubernetes-validations": []interface{}{map[string]interface{}{"message": "string message", "rule": "self.length() > 0"}, map[string]interface{}{"messageExpression": "self + ' hello'", "optionalOldSelf": true, "rule": "self.length() % 2 == 0"}},
								},
							},
							SchemaProps: spec.SchemaProps{
								Default: "",
								Type:    []string{"string"},
								Format:  "",
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validations": []interface{}{map[string]interface{}{"message": "message1", "rule": "self == oldSelf"}},
					},
				},
			},
		}
	}

`))
		if ree != nil {
			t.Fatal(ree)
		}
		assert.Equal(string(formatted_expected), string(formatted))
	}
}

func TestMultilineCELMarkerComments(t *testing.T) {

	callErr, funcErr, assert, _, funcBuffer, imports := testOpenAPITypeWriter(t, `
		package foo

		// +k8s:openapi-gen=true
		// +k8s:validation:cel[0]:rule="self == oldSelf"
		// +k8s:validation:cel[0]:message="message1"
		// +k8s:validation:cel[0]:fieldPath="field"
		type Blah struct {
			// +k8s:validation:cel[0]:rule="self.length() > 0"
			// +k8s:validation:cel[0]:message="string message"
			// +k8s:validation:cel[0]:reason="Invalid"
			// +k8s:validation:cel[1]:rule>  !oldSelf.hasValue() || self.length() % 2 == 0
			// +k8s:validation:cel[1]:rule>     ? self.field == "even"
			// +k8s:validation:cel[1]:rule>     : self.field == "odd"
			// +k8s:validation:cel[1]:messageExpression="field must be whether the length of the string is even or odd"
			// +k8s:validation:cel[1]:optionalOldSelf
			// +optional
			Field string
		}
	`)

	assert.NoError(funcErr)
	assert.NoError(callErr)
	assert.ElementsMatch(imports, []string{`foo "base/foo"`, `common "k8s.io/kube-openapi/pkg/common"`, `spec "k8s.io/kube-openapi/pkg/validation/spec"`})

	if formatted, err := format.Source(funcBuffer.Bytes()); err != nil {
		t.Fatalf("%v\n%v", err, string(funcBuffer.Bytes()))
	} else {
		formatted_expected, ree := format.Source([]byte(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
		return common.OpenAPIDefinition{
			Schema: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: 			  []string{"object"},
					Properties: map[string]spec.Schema{
						"Field": {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									"x-kubernetes-validations": []interface{}{map[string]interface{}{"message": "string message", "reason": "Invalid", "rule": "self.length() > 0"}, map[string]interface{}{"messageExpression": "field must be whether the length of the string is even or odd", "optionalOldSelf": true, "rule": "!oldSelf.hasValue() || self.length() % 2 == 0\n? self.field == \"even\"\n: self.field == \"odd\""}},
								},
							},
							SchemaProps: spec.SchemaProps{
								Default: "",
								Type:    []string{"string"},
								Format:  "",
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validations": []interface{}{map[string]interface{}{"fieldPath": "field", "message": "message1", "rule": "self == oldSelf"}},
					},
				},
			},
		}
	}

`))
		if ree != nil {
			t.Fatal(ree)
		}
		assert.Equal(string(formatted_expected), string(formatted))
	}
}

func TestMarkerCommentsCustomDefsV3(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

import openapi "k8s.io/kube-openapi/pkg/common"

// +k8s:validation:maxProperties=10
type Blah struct {
}

func (_ Blah) OpenAPIV3Definition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"object"},
				MaxProperties: ptr.To[int64](10),
				Format: "ipv4",
			},
		},
	}
}

func (_ Blah) OpenAPISchemaType() []string { return []string{"object"} }
func (_ Blah) OpenAPISchemaFormat() string { return "ipv4" }
`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.EmbedOpenAPIDefinitionIntoV2Extension(foo.Blah{}.OpenAPIV3Definition(), common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Type:foo.Blah{}.OpenAPISchemaType(),
Format:foo.Blah{}.OpenAPISchemaFormat(),
MaxProperties: ptr.To[int64](10),
},
},
})
}

`, funcBuffer.String())
}

func TestMarkerCommentsV3OneOfTypes(t *testing.T) {
	callErr, funcErr, assert, callBuffer, funcBuffer, _ := testOpenAPITypeWriter(t, `
package foo

// +k8s:validation:maxLength=10
type Blah struct {
}

func (_ Blah) OpenAPISchemaType() []string { return []string{"string"} }
func (_ Blah) OpenAPIV3OneOfTypes() []string { return []string{"string", "array"} }
func (_ Blah) OpenAPISchemaFormat() string { return "ipv4" }

`)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if funcErr != nil {
		t.Fatal(funcErr)
	}
	assert.Equal(`"base/foo.Blah": schema_base_foo_Blah(ref),
`, callBuffer.String())
	assert.Equal(`func schema_base_foo_Blah(ref common.ReferenceCallback) common.OpenAPIDefinition {
return common.EmbedOpenAPIDefinitionIntoV2Extension(common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
OneOf:common.GenerateOpenAPIV3OneOfSchema(foo.Blah{}.OpenAPIV3OneOfTypes()),
Format:foo.Blah{}.OpenAPISchemaFormat(),
MaxLength: ptr.To[int64](10),
},
},
},common.OpenAPIDefinition{
Schema: spec.Schema{
SchemaProps: spec.SchemaProps{
Type:foo.Blah{}.OpenAPISchemaType(),
Format:foo.Blah{}.OpenAPISchemaFormat(),
MaxLength: ptr.To[int64](10),
},
},
})
}

`, funcBuffer.String())
}
