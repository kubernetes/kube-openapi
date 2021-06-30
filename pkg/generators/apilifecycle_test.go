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

package generators

import (
	"reflect"
	"testing"
)

func TestParseAPILifecycleComments(t *testing.T) {
	var tests = []struct {
		name          string
		comments      []string
		lifecycleInfo map[string]string
		expectError   bool
	}{
		{
			name:          "valid tag",
			comments:      []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate"},
			lifecycleInfo: map[string]string{"component": "kubernetes", "minVersion": "v1.20", "status": "alpha", "featureGate": "AlphaFeatureGate"},
		},
		{
			name: "two tags specified, only first considered",
			comments: []string{
				"+lifecycle:component=kubernetes,minVersion=v1.21,status=beta,featureGate=BetaFeatureGate",
				"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate",
			},
			lifecycleInfo: map[string]string{"component": "kubernetes", "minVersion": "v1.21", "status": "beta", "featureGate": "BetaFeatureGate"},
		},
		{
			name:        "invalid syntax for minVersion",
			comments:    []string{"+lifecycle:component=kubernetes,minVersion:v1.20,status=alpha,featureGate=AlphaFeatureGate"},
			expectError: true,
		},
		{
			name:        "invalid featureGate key",
			comments:    []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGateName=AlphaFeatureGate"},
			expectError: true,
		},
		{
			name:        "extra invalid key",
			comments:    []string{"+lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate,extra=key"},
			expectError: true,
		},
	}
	for _, test := range tests {
		actualInfo, errs := parseAPILifecycle(test.comments)
		if !test.expectError && len(errs) > 0 {
			for _, e := range errs {
				t.Errorf("%s: error while parsing: %v", test.name, e)
			}
		}
		if !reflect.DeepEqual(actualInfo, test.lifecycleInfo) {
			t.Errorf("%s: expected (%s), actual (%s)\n", test.name, test.lifecycleInfo, actualInfo)
		}
	}
}

func TestValidateAPILifecycleComments(t *testing.T) {
	var tests = []struct {
		name        string
		keys        map[string]string
		expectError bool
	}{
		{
			name: "valid keys",
			keys: map[string]string{"component": "kubernetes", "minVersion": "v1.20", "status": "alpha", "featureGate": "AlphaFeatureGate"},
		},
		{
			name:        "invalid featureGate key",
			keys:        map[string]string{"component": "kubernetes", "minVersion": "v1.20", "status": "alpha", "featureGateName": "AlphaFeatureGate"},
			expectError: true,
		},
		{
			name:        "extra invalid tag key",
			keys:        map[string]string{"component": "kubernetes", "minVersion": "v1.20", "status": "alpha", "featureGate": "AlphaFeatureGate", "Extra": "Key"},
			expectError: true,
		},
		{
			name:        "invalid status value",
			keys:        map[string]string{"component": "kubernetes", "minVersion": "v1.20", "status": "stable", "featureGate": "AlphaFeatureGate"},
			expectError: true,
		},
	}
	for _, test := range tests {
		errs := validateAPILifecycleKeys(test.keys)
		if !test.expectError && len(errs) > 0 {
			for _, e := range errs {
				t.Errorf("%s: error while parsing: %v", test.name, e)
			}
		}
	}
}
