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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	openapi_v3 "github.com/google/gnostic/openapiv3"
)

// Fake opens and returns a openapi swagger from a file Path. It will
// parse only once and then return the same copy everytime.
type Fake struct {
	Path string

	once     sync.Once
	document *openapi_v2.Document
	err      error

	v3DocumentsLock sync.Mutex
	v3Documents     map[string]*openapi_v3.Document
	v3Errors        map[string]error
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

func (f *Fake) OpenAPIV3Schema(groupVersion string) (*openapi_v3.Document, error) {
	f.v3DocumentsLock.Lock()
	defer f.v3DocumentsLock.Unlock()

	if existing, ok := f.v3Documents[groupVersion]; ok {
		return existing, nil
	} else if existingError, ok := f.v3Errors[groupVersion]; ok {
		return nil, existingError
	}

	_, err := os.Stat(f.Path)
	if err != nil {
		return nil, err
	}
	spec, err := ioutil.ReadFile(filepath.Join(f.Path, groupVersion+".json"))
	if err != nil {
		return nil, err
	}

	if f.v3Documents == nil {
		f.v3Documents = make(map[string]*openapi_v3.Document)
	}

	if f.v3Errors == nil {
		f.v3Errors = make(map[string]error)
	}

	result, err := openapi_v3.ParseDocument(spec)
	if err != nil {
		f.v3Errors[groupVersion] = err
		return nil, err
	}

	f.v3Documents[groupVersion] = result
	return result, nil
}

type Empty struct{}

func (Empty) OpenAPISchema() (*openapi_v2.Document, error) {
	return nil, nil
}
