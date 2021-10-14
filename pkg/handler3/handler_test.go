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

package handler3

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"encoding/json"
	"k8s.io/kube-openapi/pkg/spec3"
)

var returnedGroupVersionListJSON = []byte(`{"Paths":["apis/apps/v1"]}`)

var returnedOpenAPI = []byte(`{
  "openapi": "3.0",
  "info": {
   "title": "Kubernetes",
   "version": "v1.23.0"
  },
  "paths": {}}`)

func TestRegisterOpenAPIVersionedService(t *testing.T) {
	var s *spec3.OpenAPI
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, returnedOpenAPI); err != nil {
		t.Errorf("%v", err)
	}
	compactOpenAPI := buffer.Bytes()

	json.Unmarshal(compactOpenAPI, &s)

	returnedJSON, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedJSON: %v", err)
	}

	returnedPb, err := ToV3ProtoBinary(compactOpenAPI)
	_ = returnedPb

	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedPb: %v", err)
	}

	mux := http.NewServeMux()
	o, err := NewOpenAPIService(nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.Handle("/openapi/v3", http.HandlerFunc(o.HandleDiscovery))
	mux.Handle("/openapi/v3/apis/apps/v1", http.HandlerFunc(o.HandleGroupVersion))

	o.UpdateGroupVersion("apis/apps/v1", s)

	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()

	tcs := []struct {
		acceptHeader string
		respStatus   int
		urlPath      string
		respBody     []byte
	}{
		{"", 200, "openapi/v3", returnedGroupVersionListJSON},
		{"", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"*/*", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"application/json", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"application/*", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"test/test", 406, "openapi/v3/apis/apps/v1", []byte{}},
		{"application/test", 406, "openapi/v3/apis/apps/v1", []byte{}},
		{"application/test, */*", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"application/com.github.proto-openapi.spec.v3@v1.0+protobuf", 200, "openapi/v3/apis/apps/v1", returnedPb},
		{"application/json, application/com.github.proto-openapi.spec.v2@v1.0+protobuf", 200, "openapi/v3/apis/apps/v1", returnedJSON},
		{"application/com.github.proto-openapi.spec.v3@v1.0+protobuf, application/json", 200, "openapi/v3/apis/apps/v1", returnedPb},
		{"application/com.github.proto-openapi.spec.v3@v1.0+protobuf; q=0.5, application/json", 200, "openapi/v3/apis/apps/v1", returnedJSON},
	}

	for _, tc := range tcs {
		req, err := http.NewRequest("GET", server.URL+"/"+tc.urlPath, nil)
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
