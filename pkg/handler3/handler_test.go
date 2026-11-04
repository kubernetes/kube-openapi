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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"k8s.io/kube-openapi/pkg/spec3"
)

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
	var hash = computeETag(compactOpenAPI)

	var returnedGroupVersionListJSON = []byte(`{"paths":{"apis/apps/v1":{"serverRelativeURL":"/openapi/v3/apis/apps/v1?hash=` + hash + `"}}}`)

	json.Unmarshal(compactOpenAPI, &s)

	returnedJSON, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedJSON: %v", err)
	}

	returnedPb, err := ToV3ProtoBinary(compactOpenAPI)

	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedPb: %v", err)
	}

	mux := http.NewServeMux()
	o := NewOpenAPIService()
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
		acceptHeader              string
		respStatus                int
		urlPath                   string
		respBody                  []byte
		expectedETag              string
		sendETag                  bool
		responseContentTypeHeader string
	}{
		{
			acceptHeader:              "",
			respStatus:                200,
			urlPath:                   "openapi/v3",
			respBody:                  returnedGroupVersionListJSON,
			expectedETag:              computeETag(returnedGroupVersionListJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader: "",
			respStatus:   304,
			urlPath:      "openapi/v3",
			respBody:     returnedGroupVersionListJSON,
			expectedETag: computeETag(returnedGroupVersionListJSON),
			sendETag:     true,
		}, {
			acceptHeader:              "",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader: "",
			respStatus:   304,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     returnedJSON,
			expectedETag: computeETag(returnedJSON),
			sendETag:     true,
		}, {
			acceptHeader:              "*/*",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader:              "application/json",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader:              "application/*",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader: "test/test",
			respStatus:   406,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     []byte{},
		}, {
			acceptHeader: "application/test",
			respStatus:   406,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     []byte{},
		}, {
			acceptHeader:              "application/test,  */*",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedPb,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
		}, {
			acceptHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
			respStatus:   304,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     returnedPb,
			expectedETag: computeETag(returnedJSON),
			sendETag:     true,
		}, {
			acceptHeader:              "application/json, application/com.github.proto-openapi.spec.v2.v1.0+protobuf",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3.v1.0+protobuf, application/json",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedPb,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
		}, {
			acceptHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf, application/json",
			respStatus:   304,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     returnedPb,
			expectedETag: computeETag(returnedJSON),
			sendETag:     true,
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3.v1.0+protobuf; q=0.5, application/json",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3@v1.0+protobuf",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedPb,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
		}, {
			acceptHeader: "application/com.github.proto-openapi.spec.v3@v1.0+protobuf",
			respStatus:   304,
			urlPath:      "openapi/v3/apis/apps/v1",
			respBody:     returnedPb,
			expectedETag: computeETag(returnedJSON),
			sendETag:     true,
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3@v1.0+protobuf, application/json",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedPb,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
		}, {
			acceptHeader:              "application/com.github.proto-openapi.spec.v3@v1.0+protobuf; q=0.5, application/json",
			respStatus:                200,
			urlPath:                   "openapi/v3/apis/apps/v1",
			respBody:                  returnedJSON,
			expectedETag:              computeETag(returnedJSON),
			responseContentTypeHeader: "application/json",
		},
	}

	for _, tc := range tcs {
		req, err := http.NewRequest("GET", server.URL+"/"+tc.urlPath, nil)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in creating new request: %v", tc.acceptHeader, err)
		}

		req.Header.Add("Accept", tc.acceptHeader)
		if tc.sendETag {
			req.Header.Add("If-None-Match", strconv.Quote(tc.expectedETag))
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in serving HTTP request: %v", tc.acceptHeader, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != tc.respStatus {
			t.Errorf("Accept: %v: Unexpected response status code, want: %v, got: %v", tc.acceptHeader, tc.respStatus, resp.StatusCode)
		}

		if tc.respStatus == 304 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Accept: %v: Unexpected error in reading response body: %v", tc.acceptHeader, err)
			}
			if len(body) != 0 {
				t.Errorf("Response Body length must be 0 if 304 is returned.")
			}
		}
		if tc.respStatus != 200 {
			continue
		}

		responseContentType := resp.Header.Get("Content-Type")
		if responseContentType != tc.responseContentTypeHeader {
			t.Errorf("Accept: %v: Unexpected content type in response, want: %v, got: %v", tc.acceptHeader, tc.responseContentTypeHeader, responseContentType)
		}
		_, _, err = mime.ParseMediaType(responseContentType)
		if err != nil {
			t.Errorf("Unexpected error in parsing response content type: %v, err: %v", responseContentType, err)
		}

		gotETag := resp.Header.Get("ETag")
		if strconv.Quote(tc.expectedETag) != gotETag {
			t.Errorf("Expect ETag %s, got %s", strconv.Quote(tc.expectedETag), gotETag)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in reading response body: %v", tc.acceptHeader, err)
		}
		if !reflect.DeepEqual(body, tc.respBody) {
			t.Errorf("Accept: %v: Response body mismatches, \nwant: %s, \ngot:  %s", tc.acceptHeader, string(tc.respBody), string(body))
		}
	}
}

