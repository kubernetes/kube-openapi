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
	"os"
	"path/filepath"
	"strings"
	"testing"

	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"github.com/google/go-cmp/cmp"
	yaml "go.yaml.in/yaml/v2"
	"sigs.k8s.io/structured-merge-diff/v7/schema"

	"k8s.io/kube-openapi/pkg/util/proto"
	prototesting "k8s.io/kube-openapi/pkg/util/proto/testing"
)

func TestToSchema(t *testing.T) {
	tests := []struct {
		name                   string
		openAPIFilename        string
		expectedSchemaFilename string
	}{
		{
			name:                   "kubernetes",
			openAPIFilename:        "swagger.json",
			expectedSchemaFilename: "new-schema.yaml",
		},
		{
			name:                   "atomics",
			openAPIFilename:        "atomic-types.json",
			expectedSchemaFilename: "atomic-types.yaml",
		},
		{
			name:                   "defaults",
			openAPIFilename:        "defaults.json",
			expectedSchemaFilename: "defaults.yaml",
		},
		{
			name:                   "preserve-unknown",
			openAPIFilename:        "preserve-unknown.json",
			expectedSchemaFilename: "preserve-unknown.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			openAPIPath := filepath.Join("testdata", tc.openAPIFilename)
			expectedNewSchemaPath := filepath.Join("testdata", tc.expectedSchemaFilename)
			testToSchema(t, openAPIPath, expectedNewSchemaPath)
		})
	}
}

func testToSchema(t *testing.T, openAPIPath, expectedNewSchemaPath string) {
	fakeSchema := prototesting.Fake{Path: openAPIPath}
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to get schema for %s: %v", openAPIPath, err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := ToSchema(models)
	if err != nil {
		t.Fatal(err)
	}
	got, err := yaml.Marshal(ns)
	if err != nil {
		t.Fatal(err)
	}

	expect, err := os.ReadFile(expectedNewSchemaPath)
	if err != nil {
		t.Fatalf("Unable to read golden data file %q: %v", expectedNewSchemaPath, err)
	}

	if string(expect) != string(got) {
		t.Errorf("Computed schema did not match %q.", expectedNewSchemaPath)
		t.Logf("To recompute this file, run:\n\tgo run ./cmd/openapi2smd/openapi2smd.go < %q > %q",
			filepath.Join("pkg", "schemaconv", openAPIPath),
			filepath.Join("pkg", "schemaconv", expectedNewSchemaPath),
		)
		t.Log("You can then use `git diff` to see the changes.")
		t.Error(cmp.Diff(string(expect), string(got)))
	}
}

func TestFieldLevelAnnotation(t *testing.T) {
	openAPIPath := filepath.Join("testdata", "field-level-annotation.json")
	fakeSchema := prototesting.Fake{Path: openAPIPath}
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to get schema for %s: %v", openAPIPath, err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := ToSchema(models)
	if err != nil {
		t.Fatal(err)
	}

	_ = ns

	// Test to make sure that MapElementRelationship is populated correctly
	// after being converted from proto type
	endpointAddress, ok := ns.FindNamedType("io.k8s.api.core.v1.EndpointAddress")
	if !ok {
		t.Fatalf("expected to find EndpointAddress")
	}

	targetRef, ok := endpointAddress.FindField("targetRef")
	if !ok {
		t.Fatalf("expected to find EndpointAddress field 'targetRef'")
	}

	if targetRef.Type.ElementRelationship == nil {
		t.Fatalf("expected targetRef MapElementRelationship to be atomic")
	}

	// Test to make sure that schema.Resolve overrides the ElementRelationship
	// when asked to resolve a reference
	resolved, ok := ns.Resolve(targetRef.Type)
	if !ok {
		t.Fatalf("failed to resolve targetRef type")
	}

	if resolved.Map == nil {
		t.Fatalf("expected to resolve to a Map")
	}

	if resolved.Map.ElementRelationship != *targetRef.Type.ElementRelationship {
		t.Fatalf("resolved element relationship not converted")
	}

	// Make sure our test is actually testing something by ensuring the original
	// relationship is different from what we are changing it to.
	targetRefWithoutRelationship := targetRef
	targetRefWithoutRelationship.Type.ElementRelationship = nil

	originalResolved, ok := ns.Resolve(targetRefWithoutRelationship.Type)
	if !ok {
		t.Fatalf("failed to resolve targetRef type")
	}

	if originalResolved.Map.ElementRelationship == *targetRef.Type.ElementRelationship {
		t.Fatalf("expected original element relationship to differ from field-level override for test")
	}
}

