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

package spec_test

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/jsonreference"
	"github.com/google/gnostic/compiler"
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
	. "k8s.io/kube-openapi/pkg/validation/spec"
)

var SpecV2DiffOptions = []cmp.Option{
	// cmp.Diff panics on Ref since jsonreference.Ref uses unexported fields
	cmp.Comparer(func(a Ref, b Ref) bool {
		return a.String() == b.String()
	}),
}

func gnosticCommonTest(t testing.TB, fuzzer *fuzz.Fuzzer) {
	fuzzer.Funcs(
		func(v *Responses, c fuzz.Continue) {
			c.FuzzNoCustom(v)
			if v.Default != nil {
				// Check if we hit maxDepth and left an incomplete value
				if v.Default.Description == "" {
					v.Default = nil
					v.StatusCodeResponses = nil
				}
			}

			// conversion has no way to discern empty statusCodeResponses from
			// nil, since "default" is always included in the map.
			// So avoid empty responses list
			if len(v.StatusCodeResponses) == 0 {
				v.StatusCodeResponses = nil
			}
		},
		func(v *Operation, c fuzz.Continue) {
			c.FuzzNoCustom(v)

			if v != nil {
				// force non-nil
				v.Responses = &Responses{}
				c.Fuzz(v.Responses)

				v.Schemes = nil
				if c.RandBool() {
					v.Schemes = append(v.Schemes, "http")
				}

				if c.RandBool() {
					v.Schemes = append(v.Schemes, "https")
				}

				if c.RandBool() {
					v.Schemes = append(v.Schemes, "ws")
				}

				if c.RandBool() {
					v.Schemes = append(v.Schemes, "wss")
				}

				// Gnostic unconditionally makes security values non-null
				// So do not fuzz null values into the array.
				for i, val := range v.Security {
					if val == nil {
						v.Security[i] = make(map[string][]string)
					}

					for k, v := range val {
						if v == nil {
							val[k] = make([]string, 0)
						}
					}
				}
			}
		},
		func(v map[int]Response, c fuzz.Continue) {
			n := 0
			c.Fuzz(&n)
			if n == 0 {
				// Test that fuzzer is not at maxDepth so we do not
				// end up with empty elements
				return
			}

			// Prevent negative numbers
			num := c.Intn(4)
			for i := 0; i < num+2; i++ {
				val := Response{}
				c.Fuzz(&val)

				val.Description = c.RandString() + "x"
				v[100*(i+1)+c.Intn(100)] = val
			}
		},
		func(v map[string]PathItem, c fuzz.Continue) {
			n := 0
			c.Fuzz(&n)
			if n == 0 {
				// Test that fuzzer is not at maxDepth so we do not
				// end up with empty elements
				return
			}

			num := c.Intn(5)
			for i := 0; i < num+2; i++ {
				val := PathItem{}
				c.Fuzz(&val)

				// Ref params are only allowed in certain locations, so
				// possibly add a few to PathItems
				numRefsToAdd := c.Intn(5)
				for i := 0; i < numRefsToAdd; i++ {
					theRef := Parameter{}
					c.Fuzz(&theRef.Refable)

					val.Parameters = append(val.Parameters, theRef)
				}

				v["/"+c.RandString()] = val
			}
		},
		func(v *SchemaOrArray, c fuzz.Continue) {
			*v = SchemaOrArray{}
			// gnostic parser just doesn't support more
			// than one Schema here
			v.Schema = &Schema{}
			c.Fuzz(&v.Schema)

		},
		func(v *SchemaOrBool, c fuzz.Continue) {
			*v = SchemaOrBool{}

			if c.RandBool() {
				v.Allows = c.RandBool()
			} else {
				v.Schema = &Schema{}
				c.Fuzz(&v.Schema)
			}
		},
		func(v map[string]Response, c fuzz.Continue) {
			n := 0
			c.Fuzz(&n)
			if n == 0 {
				// Test that fuzzer is not at maxDepth so we do not
				// end up with empty elements
				return
			}

			// Response definitions are not allowed to
			// be refs
			for i := 0; i < c.Intn(5)+1; i++ {
				resp := &Response{}

				c.Fuzz(resp)
				resp.Ref = Ref{}
				resp.Description = c.RandString() + "x"

				// Response refs are not vendor extensible by gnostic
				resp.VendorExtensible.Extensions = nil
				v[c.RandString()+"x"] = *resp
			}
		},
		func(v *Header, c fuzz.Continue) {
			if v != nil {
				c.FuzzNoCustom(v)

				// descendant Items of Header may not be refs
				cur := v.Items
				for cur != nil {
					cur.Ref = Ref{}
					cur = cur.Items
				}
			}
		},
		func(v *Ref, c fuzz.Continue) {
			*v = Ref{}
			v.Ref, _ = jsonreference.New("http://asd.com/" + c.RandString())
		},
		func(v *Response, c fuzz.Continue) {
			*v = Response{}
			if c.RandBool() {
				v.Ref = Ref{}
				v.Ref.Ref, _ = jsonreference.New("http://asd.com/" + c.RandString())
			} else {
				c.Fuzz(&v.VendorExtensible)
				c.Fuzz(&v.Schema)
				c.Fuzz(&v.ResponseProps)

				v.Headers = nil
				v.Ref = Ref{}

				n := 0
				c.Fuzz(&n)
				if n != 0 {
					// Test that fuzzer is not at maxDepth so we do not
					// end up with empty elements
					num := c.Intn(4)
					for i := 0; i < num; i++ {
						if v.Headers == nil {
							v.Headers = make(map[string]Header)
						}
						hdr := Header{}
						c.Fuzz(&hdr)
						if hdr.Type == "" {
							// hit maxDepth, just abort trying to make haders
							v.Headers = nil
							break
						}
						v.Headers[c.RandString()+"x"] = hdr
					}
				} else {
					v.Headers = nil
				}
			}

			v.Description = c.RandString() + "x"

			// Gnostic parses empty as nil, so to keep avoid putting empty
			if len(v.Headers) == 0 {
				v.Headers = nil
			}
		},
		func(v **Info, c fuzz.Continue) {
			// Info is never nil
			*v = &Info{}
			c.FuzzNoCustom(*v)

			(*v).Title = c.RandString() + "x"
		},
		func(v *Extensions, c fuzz.Continue) {
			// gnostic parser only picks up x- vendor extensions
			numChildren := c.Intn(5)
			for i := 0; i < numChildren; i++ {
				if *v == nil {
					*v = Extensions{}
				}
				(*v)["x-"+c.RandString()] = c.RandString()
			}
		},
		func(v *Swagger, c fuzz.Continue) {
			c.FuzzNoCustom(v)

			if v.Paths == nil {
				// Force paths non-nil since it does not have omitempty in json tag.
				// This means a perfect roundtrip (via json) is impossible,
				// since we can't tell the difference between empty/unspecified paths
				v.Paths = &Paths{}
				c.Fuzz(v.Paths)
			}

			v.Swagger = "2.0"

			// Gnostic support serializing ID at all
			// unavoidable data loss
			v.ID = ""

			v.Schemes = nil
			if c.RandUint64()%2 == 1 {
				v.Schemes = append(v.Schemes, "http")
			}

			if c.RandUint64()%2 == 1 {
				v.Schemes = append(v.Schemes, "https")
			}

			if c.RandUint64()%2 == 1 {
				v.Schemes = append(v.Schemes, "ws")
			}

			if c.RandUint64()%2 == 1 {
				v.Schemes = append(v.Schemes, "wss")
			}

			// Gnostic unconditionally makes security values non-null
			// So do not fuzz null values into the array.
			for i, val := range v.Security {
				if val == nil {
					v.Security[i] = make(map[string][]string)
				}

				for k, v := range val {
					if v == nil {
						val[k] = make([]string, 0)
					}
				}
			}
		},
		func(v *SecurityScheme, c fuzz.Continue) {
			v.Description = c.RandString() + "x"
			c.Fuzz(&v.VendorExtensible)

			switch c.Intn(3) {
			case 0:
				v.Type = "basic"
			case 1:
				v.Type = "apiKey"
				switch c.Intn(2) {
				case 0:
					v.In = "header"
				case 1:
					v.In = "query"
				default:
					panic("unreachable")
				}
				v.Name = "x" + c.RandString()
			case 2:
				v.Type = "oauth2"

				switch c.Intn(4) {
				case 0:
					v.Flow = "accessCode"
					v.TokenURL = "https://" + c.RandString()
					v.AuthorizationURL = "https://" + c.RandString()
				case 1:
					v.Flow = "application"
					v.TokenURL = "https://" + c.RandString()
				case 2:
					v.Flow = "implicit"
					v.AuthorizationURL = "https://" + c.RandString()
				case 3:
					v.Flow = "password"
					v.TokenURL = "https://" + c.RandString()
				default:
					panic("unreachable")
				}
				c.Fuzz(&v.Scopes)
			default:
				panic("unreachable")
			}
		},
		func(v *interface{}, c fuzz.Continue) {
			*v = c.RandString() + "x"
		},
		func(v *string, c fuzz.Continue) {
			*v = c.RandString() + "x"
		},
		func(v *ExternalDocumentation, c fuzz.Continue) {
			v.Description = c.RandString() + "x"
			v.URL = c.RandString() + "x"
		},
		func(v *SimpleSchema, c fuzz.Continue) {
			c.FuzzNoCustom(v)

			switch c.Intn(5) {
			case 0:
				v.Type = "string"
			case 1:
				v.Type = "number"
			case 2:
				v.Type = "boolean"
			case 3:
				v.Type = "integer"
			case 4:
				v.Type = "array"
			default:
				panic("unreachable")
			}

			switch c.Intn(5) {
			case 0:
				v.CollectionFormat = "csv"
			case 1:
				v.CollectionFormat = "ssv"
			case 2:
				v.CollectionFormat = "tsv"
			case 3:
				v.CollectionFormat = "pipes"
			case 4:
				v.CollectionFormat = ""
			default:
				panic("unreachable")
			}

			// None of the types which include SimpleSchema in our definitions
			// actually support "example" in the official spec
			v.Example = nil

			// unsupported by openapi
			v.Nullable = false
		},
		func(v *int64, c fuzz.Continue) {
			c.Fuzz(v)

			// Gnostic does not differentiate between 0 and non-specified
			// so avoid using 0 for fuzzer
			if *v == 0 {
				*v = 1
			}
		},
		func(v *float64, c fuzz.Continue) {
			c.Fuzz(v)

			// Gnostic does not differentiate between 0 and non-specified
			// so avoid using 0 for fuzzer
			if *v == 0.0 {
				*v = 1.0
			}
		},
		func(v *Parameter, c fuzz.Continue) {
			if v == nil {
				return
			}
			c.Fuzz(&v.VendorExtensible)
			if c.RandBool() {
				// body param
				v.Description = c.RandString() + "x"
				v.Name = c.RandString() + "x"
				v.In = "body"
				c.Fuzz(&v.Description)
				c.Fuzz(&v.Required)

				v.Schema = &Schema{}
				c.Fuzz(&v.Schema)

			} else {
				c.Fuzz(&v.SimpleSchema)
				c.Fuzz(&v.CommonValidations)
				v.AllowEmptyValue = false
				v.Description = c.RandString() + "x"
				v.Name = c.RandString() + "x"

				switch c.Intn(4) {
				case 0:
					// Header param
					v.In = "header"
				case 1:
					// Form data param
					v.In = "formData"
					v.AllowEmptyValue = c.RandBool()
				case 2:
					// Query param
					v.In = "query"
					v.AllowEmptyValue = c.RandBool()
				case 3:
					// Path param
					v.In = "path"
					v.Required = true
				default:
					panic("unreachable")
				}

				// descendant Items of Parameter may not be refs
				cur := v.Items
				for cur != nil {
					cur.Ref = Ref{}
					cur = cur.Items
				}
			}
		},
		func(v *Schema, c fuzz.Continue) {
			if c.RandBool() {
				// file schema
				c.Fuzz(&v.Default)
				c.Fuzz(&v.Description)
				c.Fuzz(&v.Example)
				c.Fuzz(&v.ExternalDocs)

				c.Fuzz(&v.Format)
				c.Fuzz(&v.ReadOnly)
				c.Fuzz(&v.Required)
				c.Fuzz(&v.Title)
				v.Type = StringOrArray{"file"}

			} else {
				// normal schema
				c.Fuzz(&v.SchemaProps)
				c.Fuzz(&v.SwaggerSchemaProps)
				c.Fuzz(&v.VendorExtensible)
				// c.Fuzz(&v.ExtraProps)
				// ExtraProps will not roundtrip - gnostic throws out
				// unrecognized keys
			}

			// Not supported by official openapi v2 spec
			// and stripped by k8s apiserver
			v.ID = ""
			v.AnyOf = nil
			v.OneOf = nil
			v.Not = nil
			v.Nullable = false
			v.AdditionalItems = nil
			v.Schema = ""
			v.PatternProperties = nil
			v.Definitions = nil
			v.Dependencies = nil
		},
	)

	expected := Swagger{}
	fuzzer.Fuzz(&expected)

	// Convert to gnostic via JSON to compare
	jsonBytes, err := json.Marshal(expected)
	require.NoError(t, err)

	t.Log("Specimen", string(jsonBytes))

	gnosticSpec, err := openapi_v2.ParseDocument(jsonBytes)
	require.NoError(t, err)

	actual := Swagger{}
	ok, err := actual.FromGnostic(gnosticSpec)
	require.NoError(t, err)
	require.True(t, ok)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual, SpecV2DiffOptions...))
	}

	newJsonBytes, err := json.Marshal(actual)
	require.NoError(t, err)
	if !reflect.DeepEqual(jsonBytes, newJsonBytes) {
		t.Fatal(cmp.Diff(string(jsonBytes), string(newJsonBytes), SpecV2DiffOptions...))
	}
}

