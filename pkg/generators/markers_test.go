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

package generators_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/gengo/types"
	"k8s.io/kube-openapi/pkg/generators"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/utils/ptr"
)

func TestParseCommentTags(t *testing.T) {

	structKind := createType("struct")

	numKind := createType("float")

	cases := []struct {
		t        *types.Type
		name     string
		comments []string
		expected generators.CommentTags

		// regex pattern matching the error, or empty string/unset if no error
		// is expected
		expectedError string
	}{
		{
			t:    &structKind,
			name: "basic example",
			comments: []string{
				"comment",
				"another + comment",
				"+k8s:validation:minimum=10.0",
				"+k8s:validation:maximum=20.0",
				"+k8s:validation:minLength=20",
				"+k8s:validation:maxLength=30",
				`+k8s:validation:pattern="asdf"`,
				"+k8s:validation:multipleOf=1.0",
				"+k8s:validation:minItems=1",
				"+k8s:validation:maxItems=2",
				"+k8s:validation:uniqueItems=true",
				"exclusiveMaximum=true",
				"not+k8s:validation:Minimum=0.0",
			},
			expected: generators.CommentTags{
				spec.SchemaProps{
					Maximum:     ptr.To(20.0),
					Minimum:     ptr.To(10.0),
					MinLength:   ptr.To[int64](20),
					MaxLength:   ptr.To[int64](30),
					Pattern:     "asdf",
					MultipleOf:  ptr.To(1.0),
					MinItems:    ptr.To[int64](1),
					MaxItems:    ptr.To[int64](2),
					UniqueItems: true,
				},
			},
		},
		{
			t:    &structKind,
			name: "empty",
		},
		{
			t:    &numKind,
			name: "single",
			comments: []string{
				"+k8s:validation:minimum=10.0",
			},
			expected: generators.CommentTags{
				spec.SchemaProps{
					Minimum: ptr.To(10.0),
				},
			},
		},
		{
			t:    &numKind,
			name: "multiple",
			comments: []string{
				"+k8s:validation:minimum=10.0",
				"+k8s:validation:maximum=20.0",
			},
			expected: generators.CommentTags{
				spec.SchemaProps{
					Maximum: ptr.To(20.0),
					Minimum: ptr.To(10.0),
				},
			},
		},
		{
			t:    &numKind,
			name: "invalid duplicate key",
			comments: []string{
				"+k8s:validation:minimum=10.0",
				"+k8s:validation:maximum=20.0",
				"+k8s:validation:minimum=30.0",
			},
			expectedError: `cannot unmarshal array into Go struct field CommentTags.minimum of type float64`,
		},
		{
			t:    &structKind,
			name: "unrecognized key is ignored",
			comments: []string{
				"+ignored=30.0",
			},
		},
		{
			t:    &numKind,
			name: "invalid: invalid value",
			comments: []string{
				"+k8s:validation:minimum=asdf",
			},
			expectedError: `invalid value for key k8s:validation:minimum`,
		},
		{

			t:    &structKind,
			name: "invalid: invalid value",
			comments: []string{
				"+k8s:validation:",
			},
			expectedError: `failed to parse marker comments: cannot have empty key for marker comment`,
		},
		{
			t:    &numKind,
			name: "ignore refs",
			comments: []string{
				"+k8s:validation:pattern=ref(asdf)",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := generators.ParseCommentTags(tc.t, tc.comments, "k8s:validation:")
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Regexp(t, tc.expectedError, err.Error())
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, actual)
		})
	}
}

