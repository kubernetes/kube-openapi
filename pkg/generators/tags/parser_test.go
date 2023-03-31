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

package tags

import (
	"reflect"
	"testing"
)

func TestFieldValuesParser(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected FieldValue
		err      error
	}{
		{
			name:  "multiple fields",
			input: `rule:"a < b",message:"just no", messageExpression : "'some expression'"`,
			expected: FieldValue{
				"rule":              "a < b",
				"message":           "just no",
				"messageExpression": "'some expression'",
			},
		},
		{
			name:  "escaped string",
			input: `f:"\"abc\""`,
			expected: FieldValue{
				"f": `"abc"`,
			},
		},
		{
			name:  "raw string value",
			input: "f1:`backticks`",
			expected: FieldValue{
				"f1": "backticks",
			},
		},
		{
			name:  "ints",
			input: "negative:-100,zero:0,positive:200",
			expected: FieldValue{
				"negative": int64(-100),
				"zero":     int64(0),
				"positive": int64(200),
			},
		},
		{
			name:  "floats",
			input: "negative:-1.5,zero:0.0,positive:2.75",
			expected: FieldValue{
				"negative": -1.5,
				"zero":     0.0,
				"positive": 2.75,
			},
		},
		{
			name:  "booleans",
			input: "true:true,false:false", // yes, this is allowed
			expected: FieldValue{
				"true":  true,
				"false": false,
			},
		},
		{
			name:  "ident values",
			input: "enum:MyEnumValue",
			expected: FieldValue{
				"enum": "MyEnumValue",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ParseFieldValues(tc.input)
			if tc.err != nil {
				if err == nil {
					t.Fatalf("expected error %v but got none", tc.err)
				}
				if tc.err.Error() != err.Error() {
					t.Fatalf("expected error %v but got err %v", tc.err, err)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(out, tc.expected) {
				t.Errorf("expected %+v but got %+v", tc.expected, out)
			}
		})
	}
}
