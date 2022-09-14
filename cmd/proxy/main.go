package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/handler"
	"k8s.io/kube-openapi/pkg/handler3"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func ServeV2(swagger spec.Swagger, h common.PathHandler) error {
	o, err := handler.NewOpenAPIService(&swagger)
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI v2 service: %v", err)
	}
	err = o.RegisterOpenAPIVersionedService("/openapi/v2", h)
	if err != nil {
		return fmt.Errorf("failed to register OpenAPI service: %v", err)
	}
	return nil
}

func ServeV3(v3 map[XKubernetesGroupVersion]spec3.OpenAPI, h common.PathHandler) error {
	o, err := handler3.NewOpenAPIService(nil)
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI v3 service: %v", err)
	}

	h.Handle("/openapi/v3", http.HandlerFunc(o.HandleDiscovery))
	for gv := range v3 {
		h.Handle("/openapi/v3/"+gv.Path(), http.HandlerFunc(o.HandleGroupVersion))
		spec := v3[gv]
		o.UpdateGroupVersion(gv.Path(), &spec)
	}
	return nil
}

func ParseV2(swagger_json []byte) (*spec.Swagger, error) {
	swagger := spec.Swagger{}
	if err := json.Unmarshal(swagger_json, &swagger); err != nil {
		return nil, fmt.Errorf("failed to parse json: %v", err)
	}
	return &swagger, nil
}

type XKubernetesGroupVersion struct {
	Group   string
	Version string
}

func (gv XKubernetesGroupVersion) String() string {
	if gv.Group == "" {
		return gv.Version
	}
	return gv.Group + "/" + gv.Version
}

func (gv XKubernetesGroupVersion) Path() string {
	if gv.Group == "" {
		return "api/" + gv.Version
	}
	return "apis/" + gv.Group + "/" + gv.Version
}

var resourcePathRe = regexp.MustCompile(`^/api(s/(?P<group>[^/]+))?/(?P<version>[^/]+)/`)
var groupIndexRe = resourcePathRe.SubexpIndex("group")
var versionIndexRe = resourcePathRe.SubexpIndex("version")

func ParseV3(openapi_json []byte) (XKubernetesGroupVersion, spec3.OpenAPI, error) {
	openapi := spec3.OpenAPI{}
	if err := json.Unmarshal(openapi_json, &openapi); err != nil {
		return XKubernetesGroupVersion{}, spec3.OpenAPI{}, fmt.Errorf("failed to parse json: %v", err)
	}
	// Find the groupVersion.
	if openapi.Paths == nil {
		return XKubernetesGroupVersion{}, spec3.OpenAPI{}, fmt.Errorf("empty OpenAPI paths")
	}
	var gv *XKubernetesGroupVersion = nil
	for path := range openapi.Paths.Paths {
		matches := resourcePathRe.FindStringSubmatch(path)
		if matches == nil || matches[versionIndexRe] == "" {
			return XKubernetesGroupVersion{}, spec3.OpenAPI{}, fmt.Errorf("failed to parse resource path: %v", path)
		}
		newGV := XKubernetesGroupVersion{Group: matches[groupIndexRe], Version: matches[versionIndexRe]}
		if gv == nil {
			gv = &newGV
		} else if *gv != newGV {
			return XKubernetesGroupVersion{}, spec3.OpenAPI{}, fmt.Errorf("Multiple GV operations: %v/%v", *gv, newGV)
		}
	}
	if gv == nil {
		return XKubernetesGroupVersion{}, spec3.OpenAPI{}, fmt.Errorf("no group-version detected in OpenAPI")
	}
	return *gv, openapi, nil
}

func IsV2(content []byte) (bool, error) {
	t := map[string]interface{}{}
	if err := json.Unmarshal(content, &t); err != nil {
		return false, fmt.Errorf("invalid json: %v", err)
	}
	// TODO: This probably needs a tighter check, possibly using
	// https://github.com/kubernetes/kube-openapi/blob/3a31a646d5158c26e77bebb8285a6f3194cda7e2/test/integration/builder3/main.go#L68-L75
	if version, ok := t["openapi"]; ok {
		if version == "3.0.0" {
			return false, nil
		}
		return false, fmt.Errorf(`invalid "openapi" version: %v`, version)
	}
	if version, ok := t["swagger"]; ok {
		if version == "2.0" {
			return true, nil
		}
		return false, fmt.Errorf(`invalid "swagger" version: %v`, version)
	}
	return false, errors.New(`invalid openapi/swagger, no version included`)
}

func ParseAndServe(mux common.PathHandler, args ...string) error {
	var swagger_v2 *spec.Swagger
	openapi_v3 := map[XKubernetesGroupVersion]spec3.OpenAPI{}

	for _, arg := range args {
		content, err := os.ReadFile(arg)
		if err != nil {
			return fmt.Errorf("failed to load file: %v", err)
		}
		if v2, err := IsV2(content); err != nil {
			return fmt.Errorf("failed to detect openapi version (%v): %v", arg, err)
		} else if v2 {
			if swagger_v2 != nil {
				return fmt.Errorf("second OpenAPI v2 detected (only one supported): %v", arg)
			}
			var err error
			swagger_v2, err = ParseV2(content)
			if err != nil {
				return fmt.Errorf("failed to parse OpenAPI v2 (%v): %v", arg, err)
			}
		} else {
			groupVersion, openapi, err := ParseV3(content)
			if err != nil {
				return fmt.Errorf("failed to parse OpenAPI v3 (%v): %v", arg, err)
			}
			if _, ok := openapi_v3[groupVersion]; ok {
				return fmt.Errorf("second OpenAPI v3 detected for %v (only one supported): %v", groupVersion, arg)
			}
			openapi_v3[groupVersion] = openapi
		}
	}

	if swagger_v2 != nil {
		ServeV2(*swagger_v2, mux)
	}
	ServeV3(openapi_v3, mux)
	return nil
}

func main() {
	mux := http.NewServeMux()

	if err := ParseAndServe(mux, os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process arguments: %v", err)
		os.Exit(1)
	}

	fmt.Println("Listening on :8080")
	panic(http.ListenAndServe(":8080", mux))
}
