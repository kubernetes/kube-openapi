package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/handler"
	"k8s.io/kube-openapi/pkg/handler3"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func serve_v2(swagger spec.Swagger, h common.PathHandler) error {
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

func serve_v3(v3 map[XKubernetesGroupVersion]spec3.OpenAPI, h common.PathHandler) error {
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

func parse_v2(swagger_json []byte) (*spec.Swagger, error) {
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

func parse_v3(openapi_json []byte) (XKubernetesGroupVersion, spec3.OpenAPI, error) {
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

func is_v2(content []byte) (bool, error) {
	t := map[string]interface{}{}
	if err := json.Unmarshal(content, &t); err != nil {
		return false, fmt.Errorf("invalid json: %v", err)
	}
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

func main() {
	cmd := &cobra.Command{
		Use:     "kube-openapi-proxy",
		Short:   "Creates a web-server to load, proxy and/or serve OpenAPI definitions",
		Version: "0.0.1",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()

			var swagger_v2 *spec.Swagger
			openapi_v3 := map[XKubernetesGroupVersion]spec3.OpenAPI{}

			for _, arg := range args {
				content, err := os.ReadFile(arg)
				if err != nil {
					return fmt.Errorf("failed to load file: %v", err)
				}
				if v2, err := is_v2(content); err != nil {
					return fmt.Errorf("failed to detect openapi version (%v): %v", arg, err)
				} else if v2 {
					if swagger_v2 != nil {
						return fmt.Errorf("second OpenAPI v2 detected (only one supported): %v", arg)
					}
					var err error
					swagger_v2, err = parse_v2(content)
					if err != nil {
						return fmt.Errorf("failed to parse OpenAPI v2 (%v): %v", arg, err)
					}
				} else {
					groupVersion, openapi, err := parse_v3(content)
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
				serve_v2(*swagger_v2, mux)
			}
			serve_v3(openapi_v3, mux)

			fmt.Println("Listening on :8080")
			return http.ListenAndServe(":8080", mux)
		},
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failure: %s\n", err)
		os.Exit(1)
	}
}
