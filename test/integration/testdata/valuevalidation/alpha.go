package valuevalidation

import (
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// Dummy type to test the openapi-gen API rule checker.
// The API rule violations are in format of:
// -> +k8s:validation:[validation rule]=[value]

// +k8s:validation:maxProperties=5
// +k8s:validation:minProperties=1
// +k8s:openapi-gen=true
// +k8s:validation:cel[0]:rule="self == oldSelf"
// +k8s:validation:cel[0]:message="foo"
type Foo struct {
	// +k8s:validation:maxLength=5
	// +k8s:validation:minLength=1
	// +k8s:validation:pattern="^a.*b$"
	StringValue string

	// +k8s:validation:maximum=5.0
	// +k8s:validation:minimum=1.0
	// +k8s:validation:exclusiveMinimum=true
	// +k8s:validation:exclusiveMaximum=true
	// +k8s:validation:multipleOf=2.0
	NumberValue float64

	// +k8s:validation:maxItems=5
	// +k8s:validation:minItems=1
	// +k8s:validation:uniqueItems=true
	ArrayValue []string

	// +k8s:validation:minProperties=1
	// +k8s:validation:maxProperties=5
	MapValue map[string]string

	// +k8s:validation:cel[0]:rule="self.length() > 0"
	// +k8s:validation:cel[0]:message="string message"
	// +k8s:validation:cel[1]:rule="self.length() % 2 == 0"
	// +k8s:validation:cel[1]:messageExpression="self + ' hello'"
	// +optional
	CELField string `json:"celField"`
}

// This one has an open API v3 definition
// +k8s:validation:maxProperties=5
// +k8s:openapi-gen=true
// +k8s:validation:cel[0]:rule="self == oldSelf"
// +k8s:validation:cel[0]:message="foo2"
type Foo2 struct{}

func (Foo2) OpenAPIV3Definition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		},
	}
}

func (Foo2) OpenAPISchemaType() []string {
	return []string{"test-type"}
}

func (Foo2) OpenAPISchemaFormat() string {
	return "test-format"
}

// This one has a OneOf
// +k8s:openapi-gen=true
// +k8s:validation:maxProperties=5
// +k8s:openapi-gen=true
// +k8s:validation:cel[0]:rule="self == oldSelf"
// +k8s:validation:cel[0]:message="foo3"
type Foo3 struct{}

func (Foo3) OpenAPIV3OneOfTypes() []string {
	return []string{"number", "string"}
}
func (Foo3) OpenAPISchemaType() []string {
	return []string{"string"}
}
func (Foo3) OpenAPISchemaFormat() string {
	return "string"
}

// this one should ignore marker comments
// +k8s:openapi-gen=true
// +k8s:validation:maximum=6
// +k8s:validation:cel[0]:rule="self == oldSelf"
// +k8s:validation:cel[0]:message="foo4"
type Foo4 struct{}

func (Foo4) OpenAPIDefinition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"integer"},
			},
		},
	}
}

// +k8s:openapi-gen=true
// +k8s:validation:maxProperties=5
// +k8s:validation:minProperties=1
// +k8s:validation:cel[0]:rule="self == oldSelf"
// +k8s:validation:cel[0]:message="foo5"
type Foo5 struct{}

func (Foo5) OpenAPIV3Definition() common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		},
	}
}

func (Foo5) OpenAPISchemaType() []string {
	return []string{"test-type"}
}

func (Foo5) OpenAPISchemaFormat() string {
	return "test-format"
}
