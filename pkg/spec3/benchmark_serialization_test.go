package spec3

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/kube-openapi/pkg/internal"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// cmp.Diff panics when reflecting unexported fields under jsonreference.Ref
// a custom comparator is required
var swaggerDiffOptions = []cmp.Option{cmp.Comparer(func(a spec.Ref, b spec.Ref) bool {
	return a.String() == b.String()
})}

func TestOpenAPIV3Deserialize(t *testing.T) {
	swagFile, err := os.Open("./testdata/appsv1spec.json")
	if err != nil {
		t.Fatal(err)
	}
	defer swagFile.Close()
	originalJSON, err := io.ReadAll(swagFile)
	if err != nil {
		t.Fatal(err)
	}
	internal.UseOptimizedJSONUnmarshalingV3 = false

	var result1 *OpenAPI

	if err := json.Unmarshal(originalJSON, &result1); err != nil {
		t.Fatal(err)
	}
	internal.UseOptimizedJSONUnmarshalingV3 = true
	var result2 *OpenAPI
	if err := json.Unmarshal(originalJSON, &result2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result1, result2) {
		t.Fatal(cmp.Diff(result1, result2, swaggerDiffOptions...))
	}
}

func BenchmarkOpenAPIV3Deserialize(b *testing.B) {
	benchcases := []struct {
		file string
	}{
		{
			file: "appsv1spec.json",
		},
		{
			file: "authorizationv1spec.json",
		},
	}
	for _, bc := range benchcases {
		swagFile, err := os.Open("./testdata/" + bc.file)
		if err != nil {
			b.Fatal(err)
		}
		defer swagFile.Close()
		originalJSON, err := io.ReadAll(swagFile)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		b.Run(fmt.Sprintf("%s jsonv1", bc.file), func(b2 *testing.B) {
			b2.ReportAllocs()
			internal.UseOptimizedJSONUnmarshaling = false
			internal.UseOptimizedJSONUnmarshalingV3 = false
			for i := 0; i < b2.N; i++ {
				var result *OpenAPI
				if err := json.Unmarshal(originalJSON, &result); err != nil {
					b2.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("%s jsonv2 via jsonv1 schema only", bc.file), func(b2 *testing.B) {
			b2.ReportAllocs()
			internal.UseOptimizedJSONUnmarshaling = true
			internal.UseOptimizedJSONUnmarshalingV3 = false
			for i := 0; i < b2.N; i++ {
				var result *OpenAPI
				if err := json.Unmarshal(originalJSON, &result); err != nil {
					b2.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("%s jsonv2 via jsonv1 full spec", bc.file), func(b2 *testing.B) {
			b2.ReportAllocs()
			internal.UseOptimizedJSONUnmarshaling = true
			internal.UseOptimizedJSONUnmarshalingV3 = true
			for i := 0; i < b2.N; i++ {
				var result *OpenAPI
				if err := json.Unmarshal(originalJSON, &result); err != nil {
					b2.Fatal(err)
				}
			}
		})

		b.Run("jsonv2", func(b2 *testing.B) {
			b2.ReportAllocs()
			internal.UseOptimizedJSONUnmarshaling = true
			internal.UseOptimizedJSONUnmarshalingV3 = true
			for i := 0; i < b2.N; i++ {
				var result *OpenAPI
				if err := result.CustomUnmarshalJSON(originalJSON); err != nil {
					b2.Fatal(err)
				}
			}
		})
	}
}
