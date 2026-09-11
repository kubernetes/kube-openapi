/*
Copyright 2026 The Kubernetes Authors.

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
	"k8s.io/gengo/v2/types"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// PackageOverride patches the schema of types in a single Go package while
// openapi-gen is constructing it.
type PackageOverride func(t *types.Type, schema *spec.Schema)

// KnownPackages overrides types in some packages that have semantics the
// generator cannot infer from their Go shape (e.g. runtime.RawExtension should
// carry x-kubernetes-preserve-unknown-fields: true).
//
// Author-provided values for the same key (e.g. via comment markers) are not overwritten.
var KnownPackages = map[string]PackageOverride{
	"k8s.io/apimachinery/pkg/runtime": func(t *types.Type, s *spec.Schema) {
		switch t.Name.Name {
		case "RawExtension":
			if _, ok := s.Extensions["x-kubernetes-preserve-unknown-fields"]; !ok {
				s.AddExtension("x-kubernetes-preserve-unknown-fields", true)
			}
		}
	},
}

// applyKnownPackages runs any registered override for t against schema.
func applyKnownPackages(t *types.Type, schema *spec.Schema) {
	if override, ok := KnownPackages[t.Name.Package]; ok {
		override(t, schema)
	}
}