func TestGnosticConversionSmallDeterministic(t *testing.T) {
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(15).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic2(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1646770841).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic3(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1646772024).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic4(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1646791953).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic5(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1646940131).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic6(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1646941926).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallDeterministic7(t *testing.T) {
	// A failed case of TestGnosticConversionSmallRandom
	// which failed during development/testing loop
	// This case did not convert nil/empty array within OperationProps.Security
	// correctly
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(1647297721085690000).
			NilChance(0.8).
			MaxDepth(10).
			NumElements(1, 2),
	)
}

func TestGnosticConversionSmallRandom(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Log("Using seed: ", seed)
	fuzzer := fuzz.
		NewWithSeed(seed).
		NilChance(0.8).
		MaxDepth(10).
		NumElements(1, 2)

	for i := 0; i <= 50; i++ {
		gnosticCommonTest(
			t,
			fuzzer,
		)
	}
}

func TestGnosticConversionMediumDeterministic(t *testing.T) {
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(15).
			NilChance(0.4).
			MaxDepth(12).
			NumElements(3, 5),
	)
}

func TestGnosticConversionLargeDeterministic(t *testing.T) {
	gnosticCommonTest(
		t,
		fuzz.
			NewWithSeed(15).
			NilChance(0.1).
			MaxDepth(15).
			NumElements(3, 5),
	)
}

