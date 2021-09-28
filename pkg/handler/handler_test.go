package handler

import (
	json "encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

var returnedSwagger = []byte(`{
  "swagger": "2.0",
  "info": {
   "title": "Kubernetes",
   "version": "v1.11.0"
  }}`)

func TestRegisterOpenAPIVersionedService(t *testing.T) {
	var s spec.Swagger
	err := s.UnmarshalJSON(returnedSwagger)
	if err != nil {
		t.Errorf("Unexpected error in unmarshalling SwaggerJSON: %v", err)
	}

	returnedJSON, err := json.Marshal(s)
	if err != nil {
		t.Errorf("Unexpected error in preparing returnedJSON: %v", err)
	}
	var decodedJSON map[string]interface{}
	if err := json.Unmarshal(returnedJSON, &decodedJSON); err != nil {
		t.Fatal(err)
	}
	returnedPb, err := ToProtoBinary(returnedJSON)
	if err != nil {
		t.Errorf("Unexpected error in preparing returnedPb: %v", err)
	}

	mux := http.NewServeMux()
	o, err := NewOpenAPIService(&s)
	if err != nil {
		t.Fatal(err)
	}
	if err = o.RegisterOpenAPIVersionedService("/openapi/v2", mux); err != nil {
		t.Errorf("Unexpected error in register OpenAPI versioned service: %v", err)
	}
	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()

	tcs := []struct {
		acceptHeader string
		respStatus   int
		respBody     []byte
	}{
		{"", 200, returnedJSON},
		{"*/*", 200, returnedJSON},
		{"application/*", 200, returnedJSON},
		{"application/json", 200, returnedJSON},
		{"test/test", 406, []byte{}},
		{"application/test", 406, []byte{}},
		{"application/test, */*", 200, returnedJSON},
		{"application/test, application/json", 200, returnedJSON},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf", 200, returnedPb},
		{"application/json, application/com.github.proto-openapi.spec.v2@v1.0+protobuf", 200, returnedJSON},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf, application/json", 200, returnedPb},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf; q=0.5, application/json", 200, returnedJSON},
	}

	for _, tc := range tcs {
		req, err := http.NewRequest("GET", server.URL+"/openapi/v2", nil)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in creating new request: %v", tc.acceptHeader, err)
		}

		req.Header.Add("Accept", tc.acceptHeader)
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in serving HTTP request: %v", tc.acceptHeader, err)
		}

		if resp.StatusCode != tc.respStatus {
			t.Errorf("Accept: %v: Unexpected response status code, want: %v, got: %v", tc.acceptHeader, tc.respStatus, resp.StatusCode)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in reading response body: %v", tc.acceptHeader, err)
		}
		if !reflect.DeepEqual(body, tc.respBody) {
			t.Errorf("Accept: %v: Response body mismatches, \nwant: %s, \ngot:  %s", tc.acceptHeader, string(tc.respBody), string(body))
		}
	}
}

func TestJsonToYAML(t *testing.T) {
	intOrInt64 := func(i64 int64) interface{} {
		if i := int(i64); i64 == int64(i) {
			return i
		}
		return i64
	}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected yaml.MapSlice
	}{
		{"nil", nil, nil},
		{"empty", map[string]interface{}{}, yaml.MapSlice{}},
		{
			"values",
			map[string]interface{}{
				"bool":         true,
				"float64":      float64(42.1),
				"fractionless": float64(42),
				"int":          int(42),
				"int64":        int64(42),
				"int64 big":    float64(math.Pow(2, 62)),
				"map":          map[string]interface{}{"foo": "bar"},
				"slice":        []interface{}{"foo", "bar"},
				"string":       string("foo"),
				"uint64 big":   float64(math.Pow(2, 63)),
			},
			yaml.MapSlice{
				{"bool", true},
				{"float64", float64(42.1)},
				{"fractionless", int(42)},
				{"int", int(42)},
				{"int64", int(42)},
				{"int64 big", intOrInt64(int64(1) << 62)},
				{"map", yaml.MapSlice{{"foo", "bar"}}},
				{"slice", []interface{}{"foo", "bar"}},
				{"string", string("foo")},
				{"uint64 big", uint64(1) << 63},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonToYAML(tt.input)
			sortMapSlicesInPlace(tt.expected)
			sortMapSlicesInPlace(got)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("jsonToYAML() = %v, want %v", spew.Sdump(got), spew.Sdump(tt.expected))
			}
		})
	}
}

