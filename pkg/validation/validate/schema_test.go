// Copyright 2015 go-swagger maintainers
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

package validate

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-openapi/swag/conv"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
)

func TestSchemaValidator_Validate_Pattern(t *testing.T) {
	var schemaJSON = `
{
    "properties": {
        "name": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        },
        "place": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        }
    },
    "required": [
        "name"
    ]
}`

	schema := new(spec.Schema)
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), schema))

	var input map[string]interface{}
	var inputJSON = `{"name": "Ivan"}`

	require.NoError(t, json.Unmarshal([]byte(inputJSON), &input))
	assert.NoError(t, AgainstSchema(schema, input, strfmt.Default))

	input["place"] = json.Number("10")

	assert.Error(t, AgainstSchema(schema, input, strfmt.Default))

}

func TestSchemaValidator_PatternProperties(t *testing.T) {
	var schemaJSON = `
{
    "properties": {
        "name": {
            "type": "string",
            "pattern": "^[A-Za-z]+$",
            "minLength": 1
        }
	},
    "patternProperties": {
	  "address-[0-9]+": {
         "type": "string",
         "pattern": "^[\\s|a-z]+$"
	  }
    },
    "required": [
        "name"
    ],
	"additionalProperties": false
}`

	schema := new(spec.Schema)
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), schema))

	var input map[string]interface{}

	// ok
	var inputJSON = `{"name": "Ivan","address-1": "sesame street"}`
	require.NoError(t, json.Unmarshal([]byte(inputJSON), &input))
	assert.NoError(t, AgainstSchema(schema, input, strfmt.Default))

	// fail pattern regexp
	input["address-1"] = "1, Sesame Street"
	assert.Error(t, AgainstSchema(schema, input, strfmt.Default))

	// fail patternProperties regexp
	inputJSON = `{"name": "Ivan","address-1": "sesame street","address-A": "address"}`
	require.NoError(t, json.Unmarshal([]byte(inputJSON), &input))
	assert.Error(t, AgainstSchema(schema, input, strfmt.Default))

}

func TestSchemaValidator_ReferencePanic(t *testing.T) {
	assert.PanicsWithValue(t, `schema references not supported: http://localhost:1234/integer.json`, schemaRefValidator)
}

func schemaRefValidator() {
	var schemaJSON = `
{
    "$ref": "http://localhost:1234/integer.json"
}`

	schema := new(spec.Schema)
	_ = json.Unmarshal([]byte(schemaJSON), schema)

	var input map[string]interface{}

	// ok
	var inputJSON = `{"name": "Ivan","address-1": "sesame street"}`
	_ = json.Unmarshal([]byte(inputJSON), &input)
	// panics
	_ = AgainstSchema(schema, input, strfmt.Default)
}

// Test edge cases in schemaValidator which are difficult
// to simulate with specs
func TestSchemaValidator_EdgeCases(t *testing.T) {
	var s *SchemaValidator

	res := s.Validate("123")
	assert.NotNil(t, res)
	assert.True(t, res.IsValid())

	s = NewSchemaValidator(nil, nil, "", strfmt.Default)
	assert.Nil(t, s)

	v := "ABC"
	b := s.Applies(v, reflect.String)
	assert.False(t, b)

	sp := spec.Schema{}
	b = s.Applies(&sp, reflect.Struct)
	assert.True(t, b)

	spp := spec.Float64Property()

	s = NewSchemaValidator(spp, nil, "", strfmt.Default)

	s.SetPath("path")
	assert.Equal(t, "path", s.Path)

	r := s.Validate(nil)
	assert.NotNil(t, r)
	assert.False(t, r.IsValid())

	// Validating json.Number data against number|float64
	j := json.Number("123")
	r = s.Validate(j)
	assert.True(t, r.IsValid())

	// Validating json.Number data against integer|int32
	spp = spec.Int32Property()
	s = NewSchemaValidator(spp, nil, "", strfmt.Default)
	j = json.Number("123")
	r = s.Validate(j)
	assert.True(t, r.IsValid())

	bignum := conv.FormatFloat(math.MaxFloat64)
	j = json.Number(bignum)
	r = s.Validate(j)
	assert.False(t, r.IsValid())

	// Validating incorrect json.Number data
	spp = spec.Float64Property()
	s = NewSchemaValidator(spp, nil, "", strfmt.Default)
	j = json.Number("AXF")
	r = s.Validate(j)
	assert.False(t, r.IsValid())
}

