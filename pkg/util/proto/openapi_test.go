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

// loadV18Models loads the v1.8 swagger schema and returns proto.Models.
func loadV18Models(t *testing.T) proto.Models {
	t.Helper()
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open v1.8 schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}
	return models
}

// loadV18DeploymentModel loads the v1.8 schema and returns the Deployment Kind.
func loadV18DeploymentModel(t *testing.T) (proto.Models, proto.Schema, *proto.Kind) {
	t.Helper()
	models := loadV18Models(t)
	schema := models.LookupModel("io.k8s.api.apps.v1beta1.Deployment")
	if schema == nil {
		t.Fatal("model io.k8s.api.apps.v1beta1.Deployment not found")
	}
	deployment, ok := schema.(*proto.Kind)
	if !ok || deployment == nil {
		t.Fatal("expected schema to be *proto.Kind")
	}
	return models, schema, deployment
}

func TestV18Deployment(t *testing.T) {
	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		models := loadV18Models(t)
		schema := models.LookupModel("io.k8s.api.apps.v1beta1.Deployment")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1beta1.Deployment not found")
		}
	})

	t.Run("should be a Kind", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
		if deployment == nil {
			t.Fatal("expected deployment to be non-nil")
		}
	})

	t.Run("should have a path", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
		got := deployment.GetPath().Get()
		want := []string{"io.k8s.api.apps.v1beta1.Deployment"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("path = %v, want %v", got, want)
		}
	})

	t.Run("should have a kind key of type string", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
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
		want := []string{"io.k8s.api.apps.v1beta1.Deployment", ".kind"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("kind path = %v, want %v", got, want)
		}
	})

	t.Run("should have a apiVersion key of type string", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
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
		want := []string{"io.k8s.api.apps.v1beta1.Deployment", ".apiVersion"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("apiVersion path = %v, want %v", got, want)
		}
	})

	t.Run("should have a metadata key of type Reference", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
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
		_, _, deployment := loadV18DeploymentModel(t)
		if _, ok := deployment.Fields["status"]; !ok {
			t.Fatal("missing 'status' field")
		}
		key, ok := deployment.Fields["status"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'status' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.apps.v1beta1.DeploymentStatus" {
			t.Errorf("status reference = %q, want %q", key.Reference(), "io.k8s.api.apps.v1beta1.DeploymentStatus")
		}
		status, ok := key.SubSchema().(*proto.Kind)
		if !ok || status == nil {
			t.Fatal("expected status sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a valid DeploymentStatus", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
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
		wantName := `Array of Reference to "io.k8s.api.apps.v1beta1.DeploymentCondition"`
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
		if condition.Reference() != "io.k8s.api.apps.v1beta1.DeploymentCondition" {
			t.Errorf("condition reference = %q, want %q", condition.Reference(), "io.k8s.api.apps.v1beta1.DeploymentCondition")
		}
	})

	t.Run("should have a spec key of type Reference", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
		if _, ok := deployment.Fields["spec"]; !ok {
			t.Fatal("missing 'spec' field")
		}
		key, ok := deployment.Fields["spec"].(proto.Reference)
		if !ok {
			t.Fatal("expected 'spec' to be proto.Reference")
		}
		if key.Reference() != "io.k8s.api.apps.v1beta1.DeploymentSpec" {
			t.Errorf("spec reference = %q, want %q", key.Reference(), "io.k8s.api.apps.v1beta1.DeploymentSpec")
		}
		spec, ok := key.SubSchema().(*proto.Kind)
		if !ok || spec == nil {
			t.Fatal("expected spec sub-schema to be *proto.Kind")
		}
	})

	t.Run("should have a spec with no gvk", func(t *testing.T) {
		_, _, deployment := loadV18DeploymentModel(t)
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
		_, _, deployment := loadV18DeploymentModel(t)
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

func TestV111Deployment(t *testing.T) {
	s, err := fakeSchemaNext.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open v1.11 schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}

	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1beta1.Deployment")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1beta1.Deployment not found")
		}
	})

	t.Run("should be a Kind", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1beta1.Deployment")
		if schema == nil {
			t.Fatal("model not found")
		}
		deployment, ok := schema.(*proto.Kind)
		if !ok || deployment == nil {
			t.Fatal("expected schema to be *proto.Kind")
		}
	})
}

