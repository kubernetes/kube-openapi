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

package features

import "k8s.io/kube-openapi/test/integration/testdata/featuregatetype/featuregate"

const (
	// owner: @foo
	// alpha: v1.20
	//
	// AlphaFeatureGate is alpha in v1.20
	AlphaFeatureGate featuregate.Feature = "AlphaFeatureGate"

	// owner: @foo
	// alpha: v1.20
	// beta: v1.21
	//
	// BetaFeatureGate is beta in v1.21
	BetaFeatureGate featuregate.Feature = "BetaFeatureGate"

	// owner: @foo
	// alpha: v1.20
	// beta: v1.21
	// stable: v1.22
	//
	// StableFeatureGate is stable in v1.22
	StableFeatureGate featuregate.Feature = "StableFeatureGate"

	// owner: @foo
	// alpha: v1.20
	// beta: v1.21
	// deprecated: v1.22
	//
	// DeprecatedFeatureGate is deprecated in v1.22
	DeprecatedFeatureGate featuregate.Feature = "DeprecatedFeatureGate"

	// NotAFeatureGate is a const not of the type featuregate.Feature
	NotAFeatureGate = "NotAFeatureGate"
)
