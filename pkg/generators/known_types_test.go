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
	"strings"
	"testing"

	"golang.org/x/tools/go/packages/packagestest"

	"k8s.io/gengo/v2/types"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func TestKnownPackagesRawExtension(t *testing.T) {
	rawExt := &types.Type{
		Name: types.Name{Package: "k8s.io/apimachinery/pkg/runtime", Name: "RawExtension"},
		Kind: types.Struct,
	}

	var s spec.Schema
	applyKnownPackages(rawExt, &s)
	if got := s.Extensions["x-kubernetes-preserve-unknown-fields"]; got != true {
		t.Errorf("override did not set x-kubernetes-preserve-unknown-fields: got %#v, want true", got)
	}

	explicit := spec.Schema{VendorExtensible: spec.VendorExtensible{Extensions: spec.Extensions{
		"x-kubernetes-preserve-unknown-fields": false,
	}}}
	applyKnownPackages(rawExt, &explicit)
	if got := explicit.Extensions["x-kubernetes-preserve-unknown-fields"]; got != false {
		t.Errorf("explicit value was overwritten: got %#v, want false", got)
	}
}

func TestKnownPackagesGenerate(t *testing.T) {
	const fixturePkg = "example.com/base/foo"
	orig, hadOrig := KnownPackages[fixturePkg]
	KnownPackages[fixturePkg] = func(typ *types.Type, s *spec.Schema) {
		if typ.Name.Name == "Blah" {
			s.AddExtension("x-kubernetes-preserve-unknown-fields", true)
		}
	}
	defer func() {
		if hadOrig {
			KnownPackages[fixturePkg] = orig
		} else {
			delete(KnownPackages, fixturePkg)
		}
	}()

	inputFile := `
		package foo

		type Blah struct {
		}`

	packagestest.TestAll(t, func(t *testing.T, x packagestest.Exporter) {
		e := packagestest.Export(t, x, []packagestest.Module{{
			Name: "example.com/base/foo",
			Files: map[string]interface{}{
				"foo.go": inputFile,
			},
		}})
		defer e.Cleanup()

		callErr, funcErr, _, funcBuffer, _ := testOpenAPITypeWriter(t, e.Config)
		if callErr != nil {
			t.Fatal(callErr)
		}
		if funcErr != nil {
			t.Fatal(funcErr)
		}
		want := `"x-kubernetes-preserve-unknown-fields": true,`
		if !strings.Contains(funcBuffer.String(), want) {
			t.Errorf("missing %q in generated source:\n%s", want, funcBuffer.String())
		}
	})
}
