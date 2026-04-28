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

package proto_test

import (
	"path/filepath"
	"reflect"
	"testing"

	openapi_v3 "github.com/google/gnostic-models/openapiv3"

	"k8s.io/kube-openapi/pkg/util/proto"
	prototesting "k8s.io/kube-openapi/pkg/util/proto/testing"
)

var fakeSchema = prototesting.Fake{Path: filepath.Join("testdata", "swagger.json")}
var fakeSchemaNext = prototesting.Fake{Path: filepath.Join("testdata", "swagger_next.json")}
var fakeSchemaV300 = prototesting.FakeV3{Path: filepath.Join("testdata", "openapi_v3_0_0")}

// loadModels loads the swagger schema and returns proto.Models.
func loadModels(t *testing.T) proto.Models {
	t.Helper()
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}
	return models
}

// loadDeploymentModel loads the schema and returns the Deployment Kind.
func loadDeploymentModel(t *testing.T) (proto.Models, proto.Schema, *proto.Kind) {
	t.Helper()
	models := loadModels(t)
	schema := models.LookupModel("io.k8s.api.apps.v1.Deployment")
	if schema == nil {
		t.Fatal("model io.k8s.api.apps.v1.Deployment not found")
	}
	deployment, ok := schema.(*proto.Kind)
	if !ok || deployment == nil {
		t.Fatal("expected schema to be *proto.Kind")
	}
	return models, schema, deployment
}