func TestGnosticConversionLargeRandom(t *testing.T) {
	var seed int64 = time.Now().UnixNano()
	t.Log("Using seed: ", seed)
	fuzzer := fuzz.
		NewWithSeed(seed).
		NilChance(0).
		MaxDepth(15).
		NumElements(3, 5)

	for i := 0; i < 5; i++ {
		gnosticCommonTest(
			t,
			fuzzer,
		)
	}
}

func BenchmarkGnosticConversion(b *testing.B) {
	// Download kube-openapi swagger json
	swagFile, err := os.Open("../../schemaconv/testdata/swagger.json")
	if err != nil {
		b.Fatal(err)
	}
	defer swagFile.Close()

	originalJSON, err := io.ReadAll(swagFile)
	if err != nil {
		b.Fatal(err)
	}

	// Parse into kube-openapi types
	var result *Swagger
	b.Run("json->swagger", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			if err := json.Unmarshal(originalJSON, &result); err != nil {
				b2.Fatal(err)
			}
		}
	})

	// Convert to JSON
	var encodedJSON []byte
	b.Run("swagger->json", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			encodedJSON, err = json.Marshal(result)
			if err != nil {
				b2.Fatal(err)
			}
		}
	})

	// Convert to gnostic
	var originalGnostic *openapi_v2.Document
	b.Run("json->gnostic", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			originalGnostic, err = openapi_v2.ParseDocument(encodedJSON)
			if err != nil {
				b2.Fatal(err)
			}
		}
	})

	// Convert to PB
	var encodedProto []byte
	b.Run("gnostic->pb", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			encodedProto, err = proto.Marshal(originalGnostic)
			if err != nil {
				b2.Fatal(err)
			}
		}
	})

	// Convert to gnostic
	var backToGnostic openapi_v2.Document
	b.Run("pb->gnostic", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			if err := proto.Unmarshal(encodedProto, &backToGnostic); err != nil {
				b2.Fatal(err)
			}
		}
	})

	for i := 0; i < b.N; i++ {
		b.Run("gnostic->kube", func(b2 *testing.B) {
			for i := 0; i < b2.N; i++ {
				decodedSwagger := &Swagger{}
				if ok, err := decodedSwagger.FromGnostic(&backToGnostic); err != nil {
					b2.Fatal(err)
				} else if !ok {
					b2.Fatal("conversion lost data")
				}
			}
		})
	}
}

