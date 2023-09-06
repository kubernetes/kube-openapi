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

package builder3

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/kube-openapi/pkg/spec3"
)

func TestCollectSharedParameters(t *testing.T) {
	tests := []struct {
		name string
		spec string
		want map[string]string
	}{
		{
			name: "empty",
			spec: "",
			want: nil,
		},
		{
			name: "no shared",
			spec: `{
  "parameters": {"pre": {"in": "body", "name": "body", "required": true, "schema": {}}},
  "paths": {
    "/api/v1/a/{name}": {"get": {"parameters": [
		  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/a/{name}/foo": {"get": {"parameters": [
		  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}": {"get": {"parameters": [
		  {"description": "x","in":"query","name": "x2","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y2","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}/foo": {"get": {"parameters": [
		  {"description": "z","in":"query","name": "z2","schema": {"type": "boolean", "uniqueItems": true}}
    ]}}
  }
}`,
			want: map[string]string{
				`{"name":"x","in":"query","description":"x","schema":{"type":"boolean","uniqueItems":true}}`:  "x-16u_rOM9",
				`{"name":"y","in":"query","description":"y","schema":{"type":"boolean","uniqueItems":true}}`:  "y-bABmtr22",
				`{"name":"z","in":"query","description":"z","schema":{"type":"boolean","uniqueItems":true}}`:  "z-1LKrqY5S",
				`{"name":"x2","in":"query","description":"x","schema":{"type":"boolean","uniqueItems":true}}`: "x2-RXlVCxel",
				`{"name":"y2","in":"query","description":"y","schema":{"type":"boolean","uniqueItems":true}}`: "y2-LP5FDecW",
				`{"name":"z2","in":"query","description":"z","schema":{"type":"boolean","uniqueItems":true}}`: "z2-eqTbEYig",
			},
		},
		{
			name: "shared per operation",
			spec: `{
  "parameters": {"pre": {"in": "body", "name": "body", "required": true, "schema": {}}},
  "paths": {
    "/api/v1/a/{name}": {"get": {"parameters": [
		  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/a/{name}/foo": {"get": {"parameters": [
		  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}": {"get": {"parameters": [
		  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}/foo": {"get": {"parameters": [
		  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}}
    ]}}
  }
}`,
			want: map[string]string{
				`{"name":"x","in":"query","description":"x","schema":{"type":"boolean","uniqueItems":true}}`: "x-16u_rOM9",
				`{"name":"y","in":"query","description":"y","schema":{"type":"boolean","uniqueItems":true}}`: "y-bABmtr22",
				`{"name":"z","in":"query","description":"z","schema":{"type":"boolean","uniqueItems":true}}`: "z-1LKrqY5S",
			},
		},
		{
			name: "shared per path",
			spec: `{
  "parameters": {"pre": {"in": "body", "name": "body", "required": true, "schema": {}}},
  "paths": {
    "/api/v1/a/{name}": {"get": {},
      "parameters": [
		  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
      ]
    },
    "/api/v1/a/{name}/foo": {"get": {"parameters": [
		  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}},
		  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}": {"get": {},
      "parameters": [
		  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}}
      ]
    },
    "/api/v1/b/{name}/foo": {"get": {"parameters": [
		  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}}
    ]}}
  }
}`,
			want: map[string]string{
				`{"name":"x","in":"query","description":"x","schema":{"type":"boolean","uniqueItems":true}}`: "x-16u_rOM9",
				`{"name":"y","in":"query","description":"y","schema":{"type":"boolean","uniqueItems":true}}`: "y-bABmtr22",
				`{"name":"z","in":"query","description":"z","schema":{"type":"boolean","uniqueItems":true}}`: "z-1LKrqY5S",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sp *spec3.OpenAPI
			if tt.spec != "" {
				err := json.Unmarshal([]byte(tt.spec), &sp)
				require.NoError(t, err)
			}

			gotNamesByJSON, _, err := collectSharedParameters(sp)
			require.NoError(t, err)
			require.Equalf(t, tt.want, gotNamesByJSON, "unexpected shared parameters")
		})
	}
}

