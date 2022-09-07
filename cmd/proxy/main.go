package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/handler"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func serve_openapi_v2(v2_path string, h common.PathHandler) error {
	swagger := spec.Swagger{}
	content, err := os.ReadFile(v2_path)
	if err != nil {
		return fmt.Errorf("Failed to load file: %v", err)
	}
	err = json.Unmarshal(content, &swagger)
	if err != nil {
		return fmt.Errorf("Failed to parse json: %v", err)
	}
	o, err := handler.NewOpenAPIService(&swagger)
	if err != nil {
		return fmt.Errorf("Failed to create OpenAPI service: %v", err)
	}
	err = o.RegisterOpenAPIVersionedService("/openapi/v2", h)
	if err != nil {
		return fmt.Errorf("Failed to register OpenAPI service: %v", err)
	}
	return nil
}

func main() {
	var openapi_v2_path string
	cmd := &cobra.Command{
		Use:     "kube-openapi-proxy",
		Short:   "Creates a web-server to load, proxy and/or serve OpenAPI definitions",
		Version: "0.0.1",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()
			if openapi_v2_path != "" {
				fmt.Println("Setting-up OpenAPI v2 for", openapi_v2_path)
				serve_openapi_v2(openapi_v2_path, mux)
			}

			fmt.Println("Listening on :8080")
			return http.ListenAndServe(":8080", mux)
		},
	}
	cmd.Flags().StringVar(&openapi_v2_path, "v2", "", "Path to OpenAPI v2 swagger json file")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}