// TestMakeUnionsMapShapes verifies that makeUnions and toMap handle extension
// maps decoded from both go.yaml.in/yaml/v2 (map[interface{}]interface{}, used
// by OpenAPI v2 proto.Models) and go.yaml.in/yaml/v3 or encoding/json
// (map[string]interface{}, used by OpenAPI v3 proto.Models), and reject
// malformed union or discriminator map structures with descriptive errors.
func TestMakeUnionsMapShapes(t *testing.T) {
	t.Run("produces identical SMD unions from yaml.v2 map[interface{}]interface{} and yaml.v3/JSON map[string]interface{}", func(t *testing.T) {
		fromYAMLv2 := map[string]interface{}{
			"x-kubernetes-unions": []interface{}{
				map[interface{}]interface{}{
					"discriminator":            "type",
					"fields-to-discriminateBy": map[interface{}]interface{}{"a": "A", "b": "B"},
				},
			},
		}
		fromJSONOrYAMLv3 := map[string]interface{}{
			"x-kubernetes-unions": []interface{}{
				map[string]interface{}{
					"discriminator":            "type",
					"fields-to-discriminateBy": map[string]interface{}{"a": "A", "b": "B"},
				},
			},
		}
		gotYAMLv2, err := makeUnions(fromYAMLv2)
		if err != nil {
			t.Fatalf("unexpected error for yaml.v2 map[interface{}]interface{} shape: %v", err)
		}
		gotJSONOrYAMLv3, err := makeUnions(fromJSONOrYAMLv3)
		if err != nil {
			t.Fatalf("unexpected error for yaml.v3/JSON map[string]interface{} shape: %v", err)
		}
		if len(gotYAMLv2) != 1 || gotYAMLv2[0].Discriminator == nil || *gotYAMLv2[0].Discriminator != "type" || len(gotYAMLv2[0].Fields) != 2 {
			t.Fatalf("unexpected union decoded from yaml.v2 shape: %#v", gotYAMLv2)
		}
		if diff := cmp.Diff(gotYAMLv2, gotJSONOrYAMLv3); diff != "" {
			t.Errorf("unions differ between yaml.v2 and yaml.v3/JSON map shapes (-yaml.v2 +yaml.v3):\n%s", diff)
		}
	})

	t.Run("returns error when x-kubernetes-unions item is neither map[string]interface{} nor map[interface{}]interface{}", func(t *testing.T) {
		nonMapUnionItem := map[string]interface{}{
			"x-kubernetes-unions": []interface{}{
				"not-a-map-item",
			},
		}
		_, err := makeUnions(nonMapUnionItem)
		if err == nil {
			t.Fatal("expected error when x-kubernetes-unions item is a string, got nil")
		}
		if !strings.Contains(err.Error(), "items should be a map of string to interface") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("returns error when yaml.v2 map[interface{}]interface{} union entry contains a non-string key", func(t *testing.T) {
		nonStringKeyInUnionMap := map[string]interface{}{
			"x-kubernetes-unions": []interface{}{
				map[interface{}]interface{}{
					123: "non-string-key",
				},
			},
		}
		_, err := makeUnions(nonStringKeyInUnionMap)
		if err == nil {
			t.Fatal("expected error when union map[interface{}]interface{} contains integer key, got nil")
		}
		if !strings.Contains(err.Error(), "has non-string key: 123") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("returns error when fields-to-discriminateBy is not a valid string-keyed map", func(t *testing.T) {
		nonStringKeyInDiscriminateBy := map[string]interface{}{
			"x-kubernetes-unions": []interface{}{
				map[string]interface{}{
					"discriminator": "type",
					"fields-to-discriminateBy": map[interface{}]interface{}{
						456: "FieldA",
					},
				},
			},
		}
		_, err := makeUnions(nonStringKeyInDiscriminateBy)
		if err == nil {
			t.Fatal("expected error when fields-to-discriminateBy contains non-string key, got nil")
		}
		if !strings.Contains(err.Error(), `"fields-to-discriminateBy" must be a map[string]string`) || !strings.Contains(err.Error(), "has non-string key: 456") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// TestToSchemaFromOpenAPIV3Models verifies the end-to-end conversion from an
// OpenAPI v3 document (openapi_v3.Document) through proto.NewOpenAPIV3Data into
// a structured-merge-diff schema via schemaconv.ToSchema:
//  1. Fields using an OpenAPI 3.0 allOf-wrapped $ref with sibling metadata
//     (e.g. PodSecurityContext.seccompProfile) resolve as SMD NamedType
//     references rather than collapsing to opaque untyped scalars.
//  2. Models defining x-kubernetes-unions (decoded by SpecificationExtensionToMap
//     via go.yaml.in/yaml/v3 into map[string]interface{}) populate SMD Map.Unions.
func TestToSchemaFromOpenAPIV3Models(t *testing.T) {
	v3Spec := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Kubernetes", "version": "v1"},
		"paths": {},
		"components": {
			"schemas": {
				"io.k8s.api.core.v1.PodSecurityContext": {
					"type": "object",
					"properties": {
						"seccompProfile": {
							"description": "The seccomp options to use by the containers in this pod.",
							"default": {},
							"allOf": [
								{
									"$ref": "#/components/schemas/io.k8s.api.core.v1.SeccompProfile"
								}
							]
						}
					},
					"x-kubernetes-group-version-kind": [
						{"group": "", "version": "v1", "kind": "PodSecurityContext"}
					]
				},
				"io.k8s.api.core.v1.SeccompProfile": {
					"type": "object",
					"properties": {
						"type": {"type": "string"},
						"localhostProfile": {"type": "string"}
					},
					"x-kubernetes-unions": [
						{
							"discriminator": "type",
							"fields-to-discriminateBy": {
								"localhostProfile": "LocalhostProfile"
							}
						}
					]
				}
			}
		}
	}`)
	doc, err := openapi_v3.ParseDocument(v3Spec)
	if err != nil {
		t.Fatalf("failed to parse OpenAPI v3 document: %v", err)
	}
	models, err := proto.NewOpenAPIV3Data(doc)
	if err != nil {
		t.Fatalf("failed to build OpenAPI v3 proto.Models: %v", err)
	}

	smdSchema, err := ToSchema(models)
	if err != nil {
		t.Fatalf("ToSchema(v3Models) failed: %v", err)
	}

	t.Run("allOf-wrapped $ref field converts to SMD NamedType reference", func(t *testing.T) {
		psc, ok := smdSchema.FindNamedType("io.k8s.api.core.v1.PodSecurityContext")
		if !ok {
			t.Fatalf("expected to find PodSecurityContext in SMD schema")
		}
		seccompField, ok := psc.FindField("seccompProfile")
		if !ok {
			t.Fatalf("expected to find seccompProfile field on PodSecurityContext")
		}
		if seccompField.Type.NamedType == nil || *seccompField.Type.NamedType != "io.k8s.api.core.v1.SeccompProfile" {
			t.Fatalf("expected seccompProfile to reference NamedType io.k8s.api.core.v1.SeccompProfile, got %#v", seccompField.Type)
		}
	})

	t.Run("x-kubernetes-unions extension decoded from OpenAPI v3 populates SMD Map.Unions", func(t *testing.T) {
		sp, ok := smdSchema.FindNamedType("io.k8s.api.core.v1.SeccompProfile")
		if !ok || sp.Map == nil {
			t.Fatalf("expected to find SeccompProfile Map in SMD schema, got %#v", sp)
		}
		disc := "type"
		wantUnions := []schema.Union{
			{
				Discriminator: &disc,
				Fields: []schema.UnionField{
					{
						FieldName:          "localhostProfile",
						DiscriminatorValue: "LocalhostProfile",
					},
				},
			},
		}
		if diff := cmp.Diff(wantUnions, sp.Map.Unions); diff != "" {
			t.Errorf("unexpected SeccompProfile unions (-want +got):\n%s", diff)
		}
	})
}
