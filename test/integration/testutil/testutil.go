/*
Copyright 2021 The Kubernetes Authors.

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

package testutil

import (
	"fmt"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/util"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// CreateOpenAPIBuilderConfig hard-codes some values in the API builder
// config for testing.
func CreateOpenAPIBuilderConfig() *common.Config {
	return &common.Config{
		ProtocolList:   []string{"https"},
		IgnorePrefixes: []string{"/swaggerapi"},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "Integration Test",
				Version: "1.0",
			},
		},
		ResponseDefinitions: map[string]spec.Response{
			"NotFound": {
				ResponseProps: spec.ResponseProps{
					Description: "Entity not found.",
				},
			},
		},
		CommonResponses: map[int]spec.Response{
			404: *spec.ResponseRef("#/responses/NotFound"),
		},
	}
}

func CreateOpenAPIV3BuilderConfig() *common.OpenAPIV3Config {
	return &common.OpenAPIV3Config{
		IgnorePrefixes: []string{"/swaggerapi"},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "Integration Test",
				Version: "1.0",
			},
		},
		ResponseDefinitions: map[string]*spec3.Response{
			"NotFound": {
				ResponseProps: spec3.ResponseProps{
					Description: "Entity not found.",
				},
			},
		},
		CommonResponses: map[int]*spec3.Response{
			404: {
				Refable: spec.Refable{Ref: spec.MustCreateRef("#/components/responses/NotFound")},
			},
		},
	}

}

// CreateWebServices hard-codes a simple WebService which only defines a GET and POST paths
// for testing.
func CreateWebServices(includeV2SchemaAnnotation bool) []*restful.WebService {
	w := new(restful.WebService)
	addRoutes(w, buildRouteForType(w, "dummytype", "Foo")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Bar")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Baz")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Waldo")...)
	addRoutes(w, buildRouteForType(w, "listtype", "AtomicList")...)
	addRoutes(w, buildRouteForType(w, "listtype", "MapList")...)
	addRoutes(w, buildRouteForType(w, "listtype", "SetList")...)
	addRoutes(w, buildRouteForType(w, "uniontype", "TopLevelUnion")...)
	addRoutes(w, buildRouteForType(w, "uniontype", "InlinedUnion")...)
	addRoutes(w, buildRouteForType(w, "custom", "Bal")...)
	addRoutes(w, buildRouteForType(w, "custom", "Bak")...)
	if includeV2SchemaAnnotation {
		addRoutes(w, buildRouteForType(w, "custom", "Bac")...)
		addRoutes(w, buildRouteForType(w, "custom", "Bah")...)
		addRoutes(w, buildRouteForType(w, "valuevalidation", "Foo2")...)
		addRoutes(w, buildRouteForType(w, "valuevalidation", "Foo3")...)
		addRoutes(w, buildRouteForType(w, "valuevalidation", "Foo4")...)
		addRoutes(w, buildRouteForType(w, "valuevalidation", "Foo5")...)
	}
	addRoutes(w, buildRouteForType(w, "maptype", "GranularMap")...)
	addRoutes(w, buildRouteForType(w, "maptype", "AtomicMap")...)
	addRoutes(w, buildRouteForType(w, "structtype", "GranularStruct")...)
	addRoutes(w, buildRouteForType(w, "structtype", "FieldLevelOverrideStruct")...)
	addRoutes(w, buildRouteForType(w, "structtype", "AtomicStruct")...)
	addRoutes(w, buildRouteForType(w, "structtype", "DeclaredAtomicStruct")...)
	addRoutes(w, buildRouteForType(w, "defaults", "Defaulted")...)
	addRoutes(w, buildRouteForType(w, "valuevalidation", "Foo")...)
	return []*restful.WebService{w}
}

func addRoutes(ws *restful.WebService, routes ...*restful.RouteBuilder) {
	for _, r := range routes {
		ws.Route(r)
	}
}

// Implements OpenAPICanonicalTypeNamer
var _ = util.OpenAPICanonicalTypeNamer(&typeNamer{})

type typeNamer struct {
	pkg  string
	name string
}

func (t *typeNamer) OpenAPICanonicalTypeName() string {
	return fmt.Sprintf("k8s.io/kube-openapi/test/integration/testdata/%s.%s", t.pkg, t.name)
}

func buildRouteForType(ws *restful.WebService, pkg, name string) []*restful.RouteBuilder {
	namer := typeNamer{
		pkg:  pkg,
		name: name,
	}

	routes := []*restful.RouteBuilder{
		ws.GET(fmt.Sprintf("/test/%s/%s", pkg, strings.ToLower(name))).
			Operation(fmt.Sprintf("get-%s.%s", pkg, name)).
			Produces("application/json").
			To(func(*restful.Request, *restful.Response) {}).
			Writes(&namer),
		ws.POST(fmt.Sprintf("/test/%s", pkg)).
			Operation(fmt.Sprintf("create-%s.%s", pkg, name)).
			Produces("application/json").
			To(func(*restful.Request, *restful.Response) {}).
			Returns(201, "Created", &namer).
			Writes(&namer),
	}

	if pkg == "dummytype" {
		statusErrType := typeNamer{
			pkg:  "dummytype",
			name: "StatusError",
		}

		for _, route := range routes {
			route.Returns(500, "Internal Service Error", &statusErrType)
		}
	}

	return routes
}