func TestCacheBusting(t *testing.T) {
	var s *spec3.OpenAPI
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, returnedOpenAPI); err != nil {
		t.Errorf("%v", err)
	}
	compactOpenAPI := buffer.Bytes()
	var hash = computeETag(compactOpenAPI)

	json.Unmarshal(compactOpenAPI, &s)

	returnedJSON, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedJSON: %v", err)
	}

	returnedPb, err := ToV3ProtoBinary(compactOpenAPI)

	if err != nil {
		t.Fatalf("Unexpected error in preparing returnedPb: %v", err)
	}

	mux := http.NewServeMux()
	o := NewOpenAPIService()
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
		expectedHash string
		cacheControl string
	}{
		// Correct hash should yield the proper expiry and Cache Control headers
		{"application/json",
			200,
			"openapi/v3/apis/apps/v1?hash=" + hash,
			returnedJSON,
			hash,
			"public, immutable",
		},
		{"application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
			200,
			"openapi/v3/apis/apps/v1?hash=" + hash,
			returnedPb,
			hash,
			"public, immutable",
		},
		// Incorrect hash should redirect to the page with the correct hash
		{"application/json",
			200,
			"openapi/v3/apis/apps/v1?hash=OUTDATEDHASH",
			returnedJSON,
			hash,
			"public, immutable",
		},
		{"application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
			200,
			"openapi/v3/apis/apps/v1?hash=OUTDATEDHASH",
			returnedPb,
			hash,
			"public, immutable",
		},
		// No hash should not return Cache Control information
		{"application/json",
			200,
			"openapi/v3/apis/apps/v1",
			returnedJSON,
			"",
			"",
		},
		{"application/com.github.proto-openapi.spec.v3.v1.0+protobuf",
			200,
			"openapi/v3/apis/apps/v1",
			returnedPb,
			"",
			"",
		},
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

		if resp.StatusCode != 200 {
			t.Errorf("Accept: Unexpected response status code, want: %v, got: %v", 200, resp.StatusCode)
		}

		if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != tc.cacheControl {
			t.Errorf("Expected Cache Control %v, got %v", tc.cacheControl, cacheControl)
		}

		if tc.expectedHash != "" {
			if hash := resp.Request.URL.Query().Get("hash"); hash != tc.expectedHash {
				t.Errorf("Expected Hash: %s, got %s", tc.expectedHash, hash)
			}

			expires := resp.Header.Get("Expires")
			parsedTime, err := time.Parse(time.RFC1123, expires)
			if err != nil {
				t.Errorf("Could not parse cache expiry %v", expires)
			}

			difference := parsedTime.Sub(time.Now()).Hours()
			if difference <= 0 {
				t.Errorf("Expected cache expiry to be in the future")
			}
		} else {
			hash := resp.Request.URL.Query()["hash"]
			if len(hash) != 0 {
				t.Errorf("Expect no redirect and empty hash if the hash is not provide")
			}
			expires := resp.Header.Get("Expires")
			if expires != "" {
				t.Errorf("Expected an empty Expiry if hash is not provided,  got %v", expires)
			}
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Accept: %v: Unexpected error in reading response body: %v", tc.acceptHeader, err)
		}
		if !reflect.DeepEqual(body, tc.respBody) {
			t.Errorf("Accept: %v: Response body mismatches, \nwant: %s, \ngot:  %s", tc.acceptHeader, string(tc.respBody), string(body))
		}
	}
}

func openAPIOrDie(name string) *spec3.OpenAPI {
	openapi := fmt.Sprintf(`{
  "openapi": "3.0",
  "info": {
   "title": "%s",
   "version": "v1.23.0"
  },
  "paths": {}}`, name)
	spec := spec3.OpenAPI{}
	if err := json.Unmarshal([]byte(openapi), &spec); err != nil {
		panic(err)
	}
	return &spec
}

