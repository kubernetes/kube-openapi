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

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

const (
	schemeGroupVersion = "apidefinitions.k8s.io/v1alpha1"
	kindAPIVersion     = "APIVersion"

	fileName = "apiversion.yaml"
)

// LoadAPIVersion reads apiversion.yaml from a Go package dir, returning
// nil if no such file exists.
func LoadAPIVersion(dir string) (*APIVersion, error) {
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	av := &APIVersion{}
	if err := yaml.Unmarshal(data, av); err != nil {
		return nil, fmt.Errorf("%s: decoding %s: %w", filepath.Join(dir, fileName), kindAPIVersion, err)
	}
	if av.APIVersion != schemeGroupVersion {
		return nil, fmt.Errorf("%s: expected apiVersion %s but got %s", filepath.Join(dir, fileName), schemeGroupVersion, av.APIVersion)
	}
	if av.Kind != kindAPIVersion {
		return nil, fmt.Errorf("%s: expected kind %s but got %s", filepath.Join(dir, fileName), kindAPIVersion, av.Kind)
	}
	if av.Metadata.Name == "" {
		return nil, fmt.Errorf("%s: metadata.name is required", filepath.Join(dir, fileName))
	}
	if av.Spec.ModelPackage == "" {
		return nil, fmt.Errorf("%s: spec.modelPackage is required", filepath.Join(dir, fileName))
	}
	return av, nil
}

// Group returns the API group declared by an APIVersion. metadata.name is
// parsed as "<group>/<version>" or "<version>" (core group, empty group).
func (av *APIVersion) Group() string {
	if i := strings.LastIndex(av.Metadata.Name, "/"); i >= 0 {
		return av.Metadata.Name[:i]
	}
	return ""
}

// Version returns the API version declared by an APIVersion.
func (av *APIVersion) Version() string {
	if i := strings.LastIndex(av.Metadata.Name, "/"); i >= 0 {
		return av.Metadata.Name[i+1:]
	}
	return av.Metadata.Name
}
