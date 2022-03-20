/*
Copyright 2017 The Kubernetes Authors.

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

package testing

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"k8s.io/kube-openapi/pkg/handler3"
)

// Fake opens and returns a openapi swagger from a file Path. It will
// parse only once and then return the same copy everytime.
type Fake struct {
	Path string

	once     sync.Once
	document *openapi_v2.Document
	err      error
}

// OpenAPISchema returns the openapi document and a potential error.
func (f *Fake) OpenAPISchema() (*openapi_v2.Document, error) {
	f.once.Do(func() {
		_, err := os.Stat(f.Path)
		if err != nil {
			f.err = err
			return
		}
		spec, err := ioutil.ReadFile(f.Path)
		if err != nil {
			f.err = err
			return
		}
		f.document, f.err = openapi_v2.ParseDocument(spec)
	})
	return f.document, f.err
}

func (f *Fake) OpenAPIV3Discovery() (*handler3.OpenAPIV3Discovery, error) {
	// Read directory to determine groups
	res := &handler3.OpenAPIV3Discovery{}
	filepath.WalkDir(f.Path, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) != "json" {
			return nil
		}

		res.Paths[path] = handler3.OpenAPIV3DiscoveryGroupVersion{
			URL: "/openapi/v3/" + path,
		}
		return nil
	})

	return res, nil
}

func (f *Fake) OpenAPIV3Schema(groupVersion string) (*openapi_v3.Document, error) {
	_, err := os.Stat(f.Path)
	if err != nil {
		return nil, err
	}
	spec, err := ioutil.ReadFile(filepath.Join(f.Path, groupVersion+".json"))
	if err != nil {
		return nil, err
	}
	return openapi_v3.ParseDocument(spec)
}

type Empty struct{}

func (Empty) OpenAPISchema() (*openapi_v2.Document, error) {
	return nil, nil
}
