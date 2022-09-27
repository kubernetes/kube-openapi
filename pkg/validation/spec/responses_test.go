/*
 Copyright 2022 The Kubernetes Authors.

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

package spec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

var specimen *Responses = &Responses{
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{
		"x-<vBŤç,ʡËdSS暙ɑɮ": string("鄲(兴ȑʦ衈覻鋕嚮峡jw逓:鮕虫F迢."),
		"x-h":              string(""), "x-岡ʍ": string("Đɻ/nǌo鿻曑Œ TĀyĢ"),
		"x-绅ƄȆ疩ã[魑銒;苎#砠zPȺ5Aù": string("閲ǉǠyư")},
	},
	ResponsesProps: ResponsesProps{
		Default: &Response{
			Refable:       Refable{Ref: MustCreateRef("Dog")},
			ResponseProps: ResponseProps{Description: "梱bȿF)渽Ɲō-%x"},
		},
		StatusCodeResponses: map[int]Response{
			200: {
				Refable:       Refable{Ref: MustCreateRef("Cat")},
				ResponseProps: ResponseProps{Description: "梱bȿF)渽Ɲō-%x"},
			},
		},
	},
}

func TestResponsesRoundtrip(t *testing.T) {
	jsonText, err := json.Marshal(specimen)
	require.NoError(t, err)

	var decoded Responses
	err = json.Unmarshal(jsonText, &decoded)

	if !reflect.DeepEqual(specimen, &decoded) {
		t.Fatal(cmp.Diff(specimen, &decoded))
	}
}
