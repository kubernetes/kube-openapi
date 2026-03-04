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

package generators

import (
	"testing"

	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/types"
	"k8s.io/kube-openapi/cmd/openapi-gen/args"
)

func TestIsReadOnlyPkg(t *testing.T) {
	tests := []struct {
		name         string
		pkgPath      string
		readOnlyPkgs []string
		want         bool
	}{
		{
			name:         "nil readonly pkgs matches nothing",
			pkgPath:      "k8s.io/apimachinery/pkg/runtime",
			readOnlyPkgs: nil,
			want:         false,
		},
		{
			name:         "empty readonly pkgs matches nothing",
			pkgPath:      "k8s.io/apimachinery/pkg/runtime",
			readOnlyPkgs: []string{},
			want:         false,
		},
		{
			name:         "exact match",
			pkgPath:      "k8s.io/apimachinery/pkg/apis/meta/v1",
			readOnlyPkgs: []string{"k8s.io/apimachinery/pkg/apis/meta/v1"},
			want:         true,
		},
		{
			name:         "no prefix match",
			pkgPath:      "k8s.io/apimachinery/pkg/apis/meta/v1",
			readOnlyPkgs: []string{"k8s.io/apimachinery"},
			want:         false,
		},
		{
			name:         "non-readonly package",
			pkgPath:      "k8s.io/sample-apiserver/pkg/apis/wardle/v1beta1",
			readOnlyPkgs: []string{"k8s.io/apimachinery/pkg/apis/meta/v1"},
			want:         false,
		},
		{
			name:         "multiple readonly pkgs",
			pkgPath:      "k8s.io/apimachinery/pkg/runtime",
			readOnlyPkgs: []string{"k8s.io/apimachinery/pkg/apis/meta/v1", "k8s.io/apimachinery/pkg/runtime"},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadOnlyPkg(tt.pkgPath, tt.readOnlyPkgs)
			if got != tt.want {
				t.Errorf("isReadOnlyPkg(%q, %v) = %v, want %v", tt.pkgPath, tt.readOnlyPkgs, got, tt.want)
			}
		})
	}
}

func TestGetModelNameTargets_ReadOnlyPkgs(t *testing.T) {
	localPkg := "k8s.io/sample-apiserver/pkg/apis/wardle/v1beta1"
	depPkg := "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Build a minimal generator.Context with two input packages:
	// one "local" package and one "dependency" package. Both have
	// the +k8s:openapi-model-package tag and a public struct type,
	// so without readonly pkgs both would produce targets.
	universe := types.Universe{
		localPkg: {
			Path: localPkg,
			Dir:  "/fake/local/wardle/v1beta1",
			Name: "v1beta1",
			Comments: []string{
				"+k8s:openapi-model-package=io.k8s.sample-apiserver.pkg.apis.wardle.v1beta1",
			},
			Types: map[string]*types.Type{
				"Flunder": {
					Name: types.Name{Package: localPkg, Name: "Flunder"},
					Kind: types.Struct,
				},
			},
		},
		depPkg: {
			Path: depPkg,
			Dir:  "/fake/gomodcache/k8s.io/apimachinery/pkg/apis/meta/v1",
			Name: "v1",
			Comments: []string{
				"+k8s:openapi-model-package=io.k8s.apimachinery.pkg.apis.meta.v1",
			},
			Types: map[string]*types.Type{
				"ObjectMeta": {
					Name: types.Name{Package: depPkg, Name: "ObjectMeta"},
					Kind: types.Struct,
				},
			},
		},
	}

	ctx := &generator.Context{
		Inputs:   []string{localPkg, depPkg},
		Universe: universe,
	}

	tests := []struct {
		name           string
		readOnlyPkgs   []string
		wantTargetPkgs []string
	}{
		{
			name:           "no readonly pkgs generates for all packages",
			readOnlyPkgs:   nil,
			wantTargetPkgs: []string{localPkg, depPkg},
		},
		{
			name:           "readonly dep excludes dependency",
			readOnlyPkgs:   []string{depPkg},
			wantTargetPkgs: []string{localPkg},
		},
		{
			name:           "readonly local excludes local",
			readOnlyPkgs:   []string{localPkg},
			wantTargetPkgs: []string{depPkg},
		},
		{
			name:           "both readonly excludes all",
			readOnlyPkgs:   []string{localPkg, depPkg},
			wantTargetPkgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &args.Args{
				OutputModelNameFile: "zz_generated.model_name.go",
				ReadOnlyPkgs:        tt.readOnlyPkgs,
			}

			targets := GetModelNameTargets(ctx, a, nil)

			gotPkgs := make([]string, 0, len(targets))
			for _, tgt := range targets {
				st, ok := tgt.(*generator.SimpleTarget)
				if !ok {
					t.Fatalf("unexpected target type %T", tgt)
				}
				gotPkgs = append(gotPkgs, st.PkgPath)
			}

			if len(gotPkgs) != len(tt.wantTargetPkgs) {
				t.Fatalf("got %d targets %v, want %d targets %v", len(gotPkgs), gotPkgs, len(tt.wantTargetPkgs), tt.wantTargetPkgs)
			}

			for i, want := range tt.wantTargetPkgs {
				if gotPkgs[i] != want {
					t.Errorf("target[%d] = %q, want %q", i, gotPkgs[i], want)
				}
			}
		})
	}
}
