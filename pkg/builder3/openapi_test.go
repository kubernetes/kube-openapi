/*
Copyright 2022 The Kubernetes Authors.

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

package builder3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openapi "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/common/restfuladapter"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/util/jsontesting"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// setUp is a convenience function for setting up for (most) tests.
func setUp(t *testing.T, fullMethods bool) (*openapi.OpenAPIV3Config, *restful.Container, *assert.Assertions) {
	assert := assert.New(t)
	config, container := getConfig(fullMethods)
	return config, container, assert
}

func noOp(request *restful.Request, response *restful.Response) {}

// Test input
type TestInput struct {
	// Name of the input
	Name string `json:"name,omitempty"`
	// ID of the input
	ID   int      `json:"id,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

// Test output
type TestOutput struct {
	// Name of the output
	Name string `json:"name,omitempty"`
	// Number of outputs
	Count int `json:"count,omitempty"`
}

func (_ TestInput) OpenAPIDefinition() openapi.OpenAPIDefinition {
	schema := spec.Schema{}
	schema.Description = "Test input"
	schema.Properties = map[string]spec.Schema{
		"name": {
			SchemaProps: spec.SchemaProps{
				Description: "Name of the input",
				Type:        []string{"string"},
				Format:      "",
			},
		},
		"id": {
			SchemaProps: spec.SchemaProps{
				Description: "ID of the input",
				Type:        []string{"integer"},
				Format:      "int32",
			},
		},
		"tags": {
			SchemaProps: spec.SchemaProps{
				Description: "",
				Type:        []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		"reference-extension": {
			VendorExtensible: spec.VendorExtensible{
				Extensions: map[string]interface{}{"extension": "value"},
			},
			SchemaProps: spec.SchemaProps{
				Ref: spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
			},
		},
		"reference-nullable": {
			SchemaProps: spec.SchemaProps{
				Ref:      spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
				Nullable: true,
			},
		},
		"reference-default": {
			SchemaProps: spec.SchemaProps{
				Ref:     spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
				Default: map[string]interface{}{},
			},
		},
	}
	schema.Extensions = spec.Extensions{"x-test": "test"}
	def := openapi.EmbedOpenAPIDefinitionIntoV2Extension(openapi.OpenAPIDefinition{
		Schema:       schema,
		Dependencies: []string{"k8s.io/kube-openapi/pkg/builder3.TestOutput"},
	}, openapi.OpenAPIDefinition{
		// this empty embedded v2 definition should not appear in the result
	})
	return def
}

func (_ TestOutput) OpenAPIDefinition() openapi.OpenAPIDefinition {
	schema := spec.Schema{}
	schema.Description = "Test output"
	schema.Properties = map[string]spec.Schema{
		"name": {
			SchemaProps: spec.SchemaProps{
				Description: "Name of the output",
				Type:        []string{"string"},
				Format:      "",
			},
		},
		"count": {
			SchemaProps: spec.SchemaProps{
				Description: "Number of outputs",
				Type:        []string{"integer"},
				Format:      "int32",
			},
		},
	}
	return openapi.OpenAPIDefinition{
		Schema:       schema,
		Dependencies: []string{},
	}
}

var _ openapi.OpenAPIDefinitionGetter = TestInput{}
var _ openapi.OpenAPIDefinitionGetter = TestOutput{}

// go-restful stores responses in a map and cannot declare the same status twice.
type routeWithResponses struct {
	*restfuladapter.RouteAdapter
	responses []openapi.StatusCodeResponse
}

func (r routeWithResponses) StatusCodeResponses() []openapi.StatusCodeResponse { return r.responses }

func newResponseRoute(responses []restful.ResponseError) routeWithResponses {
	r := routeWithResponses{RouteAdapter: &restfuladapter.RouteAdapter{Route: &restful.Route{
		Produces:    []string{restful.MIME_JSON, "application/yaml"},
		WriteSample: true,
	}}}
	for i := range responses {
		r.responses = append(r.responses, &restfuladapter.ResponseErrorAdapter{Err: &responses[i]})
	}
	return r
}

func TestResponseAlternatives(t *testing.T) {
	config, _ := getConfig(false)
	o := newOpenAPI(config)
	route := newResponseRoute([]restful.ResponseError{
		{Code: 200, Model: (*TestInput)(nil)},
		{Code: 202, Model: int(0)},
		{Code: 200, Model: int32(0)},
		{Code: 200, Model: int64(0)},
		{Code: 202, Model: int32(0)},
		{Code: 200, Model: TestInput{}},
		{Code: 200, Message: "last"},
	})
	op, err := o.buildOperations(route, nil)
	require.NoError(t, err)
	responses := op.Responses.StatusCodeResponses
	require.Len(t, responses, 2)
	assert.Equal(t, "last", responses[200].Description)
	want := &spec.Schema{SchemaProps: spec.SchemaProps{AnyOf: []spec.Schema{
		*spec.RefSchema("#/components/schemas/builder3.TestInput"), *spec.Int32Property(), *spec.Int64Property(),
	}}}
	for _, mime := range route.Produces() {
		assert.Equal(t, want, responses[200].Content[mime].Schema)
		assert.Equal(t, spec.Int32Property(), responses[202].Content[mime].Schema)
	}
	assert.Contains(t, o.spec.Components.Schemas, "builder3.TestInput")
	assert.Contains(t, o.spec.Components.Schemas, "builder3.TestOutput")
}

func TestWriteSampleFallback(t *testing.T) {
	config, _ := getConfig(false)
	o := newOpenAPI(config)
	responses := func(route routeWithResponses) map[int]*spec3.Response {
		t.Helper()
		op, err := o.buildOperations(route, nil)
		require.NoError(t, err)
		return op.Responses.StatusCodeResponses
	}

	route := newResponseRoute(nil)
	assert.Equal(t, spec.BoolProperty(), responses(route)[200].Content[restful.MIME_JSON].Schema)
	route.Route.WriteSample = nil
	assert.Empty(t, responses(route))

	bodyless := responses(newResponseRoute([]restful.ResponseError{{Code: 204}}))
	require.Len(t, bodyless, 1)
	require.NotNil(t, bodyless[204])
	assert.Empty(t, bodyless[204].Content)
}

type responseModelName string

func (m responseModelName) OpenAPIModelName() string { return string(m) }

func TestResponseModelOrder(t *testing.T) {
	// The default component names omit the package path. The first declaration
	// of each name supplies its definition, even across different status codes.
	const (
		resource1 responseModelName = "first/v1.Resource"
		status1   responseModelName = "first/v1.Status"
		resource2 responseModelName = "second/v1.Resource"
		status2   responseModelName = "second/v1.Status"
	)
	definitions := map[string]openapi.OpenAPIDefinition{}
	for _, name := range []responseModelName{resource1, status1, resource2, status2} {
		definitions[string(name)] = openapi.OpenAPIDefinition{Schema: *new(spec.Schema).WithDescription(string(name))}
	}
	o := newOpenAPI(&openapi.OpenAPIV3Config{Definitions: definitions})
	op, err := o.buildOperations(newResponseRoute([]restful.ResponseError{
		{Code: 200, Model: resource1},
		{Code: 202, Model: status1},
		{Code: 202, Model: resource2},
		{Code: 200, Model: status2},
		{Code: 200, Model: resource2},
	}), nil)
	require.NoError(t, err)
	assert.Equal(t, string(resource1), o.spec.Components.Schemas["v1.Resource"].Description)
	assert.Equal(t, string(status1), o.spec.Components.Schemas["v1.Status"].Description)
	assert.Equal(t, []spec.Schema{
		*spec.RefSchema("#/components/schemas/v1.Resource"), *spec.RefSchema("#/components/schemas/v1.Status"),
	}, op.Responses.StatusCodeResponses[200].Content[restful.MIME_JSON].Schema.AnyOf)
}

func TestUndefinedResponseAlternative(t *testing.T) {
	config, _ := getConfig(false)
	o := newOpenAPI(config)
	_, err := o.buildOperations(newResponseRoute([]restful.ResponseError{
		{Code: 200, Model: ""},
		{Code: 200, Model: TestRecursiveType{}},
		{Code: 200, Model: true},
	}), nil)
	require.ErrorContains(t, err, "cannot find model definition for k8s.io/kube-openapi/pkg/builder3.TestRecursiveType")
}

func getTestRoute(ws *restful.WebService, method string, opPrefix string) *restful.RouteBuilder {
	ret := ws.Method(method).
		Path("/test/{path:*}").
		Doc(fmt.Sprintf("%s test input", method)).
		Operation(fmt.Sprintf("%s%sTestInput", method, opPrefix)).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON).
		Param(ws.PathParameter("path", "path to the resource").DataType("string")).
		Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
		Reads(TestInput{}).
		Returns(200, "OK", TestOutput{}).
		Writes(TestOutput{}).
		To(noOp)
	return ret
}

func getConfig(fullMethods bool) (*openapi.OpenAPIV3Config, *restful.Container) {
	mux := http.NewServeMux()
	container := restful.NewContainer()
	container.ServeMux = mux
	ws := new(restful.WebService)
	ws.Path("/foo")
	ws.Route(getTestRoute(ws, "get", "foo"))
	if fullMethods {
		ws.Route(getTestRoute(ws, "post", "foo")).
			Route(getTestRoute(ws, "put", "foo")).
			Route(getTestRoute(ws, "head", "foo")).
			Route(getTestRoute(ws, "patch", "foo")).
			Route(getTestRoute(ws, "options", "foo")).
			Route(getTestRoute(ws, "delete", "foo"))

	}
	ws.Path("/bar")
	ws.Route(getTestRoute(ws, "get", "bar"))
	if fullMethods {
		ws.Route(getTestRoute(ws, "post", "bar")).
			Route(getTestRoute(ws, "put", "bar")).
			Route(getTestRoute(ws, "head", "bar")).
			Route(getTestRoute(ws, "patch", "bar")).
			Route(getTestRoute(ws, "options", "bar")).
			Route(getTestRoute(ws, "delete", "bar"))

	}
	container.Add(ws)
	return &openapi.OpenAPIV3Config{
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       "TestAPI",
				Description: "Test API",
				Version:     "unversioned",
			},
		},
		GetDefinitions: func(_ openapi.ReferenceCallback) map[string]openapi.OpenAPIDefinition {
			return map[string]openapi.OpenAPIDefinition{
				"k8s.io/kube-openapi/pkg/builder3.TestInput":  TestInput{}.OpenAPIDefinition(),
				"k8s.io/kube-openapi/pkg/builder3.TestOutput": TestOutput{}.OpenAPIDefinition(),
			}
		},
		GetDefinitionName: func(name string) (string, spec.Extensions) {
			friendlyName := name[strings.LastIndex(name, "/")+1:]
			return friendlyName, spec.Extensions{"x-test2": "test2"}
		},
	}, container
}

func getTestOperation(method string, opPrefix string) *spec3.Operation {
	return &spec3.Operation{
		OperationProps: spec3.OperationProps{
			Description: fmt.Sprintf("%s test input", method),
			Parameters:  []*spec3.Parameter{},
			Responses:   getTestResponses(),
			OperationId: fmt.Sprintf("%s%sTestInput", method, opPrefix),
		},
	}
}

func getTestPathItem(opPrefix string) *spec3.Path {
	ret := &spec3.Path{
		PathProps: spec3.PathProps{
			Get:        getTestOperation("get", opPrefix),
			Parameters: getTestCommonParameters(),
		},
	}
	ret.Get.RequestBody = getTestRequestBody()
	ret.Put = getTestOperation("put", opPrefix)
	ret.Put.RequestBody = getTestRequestBody()
	ret.Post = getTestOperation("post", opPrefix)
	ret.Post.RequestBody = getTestRequestBody()
	ret.Head = getTestOperation("head", opPrefix)
	ret.Head.RequestBody = getTestRequestBody()
	ret.Patch = getTestOperation("patch", opPrefix)
	ret.Patch.RequestBody = getTestRequestBody()
	ret.Delete = getTestOperation("delete", opPrefix)
	ret.Delete.RequestBody = getTestRequestBody()
	ret.Options = getTestOperation("options", opPrefix)
	ret.Options.RequestBody = getTestRequestBody()
	return ret
}

func getRefSchema(ref string) *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef(ref),
		},
	}
}

func getTestResponses() *spec3.Responses {
	ret := &spec3.Responses{
		ResponsesProps: spec3.ResponsesProps{
			StatusCodeResponses: map[int]*spec3.Response{},
		},
	}
	ret.StatusCodeResponses[200] = &spec3.Response{
		ResponseProps: spec3.ResponseProps{
			Description: "OK",
			Content:     map[string]*spec3.MediaType{},
		},
	}

	ret.StatusCodeResponses[200].Content[restful.MIME_JSON] = &spec3.MediaType{
		MediaTypeProps: spec3.MediaTypeProps{
			Schema: getRefSchema("#/components/schemas/builder3.TestOutput"),
		},
	}

	return ret
}

func getTestCommonParameters() []*spec3.Parameter {
	ret := make([]*spec3.Parameter, 2)
	ret[0] = &spec3.Parameter{
		ParameterProps: spec3.ParameterProps{
			Description: "path to the resource",
			Name:        "path",
			In:          "path",
			Required:    true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{"string"},
					UniqueItems: true,
				},
			},
		},
	}
	ret[1] = &spec3.Parameter{
		ParameterProps: spec3.ParameterProps{
			Description: "If 'true', then the output is pretty printed.",
			Name:        "pretty",
			In:          "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{"string"},
					UniqueItems: true,
				},
			},
		},
	}
	return ret
}

func getTestRequestBody() *spec3.RequestBody {
	ret := &spec3.RequestBody{
		RequestBodyProps: spec3.RequestBodyProps{
			Content: map[string]*spec3.MediaType{
				restful.MIME_JSON: {
					MediaTypeProps: spec3.MediaTypeProps{
						Schema: getRefSchema("#/components/schemas/builder3.TestInput"),
					},
				},
			},
			Required: true,
		},
	}
	return ret
}

func getTestInputDefinition() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Test input",
			Properties: map[string]spec.Schema{
				"id": {
					SchemaProps: spec.SchemaProps{
						Description: "ID of the input",
						Type:        spec.StringOrArray{"integer"},
						Format:      "int32",
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						Description: "Name of the input",
						Type:        spec.StringOrArray{"string"},
					},
				},
				"tags": {
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{"array"},
						Items: &spec.SchemaOrArray{
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type: spec.StringOrArray{"string"},
								},
							},
						},
					},
				},
				"reference-extension": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: map[string]interface{}{"extension": "value"},
					},
					SchemaProps: spec.SchemaProps{
						AllOf: []spec.Schema{{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
							},
						}},
					},
				},
				"reference-nullable": {
					SchemaProps: spec.SchemaProps{
						Nullable: true,
						AllOf: []spec.Schema{{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
							},
						}},
					},
				},
				"reference-default": {
					SchemaProps: spec.SchemaProps{
						AllOf: []spec.Schema{{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef("/components/schemas/builder3.TestOutput"),
							},
						}},
						Default: map[string]interface{}{},
					},
				},
			},
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: spec.Extensions{
				"x-test":  "test",
				"x-test2": "test2",
			},
		},
	}
}

func getTestOutputDefinition() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Test output",
			Properties: map[string]spec.Schema{
				"count": {
					SchemaProps: spec.SchemaProps{
						Description: "Number of outputs",
						Type:        spec.StringOrArray{"integer"},
						Format:      "int32",
					},
				},
				"name": {
					SchemaProps: spec.SchemaProps{
						Description: "Name of the output",
						Type:        spec.StringOrArray{"string"},
					},
				},
			},
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: spec.Extensions{
				"x-test2": "test2",
			},
		},
	}
}

func TestBuildOpenAPISpec(t *testing.T) {
	config, container, assert := setUp(t, true)
	expected := &spec3.OpenAPI{
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       "TestAPI",
				Description: "Test API",
				Version:     "unversioned",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: map[string]any{
					"hello": "world", // set from PostProcessSpec callback
				},
			},
		},
		Version: "3.0.0",
		Paths: &spec3.Paths{
			Paths: map[string]*spec3.Path{
				"/foo/test/{path}": getTestPathItem("foo"),
				"/bar/test/{path}": getTestPathItem("bar"),
			},
		},
		Components: &spec3.Components{
			Schemas: map[string]*spec.Schema{
				"builder3.TestInput":  getTestInputDefinition(),
				"builder3.TestOutput": getTestOutputDefinition(),
			},
		},
	}
	config.PostProcessSpec = func(s *spec3.OpenAPI) (*spec3.OpenAPI, error) {
		s.Info.Extensions = map[string]any{
			"hello": "world",
		}
		return s, nil
	}
	swagger, err := BuildOpenAPISpec(container.RegisteredWebServices(), config)
	if !assert.NoError(err) {
		return
	}
	expected_json, err := json.Marshal(expected)
	if !assert.NoError(err) {
		return
	}
	actual_json, err := json.Marshal(swagger)
	if !assert.NoError(err) {
		return
	}
	if err := jsontesting.JsonCompare(expected_json, actual_json); err != nil {
		t.Error(err)
	}
}

type TestRecursiveType struct {
	Value    string              `json:"value,omitempty"`
	Children []TestRecursiveType `json:"children,omitempty"`
}

func makeRecursiveTypeDefinition(depName string) openapi.OpenAPIDefinition {
	schema := spec.Schema{}
	schema.Description = "Test recursive type for JSON pointer escaping"
	schema.Properties = map[string]spec.Schema{
		"value": {
			SchemaProps: spec.SchemaProps{
				Description: "A test value",
				Type:        []string{"string"},
			},
		},
		"children": {
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Ref: spec.MustCreateRef("#/components/schemas/" + depName),
						},
					},
				},
			},
		},
	}
	return openapi.OpenAPIDefinition{
		Schema:       schema,
		Dependencies: []string{depName},
	}
}

func TestEscapeJsonPointerInSchemaName(t *testing.T) {
	testCases := []struct {
		name           string
		definitionName string
		expectedName   string
		shouldNotExist string
	}{
		{
			name:           "both slash and tilde",
			definitionName: "io.k8s/api~v1.TestRecursiveType",
			expectedName:   "io.k8s~1api~0v1.TestRecursiveType",
			shouldNotExist: "io.k8s/api~v1.TestRecursiveType",
		},
		{
			name:           "only slashes",
			definitionName: "io.k8s/api/v1.TestRecursiveType",
			expectedName:   "io.k8s~1api~1v1.TestRecursiveType",
			shouldNotExist: "io.k8s/api/v1.TestRecursiveType",
		},
		{
			name:           "only tildes",
			definitionName: "io.k8s~api~v1.TestRecursiveType",
			expectedName:   "io.k8s~0api~0v1.TestRecursiveType",
			shouldNotExist: "io.k8s~api~v1.TestRecursiveType",
		},
		{
			name:           "no special characters",
			definitionName: "io.k8s.api.v1.TestRecursiveType",
			expectedName:   "io.k8s.api.v1.TestRecursiveType",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			config := &openapi.OpenAPIV3Config{
				Info: &spec.Info{
					InfoProps: spec.InfoProps{
						Title:       "TestAPI",
						Description: "Test API",
						Version:     "v1",
					},
				},
				GetDefinitions: func(_ openapi.ReferenceCallback) map[string]openapi.OpenAPIDefinition {
					return map[string]openapi.OpenAPIDefinition{
						tc.definitionName: makeRecursiveTypeDefinition(tc.definitionName),
					}
				},
				GetDefinitionName: func(name string) (string, spec.Extensions) {
					return name, nil
				},
			}

			schemas, err := BuildOpenAPIDefinitionsForResources(config, tc.definitionName)
			assert.NoError(err)
			assert.NotNil(schemas)

			_, exists := schemas[tc.expectedName]
			assert.True(exists, "Schema should exist with escaped name: %s", tc.expectedName)

			if tc.shouldNotExist != "" {
				_, exists = schemas[tc.shouldNotExist]
				assert.False(exists, "Schema should not exist with unescaped name: %s", tc.shouldNotExist)
			}
		})
	}
}
