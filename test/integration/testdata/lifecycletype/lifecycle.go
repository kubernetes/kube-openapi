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

package lifecycletype

// +k8s:openapi-gen=true
type LifecycleStruct struct {
	// +lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=AlphaFeatureGate
	AlphaField string
	// +lifecycle:component=kubernetes,minVersion=v1.21,status=beta,featureGate=BetaFeatureGate
	BetaField Item
	// +lifecycle:component=kubernetes,minVersion=v1.22,status=deprecated,featureGate=DeprecatedFeatureGate
	// +listType=atomic
	DeprecatedField []string
	StableField     int

	// +lifecycle:component=kubernetes,minVersion=v1.20,status=alpha,featureGate=InvalidFeatureGate
	FieldWithInvalidFeatureGate string

	// FieldWithoutLifecycleTag is a GA field
	FieldWithoutLifecycleTag string
}

// +k8s:openapi-gen=true
type Item struct {
	Protocol string
	Port     int
}
