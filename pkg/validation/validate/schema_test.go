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

	"github.com/go-openapi/swag"
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

	bignum := swag.FormatFloat64(math.MaxFloat64)
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

func TestCelExpressionValidator(t *testing.T) {
	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"spec": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: map[string]interface{}{
							"x-kubernetes-validator": []interface{}{
								map[string]interface{}{
									"rule": "minReplicas < maxReplicas",
								},
							},
						},
					},
					SchemaProps: spec.SchemaProps{
						Type: []string{"object"},
						Properties: map[string]spec.Schema{
							"minReplicas": {
								SchemaProps: spec.SchemaProps{
									Type:   []string{"integer"},
									Format: "int64",
								},
								VendorExtensible: spec.VendorExtensible{
									Extensions: map[string]interface{}{
										"x-kubernetes-validator": []interface{}{
											map[string]interface{}{
												"rule": "self >= 0",
											},
										},
									},
								},
							},
							"maxReplicas": {
								SchemaProps: spec.SchemaProps{
									Type:   []string{"integer"},
									Format: "int64",
								},
								VendorExtensible: spec.VendorExtensible{
									Extensions: map[string]interface{}{
										"x-kubernetes-validator": []interface{}{
											map[string]interface{}{
												"rule": "self >= 0",
											},
										},
									},
								},
							},
							"nestedObj": {
								VendorExtensible: spec.VendorExtensible{
									Extensions: map[string]interface{}{
										"x-kubernetes-validator": []interface{}{
											map[string]interface{}{
												"rule":    "val < 10",
												"message": "val is a bit too big",
											},
											map[string]interface{}{
												"rule":    "val < 1000",
												"message": "val is way too big",
											},
										},
									},
								},
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									Properties: map[string]spec.Schema{
										"val": {
											SchemaProps: spec.SchemaProps{
												Type:   []string{"integer"},
												Format: "int64",
											},
										},
									},
								},
							},
							"objMap": {
								VendorExtensible: spec.VendorExtensible{
									Extensions: map[string]interface{}{
										"x-kubernetes-validator": []interface{}{
											map[string]interface{}{
												"rule": "self.all(m, self[m].val > 10)",
											},
										},
									},
								},
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									AdditionalProperties: &spec.SchemaOrBool{
										Allows: true,
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type: []string{"object"},
												Properties: map[string]spec.Schema{
													"val": {
														SchemaProps: spec.SchemaProps{
															Type:   []string{"integer"},
															Format: "int64",
														},
													},
												},
											},
										},
									},
								},
							},
							"details": {
								VendorExtensible: spec.VendorExtensible{
									Extensions: map[string]interface{}{
										"x-kubernetes-validator": []interface{}{
											map[string]interface{}{
												"rule": "messages.all(m, m.val > 10)",
											},
											map[string]interface{}{
												// TODO: make it easier to dereference into an associative list
												"rule": "messages.filter(m, m.key == 'a').all(m, m.val == 100)",
											},
										},
									},
								},
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									Properties: map[string]spec.Schema{
										"messages": {
											VendorExtensible: spec.VendorExtensible{
												Extensions: map[string]interface{}{
													"x-kubernetes-list-type":     "map",
													"x-kubernetes-list-map-keys": []interface{}{"key"},
												},
											},
											SchemaProps: spec.SchemaProps{
												Type: []string{"array"},
												Items: &spec.SchemaOrArray{
													Schema: &spec.Schema{
														SchemaProps: spec.SchemaProps{
															Type: []string{"object"},
															Properties: map[string]spec.Schema{
																"key": {
																	SchemaProps: spec.SchemaProps{
																		Type: []string{"string"},
																	},
																},
																"val": {
																	SchemaProps: spec.SchemaProps{
																		Type:   []string{"integer"},
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
								},
							},
						},
					},
				},
			},
		},
	}

	cases := []struct {
		name         string
		input        map[string]interface{}
		expectErrors []string
	}{
		{
			name: "valid object rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(5),
					"maxReplicas": int64(10),
				},
			},
		},
		{
			name: "invalid field rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(-2),
					"maxReplicas": int64(-1),
				},
			},
			expectErrors: []string{
				"spec.minReplicas failed validator rule 'self >= 0'",
				"spec.maxReplicas failed validator rule 'self >= 0'",
			},
		},
		{
			name: "valid object rule with nested field access",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"nestedObj": map[string]interface{}{
						"val": int64(5),
					},
				},
			},
		},
		{
			name: "invalid object rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(11),
					"maxReplicas": int64(10),
				},
			},
			expectErrors: []string{
				"spec failed validator rule 'minReplicas < maxReplicas'",
			},
		},
		{
			name: "invalid object rule with nested field access",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"nestedObj": map[string]interface{}{
						"val": int64(10000),
					},
				},
			},
			expectErrors: []string{
				"spec.nestedObj: val is a bit too big",
				"spec.nestedObj: val is way too big",
			},
		},
		{
			name: "valid map rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"objMap": map[string]interface{}{
						"a": map[string]interface{}{
							"val": int64(100),
						},
					},
				},
			},
		},
		{
			name: "invalid map rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"objMap": map[string]interface{}{
						"a": map[string]interface{}{
							"val": int64(100),
						},
						"b": map[string]interface{}{
							"val": int64(5),
						},
					},
				},
			},
			expectErrors: []string{
				"spec.objMap failed validator rule 'self.all(m, self[m].val > 10)'",
			},
		},
		{
			name: "valid associative list rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"details": map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"key": "a",
								"val": int64(100),
							},
							map[string]interface{}{
								"key": "b",
								"val": int64(200),
							},
						},
					},
				},
			},
		},
		{
			name: "invalid associative list rule",
			input: map[string]interface{}{
				"spec": map[string]interface{}{
					"minReplicas": int64(1),
					"maxReplicas": int64(2),
					"details": map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"key": "a",
								"val": int64(5),
							},
							map[string]interface{}{
								"key": "b",
								"val": int64(5),
							},
						},
					},
				},
			},
			expectErrors: []string{
				"spec.details failed validator rule 'messages.all(m, m.val > 10)'",
				"spec.details failed validator rule 'messages.filter(m, m.key == 'a').all(m, m.val == 100)'",
			},
		},
		// TODO: test all scalar types
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			validator := NewSchemaValidator(schema, nil, "", strfmt.Default, ValidationRulesEnabled)
			r := validator.Validate(tc.input)

			actualErrors := map[string]struct{}{}
			for _, e := range r.Errors {
				actualErrors[e.Error()] = struct{}{}
			}
			for _, e := range tc.expectErrors {
				assert.Contains(t, actualErrors, e, "Expected errors to contain: %v", e)
				delete(actualErrors, e)
			}
			for e := range actualErrors {
				assert.Failf(t, "Unexpected error", "Did not expect errors to contain: %v", e)
			}
			assert.Equal(t, len(tc.expectErrors) == 0, r.IsValid())
		})
	}
}
