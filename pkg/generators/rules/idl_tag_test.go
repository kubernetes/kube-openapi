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

package rules

import (
	"reflect"
	"testing"

	"k8s.io/gengo/types"
)

func TestListTypeMissing(t *testing.T) {
	tcs := []struct {
		// name of test case
		name string
		t    *types.Type

		// expected list of violation fields
		expected []string
	}{
		{
			name:     "none",
			t:        &types.Type{},
			expected: []string{},
		},
		{
			name: "simple missing",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					types.Member{
						Name: "Containers",
						Type: &types.Type{
							Kind: types.Slice,
						},
					},
				},
			},
			expected: []string{"Containers"},
		},
		{
			name: "simple passing",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					types.Member{
						Name: "Containers",
						Type: &types.Type{
							Kind: types.Slice,
						},
						CommentLines: []string{"+listType=map"},
					},
				},
			},
			expected: []string{},
		},

		{
			name: "list Items field should not be annotated",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					types.Member{
						Name: "Items",
						Type: &types.Type{
							Kind: types.Slice,
						},
						CommentLines: []string{"+listType=map"},
					},
					types.Member{
						Name:     "ListMeta",
						Embedded: true,
						Type: &types.Type{
							Kind: types.Struct,
						},
					},
				},
			},
			expected: []string{"Items"},
		},

		{
			name: "list Items field without annotation should pass validation",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					types.Member{
						Name: "Items",
						Type: &types.Type{
							Kind: types.Slice,
						},
					},
					types.Member{
						Name:     "ListMeta",
						Embedded: true,
						Type: &types.Type{
							Kind: types.Struct,
						},
					},
				},
			},
			expected: []string{},
		},

		{
			name: "a list that happens to be called Items (i.e. nested, not top-level list) needs annotations",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					types.Member{
						Name: "Items",
						Type: &types.Type{
							Kind: types.Slice,
						},
					},
				},
			},
			expected: []string{"Items"},
		},
	}

	rule := &ListTypeMissing{}
	for _, tc := range tcs {
		if violations, _ := rule.Validate(tc.t); !reflect.DeepEqual(violations, tc.expected) {
			t.Errorf("unexpected validation result: test name %v, want: %v, got: %v",
				tc.name, tc.expected, violations)
		}
	}
}