// Ensure all variants of SecurityDefinition are being exercised by tests
func TestSecurityDefinitionVariants(t *testing.T) {
	type TestPattern struct {
		Name    string
		Pattern string
	}

	patterns := []TestPattern{
		{
			Name:    "Basic Authentication",
			Pattern: `{"type": "basic", "description": "cool basic auth"}`,
		},
		{
			Name:    "API Key Query",
			Pattern: `{"type": "apiKey", "description": "cool api key auth", "in": "query", "name": "coolAuth"}`,
		},
		{
			Name:    "API Key Header",
			Pattern: `{"type": "apiKey", "description": "cool api key auth", "in": "header", "name": "coolAuth"}`,
		},
		{
			Name:    "OAuth2 Implicit",
			Pattern: `{"type": "oauth2", "flow": "implicit", "authorizationUrl": "https://google.com", "scopes": {"scope1": "a scope", "scope2": "a scope"}, "description": "cool oauth2 auth"}`,
		},
		{
			Name:    "OAuth2 Password",
			Pattern: `{"type": "oauth2", "flow": "password", "tokenUrl": "https://google.com", "scopes": {"scope1": "a scope", "scope2": "a scope"}, "description": "cool oauth2 auth"}`,
		},
		{
			Name:    "OAuth2 Application",
			Pattern: `{"type": "oauth2", "flow": "application", "tokenUrl": "https://google.com", "scopes": {"scope1": "a scope", "scope2": "a scope"}, "description": "cool oauth2 auth"}`,
		},
		{
			Name:    "OAuth2 Access Code",
			Pattern: `{"type": "oauth2", "flow": "accessCode", "authorizationUrl": "https://google.com", "tokenUrl": "https://google.com", "scopes": {"scope1": "a scope", "scope2": "a scope"}, "description": "cool oauth2 auth"}`,
		},
	}

	for _, p := range patterns {
		t.Run(p.Name, func(t *testing.T) {
			// Parse JSON into yaml
			var nodes yaml.Node
			if err := yaml.Unmarshal([]byte(p.Pattern), &nodes); err != nil {
				t.Error(err)
				return
			} else if len(nodes.Content) != 1 {
				t.Errorf("unexpected yaml parse result")
				return
			}

			root := nodes.Content[0]

			parsed, err := openapi_v2.NewSecurityDefinitionsItem(root, compiler.NewContextWithExtensions("$root", root, nil, nil))
			if err != nil {
				t.Error(err)
				return
			}

			converted := SecurityScheme{}
			if err := converted.FromGnostic(parsed); err != nil {
				t.Error(err)
				return
			}

			// Ensure that the same JSON parsed via kube-openapi gives the same
			// result
			var expected SecurityScheme
			if err := json.Unmarshal([]byte(p.Pattern), &expected); err != nil {
				t.Error(err)
				return
			} else if !reflect.DeepEqual(expected, converted) {
				t.Errorf("expected equal values: %v", cmp.Diff(expected, converted, SpecV2DiffOptions...))
				return
			}
		})
	}
}

