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

package parsing_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/util/parsing"
	tst "k8s.io/kube-openapi/pkg/util/parsing/testing"
)

var fakeSchema = tst.Fake{Path: filepath.Join("testing", "swagger.json")}

var _ = Describe("Reading apps/v1beta1/Deployment from openAPIData", func() {
	var resources parsing.Resources
	BeforeEach(func() {
		s, err := fakeSchema.OpenAPISchema()
		Expect(err).To(BeNil())
		resources, err = parsing.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	gvk := schema.GroupVersionKind{
		Kind:    "Deployment",
		Version: "v1beta1",
		Group:   "apps",
	}

	var schema parsing.Schema
	It("should lookup the Schema by its GroupVersionKind", func() {
		schema = resources.LookupResource(gvk)
		Expect(schema).ToNot(BeNil())
	})

	var deployment *parsing.Kind
	It("should be a Kind", func() {
		deployment = schema.(*parsing.Kind)
		Expect(deployment).ToNot(BeNil())
	})

	It("should have a path", func() {
		Expect(deployment.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment"}))
	})

	It("should have a kind key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("kind"))
		key := deployment.Fields["kind"].(*parsing.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment", ".kind"}))
	})

	It("should have a apiVersion key of type string", func() {
		Expect(deployment.Fields).To(HaveKey("apiVersion"))
		key := deployment.Fields["apiVersion"].(*parsing.Primitive)
		Expect(key).ToNot(BeNil())
		Expect(key.Type).To(Equal("string"))
		Expect(key.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.Deployment", ".apiVersion"}))
	})

	It("should have a metadata key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("metadata"))
		key := deployment.Fields["metadata"].(parsing.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"))
		subSchema := key.SubSchema().(*parsing.Kind)
		Expect(subSchema).ToNot(BeNil())
	})

	var status *parsing.Kind
	It("should have a status key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("status"))
		key := deployment.Fields["status"].(parsing.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentStatus"))
		status = key.SubSchema().(*parsing.Kind)
		Expect(status).ToNot(BeNil())
	})

	It("should have a valid DeploymentStatus", func() {
		By("having availableReplicas key")
		Expect(status.Fields).To(HaveKey("availableReplicas"))
		replicas := status.Fields["availableReplicas"].(*parsing.Primitive)
		Expect(replicas).ToNot(BeNil())
		Expect(replicas.Type).To(Equal("integer"))

		By("having conditions key")
		Expect(status.Fields).To(HaveKey("conditions"))
		conditions := status.Fields["conditions"].(*parsing.Array)
		Expect(conditions).ToNot(BeNil())
		Expect(conditions.GetName()).To(Equal(`Array of Reference to "io.k8s.api.apps.v1beta1.DeploymentCondition"`))
		Expect(conditions.GetExtensions()).To(Equal(map[string]interface{}{
			"x-kubernetes-patch-merge-key": "type",
			"x-kubernetes-patch-strategy":  "merge",
		}))
		condition := conditions.SubType.(parsing.Reference)
		Expect(condition.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentCondition"))
	})

	var spec *parsing.Kind
	It("should have a spec key of type Reference", func() {
		Expect(deployment.Fields).To(HaveKey("spec"))
		key := deployment.Fields["spec"].(parsing.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.apps.v1beta1.DeploymentSpec"))
		spec = key.SubSchema().(*parsing.Kind)
		Expect(spec).ToNot(BeNil())
	})

	It("should have a spec with no gvk", func() {
		_, found := spec.GetExtensions()["x-kubernetes-group-version-kind"]
		Expect(found).To(BeFalse())
	})

	It("should have a spec with a PodTemplateSpec sub-field", func() {
		Expect(spec.Fields).To(HaveKey("template"))
		key := spec.Fields["template"].(parsing.Reference)
		Expect(key).ToNot(BeNil())
		Expect(key.Reference()).To(Equal("io.k8s.api.core.v1.PodTemplateSpec"))
	})
})

var _ = Describe("Reading authorization.k8s.io/v1/SubjectAccessReview from openAPIData", func() {
	var resources parsing.Resources
	BeforeEach(func() {
		s, err := fakeSchema.OpenAPISchema()
		Expect(err).To(BeNil())
		resources, err = parsing.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	gvk := schema.GroupVersionKind{
		Kind:    "SubjectAccessReview",
		Version: "v1",
		Group:   "authorization.k8s.io",
	}

	var schema parsing.Schema
	It("should lookup the Schema by its GroupVersionKind", func() {
		schema = resources.LookupResource(gvk)
		Expect(schema).ToNot(BeNil())
	})

	var sarspec *parsing.Kind
	It("should be a Kind and have a spec", func() {
		sar := schema.(*parsing.Kind)
		Expect(sar).ToNot(BeNil())
		Expect(sar.Fields).To(HaveKey("spec"))
		specRef := sar.Fields["spec"].(parsing.Reference)
		Expect(specRef).ToNot(BeNil())
		Expect(specRef.Reference()).To(Equal("io.k8s.api.authorization.v1.SubjectAccessReviewSpec"))
		sarspec = specRef.SubSchema().(*parsing.Kind)
		Expect(sarspec).ToNot(BeNil())
	})

	It("should have a valid SubjectAccessReviewSpec", func() {
		Expect(sarspec.Fields).To(HaveKey("extra"))
		extra := sarspec.Fields["extra"].(*parsing.Map)
		Expect(extra).ToNot(BeNil())
		Expect(extra.GetName()).To(Equal("Map of Array of string"))
		Expect(extra.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		array := extra.SubType.(*parsing.Array)
		Expect(array).ToNot(BeNil())
		Expect(array.GetName()).To(Equal("Array of string"))
		Expect(array.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		str := array.SubType.(*parsing.Primitive)
		Expect(str).ToNot(BeNil())
		Expect(str.Type).To(Equal("string"))
		Expect(str.GetName()).To(Equal("string"))
		Expect(str.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
	})
})

var _ = Describe("Path", func() {
	It("can be created by NewPath", func() {
		path := parsing.NewPath("key")
		Expect(path.String()).To(Equal("key"))
	})
	It("can create and print complex paths", func() {
		key := parsing.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.String()).To(Equal("key[12].subKey"))
	})
	It("has a length", func() {
		key := parsing.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Len()).To(Equal(3))
	})
	It("can look like an array", func() {
		key := parsing.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Get()).To(Equal([]string{"key", "[12]", ".subKey"}))
	})
})
