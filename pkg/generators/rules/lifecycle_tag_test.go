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

var fileWithAllGatesMap = map[string]FeatureGateInfo{
	"AlphaFeatureGate":      {status: "alpha", minVersion: "v1.20"},
	"BetaFeatureGate":       {status: "beta", minVersion: "v1.21"},
	"DeprecatedFeatureGate": {status: "deprecated", minVersion: "v1.22"},
}

var multipleFilesMap = map[string]FeatureGateInfo{
	"AlphaFeatureGate":        {status: "alpha", minVersion: "v1.20"},
	"BetaFeatureGate":         {status: "beta", minVersion: "v1.21"},
	"DeprecatedFeatureGate":   {status: "deprecated", minVersion: "v1.22"},
	"AnotherAlphaFeatureGate": {status: "alpha", minVersion: "v1.20"},
}

func TestParseFeatureGateFiles(t *testing.T) {
	var tests = []struct {
		name        string
		fileNames   []string
		expectedMap map[string]FeatureGateInfo
	}{
		{
			name:        "all gates",
			fileNames:   []string{"../../../test/integration/testdata/featuregatetype/all_featuregate.go"},
			expectedMap: fileWithAllGatesMap,
		},
		{
			name:        "multiple files",
			fileNames:   []string{"../../../test/integration/testdata/featuregatetype/all_featuregate.go", "../../../test/integration/testdata/featuregatetype/extra_featuregate.go"},
			expectedMap: multipleFilesMap,
		},
	}

	for _, test := range tests {
		actualMap, err := ParseFeatureGateFiles(test.fileNames)
		if err != nil {
			t.Errorf("%s: error parsing feature gates file: %v", test.name, err)
		}

		if !reflect.DeepEqual(actualMap, test.expectedMap) {
			t.Errorf("%s: expected (%s), actual (%s)\n", test.name, test.expectedMap, actualMap)
		}
	}
}

func TestValidateFeatureGates(t *testing.T) {
	var tests = []struct {
		name             string
		featureGateValue string
		statusValue      string
		minVersionValue  string
		valid            bool
	}{
		{
			name:             "valid keys",
			featureGateValue: "AlphaFeatureGate",
			statusValue:      "alpha",
			minVersionValue:  "v1.20",
			valid:            true,
		},
		{
			name:             "invalid feature gate",
			featureGateValue: "NotAFeatureGate",
			statusValue:      "alpha",
			minVersionValue:  "v1.20",
		},
		{
			name:             "invalid status",
			featureGateValue: "AlphaFeatureGate",
			statusValue:      "beta",
			minVersionValue:  "v1.20",
		},
		{
			name:             "invalid minVersion",
			featureGateValue: "AlphaFeatureGate",
			statusValue:      "alpha",
			minVersionValue:  "v1.22",
		},
		{
			name:             "invalid status and minVersion",
			featureGateValue: "AlphaFeatureGate",
			statusValue:      "beta",
			minVersionValue:  "v1.20",
		},
	}

	for _, test := range tests {
		featureGateMap, err := ParseFeatureGateFiles([]string{"../../../test/integration/testdata/featuregatetype/all_featuregate.go"})
		if err != nil {
			t.Errorf("%s: error parsing feature gates file: %v", test.name, err)
		}
		rule := &APILifecyleTag{
			FeatureGateMap: featureGateMap,
		}

		err = rule.validateFeatureGates(test.featureGateValue, test.statusValue, test.minVersionValue)

		if test.valid && err != nil {
			t.Errorf("%s: expected to be valid but is invalid: %v", test.name, err)
		}
		if !test.valid && err == nil {
			t.Errorf("%s: expected to be invalid but is valid", test.name)
		}
	}
}

func TestValidateLifecycle(t *testing.T) {
	var tests = []struct {
		name               string
		t                  *types.Type
		expectedViolations []string
	}{
		{
			name: "valid comment",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "AlphaField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate"},
					},
				},
			},
			expectedViolations: []string{},
		},
		{
			name: "multiple lifecycle comments on single field",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name: "AlphaField",
						CommentLines: []string{
							"+lifecycle:component=kubernetes,minVersion=v1.21,status=beta,featureGate=BetaFeatureGate",
							"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate",
						},
					},
				},
			},
			expectedViolations: []string{},
		},
		{
			name: "missing featureGate key",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "AlphaField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha"},
					},
				},
			},
			expectedViolations: []string{"AlphaField"},
		},
		{
			name: "invalid featureGate key, component name and key=value format",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "AlphaField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGateName=AlphaFeatureGate"},
					},
					{
						Name:         "BetaField",
						CommentLines: []string{"+lifecycle:component=foo,minVersion=v1.21,status=beta,featureGate=BetaFeatureGate"},
					},
					{
						Name:         "DeprecatedField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.22,status:deprecated,featureGate=DeprecatedFeatureGate"},
					},
				},
			},
			expectedViolations: []string{"AlphaField", "BetaField", "DeprecatedField"},
		},
		{
			name: "invalid values for minVersion, status and featureGate",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "AlphaField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=1.20,status=alpha,featureGate=AlphaFeatureGate"},
					},
					{
						Name:         "StableField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.22,status=stable,featureGate=StableFeatureGate"},
					},
					{
						Name:         "DeprecatedField",
						CommentLines: []string{"+lifecycle:component=kubernetes,minVersion=v1.22,status=deprecated,featureGate=InvalidFeatureGate"},
					},
				},
			},
			expectedViolations: []string{"AlphaField", "StableField", "DeprecatedField"},
		},
		{
			name: "single field with multiple errors in comment tag",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "AlphaField",
						CommentLines: []string{"+lifecycle:component=foo,minVersion=v1.20,status:alpha,featureGateName=AlphaFeatureGate"},
					},
				},
			},
			expectedViolations: []string{"AlphaField"},
		},
		{
			name: "field without lifecycle tag should be treated as an exception/violation",
			t: &types.Type{
				Kind: types.Struct,
				Members: []types.Member{
					{
						Name:         "FieldWithoutLifecycleTag",
						CommentLines: nil,
					},
				},
			},
			expectedViolations: []string{"FieldWithoutLifecycleTag"},
		},
	}

	for _, test := range tests {
		featureGateMap, err := ParseFeatureGateFiles([]string{"../../../test/integration/testdata/featuregatetype/all_featuregate.go"})
		if err != nil {
			t.Errorf("%s: error parsing feature gates file: %v", test.name, err)
		}
		rule := &APILifecyleTag{
			FeatureGateMap: featureGateMap,
		}
		actualViolations, err := rule.Validate(test.t)
		if err != nil {
			t.Errorf("%s: error validating lifecycle tag: %v", test.name, err)
		}
		if !reflect.DeepEqual(actualViolations, test.expectedViolations) {
			t.Errorf("%s: unexpected validation result: want: %v, got: %v", test.name, test.expectedViolations, actualViolations)
		}
	}
}