// Ensure all variants of Parameter are being exercised by tests
func TestParamVariants(t *testing.T) {
	type TestPattern struct {
		Name    string
		Pattern string
	}

	patterns := []TestPattern{
		{
			Name:    "Body Parameter",
			Pattern: `{"in": "body", "name": "myBodyParam", "schema": {}}`,
		},
		{
			Name:    "NonBody Header Parameter",
			Pattern: `{"in": "header", "name": "myHeaderParam", "description": "a cool parameter", "type": "string", "collectionFormat": "pipes"}`,
		},
		{
			Name:    "NonBody FormData Parameter",
			Pattern: `{"in": "formData", "name": "myFormDataParam", "description": "a cool parameter", "type": "string", "collectionFormat": "pipes"}`,
		},
		{
			Name:    "NonBody Query Parameter",
			Pattern: `{"in": "query", "name": "myQueryParam", "description": "a cool parameter", "type": "string", "collectionFormat": "pipes"}`,
		},
		{
			Name:    "NonBody Path Parameter",
			Pattern: `{"required": true, "in": "path", "name": "myPathParam", "description": "a cool parameter", "type": "string", "collectionFormat": "pipes"}`,
		},
	}

	for _, p := range patterns {
		t.Run(p.Name, func(t *testing.T) {
			// Parse JSON into yaml
			var nodes yaml.Node
			if err := yaml.Unmarshal([]byte(p.Pattern), &nodes); err != nil {
				t.Error(err)
				return
			} else if len(nodes.Content) != 1 {
				t.Errorf("unexpected yaml parse result")
				return
			}

			root := nodes.Content[0]

			ctx := compiler.NewContextWithExtensions("$root", root, nil, nil)
			parsed, err := openapi_v2.NewParameter(root, ctx)
			if err != nil {
				t.Error(err)
				return
			}

			converted := Parameter{}
			if ok, err := converted.FromGnostic(parsed); err != nil {
				t.Error(err)
				return
			} else if !ok {
				t.Errorf("expected no data loss while converting parameter: %v", p.Pattern)
				return
			}

			// Ensure that the same JSON parsed via kube-openapi gives the same
			// result
			var expected Parameter
			if err := json.Unmarshal([]byte(p.Pattern), &expected); err != nil {
				t.Error(err)
				return
			} else if !reflect.DeepEqual(expected, converted) {
				t.Errorf("expected equal values: %v", cmp.Diff(expected, converted, SpecV2DiffOptions...))
				return
			}
		})
	}
}