func TestDeployment(t *testing.T) {
	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		models := loadModels(t)
		schema := models.LookupModel("io.k8s.api.apps.v1.Deployment")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1.Deployment not found")
		}
	})

	t.Run("should be a Kind", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if deployment == nil {
			t.Fatal("expected deployment to be non-nil")
		}
	})

	t.Run("should have a path", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		got := deployment.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("path = %v, want %v", got, want)
		}
	})

	t.Run("should have a kind key of type string", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if _, ok := deployment.Fields["kind"]; !ok {
			t.Fatal("missing 'kind' field")
		}
		key, ok := deployment.Fields["kind"].(*proto.Primitive)
		if !ok || key == nil {
			t.Fatal("expected 'kind' to be *proto.Primitive")
		}
		if key.Type != "string" {
			t.Errorf("kind.Type = %q, want %q", key.Type, "string")
		}
		got := key.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment", ".kind"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("kind path = %v, want %v", got, want)
		}
	})

	t.Run("should have a apiVersion key of type string", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if _, ok := deployment.Fields["apiVersion"]; !ok {
			t.Fatal("missing 'apiVersion' field")
		}
		key, ok := deployment.Fields["apiVersion"].(*proto.Primitive)
		if !ok || key == nil {
			t.Fatal("expected 'apiVersion' to be *proto.Primitive")
		}
		if key.Type != "string" {
			t.Errorf("apiVersion.Type = %q, want %q", key.Type, "string")
		}
		got := key.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment", ".apiVersion"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("apiVersion path = %v, want %v", got, want)
		}
	})

	t.Run("should have a metadata key of type Reference", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if _, ok := deployment.Fields["metadata"]; !ok {
			t.Fatal("missing 'metadata' field")
		}
		key, ok := deployment.Fields["metadata"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'metadata' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta" {
			t.Errorf("metadata reference = %q, want %q", key.Reference(), "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta")
		}
		subSchema, ok := key.SubSchema().(*proto.Kind)
		if !ok || subSchema == nil {
			t.Fatal("expected metadata sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a status key of type Reference", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if _, ok := deployment.Fields["status"]; !ok {
			t.Fatal("missing 'status' field")
		}
		key, ok := deployment.Fields["status"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'status' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.apps.v1.DeploymentStatus" {
			t.Errorf("status reference = %q, want %q", key.Reference(), "io.k8s.api.apps.v1.DeploymentStatus")
		}
		status, ok := key.SubSchema().(*proto.Kind)
		if !ok || status == nil {
			t.Fatal("expected status sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a valid DeploymentStatus", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		statusRef, ok := deployment.Fields["status"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'status' to be proto.Reference")
		}
		status, ok := statusRef.SubSchema().(*proto.Kind)
		if !ok || status == nil {
			t.Fatal("expected status sub-schema to be *proto.Kind")
		}

		t.Log("having availableReplicas key")
		if _, ok := status.Fields["availableReplicas"]; !ok {
			t.Fatal("missing 'availableReplicas' field in status")
		}
		replicas, ok := status.Fields["availableReplicas"].(*proto.Primitive)
		if !ok || replicas == nil {
			t.Fatal("expected 'availableReplicas' to be *proto.Primitive")
		}
		if replicas.Type != "integer" {
			t.Errorf("availableReplicas.Type = %q, want %q", replicas.Type, "integer")
		}

		t.Log("having conditions key")
		if _, ok := status.Fields["conditions"]; !ok {
			t.Fatal("missing 'conditions' field in status")
		}
		conditions, ok := status.Fields["conditions"].(*proto.Array)
		if !ok || conditions == nil {
			t.Fatal("expected 'conditions' to be *proto.Array")
		}
		wantName := `Array of Reference to "io.k8s.api.apps.v1.DeploymentCondition"`
		if conditions.GetName() != wantName {
			t.Errorf("conditions name = %q, want %q", conditions.GetName(), wantName)
		}
		wantExt := map[string]interface{}{
			"x-kubernetes-list-map-keys":   []interface{}{"type"},
			"x-kubernetes-list-type":       "map",
			"x-kubernetes-patch-merge-key": "type",
			"x-kubernetes-patch-strategy":  "merge",
		}
		if !reflect.DeepEqual(conditions.GetExtensions(), wantExt) {
			t.Errorf("conditions extensions = %v, want %v", conditions.GetExtensions(), wantExt)
		}
		condition, ok := conditions.SubType.(proto.Reference)
		if !ok {
			t.Fatal("expected conditions.SubType to be proto.Reference")
		}
		if condition.Reference() != "io.k8s.api.apps.v1.DeploymentCondition" {
			t.Errorf("condition reference = %q, want %q", condition.Reference(), "io.k8s.api.apps.v1.DeploymentCondition")
		}
	})

	t.Run("should have a spec key of type Reference", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		if _, ok := deployment.Fields["spec"]; !ok {
			t.Fatal("missing 'spec' field")
		}
		key, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.apps.v1.DeploymentSpec" {
			t.Errorf("spec reference = %q, want %q", key.Reference(), "io.k8s.api.apps.v1.DeploymentSpec")
		}
		spec, ok := key.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a spec with no gvk", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		specRef, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		spec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
		if _, found := spec.GetExtensions()["x-kubernetes-group-version-kind"]; found {
			t.Error("spec should not have x-kubernetes-group-version-kind extension")
		}
	})

	t.Run("should have a spec with a PodTemplateSpec sub-field", func(t *testing.T) {
		_, _, deployment := loadDeploymentModel(t)
		specRef, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		spec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
		if _, ok := spec.Fields["template"]; !ok {
			t.Fatal("missing 'template' field in spec")
		}
		key, ok := spec.Fields["template"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'template' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.core.v1.PodTemplateSpec" {
			t.Errorf("template reference = %q, want %q", key.Reference(), "io.k8s.api.core.v1.PodTemplateSpec")
		}
	})
}

func TestNextSchemaDeployment(t *testing.T) {
	s, err := fakeSchemaNext.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open next schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}

	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1.Deployment")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1.Deployment not found")
		}
	})

	t.Run("should be a Kind", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1.Deployment")
		if schema == nil {
			t.Fatal("model not found")
		}
		deployment, ok := schema.(*proto.Kind)
		if !ok || deployment == nil {
			t.Fatal("expected schema to be *proto.Kind")
		}
	})
}

