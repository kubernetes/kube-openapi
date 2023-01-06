package handler

import (
	json "encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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
		acceptHeader        string
		respStatus          int
		respBody            []byte
		expectedContentType string
	}{
		{"", 200, returnedJSON, "application/json"},
		{"*/*", 200, returnedJSON, "application/json"},
		{"application/*", 200, returnedJSON, "application/json"},
		{"application/json", 200, returnedJSON, "application/json"},
		{"test/test", 406, []byte{}, "application/json"},
		{"application/test", 406, []byte{}, "application/json"},
		{"application/test, */*", 200, returnedJSON, "application/json"},
		{"application/test, application/json", 200, returnedJSON, "application/json"},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf", 200, returnedPb, "application/com.github.proto-openapi.spec.v2@v1.0+protobuf"},
		{"application/json, application/com.github.proto-openapi.spec.v2@v1.0+protobuf", 200, returnedJSON, "application/json"},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf, application/json", 200, returnedPb, "application/com.github.proto-openapi.spec.v2@v1.0+protobuf"},
		{"application/com.github.proto-openapi.spec.v2@v1.0+protobuf; q=0.5, application/json", 200, returnedJSON, "application/json"},
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
		if resp.StatusCode != 200 {
			continue
		}

		if resp.Header.Get("Content-Type") != tc.expectedContentType {
			t.Errorf("Content-Type: Unexpected content type, want: %v, got: %v", tc.expectedContentType, resp.Header.Get("Content-Type"))
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
