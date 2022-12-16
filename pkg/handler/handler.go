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
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/emicklei/go-restful/v3"
	"github.com/golang/protobuf/proto"
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/munnerz/goautoneg"
	klog "k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/common/restfuladapter"
	"k8s.io/kube-openapi/pkg/internal/handler"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func computeETag(data []byte) string {
	if data == nil {
		return ""
	}
	return fmt.Sprintf("%X", sha512.Sum512(data))
}

// OpenAPIService is the service responsible for serving OpenAPI spec. It has
// the ability to safely change the spec while serving it.
type OpenAPIService struct {
	// mutex protects All members of this service.
	mutex sync.Mutex

	jsonCache  handler.HandlerCache
	protoCache handler.HandlerCache
	etagCache  handler.HandlerCache
}

// NewOpenAPIService builds an OpenAPIService starting with the given spec.
func NewOpenAPIService(spec *spec.Swagger) (*OpenAPIService, error) {
	o := &OpenAPIService{}
	if err := o.UpdateSpec(spec); err != nil {
		return nil, err
	}
	return o, nil
}

// UpdateSpecLazy updates the handler to use the spec given. While
// `UpdateSpec` takes the actual spec, `UpdateSpecLazy` receives a
// function that will be called to prepare the spec when the first
// download request happens.
func (o *OpenAPIService) UpdateSpecLazy(specGetter func() (*spec.Swagger, error)) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.jsonCache = o.jsonCache.New(func() ([]byte, error) {
		spec, err := specGetter()
		if err != nil {
			return nil, err
		}
		return json.Marshal(spec)
	})
	o.protoCache = o.protoCache.New(func() ([]byte, error) {
		json, err := o.jsonCache.Get()
		if err != nil {
			return nil, err
		}
		return ToProtoBinary(json)
	})
	o.etagCache = o.etagCache.New(func() ([]byte, error) {
		json, err := o.jsonCache.Get()
		if err != nil {
			return nil, err
		}
		return []byte(computeETag(json)), nil
	})
	return nil
}

func (o *OpenAPIService) UpdateSpec(openapiSpec *spec.Swagger) (err error) {
	return o.UpdateSpecLazy(func() (*spec.Swagger, error) { return openapiSpec, nil })
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
func RegisterOpenAPIVersionedService(spec *spec.Swagger, servePath string, handler common.PathHandler) (*OpenAPIService, error) {
	o, err := NewOpenAPIService(spec)
	if err != nil {
		return nil, err
	}
	return o, o.RegisterOpenAPIVersionedService(servePath, handler)
}

// RegisterOpenAPIVersionedService registers a handler to provide access to provided swagger spec.
func (o *OpenAPIService) RegisterOpenAPIVersionedService(servePath string, handler common.PathHandler) error {
	accepted := []struct {
		Type    string
		SubType string
		GetData func() ([]byte, error)
	}{
		{"application", "json", func() ([]byte, error) { return o.jsonCache.Get() }},
		{"application", "com.github.proto-openapi.spec.v2@v1.0+protobuf", func() ([]byte, error) { return o.protoCache.Get() }},
	}

	var lastEtag string
	var lastModified time.Time
	handler.Handle(servePath, gziphandler.GzipHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			decipherableFormats := r.Header.Get("Accept")
			if decipherableFormats == "" {
				decipherableFormats = "*/*"
			}
			clauses := goautoneg.ParseAccept(decipherableFormats)
			w.Header().Add("Vary", "Accept")
			for _, clause := range clauses {
				for _, accepts := range accepted {
					if clause.Type != accepts.Type && clause.Type != "*" {
						continue
					}
					if clause.SubType != accepts.SubType && clause.SubType != "*" {
						continue
					}

					contentType := accepts.Type + "/" + accepts.SubType
					w.Header().Set("Content-Type", contentType)

					o.mutex.Lock()
					// serve the first matching media type in the sorted clause list
					data, err := accepts.GetData()
					if err != nil {
						klog.Errorf("Error in OpenAPI handler: %s", err)
						// only return a 503 if we have no older cache data to serve
						if data == nil {
							w.WriteHeader(http.StatusServiceUnavailable)
							o.mutex.Unlock()
							return
						}
					}

					etag, err := o.etagCache.Get()
					if err != nil {
						// We couldn't get the etag, don't use caching at all.
						klog.Warningf("Error in OpenAPI handler: %s", err)
						lastModified = time.Now()
					} else {
						if string(etag) != lastEtag {
							lastModified = time.Now()
							lastEtag = string(etag)
						}
						// ETag must be enclosed in double quotes:
						// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
						w.Header().Set("Etag", strconv.Quote(string(etag)))
					}
					o.mutex.Unlock()

					// ServeContent will take care of caching using eTag.
					http.ServeContent(w, r, servePath, lastModified, bytes.NewReader(data))
					return
				}
			}
			// Return 406 for not acceptable format
			w.WriteHeader(406)
			return
		}),
	))

	return nil
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
	o, err := NewOpenAPIService(spec)
	if err != nil {
		return nil, err
	}
	return o, o.RegisterOpenAPIVersionedService(servePath, handler)
}