func TestNextSchemaControllerRevision(t *testing.T) {
	s, err := fakeSchemaNext.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open next schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}

	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1.ControllerRevision")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1.ControllerRevision not found")
		}
	})

	t.Run("data property should be map[string]Arbitrary", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1.ControllerRevision")
		if schema == nil {
			t.Fatal("model not found")
		}
		cr, ok := schema.(*proto.Kind)
		if !ok || cr == nil {
			t.Fatal("expected schema to be *proto.Kind")
		}
		if _, ok := cr.Fields["data"]; !ok {
			t.Fatal("missing 'data' field")
		}

		data, ok := cr.Fields["data"].(*proto.Map)
		if !ok || data == nil {
			t.Fatal("expected 'data' to be *proto.Map")
		}
		wantName := "Map of Arbitrary value (primitive, object or array)"
		if data.GetName() != wantName {
			t.Errorf("data name = %q, want %q", data.GetName(), wantName)
		}
		wantPath := []string{"io.k8s.api.apps.v1.ControllerRevision", ".data"}
		if !reflect.DeepEqual(data.GetPath().Get(), wantPath) {
			t.Errorf("data path = %v, want %v", data.GetPath().Get(), wantPath)
		}

		arbitrary, ok := data.SubType.(*proto.Arbitrary)
		if !ok || arbitrary == nil {
			t.Fatal("expected data.SubType to be *proto.Arbitrary")
		}
		wantArbName := "Arbitrary value (primitive, object or array)"
		if arbitrary.GetName() != wantArbName {
			t.Errorf("arbitrary name = %q, want %q", arbitrary.GetName(), wantArbName)
		}
		if !reflect.DeepEqual(arbitrary.GetPath().Get(), wantPath) {
			t.Errorf("arbitrary path = %v, want %v", arbitrary.GetPath().Get(), wantPath)
		}
	})
}

func TestSubjectAccessReview(t *testing.T) {
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}

	t.Run("should lookup the Schema by its model", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.authorization.v1.LocalSubjectAccessReview")
		if schema == nil {
			t.Fatal("model io.k8s.api.authorization.v1.LocalSubjectAccessReview not found")
		}
	})

	t.Run("should be a Kind and have a spec", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.authorization.v1.LocalSubjectAccessReview")
		if schema == nil {
			t.Fatal("model not found")
		}
		sar, ok := schema.(*proto.Kind)
		if !ok || sar == nil {
			t.Fatal("expected schema to be *proto.Kind")
		}
		if _, ok := sar.Fields["spec"]; !ok {
			t.Fatal("missing 'spec' field")
		}
		specRef, ok := sar.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		if specRef.Reference() != "io.k8s.api.authorization.v1.SubjectAccessReviewSpec" {
			t.Errorf("spec reference = %q, want %q", specRef.Reference(), "io.k8s.api.authorization.v1.SubjectAccessReviewSpec")
		}
		sarspec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || sarspec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a valid SubjectAccessReviewSpec", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.authorization.v1.LocalSubjectAccessReview")
		if schema == nil {
			t.Fatal("model not found")
		}
		sar, ok := schema.(*proto.Kind)
		if !ok || sar == nil {
			t.Fatal("expected schema to be *proto.Kind")
		}
		specRef, ok := sar.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		sarspec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || sarspec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}

		if _, ok := sarspec.Fields["extra"]; !ok {
			t.Fatal("missing 'extra' field")
		}
		extra, ok := sarspec.Fields["extra"].(*proto.Map)
		if !ok || extra == nil {
			t.Fatal("expected 'extra' to be *proto.Map")
		}
		if extra.GetName() != "Map of Array of string" {
			t.Errorf("extra name = %q, want %q", extra.GetName(), "Map of Array of string")
		}
		wantExtraPath := []string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}
		if !reflect.DeepEqual(extra.GetPath().Get(), wantExtraPath) {
			t.Errorf("extra path = %v, want %v", extra.GetPath().Get(), wantExtraPath)
		}

		array, ok := extra.SubType.(*proto.Array)
		if !ok || array == nil {
			t.Fatal("expected extra.SubType to be *proto.Array")
		}
		if array.GetName() != "Array of string" {
			t.Errorf("array name = %q, want %q", array.GetName(), "Array of string")
		}
		if !reflect.DeepEqual(array.GetPath().Get(), wantExtraPath) {
			t.Errorf("array path = %v, want %v", array.GetPath().Get(), wantExtraPath)
		}

		str, ok := array.SubType.(*proto.Primitive)
		if !ok || str == nil {
			t.Fatal("expected array.SubType to be *proto.Primitive")
		}
		if str.Type != "string" {
			t.Errorf("str.Type = %q, want %q", str.Type, "string")
		}
		if str.GetName() != "string" {
			t.Errorf("str name = %q, want %q", str.GetName(), "string")
		}
		if !reflect.DeepEqual(str.GetPath().Get(), wantExtraPath) {
			t.Errorf("str path = %v, want %v", str.GetPath().Get(), wantExtraPath)
		}
	})
}