func TestNumericFormatEnforcement(t *testing.T) {
	tests := []struct {
		name          string
		schema        *spec.Schema
		value         interface{}
		expectSuccess bool
	}{
		// Integer int32 tests
		{
			name: "int32 valid value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int32",
				},
			},
			value:         int64(123),
			expectSuccess: true,
		},
		{
			name: "int32 overflow value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int32",
				},
			},
			value:         int64(2147483648), // MaxInt32 + 1
			expectSuccess: false,
		},
		{
			name: "int32 boundary max value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int32",
				},
			},
			value:         int64(2147483647), // MaxInt32
			expectSuccess: true,
		},
		{
			name: "int32 boundary min value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int32",
				},
			},
			value:         int64(-2147483648), // MinInt32
			expectSuccess: true,
		},
		{
			name: "int32 underflow value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int32",
				},
			},
			value:         int64(-2147483649), // MinInt32 - 1
			expectSuccess: false,
		},

		// Integer int64 tests
		{
			name: "int64 valid value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "int64",
				},
			},
			value:         int64(9223372036854775807), // MaxInt64
			expectSuccess: true,
		},

		// Number float (float32) tests
		{
			name: "float32 valid value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"number"},
					Format: "float",
				},
			},
			value:         float64(1.23),
			expectSuccess: true,
		},
		{
			name: "float32 overflow value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"number"},
					Format: "float",
				},
			},
			value:         float64(math.MaxFloat32) * 1.1,
			expectSuccess: false,
		},

		// Number double (float64) tests
		{
			name: "double valid value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"number"},
					Format: "double",
				},
			},
			value:         float64(1.23),
			expectSuccess: true,
		},
		{
			name: "double max value",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"number"},
					Format: "double",
				},
			},
			value:         math.MaxFloat64,
			expectSuccess: true,
		},

		// Mismatched format tests (format should be ignored)
		{
			name: "number type with int32 format (format ignored)",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"number"},
					Format: "int32",
				},
			},
			value:         float64(math.MaxInt32) + 1.0, // Should be valid as a number, ignoring int32 limit
			expectSuccess: true,
		},
		{
			name: "integer type with float format (format ignored)",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer"},
					Format: "float",
				},
			},
			value:         int64(math.MaxInt32) * 2, // Should be valid as integer (int64), ignoring float limit/check
			expectSuccess: true,
		},
		{
			name: "multiple types with int32 format (format ignored)",
			schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"integer", "string"},
					Format: "int32",
				},
			},
			value:         int64(math.MaxInt32) + 1, // Should be valid as integer (int64), ignoring int32 limit due to ambiguity
			expectSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validator := NewSchemaValidator(tc.schema, nil, "", strfmt.Default)
			res := validator.Validate(tc.value)
			if tc.expectSuccess {
				assert.Empty(t, res.Errors, "Expected success for %s", tc.name)
			} else {
				assert.NotEmpty(t, res.Errors, "Expected error for %s", tc.name)
			}
		})
	}
}

func TestAgainstSchemaStructPointer(t *testing.T) {
	schema := &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}, Required: []string{"b"}}}
	in := struct {
		A int `json:"a"`
	}{A: 1}
	byValue := AgainstSchema(schema, in, strfmt.Default)
	byPointer := AgainstSchema(schema, &in, strfmt.Default)
	require.Error(t, byValue)
	require.Error(t, byPointer)
	assert.Equal(t, byValue.Error(), byPointer.Error())
}
