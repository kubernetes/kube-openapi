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

package schemaconv

import (
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"k8s.io/kube-openapi/pkg/util/proto"
	prototesting "k8s.io/kube-openapi/pkg/util/proto/testing"
)

var (
	fakeSchema = prototesting.Fake{Path: filepath.Join("..", "util", "proto", "testdata", "swagger.json")}
)

func TestToSchema(t *testing.T) {
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatal(err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := ToSchema(models)
	if err != nil {
		t.Fatal(err)
	}
	str, err := yaml.Marshal(ns)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(str))
}