func TestPath(t *testing.T) {
	t.Run("can be created by NewPath", func(t *testing.T) {
		path := proto.NewPath("key")
		if path.String() != "key" {
			t.Errorf("path.String() = %q, want %q", path.String(), "key")
		}
	})

	t.Run("can create and print complex paths", func(t *testing.T) {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")
		if field.String() != "key[12].subKey" {
			t.Errorf("field.String() = %q, want %q", field.String(), "key[12].subKey")
		}
	})

	t.Run("has a length", func(t *testing.T) {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")
		if field.Len() != 3 {
			t.Errorf("field.Len() = %d, want %d", field.Len(), 3)
		}
	})

	t.Run("can look like an array", func(t *testing.T) {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")
		got := field.Get()
		want := []string{"key", "[12]", ".subKey"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("field.Get() = %v, want %v", got, want)
		}
	})
}

// loadV3Deployment loads the v3.0.0 schema and returns the Deployment Kind.
func loadV3Deployment(t *testing.T) *proto.Kind {
	t.Helper()
	s, err := fakeSchemaV300.OpenAPIV3Schema("apps/v1")
	if err != nil {
		t.Fatalf("failed to open v3.0.0 schema: %v", err)
	}
	models, err := proto.NewOpenAPIV3Data(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI v3 data: %v", err)
	}
	schema := models.LookupModel("io.k8s.api.apps.v1.Deployment")
	if schema == nil {
		t.Fatal("model io.k8s.api.apps.v1.Deployment not found")
	}
	deployment, ok := schema.(*proto.Kind)
	if !ok || deployment == nil {
		t.Fatal("expected schema to be *proto.Kind")
	}
	return deployment
}

func TestV3Deployment(t *testing.T) {
	t.Run("should have a path", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		got := deployment.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("path = %v, want %v", got, want)
		}
	})

	t.Run("should have a kind key of type string", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		if _, ok := deployment.Fields["kind"]; !ok {
			t.Fatal("missing 'kind' field")
		}
		key, ok := deployment.Fields["kind"].(*proto.Primitive)
		if !ok || key == nil {
			t.Fatal("expected 'kind' to be *proto.Primitive")
		}
		if key.Type != "string" {
			t.Errorf("kind.Type = %q, want %q", key.Type, "string")
		}
		got := key.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment", ".kind"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("kind path = %v, want %v", got, want)
		}
	})

	t.Run("should have a apiVersion key of type string", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		if _, ok := deployment.Fields["apiVersion"]; !ok {
			t.Fatal("missing 'apiVersion' field")
		}
		key, ok := deployment.Fields["apiVersion"].(*proto.Primitive)
		if !ok || key == nil {
			t.Fatal("expected 'apiVersion' to be *proto.Primitive")
		}
		if key.Type != "string" {
			t.Errorf("apiVersion.Type = %q, want %q", key.Type, "string")
		}
		got := key.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1.Deployment", ".apiVersion"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("apiVersion path = %v, want %v", got, want)
		}
	})

	t.Run("should have a metadata key of type Reference", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		if _, ok := deployment.Fields["metadata"]; !ok {
			t.Fatal("missing 'metadata' field")
		}
		key, ok := deployment.Fields["metadata"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'metadata' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta" {
			t.Errorf("metadata reference = %q, want %q", key.Reference(), "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta")
		}
		subSchema, ok := key.SubSchema().(*proto.Kind)
		if !ok || subSchema == nil {
			t.Fatal("expected metadata sub-schema to be *proto.Kind")
		}
	})

	t.Run("status/should have a valid DeploymentStatus", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		if _, ok := deployment.Fields["status"]; !ok {
			t.Fatal("missing 'status' field")
		}
		statusRef, ok := deployment.Fields["status"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'status' to be proto.Reference")
		}
		if statusRef.Reference() != "io.k8s.api.apps.v1.DeploymentStatus" {
			t.Errorf("status reference = %q, want %q", statusRef.Reference(), "io.k8s.api.apps.v1.DeploymentStatus")
		}
		status, ok := statusRef.SubSchema().(*proto.Kind)
		if !ok || status == nil {
			t.Fatal("expected status sub-schema to be *proto.Kind")
		}

		t.Log("having availableReplicas key")
		if _, ok := status.Fields["availableReplicas"]; !ok {
			t.Fatal("missing 'availableReplicas' field in status")
		}
		replicas, ok := status.Fields["availableReplicas"].(*proto.Primitive)
		if !ok || replicas == nil {
			t.Fatal("expected 'availableReplicas' to be *proto.Primitive")
		}
		if replicas.Type != "integer" {
			t.Errorf("availableReplicas.Type = %q, want %q", replicas.Type, "integer")
		}

		t.Log("having conditions key")
		if _, ok := status.Fields["conditions"]; !ok {
			t.Fatal("missing 'conditions' field in status")
		}
		conditions, ok := status.Fields["conditions"].(*proto.Array)
		if !ok || conditions == nil {
			t.Fatal("expected 'conditions' to be *proto.Array")
		}
		wantName := `Array of Reference to "io.k8s.api.apps.v1.DeploymentCondition"`
		if conditions.GetName() != wantName {
			t.Errorf("conditions name = %q, want %q", conditions.GetName(), wantName)
		}
		wantExt := map[string]interface{}{
			"x-kubernetes-patch-merge-key": "type",
			"x-kubernetes-patch-strategy":  "merge",
		}
		if !reflect.DeepEqual(conditions.GetExtensions(), wantExt) {
			t.Errorf("conditions extensions = %v, want %v", conditions.GetExtensions(), wantExt)
		}
		condition, ok := conditions.SubType.(proto.Reference)
		if !ok {
			t.Fatal("expected conditions.SubType to be proto.Reference")
		}
		if condition.Reference() != "io.k8s.api.apps.v1.DeploymentCondition" {
			t.Errorf("condition reference = %q, want %q", condition.Reference(), "io.k8s.api.apps.v1.DeploymentCondition")
		}
	})

	t.Run("spec/should have no gvk", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		if _, ok := deployment.Fields["spec"]; !ok {
			t.Fatal("missing 'spec' field")
		}
		specRef, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		if specRef.Reference() != "io.k8s.api.apps.v1.DeploymentSpec" {
			t.Errorf("spec reference = %q, want %q", specRef.Reference(), "io.k8s.api.apps.v1.DeploymentSpec")
		}
		spec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
		if _, found := spec.GetExtensions()["x-kubernetes-group-version-kind"]; found {
			t.Error("spec should not have x-kubernetes-group-version-kind extension")
		}
	})

	t.Run("spec/should have a PodTemplateSpec sub-field", func(t *testing.T) {
		deployment := loadV3Deployment(t)
		specRef, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		spec, ok := specRef.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
		if _, ok := spec.Fields["template"]; !ok {
			t.Fatal("missing 'template' field in spec")
		}
		key, ok := spec.Fields["template"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'template' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.core.v1.PodTemplateSpec" {
			t.Errorf("template reference = %q, want %q", key.Reference(), "io.k8s.api.core.v1.PodTemplateSpec")
		}
	})
}

