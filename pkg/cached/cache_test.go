/*
Copyright 2023 The Kubernetes Authors.

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

package cached_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"k8s.io/kube-openapi/pkg/cached"
)

func TestSource(t *testing.T) {
	count := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		count += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected function called twice, called: %v", count)
	}
}

func TestSourceError(t *testing.T) {
	count := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		count += 1
		return cached.NewResultErr[[]byte](errors.New("source error"))
	})
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if count != 2 {
		t.Fatalf("Expected function called twice, called: %v", count)
	}
}

func TestSourceAlternate(t *testing.T) {
	count := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		count += 1
		if count%2 == 0 {
			return cached.NewResultErr[[]byte](errors.New("source error"))
		} else {
			return cached.NewResultOK([]byte("source"), "source")
		}
	})
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if count != 4 {
		t.Fatalf("Expected function called 4x, called: %v", count)
	}
}

func TestStaticSource(t *testing.T) {
	count := 0
	source := cached.NewStaticSource(func() cached.Result[[]byte] {
		count += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected function called once, called: %v", count)
	}
}

func TestStaticSourceError(t *testing.T) {
	count := 0
	source := cached.NewStaticSource(func() cached.Result[[]byte] {
		count += 1
		return cached.NewResultErr[[]byte](errors.New("source error"))
	})
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if count != 1 {
		t.Fatalf("Expected function called once, called: %v", count)
	}
}

func TestResultData(t *testing.T) {
	source := cached.NewResultOK([]byte("source"), "source")
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := source.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResultError(t *testing.T) {
	source := cached.NewResultErr[[]byte](errors.New("source error"))
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	if err := source.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
}

func TestTransformer(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	transformerCount := 0
	transformer := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformerCount += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), "transformed "+result.Etag)
	}, source)
	if err := transformer.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := transformer.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sourceCount != 2 {
		t.Fatalf("Expected source function called twice, called: %v", sourceCount)
	}
	if transformerCount != 1 {
		t.Fatalf("Expected transformer function called once, called: %v", transformerCount)
	}
}

func TestTransformerChained(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	transformer1Count := 0
	transformer1 := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformer1Count += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), result.Etag)
	}, source)
	transformer2Count := 0
	transformer2 := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformer2Count += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), result.Etag)
	}, transformer1)
	transformer3Count := 0
	transformer3 := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformer3Count += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), result.Etag)
	}, transformer2)
	if err := transformer3.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := transformer3.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed transformed transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if sourceCount != 2 {
		t.Fatalf("Expected source function called twice, called: %v", sourceCount)
	}
	if transformer1Count != 1 {
		t.Fatalf("Expected transformer function called once, called: %v", transformer1Count)
	}
	if transformer2Count != 1 {
		t.Fatalf("Expected transformer function called once, called: %v", transformer2Count)
	}
	if transformer3Count != 1 {
		t.Fatalf("Expected transformer function called once, called: %v", transformer3Count)
	}
}

func TestTransformerError(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	transformerCount := 0
	transformer := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformerCount += 1
		return cached.NewResultErr[[]byte](errors.New("transformer error"))
	}, source)
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if sourceCount != 2 {
		t.Fatalf("Expected source function called twice, called: %v", sourceCount)
	}
	if transformerCount != 2 {
		t.Fatalf("Expected transformer function called twice, called: %v", transformerCount)
	}
}

func TestTransformerSourceError(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		return cached.NewResultErr[[]byte](errors.New("source error"))
	})
	transformerCount := 0
	transformer := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformerCount += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), "transformed "+result.Etag)
	}, source)
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if sourceCount != 2 {
		t.Fatalf("Expected source function called twice, called: %v", sourceCount)
	}
	if transformerCount != 2 {
		t.Fatalf("Expected transformer function called twice, called: %v", transformerCount)
	}
}

func TestTransformerAlternateSourceError(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		if sourceCount%2 == 0 {
			return cached.NewResultErr[[]byte](errors.New("source error"))
		} else {
			return cached.NewResultOK([]byte("source"), "source")
		}
	})
	transformerCount := 0
	transformer := cached.NewTransformer(func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformerCount += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), "transformed "+result.Etag)
	}, source)
	result := transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "transformed source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "transformed source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if err := transformer.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if sourceCount != 4 {
		t.Fatalf("Expected source function called 4x, called: %v", sourceCount)
	}
	if transformerCount != 4 {
		t.Fatalf("Expected transformer function called 4x, called: %v", transformerCount)
	}

}

func TestMerger(t *testing.T) {
	source1Count := 0
	source1 := cached.NewSource(func() cached.Result[[]byte] {
		source1Count += 1
		return cached.NewResultOK([]byte("source1"), "source1")
	})
	source2Count := 0
	source2 := cached.NewSource(func() cached.Result[[]byte] {
		source2Count += 1
		return cached.NewResultOK([]byte("source2"), "source2")
	})
	mergerCount := 0
	merger := cached.NewMerger(func(results map[string]cached.Result[[]byte]) cached.Result[[]byte] {
		mergerCount += 1
		d := []string{}
		e := []string{}
		for _, result := range results {
			if result.Err != nil {
				return cached.NewResultErr[[]byte](result.Err)
			}
			d = append(d, string(result.Data))
			e = append(e, result.Etag)
		}
		sort.Strings(d)
		sort.Strings(e)
		return cached.NewResultOK([]byte("merged "+strings.Join(d, " and ")), "merged "+strings.Join(e, " and "))
	}, map[string]cached.Data[[]byte]{
		"source1": source1,
		"source2": source2,
	})
	if err := merger.Get().Err; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := merger.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "merged source1 and source2"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "merged source1 and source2"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}

	if source1Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source1Count)
	}
	if source2Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source2Count)
	}
	if mergerCount != 1 {
		t.Fatalf("Expected merger function called once, called: %v", mergerCount)
	}
}

func TestMergerError(t *testing.T) {
	source1Count := 0
	source1 := cached.NewSource(func() cached.Result[[]byte] {
		source1Count += 1
		return cached.NewResultOK([]byte("source1"), "source1")
	})
	source2Count := 0
	source2 := cached.NewSource(func() cached.Result[[]byte] {
		source2Count += 1
		return cached.NewResultOK([]byte("source2"), "source2")
	})
	mergerCount := 0
	merger := cached.NewMerger(func(results map[string]cached.Result[[]byte]) cached.Result[[]byte] {
		mergerCount += 1
		return cached.NewResultErr[[]byte](errors.New("merger error"))
	}, map[string]cached.Data[[]byte]{
		"source1": source1,
		"source2": source2,
	})
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if source1Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source1Count)
	}
	if source2Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source2Count)
	}
	if mergerCount != 2 {
		t.Fatalf("Expected merger function called twice, called: %v", mergerCount)
	}
}

func TestMergerSourceError(t *testing.T) {
	source1Count := 0
	source1 := cached.NewSource(func() cached.Result[[]byte] {
		source1Count += 1
		return cached.NewResultErr[[]byte](errors.New("source1 error"))
	})
	source2Count := 0
	source2 := cached.NewSource(func() cached.Result[[]byte] {
		source2Count += 1
		return cached.NewResultOK([]byte("source2"), "source2")
	})
	mergerCount := 0
	merger := cached.NewMerger(func(results map[string]cached.Result[[]byte]) cached.Result[[]byte] {
		mergerCount += 1
		d := []string{}
		e := []string{}
		for _, result := range results {
			if result.Err != nil {
				return cached.NewResultErr[[]byte](result.Err)
			}
			d = append(d, string(result.Data))
			e = append(e, result.Etag)
		}
		sort.Strings(d)
		sort.Strings(e)
		return cached.NewResultOK([]byte("merged "+strings.Join(d, " and ")), "merged "+strings.Join(e, " and "))
	}, map[string]cached.Data[[]byte]{
		"source1": source1,
		"source2": source2,
	})
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if source1Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source1Count)
	}
	if source2Count != 2 {
		t.Fatalf("Expected source function called twice, called: %v", source2Count)
	}
	if mergerCount != 2 {
		t.Fatalf("Expected merger function called twice, called: %v", mergerCount)
	}
}

func TestMergerAlternateSourceError(t *testing.T) {
	source1Count := 0
	source1 := cached.NewSource(func() cached.Result[[]byte] {
		source1Count += 1
		if source1Count%2 == 0 {
			return cached.NewResultErr[[]byte](errors.New("source1 error"))
		} else {
			return cached.NewResultOK([]byte("source1"), "source1")
		}
	})
	source2Count := 0
	source2 := cached.NewSource(func() cached.Result[[]byte] {
		source2Count += 1
		return cached.NewResultOK([]byte("source2"), "source2")
	})
	mergerCount := 0
	merger := cached.NewMerger(func(results map[string]cached.Result[[]byte]) cached.Result[[]byte] {
		mergerCount += 1
		d := []string{}
		e := []string{}
		for _, result := range results {
			if result.Err != nil {
				return cached.NewResultErr[[]byte](result.Err)
			}
			d = append(d, string(result.Data))
			e = append(e, result.Etag)
		}
		sort.Strings(d)
		sort.Strings(e)
		return cached.NewResultOK([]byte("merged "+strings.Join(d, " and ")), "merged "+strings.Join(e, " and "))
	}, map[string]cached.Data[[]byte]{
		"source1": source1,
		"source2": source2,
	})
	result := merger.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "merged source1 and source2"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "merged source1 and source2"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	result = merger.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "merged source1 and source2"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "merged source1 and source2"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if err := merger.Get().Err; err == nil {
		t.Fatalf("expected error, none found")
	}
	if source1Count != 4 {
		t.Fatalf("Expected source function called 4x, called: %v", source1Count)
	}
	if source2Count != 4 {
		t.Fatalf("Expected source function called 4x, called: %v", source2Count)
	}
	if mergerCount != 4 {
		t.Fatalf("Expected merger function called 4x, called: %v", mergerCount)
	}
}

func TestReplaceable(t *testing.T) {
	sourceDataCount := 0
	sourceData := cached.NewSource(func() cached.Result[[]byte] {
		sourceDataCount += 1
		return cached.NewResultOK([]byte("source"), "source")
	})
	sourceData2Count := 0
	sourceData2 := cached.NewSource(func() cached.Result[[]byte] {
		sourceData2Count += 1
		return cached.NewResultOK([]byte("source2"), "source2")
	})

	sourceErrCount := 0
	sourceErr := cached.NewSource(func() cached.Result[[]byte] {
		sourceErrCount += 1
		return cached.NewResultErr[[]byte](errors.New("source error"))
	})
	replaceable := cached.Replaceable[[]byte]{}
	replaceable.Replace(sourceErr)
	if err := replaceable.Get().Err; err == nil {
		t.Fatalf("expected error, found none")
	}
	replaceable.Replace(sourceData)
	result := replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	// replace with the same thing, shouldn't change anything
	replaceable.Replace(sourceData)
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	// Even if we replace with something that fails, we continue to return the success.
	replaceable.Replace(sourceErr)
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	replaceable.Replace(sourceData2)
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source2"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source2"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if sourceDataCount != 2 {
		t.Fatalf("Expected sourceData function called twice, called: %v", sourceDataCount)
	}
	if sourceData2Count != 1 {
		t.Fatalf("Expected sourceData2 function called once, called: %v", sourceData2Count)
	}
	if sourceErrCount != 2 {
		t.Fatalf("Expected error source function called once, called: %v", sourceErrCount)
	}
}

func TestReplaceableAlternateError(t *testing.T) {
	sourceCount := 0
	source := cached.NewSource(func() cached.Result[[]byte] {
		sourceCount += 1
		if sourceCount%2 == 0 {
			return cached.NewResultErr[[]byte](errors.New("source error"))
		} else {
			return cached.NewResultOK([]byte("source"), "source")
		}
	})
	replaceable := cached.Replaceable[[]byte]{}
	replaceable.Replace(source)
	result := replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	result = replaceable.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	if sourceCount != 4 {
		t.Fatalf("Expected sourceData function called 4x, called: %v", sourceCount)
	}
}

func TestReplaceableWithTransformer(t *testing.T) {
	replaceable := cached.Replaceable[[]byte]{}
	replaceable.Replace(cached.NewResultOK([]byte("source"), "source"))
	transformerCount := 0
	transformer := cached.NewTransformer[[]byte](func(result cached.Result[[]byte]) cached.Result[[]byte] {
		transformerCount += 1
		if result.Err != nil {
			return cached.NewResultErr[[]byte](result.Err)
		}
		return cached.NewResultOK([]byte("transformed "+string(result.Data)), "transformed "+result.Etag)
	}, &replaceable)
	result := transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "transformed source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	// replace with new cache, transformer shouldn't be affected (or called)
	replaceable.Replace(cached.NewResultOK([]byte("source"), "source"))
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "transformed source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}
	// replace with failing cache, transformer should still not be affected (or called)
	replaceable.Replace(cached.NewResultErr[[]byte](errors.New("source error")))
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	result = transformer.Get()
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if want := "transformed source"; string(result.Data) != want {
		t.Fatalf("expected data = %v, got %v", want, string(result.Data))
	}
	if want := "transformed source"; result.Etag != want {
		t.Fatalf("expected etag = %v, got %v", want, result.Etag)
	}

	if transformerCount != 1 {
		t.Fatalf("Expected transformer function called once, called: %v", transformerCount)
	}
}

// Here is an example of how one can write a cache that will constantly
// be pulled, while actually recomputing the results only as needed.
func Example() {
	// Merge Json is a replaceable cache, since we'll want it to
	// change a few times.
	mergeJson := cached.Replaceable[[]byte]{}

	one := cached.NewStaticSource(func() cached.Result[[]byte] {
		// This one is computed lazily, only when requested, and only once.
		return cached.NewResultOK([]byte("one"), "one")
	})
	two := cached.NewSource(func() cached.Result[[]byte] {
		// This cache is re-computed every time.
		return cached.NewResultOK([]byte("two"), "two")
	})
	// This cache is computed once, and is not lazy at all.
	three := cached.NewResultOK([]byte("three"), "three")

	// This cache will allow us to replace a branch of the tree
	// efficiently.
	replaceable := cached.Replaceable[[]byte]{}
	replaceable.Replace(cached.NewResultOK([]byte("four"), "four"))

	merger := func(results map[string]cached.Result[[]byte]) cached.Result[[]byte] {
		var out = []json.RawMessage{}
		var resultEtag string
		for _, result := range results {
			if result.Err != nil {
				return cached.NewResultErr[[]byte](result.Err)
			}
			resultEtag += result.Etag
			out = append(out, result.Data)
		}
		data, err := json.Marshal(out)
		if err != nil {
			return cached.NewResultErr[[]byte](err)
		}
		return cached.NewResultOK(data, resultEtag)
	}

	mergeJson.Replace(cached.NewMerger(merger, map[string]cached.Data[[]byte]{
		"one":         one,
		"two":         two,
		"three":       three,
		"replaceable": &replaceable,
	}))

	// Create a new cache that indents a buffer. This should only be
	// called if the buffer has changed.
	indented := cached.NewTransformer[[]byte](func(js cached.Result[[]byte]) cached.Result[[]byte] {
		// Get the json from the previous layer of cache, before
		// we indent.
		if js.Err != nil {
			return cached.NewResultErr[[]byte](js.Err)
		}
		var out bytes.Buffer
		json.Indent(&out, js.Data, "", "\t")
		return cached.NewResultOK(out.Bytes(), js.Etag)
	}, &mergeJson)

	// We have "clients" that constantly pulls the indented format.
	go func() {
		for {
			if err := indented.Get().Err; err != nil {
				panic(fmt.Errorf("invalid error: %v", err))
			}
		}
	}()

	failure := cached.NewResultErr[[]byte](errors.New("Invalid cache!"))
	// Insert a new sub-cache that fails, it should just be ignored.
	mergeJson.Replace(cached.NewMerger(merger, map[string]cached.Data[[]byte]{
		"one":         one,
		"two":         two,
		"three":       three,
		"replaceable": &replaceable,
		"failure":     failure,
	}))

	// We can replace just a branch of the dependency tree.
	replaceable.Replace(cached.NewResultOK([]byte("five"), "five"))

	// We can replace to remove the failure and one of the sub-cached.
	mergeJson.Replace(cached.NewMerger(merger, map[string]cached.Data[[]byte]{
		"one":         one,
		"two":         two,
		"replaceable": &replaceable,
	}))
}