func getDiscovery(server *httptest.Server, path string) (*OpenAPIV3Discovery, string, error) {
	client := server.Client()
	req, err := http.NewRequest("GET", server.URL+"/"+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error in creating new request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error in serving HTTP request: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("unexpected response status code, want: %v, got: %v", 200, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to read request body: %v", err)
	}

	discovery := &OpenAPIV3Discovery{}
	if err := json.Unmarshal(body, &discovery); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal discovery: %v", err)
	}
	return discovery, resp.Header.Get("etag"), nil
}

func TestUpdateGroupVersion(t *testing.T) {
	mux := http.NewServeMux()
	o := NewOpenAPIService()

	mux.Handle("/openapi/v3", http.HandlerFunc(o.HandleDiscovery))

	o.UpdateGroupVersion("apis/apps/v1", openAPIOrDie("apps-v1"))

	server := httptest.NewServer(mux)
	defer server.Close()

	discovery, discovery_etag, err := getDiscovery(server, "/openapi/v3")
	if err != nil {
		t.Fatalf("failed to get /openapi/v3: %v", err)
	}
	etag, ok := discovery.Paths["apis/apps/v1"]
	if !ok {
		t.Fatalf("missing apis/apps/v1")
	}

	// Update with the same thing, make sure we don't update anything.
	o.UpdateGroupVersion("apis/apps/v1", openAPIOrDie("apps-v1"))

	discovery, discovery_etag_updated, err := getDiscovery(server, "/openapi/v3")
	if err != nil {
		t.Fatalf("failed to get /openapi/v3: %v", err)
	}
	if len(discovery.Paths) != 1 {
		t.Fatalf("Invalid number of Paths, expected 1: %v", discovery.Paths)
	}
	etag_updated, ok := discovery.Paths["apis/apps/v1"]
	if !ok {
		t.Fatalf("missing apis/apps/v1")
	}

	if discovery_etag_updated != discovery_etag {
		t.Fatalf("No-op update shouldn't update OpenAPI Discovery etag")
	}

	if etag_updated != etag {
		t.Fatalf("No-op update shouldn't update OpenAPI etag")
	}

	// Add one more, make sure it's in the list
	o.UpdateGroupVersion("apis/something/v1", openAPIOrDie("something-v1"))
	discovery, _, err = getDiscovery(server, "/openapi/v3")
	if err != nil {
		t.Fatalf("failed to get /openapi/v3: %v", err)
	}
	if len(discovery.Paths) != 2 {
		t.Fatalf("Invalid number of Paths, expected 2: %v", discovery.Paths)
	}

	// And remove
	o.DeleteGroupVersion("apis/apps/v1")
	discovery, _, err = getDiscovery(server, "/openapi/v3")
	if err != nil {
		t.Fatalf("failed to get /openapi/v3: %v", err)
	}
	if len(discovery.Paths) != 1 {
		t.Fatalf("Invalid number of Paths, expected 2: %v", discovery.Paths)
	}
}