func TestV3GVKExtension(t *testing.T) {
	spec := []byte(`{
	"openapi": "3.0.0",
	"info": {
		"title": "Kubernetes",
		"version": "v1.24.0"
	},
	"paths": {
		"/foo": {
			"get": {
				"responses": {
					"200": {
						"description": "OK",
						"content": {
							"application/json": {
								"schema": {
									"$ref": "#/components/schemas/Foo"
								}
							}
						}
					}
				}
			}
		}
	},
	"components": {
		"schemas": {
			"Foo": {
				"type": "object",
				"properties": {},
				"x-kubernetes-group-version-kind": [
					{
						"group": "foo",
						"kind": "Foo",
						"version": "v1"
					}
				]
			}
		}
	}
}`)

	document, err := openapi_v3.ParseDocument(spec)
	if err != nil {
		t.Fatalf("failed to parse OpenAPI v3 document: %v", err)
	}
	models, err := proto.NewOpenAPIV3Data(document)
	if err != nil {
		t.Fatalf("failed to create OpenAPI v3 data: %v", err)
	}
	schema := models.LookupModel("Foo")
	if schema == nil {
		t.Fatal("model Foo not found")
	}

	t.Run("should have an extension with gvk", func(t *testing.T) {
		if _, found := schema.GetExtensions()["x-kubernetes-group-version-kind"]; !found {
			t.Error("expected x-kubernetes-group-version-kind extension to be present")
		}
	})

	t.Run("should convert to proto.Kind type", func(t *testing.T) {
		foo, ok := schema.(*proto.Kind)
		if !ok {
			t.Fatal("expected schema to be *proto.Kind")
		}
		if foo == nil {
			t.Fatal("expected foo to be non-nil")
		}
	})
}
