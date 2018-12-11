/*
Copyright 2019 The Kubernetes Authors.

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

package aggregator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	fuzz "github.com/google/gofuzz"
	"k8s.io/kube-openapi/pkg/util/sets"
)

func TestReplaceReferences(t *testing.T) {
	re, err := regexp.Compile("\"\\$ref\":\"(http://ref-[^\"]*)\"")
	if err != nil {
		t.Fatalf("failed to compile ref regex: %v", err)
	}

	for i := 0; i < 100; i++ {
		f := fuzz.New()
		seed := time.Now().UnixNano()
		//seed = int64(1548861570869173000)
		f.RandSource(rand.New(rand.NewSource(seed)))

		refNum := 0
		f.Funcs(
			func(ref *spec.Ref, c fuzz.Continue) {
				r, err := spec.NewRef(fmt.Sprintf("http://ref-%d", refNum))
				if err != nil {
					t.Fatalf("failed to fuzz ref: %v", err)
				}
				*ref = r
				refNum++
			},
			func(sa *spec.SchemaOrStringArray, c fuzz.Continue) {
				*sa = spec.SchemaOrStringArray{}
				if c.RandBool() {
					c.Fuzz(&sa.Schema)
				} else {
					c.Fuzz(&sa.Property)
				}
				if sa.Schema == nil && len(sa.Property) == 0 {
					*sa = spec.SchemaOrStringArray{Schema: &spec.Schema{}}
				}
			},
			func(url *spec.SchemaURL, c fuzz.Continue) {
				*url = spec.SchemaURL("http://url")
			},
			func(s *spec.Swagger, c fuzz.Continue) {
				// only fuzz those fields we walk into (we skip some like extensions)
				c.Fuzz(&s.Parameters)
				c.Fuzz(&s.Responses)
				c.Fuzz(&s.Definitions)
				c.Fuzz(&s.Paths)
			},
			func(p *spec.Parameter, c fuzz.Continue) {
				// only fuzz those fields we walk into
				c.Fuzz(&p.Ref)
				c.Fuzz(&p.Schema)
				if c.RandBool() {
					p.Items = &spec.Items{}
					c.Fuzz(&p.Items.Ref)
				}
			},
			func(s *spec.Response, c fuzz.Continue) {
				c.FuzzNoCustom(s)
				// we don't walk into headers
				s.Headers = nil
			},
			func(s *spec.Dependencies, c fuzz.Continue) {
				c.FuzzNoCustom(s)
				// we don't walk into dependencies
				*s = nil
			},
			func(i *interface{}, c fuzz.Continue) {
				// do nothing for examples and defaults.
			},
		)
		f.NilChance(0.92)
		f.NumElements(0, 3)

		fuzzedRefs := sets.NewString()
		var s *spec.Swagger
		for {
			// create random swagger spec with random URL references, but at least one ref
			s = &spec.Swagger{}
			f.Fuzz(s)

			// clone spec to normalize (fuzz might generate objects which do not roundtrip json marshalling
			var err error
			s, err = cloneSwagger(s)
			if err != nil {
				t.Fatalf("failed to normalize swagger after fuzzing: %v", err)
			}

			// find refs
			bs, err := json.Marshal(s)
			if err != nil {
				t.Fatalf("failed to marshal swagger: %v", err)
			}
			findings := re.FindAllStringSubmatch(string(bs), -1)
			if len(findings) > 0 {
				for _, m := range findings {
					fuzzedRefs.Insert(m[1])
				}
				break
			}
		}

		t.Run(fmt.Sprintf("iteration %d", i), func(t *testing.T) {
			mutatedRef := fuzzedRefs.List()[rand.Intn(fuzzedRefs.Len())]
			origString, err := json.Marshal(s)
			if err != nil {
				t.Fatalf("failed to marshal swagger: %v", err)
			}
			t.Logf("created schema with %d refs, mutating %q, seed %d: %s", fuzzedRefs.Len(), mutatedRef, seed, string(origString))

			// convert to json string, replace one of the refs, and unmarshal back
			mutatedString := strings.Replace(string(origString), "\""+mutatedRef+"\"", "\"http://mutated-ref\"", -1)
			mutatedViaJSON := &spec.Swagger{}
			if err := json.Unmarshal([]byte(mutatedString), mutatedViaJSON); err != nil {
				t.Fatalf("failed to unmarshal mutated spec: %v", err)
			}

			// replay the same mutation using the mutating walker
			seenRefs := sets.NewString()
			walker := mutatingReferenceWalker{
				walkRefCallback: func(ref *spec.Ref) *spec.Ref {
					seenRefs.Insert(ref.String())
					if ref.String() == mutatedRef {
						r, err := spec.NewRef("http://mutated-ref")
						if err != nil {
							t.Fatalf("failed to create ref: %v", err)
						}
						return &r
					}
					return ref
				},
			}
			mutatedViaWalker := walker.Start(s)

			// compare that we got the same
			if !reflect.DeepEqual(mutatedViaJSON, mutatedViaWalker) {
				t.Errorf("mutation via walker differ from JSON text replacement (got A, expected B): %s", objectDiff(mutatedViaWalker, mutatedViaJSON))
			}
			if !seenRefs.Equal(fuzzedRefs) {
				t.Errorf("expected to see the same refs in the walker as during fuzzing. Not seen: %v", fuzzedRefs.Difference(seenRefs).List())
			}
		})
	}
}

func cloneSwagger(orig *spec.Swagger) (*spec.Swagger, error) {
	bs, err := json.Marshal(orig)
	if err != nil {
		return nil, err
	}
	s := &spec.Swagger{}
	if err := json.Unmarshal(bs, s); err != nil {
		return nil, err
	}
	return s, nil
}

// stringDiff diffs a and b and returns a human readable diff.
func stringDiff(a, b string) string {
	ba := []byte(a)
	bb := []byte(b)
	out := []byte{}
	i := 0
	for ; i < len(ba) && i < len(bb); i++ {
		if ba[i] != bb[i] {
			break
		}
		out = append(out, ba[i])
	}
	out = append(out, []byte("\n\nA: ")...)
	out = append(out, ba[i:]...)
	out = append(out, []byte("\n\nB: ")...)
	out = append(out, bb[i:]...)
	out = append(out, []byte("\n\n")...)
	return string(out)
}

// objectDiff writes the two objects out as JSON and prints out the identical part of
// the objects followed by the remaining part of 'a' and finally the remaining part of 'b'.
// For debugging tests.
func objectDiff(a, b interface{}) string {
	ab, err := json.Marshal(a)
	if err != nil {
		panic(fmt.Sprintf("a: %v", err))
	}
	bb, err := json.Marshal(b)
	if err != nil {
		panic(fmt.Sprintf("b: %v", err))
	}
	return stringDiff(string(ab), string(bb))
}