func TestV111ControllerRevision(t *testing.T) {
	s, err := fakeSchemaNext.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to open v1.11 schema: %v", err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatalf("failed to create OpenAPI data: %v", err)
	}

	t.Run("should lookup the Schema by its model name", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1beta1.ControllerRevision")
		if schema == nil {
			t.Fatal("model io.k8s.api.apps.v1beta1.ControllerRevision not found")
		}
	})

	t.Run("data property should be map[string]Arbitrary", func(t *testing.T) {
		schema := models.LookupModel("io.k8s.api.apps.v1beta1.ControllerRevision")
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
		wantPath := []string{"io.k8s.api.apps.v1beta1.ControllerRevision", ".data"}
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

// TestV3AllOfRef verifies how ParseSchemaV3 handles OpenAPI 3.0 schemas that use
// allOf. In OpenAPI 3.0, sibling fields alongside a $ref (such as "description",
// "default", or "x-kubernetes-*" extensions) are ignored per the specification.
// To attach field-level metadata to a referenced type, Kubernetes >= 1.35 (via
// kube-openapi's pkg/builder3/util.WrapRefs) wraps the $ref in a single-element
// allOf array on an otherwise typeless schema.
func TestV3AllOfRef(t *testing.T) {
	spec := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "v1"
		},
		"paths": {},
		"components": {
			"schemas": {
				"Parent": {
					"type": "object",
					"properties": {
						"directRef": {
							"$ref": "#/components/schemas/Child"
						},
						"allOfRef": {
							"allOf": [
								{
									"$ref": "#/components/schemas/Child"
								}
							],
							"description": "field using allOf-wrapped ref",
							"default": "some-default",
							"x-kubernetes-patch-strategy": "merge"
						},
						"allOfExternalRef": {
							"allOf": [
								{
									"$ref": "#/components/parameters/ExternalParam"
								}
							],
							"description": "field using allOf-wrapped non-component ref"
						},
						"allOfInlineSchema": {
							"allOf": [
								{
									"type": "string"
								}
							],
							"description": "field using allOf with inline schema instead of ref"
						},
						"allOfMultipleRefs": {
							"allOf": [
								{
									"$ref": "#/components/schemas/Child"
								},
								{
									"$ref": "#/components/schemas/Child"
								}
							],
							"description": "field using multi-element allOf composition"
						},
						"allOfUnknownSchemaRef": {
							"allOf": [
								{
									"$ref": "#/components/schemas/NonExistentModel"
								}
							],
							"description": "field using allOf-wrapped unknown component ref"
						}
					},
					"x-kubernetes-group-version-kind": [
						{
							"group": "test",
							"kind": "Parent",
							"version": "v1"
						}
					]
				},
				"Child": {
					"type": "object",
					"properties": {
						"name": {
							"type": "string"
						}
					},
					"x-kubernetes-group-version-kind": [
						{
							"group": "test",
							"kind": "Child",
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

	schema := models.LookupModel("Parent")
	if schema == nil {
		t.Fatal("model Parent not found")
	}
	parent, ok := schema.(*proto.Kind)
	if !ok || parent == nil {
		t.Fatal("expected Parent to be *proto.Kind")
	}

	// Baseline: a direct $ref with no sibling fields resolves to proto.Reference.
	t.Run("direct $ref without sibling fields resolves to proto.Reference", func(t *testing.T) {
		field, ok := parent.Fields["directRef"]
		if !ok {
			t.Fatal("missing 'directRef' field")
		}
		ref, ok := field.(proto.Reference)
		if !ok {
			t.Fatalf("expected 'directRef' to be proto.Reference, got %T", field)
		}
		if ref.Reference() != "Child" {
			t.Errorf("directRef reference = %q, want %q", ref.Reference(), "Child")
		}
	})

	// A single-element allOf containing a #/components/schemas/ $ref must unwrap
	// to a proto.Reference pointing at the target model and linking its SubSchema.
	t.Run("single-element allOf wrapping #/components/schemas/ $ref resolves to proto.Reference", func(t *testing.T) {
		field, ok := parent.Fields["allOfRef"]
		if !ok {
			t.Fatal("missing 'allOfRef' field")
		}
		ref, ok := field.(proto.Reference)
		if !ok {
			t.Fatalf("expected 'allOfRef' to be proto.Reference, got %T", field)
		}
		if ref.Reference() != "Child" {
			t.Errorf("allOfRef reference = %q, want %q", ref.Reference(), "Child")
		}
		if ref.SubSchema() == nil {
			t.Error("expected allOfRef.SubSchema() to resolve to non-nil Child schema")
		}
	})

	// Metadata placed on the outer schema alongside the single-element allOf
	// (description, default, and x-kubernetes-* extensions) must be preserved
	// on the resulting proto.Reference BaseSchema.
	t.Run("single-element allOf wrapping $ref preserves outer description, default, and extensions", func(t *testing.T) {
		field := parent.Fields["allOfRef"]
		if got := field.GetDescription(); got != "field using allOf-wrapped ref" {
			t.Errorf("description = %q, want %q", got, "field using allOf-wrapped ref")
		}
		if got, want := field.GetDefault(), "some-default"; got != want {
			t.Errorf("default = %v, want %v", got, want)
		}
		if got, want := field.GetExtensions()["x-kubernetes-patch-strategy"], "merge"; got != want {
			t.Errorf("x-kubernetes-patch-strategy extension = %v, want %v", got, want)
		}
	})

	// Mirrors ParseV3SchemaReference: references outside #/components/schemas
	// (such as #/components/parameters/...) cannot be resolved against d.models
	// and must fall back to *proto.Arbitrary while preserving outer BaseSchema metadata.
	t.Run("single-element allOf wrapping non-component $ref falls back to proto.Arbitrary", func(t *testing.T) {
		field, ok := parent.Fields["allOfExternalRef"]
		if !ok {
			t.Fatal("missing 'allOfExternalRef' field")
		}
		arb, ok := field.(*proto.Arbitrary)
		if !ok {
			t.Fatalf("expected 'allOfExternalRef' to fall back to *proto.Arbitrary, got %T", field)
		}
		if got, want := arb.GetDescription(), "field using allOf-wrapped non-component ref"; got != want {
			t.Errorf("description = %q, want %q", got, want)
		}
	})

	// When a single-element allOf contains an inline Schema rather than a $ref,
	// it is not a WrapRefs reference wrapper and must fall through to *proto.Arbitrary.
	t.Run("single-element allOf wrapping inline schema instead of $ref falls back to proto.Arbitrary", func(t *testing.T) {
		field, ok := parent.Fields["allOfInlineSchema"]
		if !ok {
			t.Fatal("missing 'allOfInlineSchema' field")
		}
		if _, ok := field.(*proto.Arbitrary); !ok {
			t.Fatalf("expected 'allOfInlineSchema' to fall back to *proto.Arbitrary, got %T", field)
		}
	})

	// When allOf contains more than one element, it represents schema composition
	// rather than a single wrapped $ref and must not be unwrapped into proto.Reference.
	t.Run("multi-element allOf is not unwrapped as a single reference and falls back to proto.Arbitrary", func(t *testing.T) {
		field, ok := parent.Fields["allOfMultipleRefs"]
		if !ok {
			t.Fatal("missing 'allOfMultipleRefs' field")
		}
		if _, ok := field.(*proto.Arbitrary); !ok {
			t.Fatalf("expected 'allOfMultipleRefs' to fall back to *proto.Arbitrary, got %T", field)
		}
	})

	// When a single-element allOf references a #/components/schemas/ model that
	// does not exist in the document, parseAllOfRefSchema returns false so
	// ParseSchemaV3 falls back to *proto.Arbitrary rather than introducing a new
	// failure return.
	t.Run("single-element allOf wrapping unknown #/components/schemas/ $ref falls back to proto.Arbitrary", func(t *testing.T) {
		field, ok := parent.Fields["allOfUnknownSchemaRef"]
		if !ok {
			t.Fatal("missing 'allOfUnknownSchemaRef' field")
		}
		arb, ok := field.(*proto.Arbitrary)
		if !ok {
			t.Fatalf("expected 'allOfUnknownSchemaRef' to fall back to *proto.Arbitrary, got %T", field)
		}
		if got, want := arb.GetDescription(), "field using allOf-wrapped unknown component ref"; got != want {
			t.Errorf("description = %q, want %q", got, want)
		}
	})
}

// TestSpecificationExtensionToMap verifies that OpenAPI v3 specification
// extensions (stored as raw YAML strings in gnostic-models v0.7.0+ NamedAny
// protos) are unmarshaled into native Go types with string-keyed maps, while
// malformed or empty extension entries are safely ignored.
func TestSpecificationExtensionToMap(t *testing.T) {
	// gnostic-models v0.7.0+ populates NamedAny.Value.Yaml without populating
	// the legacy Value.Value protobuf Any field. SpecificationExtensionToMap
	// unmarshals Value.Yaml via go.yaml.in/yaml/v3 so that YAML objects decode
	// into map[string]interface{} (matching JSON decoding).
	t.Run("decodes scalar, slice, and nested map YAML extension values into native Go types", func(t *testing.T) {
		extensions := []*openapi_v3.NamedAny{
			{
				Name:  "x-nil",
				Value: &openapi_v3.Any{Yaml: "null"},
			},
			{
				Name:  "x-bool",
				Value: &openapi_v3.Any{Yaml: "true"},
			},
			{
				Name:  "x-int",
				Value: &openapi_v3.Any{Yaml: "42"},
			},
			{
				Name:  "x-float",
				Value: &openapi_v3.Any{Yaml: "3.14"},
			},
			{
				Name:  "x-string",
				Value: &openapi_v3.Any{Yaml: `"hello"`},
			},
			{
				Name:  "x-slice",
				Value: &openapi_v3.Any{Yaml: `["a", 1, true, {"k": "v"}]`},
			},
			{
				Name:  "x-map",
				Value: &openapi_v3.Any{Yaml: `{"group": "apps", "version": "v1", "kind": "Deployment", "nested": {"enabled": true}}`},
			},
		}

		got := proto.SpecificationExtensionToMap(extensions)

		want := map[string]interface{}{
			"x-nil":    nil,
			"x-bool":   true,
			"x-int":    42,
			"x-float":  3.14,
			"x-string": "hello",
			"x-slice": []interface{}{
				"a",
				1,
				true,
				map[string]interface{}{"k": "v"},
			},
			"x-map": map[string]interface{}{
				"group":   "apps",
				"version": "v1",
				"kind":    "Deployment",
				"nested":  map[string]interface{}{"enabled": true},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("SpecificationExtensionToMap mismatch:\ngot:  %#v\nwant: %#v", got, want)
		}
	})

	// Defensive guards in SpecificationExtensionToMap must skip entries with
	// missing names, nil Any values, empty YAML strings, or invalid YAML syntax
	// without failing conversion of valid sibling extensions.
	t.Run("skips extension entries with empty name, nil value, empty YAML, or malformed YAML", func(t *testing.T) {
		extensions := []*openapi_v3.NamedAny{
			{
				// Empty extension name must be skipped.
				Name:  "",
				Value: &openapi_v3.Any{Yaml: `"ignored"`},
			},
			{
				// Nil Value pointer must be skipped.
				Name:  "x-nil-value-struct",
				Value: nil,
			},
			{
				// Empty raw YAML string must be skipped.
				Name:  "x-empty-yaml",
				Value: &openapi_v3.Any{Yaml: ""},
			},
			{
				// Malformed YAML that fails yaml.Unmarshal must be skipped.
				Name:  "x-malformed-yaml",
				Value: &openapi_v3.Any{Yaml: "[unclosed flow sequence"},
			},
			{
				// Valid sibling extension must still be preserved.
				Name:  "x-valid-extension",
				Value: &openapi_v3.Any{Yaml: `"preserved"`},
			},
		}

		got := proto.SpecificationExtensionToMap(extensions)
		want := map[string]interface{}{
			"x-valid-extension": "preserved",
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("SpecificationExtensionToMap guard handling mismatch:\ngot:  %#v\nwant: %#v", got, want)
		}
	})
}
