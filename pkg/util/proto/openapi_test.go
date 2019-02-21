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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kube-openapi/pkg/util/proto/testing"
)

var fakeSchema111 = testing.Fake{Path: filepath.Join("testdata", "swagger-1.11.json")}

var _ = Describe("Reading apps/v1beta1/Deployment from v1.11 openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchema111.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.apps.v1beta1.Deployment"
	var schema proto.Schema
	It("should lookup the Schema by its model name", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var deployment *proto.Kind
	It("should be a Kind", func() {
		deployment = schema.(*proto.Kind)
		Expect(deployment).ToNot(BeNil())
	})
})

var _ = Describe("Reading apps/v1beta1/ControllerRevision from v1.11 openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchema111.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.apps.v1beta1.ControllerRevision"
	var schema proto.Schema
	It("should lookup the Schema by its model name", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var cr *proto.Kind
	It("data property should be map[string]Arbitrary", func() {
		cr = schema.(*proto.Kind)
		Expect(cr).ToNot(BeNil())
		Expect(cr.Fields).To(HaveKey("data"))

		data := cr.Fields["data"].(*proto.Map)
		Expect(data).ToNot(BeNil())
		Expect(data.GetName()).To(Equal("Map of Arbitrary value (primitive, object or array)"))
		Expect(data.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.ControllerRevision", ".data"}))

		arbitrary := data.SubType.(*proto.Arbitrary)
		Expect(arbitrary).ToNot(BeNil())
		Expect(arbitrary.GetName()).To(Equal("Arbitrary value (primitive, object or array)"))
		Expect(arbitrary.GetPath().Get()).To(Equal([]string{"io.k8s.api.apps.v1beta1.ControllerRevision", ".data"}))
	})
})

var _ = Describe("Reading authorization.k8s.io/v1/SubjectAccessReview from openAPIData", func() {
	var models proto.Models
	BeforeEach(func() {
		s, err := fakeSchema111.OpenAPISchema()
		Expect(err).To(BeNil())
		models, err = proto.NewOpenAPIData(s)
		Expect(err).To(BeNil())
	})

	model := "io.k8s.api.authorization.v1.LocalSubjectAccessReview"
	var schema proto.Schema
	It("should lookup the Schema by its model", func() {
		schema = models.LookupModel(model)
		Expect(schema).ToNot(BeNil())
	})

	var sarspec *proto.Kind
	It("should be a Kind and have a spec", func() {
		sar := schema.(*proto.Kind)
		Expect(sar).ToNot(BeNil())
		Expect(sar.Fields).To(HaveKey("spec"))
		specRef := sar.Fields["spec"].(proto.Reference)
		Expect(specRef).ToNot(BeNil())
		Expect(specRef.Reference()).To(Equal("io.k8s.api.authorization.v1.SubjectAccessReviewSpec"))
		sarspec = specRef.SubSchema().(*proto.Kind)
		Expect(sarspec).ToNot(BeNil())
	})

	It("should have a valid SubjectAccessReviewSpec", func() {
		Expect(sarspec.Fields).To(HaveKey("extra"))
		extra := sarspec.Fields["extra"].(*proto.Map)
		Expect(extra).ToNot(BeNil())
		Expect(extra.GetName()).To(Equal("Map of Array of string"))
		Expect(extra.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		array := extra.SubType.(*proto.Array)
		Expect(array).ToNot(BeNil())
		Expect(array.GetName()).To(Equal("Array of string"))
		Expect(array.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
		str := array.SubType.(*proto.Primitive)
		Expect(str).ToNot(BeNil())
		Expect(str.Type).To(Equal("string"))
		Expect(str.GetName()).To(Equal("string"))
		Expect(str.GetPath().Get()).To(Equal([]string{"io.k8s.api.authorization.v1.SubjectAccessReviewSpec", ".extra"}))
	})
})

var _ = Describe("Path", func() {
	It("can be created by NewPath", func() {
		path := proto.NewPath("key")
		Expect(path.String()).To(Equal("key"))
	})
	It("can create and print complex paths", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.String()).To(Equal("key[12].subKey"))
	})
	It("has a length", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Len()).To(Equal(3))
	})
	It("can look like an array", func() {
		key := proto.NewPath("key")
		array := key.ArrayPath(12)
		field := array.FieldPath("subKey")

		Expect(field.Get()).To(Equal([]string{"key", "[12]", ".subKey"}))
	})
})
