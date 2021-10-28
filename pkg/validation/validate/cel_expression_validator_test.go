/*
Copyright 2021 The Kubernetes Authors.
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

package validate

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
	"math"
	"testing"
)

func TestCelValueValidator(t *testing.T) {
	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"minReplicas": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"integer"},
						Format: "int64",
					},
				},
				"maxReplicas": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"integer"},
						Format: "int64",
					},
				},
				"nestedObj": {
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
				"int32val": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"integer"},
						Format: "int32",
					},
				},
				"int64val": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"integer"},
						Format: "int64",
					},
				},
				"floatval": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"number"},
						Format: "float",
					},
				},
				"doubleval": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"number"},
						Format: "double",
					},
				},
				"stringval": {
					SchemaProps: spec.SchemaProps{
						Type: []string{"string"},
					},
				},
				"binaryval": {
					SchemaProps: spec.SchemaProps{
						Type:   []string{"string"},
						Format: "binary",
					},
				},
				"booleanval": {
					SchemaProps: spec.SchemaProps{
						Type: []string{"boolean"},
					},
				},
			},
		},
	}
	cases := []struct {
		name    string
		input   map[string]interface{}
		expr    string
		isValid bool
	}{
		{
			name: "valid int compare",
			input: map[string]interface{}{
				"minReplicas": int64(5),
				"maxReplicas": int64(10),
			},
			expr:    "minReplicas < maxReplicas",
			isValid: true,
		},
		{
			name: "no validator",
			input: map[string]interface{}{
				"minReplicas": int64(5),
				"maxReplicas": int64(10),
			},
			isValid: true,
		},
		{
			name: "invalid int compare",
			input: map[string]interface{}{
				"minReplicas": int64(11),
				"maxReplicas": int64(10),
			},
			expr:    "minReplicas < maxReplicas",
			isValid: false,
		},
		{
			name: "valid nested field access",
			input: map[string]interface{}{
				"nestedObj": map[string]interface{}{
					"val": int64(10),
				},
			},
			expr:    "nestedObj.val == 10",
			isValid: true,
		},
		{
			name: "int32 and int64 comparison",
			input: map[string]interface{}{
				"int32val": math.MaxInt32,
				"int64val": math.MaxInt64,
			},
			expr:    "int32val < int64val",
			isValid: true,
		},
		{
			name: "float",
			input: map[string]interface{}{
				"floatval": float32(3.1415926),
			},
			expr:    "floatval > 3.1415925 && floatval < 3.1415927",
			isValid: true,
		},
		{
			name: "float",
			input: map[string]interface{}{
				"doubleval": float64(3.141592653589793),
			},
			expr:    "doubleval > 3.141592653589792 && doubleval < 3.141592653589794",
			isValid: true,
		},
		{
			name: "true",
			input: map[string]interface{}{
				"booleanval": true,
			},
			expr:    "booleanval",
			isValid: true,
		},
		{
			name: "string",
			input: map[string]interface{}{
				"stringval": "♛",
			},
			expr:    "stringval == '♛'",
			isValid: true,
		},
		{
			name: "binary",
			input: map[string]interface{}{
				"binaryval": []byte("xyz"),
			},
			expr:    "binaryval == b'xyz'",
			isValid: true,
		},
		{
			name: "false",
			input: map[string]interface{}{
				"booleanval": false,
			},
			expr:    "!booleanval",
			isValid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				schema.Extensions = nil
			}()
			if tc.expr != "" {
				schema.Extensions = map[string]interface{}{
					"x-kubernetes-validators": []interface{}{
						map[string]interface{}{
							"rule": tc.expr,
						},
					},
				}
			}
			validator := newCelExpressionValidator("", schema)
			if validator == nil {
				if !tc.isValid {
					t.Fatalf("Expected a non-nil validator since isValid is expected to be false")
				}
				return
			}
			result := validator.Validate(tc.input)
			if result.IsValid() != tc.isValid {
				t.Fatalf("Expected isValid=%t, but got %t. Errors: %v", tc.isValid, result.IsValid(), result.Errors)
			}
		})
	}
}
