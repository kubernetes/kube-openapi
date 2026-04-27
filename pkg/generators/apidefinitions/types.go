/*
Copyright The Kubernetes Authors.

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

package apidefinitions

// Metadata is a thin subset of K8s ObjectMeta used by codegen manifests.
type Metadata struct {
	Name string `yaml:"name"`
}

// APIVersion declares an external versioned API package.
type APIVersion struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   Metadata       `yaml:"metadata"`
	Spec       APIVersionSpec `yaml:"spec"`
}

// APIVersionSpec carries codegen-relevant fields for an APIVersion.
type APIVersionSpec struct {
	// ModelPackage is the OpenAPI model package name for the schema
	// types defined in this group/version.
	ModelPackage string `yaml:"modelPackage"`

	// InternalVersions lists the InternalVersion metadata.name values
	// this APIVersion converts to/from. Used by extensions/v1beta1's
	// historical fan-out. openapi-gen does not consume this field but
	// keeps it in the schema for round-trip safety.
	InternalVersions []string `yaml:"internalVersions,omitempty"`
}