// Test that a few patterns of obvious data loss are detected
func TestCommonDataLoss(t *testing.T) {
	type TestPattern struct {
		Name          string
		BadInstance   string
		FixedInstance string
	}

	patterns := []TestPattern{
		{
			Name:          "License with Vendor Extension",
			BadInstance:   `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "license": {"name": "MIT", "x-hello": "ignored"}}, "paths": {}}`,
			FixedInstance: `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "license": {"name": "MIT"}}, "paths": {}}`,
		},
		{
			Name:          "Contact with Vendor Extension",
			BadInstance:   `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "contact": {"name": "bill", "x-hello": "ignored"}}, "paths": {}}`,
			FixedInstance: `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "contact": {"name": "bill"}}, "paths": {}}`,
		},
		{
			Name:          "External Documentation with Vendor Extension",
			BadInstance:   `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "contact": {"name": "bill", "x-hello": "ignored"}}, "paths": {}}`,
			FixedInstance: `{"swagger": "2.0", "info": {"title": "test", "version": "1.0", "contact": {"name": "bill"}}, "paths": {}}`,
		},
	}

	for _, v := range patterns {
		t.Run(v.Name, func(t *testing.T) {
			bad, err := openapi_v2.ParseDocument([]byte(v.BadInstance))
			if err != nil {
				t.Error(err)
				return
			}

			fixed, err := openapi_v2.ParseDocument([]byte(v.FixedInstance))
			if err != nil {
				t.Error(err)
				return
			}

			badConverted := Swagger{}
			if ok, err := badConverted.FromGnostic(bad); err != nil {
				t.Error(err)
				return
			} else if ok {
				t.Errorf("expected test to have data loss")
				return
			}

			fixedConverted := Swagger{}
			if ok, err := fixedConverted.FromGnostic(fixed); err != nil {
				t.Error(err)
				return
			} else if !ok {
				t.Errorf("expected fixed test to not have data loss")
				return
			}

			// Convert JSON directly into our kube-openapi type and check that
			// it is exactly equal to the converted instance
			fixedDirect := Swagger{}
			if err := json.Unmarshal([]byte(v.FixedInstance), &fixedDirect); err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(fixedConverted, badConverted) {
				t.Errorf("expected equal documents: %v", cmp.Diff(fixedConverted, badConverted, SpecV2DiffOptions...))
				return
			}

			// Make sure that they were exactly the same, except for the data loss
			//	by checking JSON encodes the some
			badConvertedJSON, err := json.Marshal(badConverted)
			if err != nil {
				t.Error(err)
				return
			}

			fixedConvertedJSON, err := json.Marshal(fixedConverted)
			if err != nil {
				t.Error(err)
				return
			}

			fixedDirectJSON, err := json.Marshal(fixedDirect)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(badConvertedJSON, fixedConvertedJSON) {
				t.Errorf("encoded json values for bad and fixed tests are not identical: %v", cmp.Diff(string(badConvertedJSON), string(fixedConvertedJSON)))
			}

			if !reflect.DeepEqual(fixedDirectJSON, fixedConvertedJSON) {
				t.Errorf("encoded json values for fixed direct and fixed-from-gnostic tests are not identical: %v", cmp.Diff(string(fixedDirectJSON), string(fixedConvertedJSON)))
			}
		})
	}
}

