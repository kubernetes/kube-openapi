package spec3_test

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/google/go-cmp/cmp"

	"k8s.io/kube-openapi/pkg/spec3"
)

func TestPathsJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.Paths
		expectedOutput string
	}{
		// scenario 1
		{
			name: "scenario1: smoke test serialization of Paths object",
			target: func() *spec3.Paths {
				p := &spec3.Paths{Paths: map[string]*spec3.Path{}}

				p1 := &spec3.Path{}
				p1.Parameters = []*spec3.Parameter{}
				{ /* GET operation for p1 path */
					o := &spec3.Operation{}
					o.Description = "Returns pets based on ID"
					o.Summary = "Find pets by ID"
					o.OperationId = "getPetsById"

					ors := &spec3.Responses{ResponsesProps: spec3.ResponsesProps{StatusCodeResponses: map[int]*spec3.Response{}}}

					// HTTP 200 response
					r200 := &spec3.Response{}
					r200.Description = "pet response"
					r200.Content = map[string]*spec3.MediaType{
						"*/*": &spec3.MediaType{
							MediaTypeProps: spec3.MediaTypeProps{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"array"},
										Items: &spec.SchemaOrArray{
											Schema: &spec.Schema{
												SchemaProps: spec.SchemaProps{
													Ref: spec.MustCreateRef("#/components/schemas/Pet"),
												},
											},
										},
									},
								},
							},
						},
					}
					ors.StatusCodeResponses[200] = r200

					// Default response
					rDef := &spec3.Response{}
					rDef.Description = "error payload"
					rDef.Content = map[string]*spec3.MediaType{
						"text/html": &spec3.MediaType{
							MediaTypeProps: spec3.MediaTypeProps{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: spec.MustCreateRef("#/components/schemas/ErrorModel"),
									},
								},
							},
						},
					}
					ors.Default = rDef

					o.Responses = ors
					p1.Get = o
				}
				{ /* Parameters for p1 path */
					pr := &spec3.Parameter{}

					pr.Name = "id"
					pr.In = "path"
					pr.Description = "ID of pet to use"
					pr.Required = true
					pr.Schema = &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
							},
						},
					}
					pr.Style = "simple"

					p1.Parameters = append(p1.Parameters, pr)
				}

				p.Paths["pets"] = p1
				return p
			}(),
			expectedOutput: `{"pets":{"get":{"summary":"Find pets by ID","description":"Returns pets based on ID","operationId":"getPetsById","responses":{"200":{"description":"pet response","content":{"*/*":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/Pet"}}}}},"default":{"description":"error payload","content":{"text/html":{"schema":{"$ref":"#/components/schemas/ErrorModel"}}}}}},"parameters":[{"name":"id","in":"path","description":"ID of pet to use","required":true,"style":"simple","schema":{"type":"array","items":{"type":"string"}}}]}}`,
		},

		{
			name: "scenario2: smoke test more than one operation serialization",
			target: &spec3.Paths{
				Paths: map[string]*spec3.Path{
					"/api/v1/pods": &spec3.Path{
						PathProps: spec3.PathProps{
							Summary: "summary1",
							Get: &spec3.Operation{
								OperationProps: spec3.OperationProps{
									OperationId: "operation1",
								},
							},
							Post: &spec3.Operation{
								OperationProps: spec3.OperationProps{
									OperationId: "operation2",
								},
							},
						},
					},
				},
			},
			expectedOutput: `{"/api/v1/pods":{"summary":"summary1","get":{"operationId":"operation1"},"post":{"operationId":"operation2"}}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawTarget, err := json.Marshal(tc.target)
			if err != nil {
				t.Fatal(err)
			}
			serializedTarget := string(rawTarget)
			if !cmp.Equal(serializedTarget, tc.expectedOutput) {
				t.Fatalf("diff %s", cmp.Diff(serializedTarget, tc.expectedOutput))
			}
		})
	}
}
