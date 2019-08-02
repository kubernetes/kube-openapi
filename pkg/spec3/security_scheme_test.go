package spec3_test

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/google/go-cmp/cmp"

	"k8s.io/kube-openapi/pkg/spec3"
)

func TestSecuritySchemaJSONSerialization(t *testing.T) {
	cases := []struct {
		name           string
		target         *spec3.SecurityScheme
		expectedOutput string
	}{
		// scenario 1
		{
			name: "scenario1: basic authentication",
			target: &spec3.SecurityScheme{
				SecuritySchemeProps: spec3.SecuritySchemeProps{
					Type:   "http",
					Scheme: "basic",
				},
			},
			expectedOutput: `{"type":"http","scheme":"basic"}`,
		},

		// scenario 2
		{
			name: "scenario2: JWT Bearer",
			target: &spec3.SecurityScheme{
				SecuritySchemeProps: spec3.SecuritySchemeProps{
					Type:         "http",
					Scheme:       "basic",
					BearerFormat: "JWT",
				},
			},
			expectedOutput: `{"type":"http","scheme":"basic","bearerFormat":"JWT"}`,
		},

		// scenario 3
		{
			name: "scenario3: implicit OAuth2",
			target: &spec3.SecurityScheme{
				SecuritySchemeProps: spec3.SecuritySchemeProps{
					Type: "oauth2",
					Flows: map[string]*spec3.OAuthFlow{
						"implicit": &spec3.OAuthFlow{
							OAuthFlowProps: spec3.OAuthFlowProps{
								AuthorizationUrl: "https://example.com/api/oauth/dialog",
								Scopes: map[string]string{
									"write:pets": "modify pets in your account",
									"read:pets":  "read your pets",
								},
							},
						},
					},
				},
			},
			expectedOutput: `{"type":"oauth2","flows":{"implicit":{"authorizationUrl":"https://example.com/api/oauth/dialog","scopes":{"read:pets":"read your pets","write:pets":"modify pets in your account"}}}}`,
		},

		// scenario 4
		{
			name: "scenario4: reference Object",
			target: &spec3.SecurityScheme{
				Refable: spec.Refable{Ref: spec.MustCreateRef("k8s.io/api/foo/v1beta1b.bar")},
			},
			expectedOutput: `{"$ref":"k8s.io/api/foo/v1beta1b.bar"}`,
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
