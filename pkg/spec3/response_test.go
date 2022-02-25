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

package spec3_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

var responseCases = []struct {
	name           string
	target         *spec3.Response
	expectedOutput string
	yaml           []byte
}{
	// scenario 1
	{
		name: "basic",
		target: &spec3.Response{
			ResponseProps: spec3.ResponseProps{
				Content: map[string]*spec3.MediaType{
					"text/plain": &spec3.MediaType{
						MediaTypeProps: spec3.MediaTypeProps{
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
		expectedOutput: `{"content":{"text/plain":{"schema":{"type":"string"}}}}`,
		yaml: []byte(`
content:
  text/plain:
    schema:
      type: string
`),
	},
}

func TestResponseJSONSerialization(t *testing.T) {
	for _, tc := range responseCases {
		t.Run(tc.name, func(t *testing.T) {
			rawTarget, err := json.Marshal(tc.target)
			if err != nil {
				t.Fatal(err)
			}
			serializedTarget := string(rawTarget)
			if !cmp.Equal(serializedTarget, tc.expectedOutput) {
				t.Fatalf("diff %s", cmp.Diff(serializedTarget, tc.expectedOutput))
			}
		})
	}
}

func TestResponseYAMLDeserialization(t *testing.T) {
	for _, tc := range responseCases {
		t.Run(tc.name, func(t *testing.T) {
			// var nodes yaml.Node
			var actual spec3.Response

			err := yaml.Unmarshal(tc.yaml, &actual)
			if err != nil {
				t.Fatal(err)
			}

			require.EqualValues(t, tc.target, &actual, "round trip")
		})
	}
}