func TestReplaceSharedParameters(t *testing.T) {
	shared := map[string]string{
		`{"name":"x","in":"query","description":"x","schema":{"type":"boolean","uniqueItems":true}}`: "x",
		`{"name":"y","in":"query","description":"y","schema":{"type":"boolean","uniqueItems":true}}`: "y",
		`{"name":"z","in":"query","description":"z","schema":{"type":"boolean","uniqueItems":true}}`: "z",
	}

	tests := []struct {
		name string
		spec string
		want string
	}{
		{
			name: "empty",
			spec: "{}",
			want: `{"info":null,"openapi":""}`,
		},
		{
			name: "existing parameters",
			spec: `{"components":{"parameters": {"a":{"schema":{"type":"boolean"}}}}}`,
			want: `{"components":{"parameters": {"a":{"schema":{"type":"boolean"}}}},"info":null,"openapi":""}`,
		},
		{
			name: "replace",
			spec: `{
  "components":{"parameters": {"pre": {"in": "body", "name": "body", "required": true, "schema": {}}}},
  "info":null,
  "openapi":"",
  "paths": {
    "/api/v1/a/{name}": {"get": {"description":"foo"},
      "parameters": [
        {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}},
	    {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
      ]
    },
    "/api/v1/a/{name}/foo": {"get": {"parameters": [
      {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}},
	  {"description": "y","in":"query","name": "y","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}": {"get": {"parameters": [
	  {"description": "z","in":"query","name": "z","schema": {"type": "boolean", "uniqueItems": true}}
    ]}},
    "/api/v1/b/{name}/foo": {"get": {"parameters": [
	  {"description": "x","in":"query","name": "x","schema": {"type": "boolean", "uniqueItems": true}},
      {"description": "w","in":"query","name": "w","schema": {"type": "boolean", "uniqueItems": true}}
    ]}}
  }
}`,
			want: `{
  "components":{"parameters": {"pre":{"in":"body","name":"body","required":true,"schema":{}}}},
  "info":null,
  "openapi":"",
  "paths": {
    "/api/v1/a/{name}": {"get": {"description":"foo"},
      "parameters": [
        {"$ref": "#/components/parameters/x"},
        {"$ref": "#/components/parameters/y"}
      ]
    },
    "/api/v1/a/{name}/foo": {"get": {"parameters": [
      {"$ref": "#/components/parameters/z"},
      {"$ref": "#/components/parameters/y"}
    ]}},
    "/api/v1/b/{name}": {"get": {"parameters": [
      {"$ref": "#/components/parameters/z"}
    ]}},
    "/api/v1/b/{name}/foo": {"get": {"parameters": [
      {"$ref":"#/components/parameters/x"},
      {"description": "w","in":"query","name": "w","schema": {"type": "boolean", "uniqueItems": true}}
    ]}}
  }
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var unmarshalled *spec3.OpenAPI
			err := json.Unmarshal([]byte(tt.spec), &unmarshalled)
			require.NoError(t, err)

			got, err := replaceSharedParameters(shared, unmarshalled)
			require.NoError(t, err)

			require.Equalf(t, normalizeJSON(t, tt.want), normalizeJSON(t, toJSON(t, got)), "unexpected result")
		})
	}
}

func toJSON(t *testing.T, x interface{}) string {
	bs, err := json.Marshal(x)
	require.NoError(t, err)

	return string(bs)
}

func normalizeJSON(t *testing.T, j string) string {
	var obj interface{}
	err := json.Unmarshal([]byte(j), &obj)
	require.NoError(t, err)
	return toJSON(t, obj)
}

func TestOperations(t *testing.T) {
	t.Log("Ensuring that operations() returns all operations in spec.PathItemProps")
	path := &spec3.Path{}
	v := reflect.ValueOf(path.PathProps)
	var rOps []any
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Ptr {
			rOps = append(rOps, v.Field(i).Interface())
		}
	}

	ops := operations(path)
	require.Equal(t, len(rOps), len(ops), "operations() should return all operations in spec.PathItemProps")
}