func sortMapSlicesInPlace(x interface{}) {
	switch x := x.(type) {
	case []interface{}:
		for i := range x {
			sortMapSlicesInPlace(x[i])
		}
	case yaml.MapSlice:
		sort.Slice(x, func(a, b int) bool {
			return x[a].Key.(string) < x[b].Key.(string)
		})
	}
}

func TestToProtoBinary(t *testing.T) {
	bs, err := ioutil.ReadFile("../../test/integration/testdata/aggregator/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ToProtoBinary(bs); err != nil {
		t.Fatal()
	}
	// TODO: add some kind of roundtrip test here
}

func TestCache(t *testing.T) {
	calledCount := 0
	expectedBytes := []byte("ABC")
	cacheObj := cache{
		BuildCache: func() ([]byte, error) {
			calledCount++
			return expectedBytes, nil
		},
	}
	bytes, _, _ := cacheObj.Get()
	if string(bytes) != string(expectedBytes) {
		t.Fatalf("got value of %q from cache (expected %q)", bytes, expectedBytes)
	}
	cacheObj.Get()
	if calledCount != 1 {
		t.Fatalf("expected BuildCache to be called once (called %d times)", calledCount)
	}
}

func TestCacheError(t *testing.T) {
	cacheObj := cache{
		BuildCache: func() ([]byte, error) {
			return nil, errors.New("cache error")
		},
	}
	_, _, err := cacheObj.Get()
	if err == nil {
		t.Fatalf("expected non-nil err from cache.Get()")
	}
}

func TestCacheRefresh(t *testing.T) {
	// check that returning an error while having no prior cached value results in a nil value from cache.Get()
	cacheObj := (&cache{}).New(func() ([]byte, error) {
		return nil, errors.New("returning nil bytes")
	})
	// make multiple calls to Get() to ensure errors are preserved across subsequent calls
	for i := 0; i < 4; i++ {
		value, _, err := cacheObj.Get()
		if value != nil {
			t.Fatalf("expected nil bytes (got %s)", value)
		}
		if err == nil {
			t.Fatalf("expected non-nil err from cache.Get()")
		}
	}
	// check that we can call New() multiple times and get the last known cache value while also returning any errors
	lastGoodVal := []byte("last good value")
	cacheObj = cacheObj.New(func() ([]byte, error) {
		return lastGoodVal, nil
	})
	// call Get() once, so lastGoodVal is cached
	_, lastGoodEtag, _ := cacheObj.Get()
	for i := 0; i < 4; i++ {
		cacheObj = cacheObj.New(func() ([]byte, error) {
			return nil, errors.New("check that c.bytes is preserved across New() calls")
		})
		value, newEtag, err := cacheObj.Get()
		if err == nil {
			t.Fatalf("expected non-nil err from cache.Get()")
		}
		if string(value) != string(lastGoodVal) {
			t.Fatalf("expected previous value for cache to be returned (got %s, expected %s)", value, lastGoodVal)
		}
		// check that etags carry over between calls to cache.New()
		if lastGoodEtag != newEtag {
			t.Fatalf("expected etags to match (got %s, expected %s", newEtag, lastGoodEtag)
		}
	}
	// check that if we successfully renew the cache the old last known value is flushed
	newVal := []byte("new good value")
	cacheObj = cacheObj.New(func() ([]byte, error) {
		return newVal, nil
	})
	value, _, err := cacheObj.Get()
	if err != nil {
		t.Fatalf("expected nil err from cache.Get()")
	}
	if string(value) != string(newVal) {
		t.Fatalf("got value of %s from cache (expected %s)", value, newVal)
	}
}
