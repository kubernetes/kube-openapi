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

package cel

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
	"strings"
	"testing"
)

func TestCelCompilation(t *testing.T) {
	cases := []struct {
		name               string
		input              spec.Schema
		wantError          bool
		checkErrorMessage  bool
		expectedErrMessage string
	}{
		{
			name: "valid object",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"minReplicas": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"integer"},
							},
						},
						"maxReplicas": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"integer"},
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "minReplicas < maxReplicas",
								"message": "minReplicas should be smaller than maxReplicas",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid for string",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "self.startsWith('s')",
								"message": "scoped field should start with 's'",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid for byte",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:   []string{"string"},
					Format: "byte",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "string(self).endsWith('s')",
								"message": "scoped field should end with 's'",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid for boolean",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"boolean"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "self == true",
								"message": "scoped field should be true",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid for integer",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"integer"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "self > 0",
								"message": "scoped field should be greater than 0",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid for number",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"number"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "self > 1.0",
								"message": "scoped field should be greater than 1.0",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid nested object of object",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
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
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "nestedObj.val == 10",
								"message": "val should be equal to 10",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid nested object of array",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"nestedObj": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
								Items: &spec.SchemaOrArray{
									Schema: &spec.Schema{
										SchemaProps: spec.SchemaProps{
											Type: []string{"array"},
											Items: &spec.SchemaOrArray{
												Schema: &spec.Schema{
													SchemaProps: spec.SchemaProps{
														Type: []string{"string"},
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
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "size(self.nestedObj[0]) == 10",
								"message": "size of first element in nestedObj should be equal to 10",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid nested array of array",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: []string{"array"},
								Items: &spec.SchemaOrArray{
									Schema: &spec.Schema{
										SchemaProps: spec.SchemaProps{
											Type: []string{"array"},
											Items: &spec.SchemaOrArray{
												Schema: &spec.Schema{
													SchemaProps: spec.SchemaProps{
														Type: []string{"string"},
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
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "size(self[0][0]) == 10",
								"message": "size of items under items of scoped field should be equal to 10",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid nested array of object",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
								Properties: map[string]spec.Schema{
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
								},
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "self[0].nestedObj.val == 10",
								"message": "val under nestedObj under properties under items should be equal to 10",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "valid map",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					AdditionalProperties: &spec.SchemaOrBool{
						Allows: true,
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type:     []string{"boolean"},
								Nullable: false,
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "size(self) > 0",
								"message": "size of scoped field should be greater than 0",
							},
						},
					},
				},
			},
			wantError:         false,
			checkErrorMessage: false,
		},
		{
			name: "invalid checking for number",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"number"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "size(self) == 10",
								"message": "size of scoped field should be equal to 10",
							},
						},
					},
				},
			},
			wantError:          true,
			checkErrorMessage:  true,
			expectedErrMessage: "size of scoped field should be equal to 10",
		},
		{
			name: "compilation failure",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"integer"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"rule":    "size(self) == 10",
								"message": "size of scoped field should be equal to 10",
							},
						},
					},
				},
			},
			wantError:          true,
			checkErrorMessage:  true,
			expectedErrMessage: "compilation failed for rule",
		},
		{
			name: "rule is not specified",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"integer"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"message": "size of scoped field should be equal to 10",
							},
						},
					},
				},
			},
			wantError:          true,
			checkErrorMessage:  true,
			expectedErrMessage: "rule is not specified",
		},
		{
			name: "rule is not specified",
			input: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"integer"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-kubernetes-validator": []interface{}{
							map[string]interface{}{
								"message": "size of scoped field should be equal to 10",
							},
						},
					},
				},
			},
			wantError:          true,
			checkErrorMessage:  true,
			expectedErrMessage: "rule is not specified",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, allErrors := Compile(&tt.input)
			if tt.checkErrorMessage {
				var pass = false
				for _, err := range allErrors {
					if strings.Contains(err.Error(), tt.expectedErrMessage) {
						pass = true
					}
				}
				if !pass {
					t.Errorf("Expected error massage contains: %v, but got error: %v", tt.expectedErrMessage, allErrors)
				}
			} else {
				if !tt.wantError && len(allErrors) > 0 {
					t.Errorf("Expected no error, but got: %v", allErrors)
				} else if tt.wantError && len(allErrors) == 0 {
					t.Error("Expected error, but got none")
				}
			}
		})
	}
}
