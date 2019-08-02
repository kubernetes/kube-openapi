package spec3

import (
	"encoding/json"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

// Header a struct that allows for additional information to be provided, for example Content-Disposition
//
// Note that allowing as stated in https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#headerObject
// the Header follows the structure of the Parameter with some changes that don't affect the structure of ParameterProps struct
type Header struct {
	spec.Refable
	ParameterProps
	spec.VendorExtensible
}

// MarshalJSON is a custom marshal function that knows how to encode Header as JSON
func (h *Header) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(h.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(h.ParameterProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(h.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}