func TestOpenAPIV3VersionedServiceGzip(t *testing.T) {
	s := buildBenchmarkV3Spec(t, "apps", 20)
	expectedJSON, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	expectedETag := strconv.Quote(computeETag(expectedJSON))
	expectedPb, err := ToV3ProtoBinary(expectedJSON)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	o := NewOpenAPIService()
	mux.Handle("/openapi/v3", http.HandlerFunc(o.HandleDiscovery))
	mux.Handle("/openapi/v3/apis/apps/v1", http.HandlerFunc(o.HandleGroupVersion))
	o.UpdateGroupVersion("apis/apps/v1", s)

	server := httptest.NewServer(mux)
	defer server.Close()

	// Disable automatic gzip negotiation in http.Transport so we can explicitly
	// verify server behavior with and without the Accept-Encoding header.
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	assertVaryContains := func(t *testing.T, h http.Header, want string) {
		t.Helper()
		for _, v := range h.Values("Vary") {
			if v == want {
				return
			}
		}
		t.Errorf("Vary header %v does not contain %q", h.Values("Vary"), want)
	}

	t.Run("JSON without Accept-Encoding: gzip returns uncompressed JSON", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
		if ce := resp.Header.Get("Content-Encoding"); ce != "" {
			t.Errorf("expected empty Content-Encoding, got %q", ce)
		}
		assertVaryContains(t, resp.Header, "Accept")
		assertVaryContains(t, resp.Header, "Accept-Encoding")
		if gotETag := resp.Header.Get("ETag"); gotETag != expectedETag {
			t.Errorf("expected ETag %s, got %s", expectedETag, gotETag)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(body, expectedJSON) {
			t.Errorf("uncompressed body does not match expected JSON")
		}
	})

	t.Run("JSON with Accept-Encoding: gzip returns pre-compressed gzip payload", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept-Encoding", "gzip")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %q", ct)
		}
		if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
			t.Fatalf("expected Content-Encoding: gzip, got %q", ce)
		}
		assertVaryContains(t, resp.Header, "Accept")
		assertVaryContains(t, resp.Header, "Accept-Encoding")
		if gotETag := resp.Header.Get("ETag"); gotETag != expectedETag {
			t.Errorf("expected ETag %s, got %s", expectedETag, gotETag)
		}

		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()
		decompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("failed to read decompressed body: %v", err)
		}
		if !bytes.Equal(decompressed, expectedJSON) {
			t.Errorf("decompressed body does not match expected JSON")
		}
	})

	t.Run("JSON with multi-header Accept-Encoding returns pre-compressed gzip payload", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Accept-Encoding", "deflate")
		req.Header.Add("Accept-Encoding", "gzip")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
		if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
			t.Fatalf("expected Content-Encoding: gzip for multi-header Accept-Encoding, got %q", ce)
		}
		assertVaryContains(t, resp.Header, "Accept-Encoding")
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()
		decompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("failed to read decompressed body: %v", err)
		}
		if !bytes.Equal(decompressed, expectedJSON) {
			t.Errorf("decompressed body does not match expected JSON")
		}
	})

	t.Run("JSON with Accept-Encoding: gzip;q=0 or invalid q returns uncompressed JSON", func(t *testing.T) {
		for _, ae := range []string{"gzip;q=0", "gzip;q=0.000", "gzip;q=invalid", "gzip;q=1.5"} {
			req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept-Encoding", ae)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Accept-Encoding %q: unexpected status code: %d", ae, resp.StatusCode)
			}
			if ce := resp.Header.Get("Content-Encoding"); ce != "" {
				t.Errorf("Accept-Encoding %q: expected empty Content-Encoding, got %q", ae, ce)
			}
			assertVaryContains(t, resp.Header, "Accept-Encoding")
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(body, expectedJSON) {
				t.Errorf("Accept-Encoding %q: body does not match expected JSON", ae)
			}
		}
	})

	t.Run("If-None-Match returns 304 Not Modified without decompressing", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("If-None-Match", expectedETag)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotModified {
			t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusNotModified)
		}
		assertVaryContains(t, resp.Header, "Accept-Encoding")
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(body) != 0 {
			t.Errorf("expected empty 304 body, got %d bytes", len(body))
		}

		// Verify directly that lazyGzipReadSeeker never invokes decompressGzip when
		// http.ServeContent short-circuits on a matching If-None-Match precondition.
		rs := &lazyGzipReadSeeker{compressed: []byte("corrupt-gzip-that-must-not-be-decompressed")}
		rec := httptest.NewRecorder()
		rec.Header().Set("Etag", expectedETag)
		condReq := httptest.NewRequest(http.MethodGet, "/openapi/v3/apis/apps/v1", nil)
		condReq.Header.Set("If-None-Match", expectedETag)
		http.ServeContent(rec, condReq, "", time.Now(), rs)
		if rec.Code != http.StatusNotModified {
			t.Fatalf("lazyGzipReadSeeker 304 status: got %d, want %d", rec.Code, http.StatusNotModified)
		}
		if rs.reader != nil || rs.err != nil {
			t.Errorf("lazyGzipReadSeeker should not have initialized on 304 Not Modified (reader=%v, err=%v)", rs.reader, rs.err)
		}
	})

	t.Run("Protobuf with and without Accept-Encoding: gzip", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/openapi/v3/apis/apps/v1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept", "application/com.github.proto-openapi.spec.v3.v1.0+protobuf")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(body, expectedPb) {
			t.Errorf("protobuf body does not match expected protobuf")
		}
	})
}