// Test comment tag validation function
func TestCommentTags_Validate(t *testing.T) {

	testCases := []struct {
		name          string
		commentParams map[string]any
		t             types.Type
		errorMessage  string
	}{
		{
			name: "invalid minimum type",
			commentParams: map[string]any{
				"minimum": 10.5,
			},
			t:            createType("string"),
			errorMessage: "minimum can only be used on numeric types",
		},
		{
			name: "invalid minLength type",
			commentParams: map[string]any{
				"minLength": 10,
			},
			t:            createType("bool"),
			errorMessage: "minLength can only be used on string types",
		},
		{
			name: "invalid minItems type",
			commentParams: map[string]any{
				"minItems": 10,
			},
			t:            createType("string"),
			errorMessage: "minItems can only be used on array types",
		},
		{
			name: "invalid minProperties type",
			commentParams: map[string]any{
				"minProperties": 10,
			},
			t:            createType("string"),
			errorMessage: "minProperties can only be used on map types",
		},
		{
			name: "invalid exclusiveMinimum type",
			commentParams: map[string]any{
				"exclusiveMinimum": true,
			},
			t:            createType("array"),
			errorMessage: "exclusiveMinimum can only be used on numeric types",
		},
		{
			name: "invalid maximum type",
			commentParams: map[string]any{
				"maximum": 10.5,
			},
			t:            createType("array"),
			errorMessage: "maximum can only be used on numeric types",
		},
		{
			name: "invalid maxLength type",
			commentParams: map[string]any{
				"maxLength": 10,
			},
			t:            createType("map"),
			errorMessage: "maxLength can only be used on string types",
		},
		{
			name: "invalid maxItems type",
			commentParams: map[string]any{
				"maxItems": 10,
			},
			t:            createType("bool"),
			errorMessage: "maxItems can only be used on array types",
		},
		{
			name: "invalid maxProperties type",
			commentParams: map[string]any{
				"maxProperties": 10,
			},
			t:            createType("bool"),
			errorMessage: "maxProperties can only be used on map types",
		},
		{
			name: "invalid exclusiveMaximum type",
			commentParams: map[string]any{
				"exclusiveMaximum": true,
			},
			t:            createType("map"),
			errorMessage: "exclusiveMaximum can only be used on numeric types",
		},
		{
			name: "invalid pattern type",
			commentParams: map[string]any{
				"pattern": ".*",
			},
			t:            createType("int"),
			errorMessage: "pattern can only be used on string types",
		},
		{
			name: "invalid multipleOf type",
			commentParams: map[string]any{
				"multipleOf": 10.5,
			},
			t:            createType("string"),
			errorMessage: "multipleOf can only be used on numeric types",
		},
		{
			name: "invalid uniqueItems type",
			commentParams: map[string]any{
				"uniqueItems": true,
			},
			t:            createType("int"),
			errorMessage: "uniqueItems can only be used on array types",
		},
		{
			name: "negative minLength",
			commentParams: map[string]any{
				"minLength": -10,
			},
			t:            createType("string"),
			errorMessage: "minLength cannot be negative",
		},
		{
			name: "negative minItems",
			commentParams: map[string]any{
				"minItems": -10,
			},
			t:            createType("array"),
			errorMessage: "minItems cannot be negative",
		},
		{
			name: "negative minProperties",
			commentParams: map[string]any{
				"minProperties": -10,
			},
			t:            createType("map"),
			errorMessage: "minProperties cannot be negative",
		},
		{
			name: "negative maxLength",
			commentParams: map[string]any{
				"maxLength": -10,
			},
			t:            createType("string"),
			errorMessage: "maxLength cannot be negative",
		},
		{
			name: "negative maxItems",
			commentParams: map[string]any{
				"maxItems": -10,
			},
			t:            createType("array"),
			errorMessage: "maxItems cannot be negative",
		},
		{
			name: "negative maxProperties",
			commentParams: map[string]any{
				"maxProperties": -10,
			},
			t:            createType("map"),
			errorMessage: "maxProperties cannot be negative",
		},
		{
			name: "minimum > maximum",
			commentParams: map[string]any{
				"minimum": 10.5,
				"maximum": 5.5,
			},
			t:            createType("float"),
			errorMessage: "minimum 10.500000 is greater than maximum 5.500000",
		},
		{
			name: "exclusiveMinimum when minimum == maximum",
			commentParams: map[string]any{
				"minimum":          10.5,
				"maximum":          10.5,
				"exclusiveMinimum": true,
			},
			t:            createType("float"),
			errorMessage: "exclusiveMinimum/Maximum cannot be set when minimum == maximum",
		},
		{
			name: "exclusiveMaximum when minimum == maximum",
			commentParams: map[string]any{
				"minimum":          10.5,
				"maximum":          10.5,
				"exclusiveMaximum": true,
			},
			t:            createType("float"),
			errorMessage: "exclusiveMinimum/Maximum cannot be set when minimum == maximum",
		},
		{
			name: "minLength > maxLength",
			commentParams: map[string]any{
				"minLength": 10,
				"maxLength": 5,
			},
			t:            createType("string"),
			errorMessage: "minLength 10 is greater than maxLength 5",
		},
		{
			name: "minItems > maxItems",
			commentParams: map[string]any{
				"minItems": 10,
				"maxItems": 5,
			},
			t:            createType("array"),
			errorMessage: "minItems 10 is greater than maxItems 5",
		},
		{
			name: "minProperties > maxProperties",
			commentParams: map[string]any{
				"minProperties": 10,
				"maxProperties": 5,
			},
			t:            createType("map"),
			errorMessage: "minProperties 10 is greater than maxProperties 5",
		},
		{
			name: "invalid pattern",
			commentParams: map[string]any{
				"pattern": "([a-z]+",
			},
			t:            createType("string"),
			errorMessage: "invalid pattern \"([a-z]+\": error parsing regexp: missing closing ): `([a-z]+`",
		},
		{
			name: "multipleOf = 0",
			commentParams: map[string]any{
				"multipleOf": 0.0,
			},
			t:            createType("int"),
			errorMessage: "multipleOf cannot be 0",
		},
		{
			name: "valid comment tags with no invalid validations",
			commentParams: map[string]any{
				"pattern": ".*",
			},
			t:            createType("string"),
			errorMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commentTags := createCommentTags(tc.commentParams)
			err := commentTags.Validate()
			if err == nil {
				err = commentTags.ValidateType(&tc.t)
			}
			if tc.errorMessage != "" {
				require.Error(t, err)
				require.Equal(t, tc.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createCommentTags(input map[string]any) generators.CommentTags {

	ct := generators.CommentTags{}

	for key, value := range input {

		switch key {
		case "minimum":
			ct.Minimum = ptr.To(value.(float64))
		case "maximum":
			ct.Maximum = ptr.To(value.(float64))
		case "minLength":
			ct.MinLength = ptr.To(int64(value.(int)))
		case "maxLength":
			ct.MaxLength = ptr.To(int64(value.(int)))
		case "pattern":
			ct.Pattern = value.(string)
		case "multipleOf":
			ct.MultipleOf = ptr.To(value.(float64))
		case "minItems":
			ct.MinItems = ptr.To(int64(value.(int)))
		case "maxItems":
			ct.MaxItems = ptr.To(int64(value.(int)))
		case "uniqueItems":
			ct.UniqueItems = value.(bool)
		case "exclusiveMaximum":
			ct.ExclusiveMaximum = value.(bool)
		case "exclusiveMinimum":
			ct.ExclusiveMinimum = value.(bool)
		case "minProperties":
			ct.MinProperties = ptr.To(int64(value.(int)))
		case "maxProperties":
			ct.MaxProperties = ptr.To(int64(value.(int)))
		}
	}

	return ct
}

func createType(name string) types.Type {
	switch name {
	case "string":
		return *types.String
	case "int":
		return *types.Int64
	case "float":
		return *types.Float64
	case "bool":
		return *types.Bool
	case "array":
		return types.Type{Kind: types.Slice, Name: types.Name{Name: "[]int"}}
	case "map":
		return types.Type{Kind: types.Map, Name: types.Name{Name: "map[string]int"}}
	}
	return types.Type{Kind: types.Struct, Name: types.Name{Name: "struct"}}
}