func TestBadStatusCode(t *testing.T) {
	const testCase = `{"swagger": "2.0", "info": {"title": "test", "version": "1.0"}, "paths": {"/": {"get": {"responses" : { "default": { "$ref": "#/definitions/a" }, "200": { "$ref": "#/definitions/b" }}}}}}`
	const dropped = `{"swagger": "2.0", "info": {"title": "test", "version": "1.0"}, "paths": {"/": {"get": {"responses" : { "200": { "$ref": "#/definitions/b" }}}}}}`
	gnosticInstance, err := openapi_v2.ParseDocument([]byte(testCase))
	if err != nil {
		t.Fatal(err)
	}

	droppedGnosticInstance, err := openapi_v2.ParseDocument([]byte(dropped))
	if err != nil {
		t.Fatal(err)
	}

	// Manually poke an response code name which gnostic's json parser would not allow
	gnosticInstance.Paths.Path[0].Value.Get.Responses.ResponseCode[0].Name = "bad"

	badConverted := Swagger{}
	droppedConverted := Swagger{}

	if ok, err := badConverted.FromGnostic(gnosticInstance); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatalf("expected data loss converting an operation with a response code 'bad'")
	}

	if ok, err := droppedConverted.FromGnostic(droppedGnosticInstance); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("expected no data loss converting a known good operation")
	}

	// Make sure that they were exactly the same, except for the data loss
	//	by checking JSON encodes the some
	badConvertedJSON, err := json.Marshal(badConverted)
	if err != nil {
		t.Error(err)
		return
	}

	droppedConvertedJSON, err := json.Marshal(droppedConverted)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(badConvertedJSON, droppedConvertedJSON) {
		t.Errorf("encoded json values for bad and fixed tests are not identical: %v", cmp.Diff(string(badConvertedJSON), string(droppedConvertedJSON)))
	}
}