func buildBenchmarkV3Spec(tb testing.TB, group string, numResources int) *spec3.OpenAPI {
	tb.Helper()
	paths := make(map[string]interface{}, numResources*2)
	schemas := make(map[string]interface{}, numResources)
	for i := 0; i < numResources; i++ {
		resName := fmt.Sprintf("Resource%d", i)
		ref := fmt.Sprintf("#/components/schemas/io.k8s.%s.%s", group, resName)
		collPath := fmt.Sprintf("/apis/%s/v1/namespaces/{namespace}/resource%ds", group, i)
		itemPath := fmt.Sprintf("/apis/%s/v1/namespaces/{namespace}/resource%ds/{name}", group, i)
		paths[collPath] = map[string]interface{}{
			"get": map[string]interface{}{
				"description": fmt.Sprintf("list or watch objects of kind %s in group %s", resName, group),
				"operationId": fmt.Sprintf("list%sNamespaced%s", group, resName),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": ref},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"description": fmt.Sprintf("create an object of kind %s in group %s", resName, group),
				"operationId": fmt.Sprintf("create%sNamespaced%s", group, resName),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": ref},
							},
						},
					},
				},
			},
		}
		paths[itemPath] = map[string]interface{}{
			"get": map[string]interface{}{
				"description": fmt.Sprintf("read the specified %s in group %s", resName, group),
				"operationId": fmt.Sprintf("read%sNamespaced%s", group, resName),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": ref},
							},
						},
					},
				},
			},
			"put": map[string]interface{}{
				"description": fmt.Sprintf("replace the specified %s in group %s", resName, group),
				"operationId": fmt.Sprintf("replace%sNamespaced%s", group, resName),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": ref},
							},
						},
					},
				},
			},
			"delete": map[string]interface{}{
				"description": fmt.Sprintf("delete the specified %s in group %s", resName, group),
				"operationId": fmt.Sprintf("delete%sNamespaced%s", group, resName),
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": ref},
							},
						},
					},
				},
			},
		}
		props := make(map[string]interface{}, 16)
		for f := 0; f < 16; f++ {
			props[fmt.Sprintf("field%d", f)] = map[string]interface{}{
				"type":        "string",
				"description": fmt.Sprintf("Standard configuration field %d for %s in group %s used by controllers and admission policies.", f, resName, group),
			}
		}
		schemas[fmt.Sprintf("io.k8s.%s.%s", group, resName)] = map[string]interface{}{
			"description": fmt.Sprintf("%s represents a declarative Kubernetes API resource in group %s.", resName, group),
			"type":        "object",
			"properties":  props,
		}
	}
	raw, err := json.Marshal(map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Kubernetes",
			"version": "v1.32.0",
		},
		"paths": paths,
		"components": map[string]interface{}{
			"schemas": schemas,
		},
	})
	if err != nil {
		tb.Fatal(err)
	}
	var s spec3.OpenAPI
	if err := json.Unmarshal(raw, &s); err != nil {
		tb.Fatal(err)
	}
	return &s
}

func BenchmarkHandler3CachedSpecMemory(b *testing.B) {
	groups := []struct {
		name string
		spec *spec3.OpenAPI
	}{
		{"api/v1", buildBenchmarkV3Spec(b, "core", 80)},
		{"apis/apps/v1", buildBenchmarkV3Spec(b, "apps", 40)},
		{"apis/admissionregistration.k8s.io/v1", buildBenchmarkV3Spec(b, "admissionregistration", 30)},
	}

	b.ReportAllocs()
	b.ResetTimer()
	var cachedBytes, cachedCapBytes int
	for i := 0; i < b.N; i++ {
		o := NewOpenAPIService()
		for _, g := range groups {
			o.UpdateGroupVersion(g.name, g.spec)
			if _, _, _, err := o.getSingleGroupBytes(subTypeJSON, g.name); err != nil {
				b.Fatal(err)
			}
		}
		if i == 0 {
			for _, grp := range o.v3Schema {
				ts, _, err := grp.jsonCache.Get()
				if err != nil {
					b.Fatal(err)
				}
				cachedBytes += len(ts.spec)
				cachedCapBytes += cap(ts.spec)
			}
		}
	}
	b.ReportMetric(float64(cachedBytes), "cached-bytes")
	b.ReportMetric(float64(cachedCapBytes), "cached-cap-bytes")
}

func BenchmarkServeGroupVersionGzip(b *testing.B) {
	o := NewOpenAPIService()
	o.UpdateGroupVersion("apis/apps/v1", buildBenchmarkV3Spec(b, "apps", 40))
	req := httptest.NewRequest(http.MethodGet, "/openapi/v3/apis/apps/v1", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	// Warm the cache before measuring per-request serving cost.
	rec := httptest.NewRecorder()
	o.HandleGroupVersion(rec, req)
	if rec.Code != http.StatusOK {
		b.Fatalf("unexpected status: %d", rec.Code)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		o.HandleGroupVersion(w, req)
	}
}
