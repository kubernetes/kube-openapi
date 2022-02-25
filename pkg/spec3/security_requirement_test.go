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
)

var securityRequirementCases = []struct {
	name           string
	target         *spec3.SecurityRequirement
	expectedOutput string
	yaml           []byte
}{
	{
		name: "Non-OAuth2 Security Requirement",
		target: &spec3.SecurityRequirement{
			SecurityRequirementProps: map[string][]string{
				"api_key": []string{},
			},
		},
		expectedOutput: `{"api_key":[]}`,
		yaml: []byte(`
api_key: []
`),
	},
	{
		name: "OAuth2 Security Requirement",
		target: &spec3.SecurityRequirement{
			SecurityRequirementProps: map[string][]string{
				"petstore_auth": []string{
					"write_pets",
					"read:pets",
				},
			},
		},
		expectedOutput: `{"petstore_auth":["write_pets","read:pets"]}`,
		yaml: []byte(`
petstore_auth:
- write_pets
- read:pets
`),
	},
}

func TestSecurityRequirementJSONSerialization(t *testing.T) {
	for _, tc := range securityRequirementCases {
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

func TestSecurityRequirementYAMLDeserialization(t *testing.T) {
	for _, tc := range securityRequirementCases {
		t.Run(tc.name, func(t *testing.T) {
			// var nodes yaml.Node
			var actual spec3.SecurityRequirement

			err := yaml.Unmarshal(tc.yaml, &actual)
			if err != nil {
				t.Fatal(err)
			}

			require.EqualValues(t, tc.target, &actual, "round trip")
		})
	}
}
