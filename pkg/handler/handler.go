/*
Copyright 2017 The Kubernetes Authors.

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

package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/google/uuid"
	"github.com/munnerz/goautoneg"
	"google.golang.org/protobuf/proto"

	klog "k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/cached"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/common/restfuladapter"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

const (
	subTypeProtobufDeprecated = "com.github.proto-openapi.spec.v2@v1.0+protobuf"
	subTypeProtobuf           = "com.github.proto-openapi.spec.v2.v1.0+protobuf"
	subTypeJSON               = "json"
)

func computeETag(data []byte) string {
	if data == nil {
		return ""
	}
	return fmt.Sprintf("%X", sha512.Sum512(data))
}

type timedSpec struct {
	spec         []byte
	gzipped      []byte
	lastModified time.Time
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(b) // bytes.Buffer never fails
	_ = zw.Close()
	return buf.Bytes()
}

func acceptsGzip(header string) bool {
	for _, enc := range strings.Split(header, ",") {
		if name, params, err := mime.ParseMediaType(enc); err == nil && name == "gzip" {
			q, err := strconv.ParseFloat(params["q"], 64)
			return params["q"] == "" || (err == nil && q > 0)
		}
	}
	return false
}

// OpenAPIService is the service responsible for serving OpenAPI spec. It has
// the ability to safely change the spec while serving it.
type OpenAPIService struct {
	specCache  cached.LastSuccess[*spec.Swagger]
	jsonCache  cached.Value[timedSpec]
	protoCache cached.Value[timedSpec]
}

// NewOpenAPIService builds an OpenAPIService starting with the given spec.
func NewOpenAPIService(swagger *spec.Swagger) *OpenAPIService {
	return NewOpenAPIServiceLazy(cached.Static(swagger, uuid.New().String()))
}

// NewOpenAPIServiceLazy builds an OpenAPIService from lazy spec.
func NewOpenAPIServiceLazy(swagger cached.Value[*spec.Swagger]) *OpenAPIService {
	o := &OpenAPIService{}
	o.UpdateSpecLazy(swagger)

	o.jsonCache = cached.Transform[*spec.Swagger](func(spec *spec.Swagger, etag string, err error) (timedSpec, string, error) {
		if err != nil {
			return timedSpec{}, "", err
		}
		json, err := spec.MarshalJSON()
		if err != nil {
			return timedSpec{}, "", err
		}
		return timedSpec{spec: json, gzipped: gzipBytes(json), lastModified: time.Now()}, computeETag(json), nil
	}, &o.specCache)
	o.protoCache = cached.Transform(func(ts timedSpec, etag string, err error) (timedSpec, string, error) {
		if err != nil {
			return timedSpec{}, "", err
		}
		proto, err := ToProtoBinary(ts.spec)
		if err != nil {
			return timedSpec{}, "", err
		}
		// We can re-use the same etag as json because of the Vary header.
		return timedSpec{spec: proto, gzipped: gzipBytes(proto), lastModified: ts.lastModified}, etag, nil
	}, o.jsonCache)
	return o
}

func (o *OpenAPIService) UpdateSpec(swagger *spec.Swagger) error {
	o.UpdateSpecLazy(cached.Static(swagger, uuid.New().String()))
	return nil
}

func (o *OpenAPIService) UpdateSpecLazy(swagger cached.Value[*spec.Swagger]) {
	o.specCache.Store(swagger)
}

func ToProtoBinary(json []byte) ([]byte, error) {
	document, err := openapi_v2.ParseDocument(json)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(document)
}

// RegisterOpenAPIVersionedService registers a handler to provide access to provided swagger spec.
//
// Deprecated: use OpenAPIService.RegisterOpenAPIVersionedService instead.
func RegisterOpenAPIVersionedService(spec *spec.Swagger, servePath string, handler common.PathHandler) *OpenAPIService {
	o := NewOpenAPIService(spec)
	o.RegisterOpenAPIVersionedService(servePath, handler)
	return o
}

// RegisterOpenAPIVersionedService registers a handler to provide access to provided swagger spec.
func (o *OpenAPIService) RegisterOpenAPIVersionedService(servePath string, handler common.PathHandler) {
	accepted := []struct {
		Type                string
		SubType             string
		ReturnedContentType string
		GetDataAndEtag      cached.Value[timedSpec]
	}{
		{"application", subTypeJSON, "application/" + subTypeJSON, o.jsonCache},
		{"application", subTypeProtobufDeprecated, "application/" + subTypeProtobuf, o.protoCache},
		{"application", subTypeProtobuf, "application/" + subTypeProtobuf, o.protoCache},
	}

	handler.Handle(servePath, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			decipherableFormats := r.Header.Get("Accept")
			if decipherableFormats == "" {
				decipherableFormats = "*/*"
			}
			clauses := goautoneg.ParseAccept(decipherableFormats)
			w.Header().Add("Vary", "Accept")
			w.Header().Add("Vary", "Accept-Encoding")
			for _, clause := range clauses {
				for _, accepts := range accepted {
					if clause.Type != accepts.Type && clause.Type != "*" {
						continue
					}
					if clause.SubType != accepts.SubType && clause.SubType != "*" {
						continue
					}
					// serve the first matching media type in the sorted clause list
					ts, etag, err := accepts.GetDataAndEtag.Get()
					if err != nil {
						klog.Errorf("Error in OpenAPI handler: %s", err)
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					body := ts.spec
					if acceptsGzip(r.Header.Get("Accept-Encoding")) {
						body, etag = ts.gzipped, etag+"-gzip"
						w.Header().Set("Content-Encoding", "gzip")
					}
					// Set Content-Type header in the reponse
					w.Header().Set("Content-Type", accepts.ReturnedContentType)

					// ETag must be enclosed in double quotes: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
					w.Header().Set("Etag", strconv.Quote(etag))
					// ServeContent will take care of caching using eTag.
					http.ServeContent(w, r, servePath, ts.lastModified, bytes.NewReader(body))
					return
				}
			}
			// Return 406 for not acceptable format
			w.WriteHeader(406)
			return
		}),
	)
}

// BuildAndRegisterOpenAPIVersionedService builds the spec and registers a handler to provide access to it.
// Use this method if your OpenAPI spec is static. If you want to update the spec, use BuildOpenAPISpec then RegisterOpenAPIVersionedService.
//
// Deprecated: BuildAndRegisterOpenAPIVersionedServiceFromRoutes should be used instead.
func BuildAndRegisterOpenAPIVersionedService(servePath string, webServices []*restful.WebService, config *common.Config, handler common.PathHandler) (*OpenAPIService, error) {
	return BuildAndRegisterOpenAPIVersionedServiceFromRoutes(servePath, restfuladapter.AdaptWebServices(webServices), config, handler)
}

// BuildAndRegisterOpenAPIVersionedServiceFromRoutes builds the spec and registers a handler to provide access to it.
// Use this method if your OpenAPI spec is static. If you want to update the spec, use BuildOpenAPISpec then RegisterOpenAPIVersionedService.
func BuildAndRegisterOpenAPIVersionedServiceFromRoutes(servePath string, routeContainers []common.RouteContainer, config *common.Config, handler common.PathHandler) (*OpenAPIService, error) {
	spec, err := builder.BuildOpenAPISpecFromRoutes(routeContainers, config)
	if err != nil {
		return nil, err
	}
	o := NewOpenAPIService(spec)
	o.RegisterOpenAPIVersionedService(servePath, handler)
	return o, nil
}
