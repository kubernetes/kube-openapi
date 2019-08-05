package spec3

import (
	"encoding/json"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

// Parameter a struct that describes a single operation parameter, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#parameterObject
//
// Note that this struct is actually a thin wrapper around ParameterProps to make it referable and extensible
type Parameter struct {
	spec.Refable
	ParameterProps
	spec.VendorExtensible
}

// MarshalJSON is a custom marshal function that knows how to encode Parameter as JSON
func (p *Parameter) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(p.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(p.ParameterProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(p.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

// ParameterProps a struct that describes a single operation parameter, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#parameterObject
type ParameterProps struct {
	// Name holds the name of the parameter
	Name string `json:"name,omitempty"`
	// In holds the location of the parameter
	In string `json:"in,omitempty"`
	// Description holds a brief description of the parameter
	Description string `json:"description,omitempty"`
	// Required determines whether this parameter is mandatory
	Required bool `json:"required,omitempty"`
	// Deprecated declares this operation to be deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// AllowEmptyValue sets the ability to pass empty-valued parameters
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	// Style describes how the parameter value will be serialized depending on the type of the parameter value
	Style string `json:"style,omitempty"`
	// Explode when true, parameter values of type array or object generate separate parameters for each value of the array or key-value pair of the map
	Explode bool `json:"explode,omitempty"`
	// AllowReserved determines whether the parameter value SHOULD allow reserved characters, as defined by RFC3986
	AllowReserved bool `json:"allowReserved,omitempty"`
	// Schema holds the schema defining the type used for the parameter
	Schema *spec.Schema `json:"schema,omitempty"`
	// Content holds a map containing the representations for the parameter
	Content map[string]*MediaType `json:"content,omitempty"`
	// the following fields are missing:
	// TODO: Example field is missing - (example	Any	Example of the media type. The example SHOULD match the specified schema and encoding properties if present. The example object is mutually exclusive of the examples object. Furthermore, if referencing a schema which contains an example, the example value SHALL override the example provided by the schema. To represent examples of media types that cannot naturally be represented in JSON or YAML, a string value can contain the example with escaping where necessary.)
	// TODO: Examples field is missing - (examples	Map[ string, Example Object | Reference Object]	Examples of the media type. Each example SHOULD contain a value in the correct format as specified in the parameter encoding. The examples object is mutually exclusive of the example object. Furthermore, if referencing a schema which contains an example, the examples value SHALL override the example provided by the schema.)
}
