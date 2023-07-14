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

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jsontesting "k8s.io/kube-openapi/pkg/util/jsontesting"
)

const infoJSON = `{
	"description": "A sample API that uses a petstore as an example to demonstrate features in ` +
	`the swagger-2.0 specification",
	"title": "Swagger Sample API",
	"termsOfService": "http://helloreverb.com/terms/",
	"contact": {
		"name": "wordnik api team",
		"url": "http://developer.wordnik.com"
	},
	"license": {
		"name": "Creative Commons 4.0 International",
		"url": "http://creativecommons.org/licenses/by/4.0/"
	},
	"version": "1.0.9-abcd",
	"x-framework": "go-swagger"
}`

var info = Info{
	InfoProps: InfoProps{
		Version: "1.0.9-abcd",
		Title:   "Swagger Sample API",
		Description: "A sample API that uses a petstore as an example to demonstrate features in " +
			"the swagger-2.0 specification",
		TermsOfService: "http://helloreverb.com/terms/",
		Contact:        &ContactInfo{Name: "wordnik api team", URL: "http://developer.wordnik.com"},
		License: &License{
			Name: "Creative Commons 4.0 International",
			URL:  "http://creativecommons.org/licenses/by/4.0/",
		},
	},
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{"x-framework": "go-swagger"}},
}

func TestIntegrationInfo_Serialize(t *testing.T) {
	b, err := json.MarshalIndent(info, "", "\t")
	if assert.NoError(t, err) {
		assert.Equal(t, infoJSON, string(b))
	}
}

func TestIntegrationInfo_Deserialize(t *testing.T) {
	actual := Info{}
	err := json.Unmarshal([]byte(infoJSON), &actual)
	if assert.NoError(t, err) {
		assert.EqualValues(t, info, actual)
	}
}

func TestInfoRoundTrip(t *testing.T) {
	cases := []jsontesting.RoundTripTestCase{
		{
			// Show at least one field from each embededd struct sitll allows
			// roundtrips successfully
			Name: "UnmarshalEmbedded",
			JSON: `{
				"x-framework": "swagger-go",
				"description": "the description of this object"
			  }`,
			Object: &Info{
				VendorExtensible{Extensions{
					"x-framework": "swagger-go",
				}},
				InfoProps{
					Description: "the description of this object",
				},
			},
		}, {
			Name:   "BasicCase",
			JSON:   infoJSON,
			Object: &info,
		},
	}

	for _, tcase := range cases {
		t.Run(tcase.Name, func(t *testing.T) {
			require.NoError(t, tcase.RoundTripTest(&Info{}))
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	type stringSliceCase struct {
		name  string
		slice any

		expectedValue []string
	}

	cases := []stringSliceCase{
		{
			name:          "interfaces",
			slice:         []interface{}{"key1", "key2", "key3"},
			expectedValue: []string{"key1", "key2", "key3"},
		},
		{
			name:          "strings",
			slice:         []string{"key1", "key2", "key3"},
			expectedValue: []string{"key1", "key2", "key3"},
		},
		{
			name:  "ints",
			slice: []int{1, 2, 3},
		},
		{
			name:  "string",
			slice: "hello world",
		},
		{
			name: "badkey",
		},
	}

	exts := VendorExtensible{}
	for _, c := range cases {
		if c.slice != nil {
			exts.AddExtension(c.name, c.slice)
		}
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			slice, ok := exts.Extensions.GetStringSlice(c.name)
			assert.Equal(t, c.expectedValue != nil, ok)
			assert.Equal(t, c.expectedValue, slice)

			// Make sure modifying the result slice doesn't touch the original
			if len(slice) > 0 {
				slice[0] = "changed value"
			} else {
				slice = append(slice, "new value")
			}

			slice, ok = exts.Extensions.GetStringSlice(c.name)
			assert.Equal(t, c.expectedValue != nil, ok)
			assert.Equal(t, c.expectedValue, slice)
		})
	}
}
