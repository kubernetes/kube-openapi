package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"k8s.io/kube-openapi/pkg/handler"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func main() {
	swagger := spec.Swagger{}
	if len(os.Args) != 2 {
		fmt.Println("usage: %v <path_to_openapi_v2>", os.Args[0])
		os.Exit(1)
	}
	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("Failed to load file: %v", err))
	}
	err = json.Unmarshal(content, &swagger)
	if err != nil {
		panic(fmt.Errorf("Failed to parse json: %v", err))
	}
	o, err := handler.NewOpenAPIService(&swagger)
	if err != nil {
		panic(fmt.Errorf("Failed to create OpenAPI service: %v", err))
	}
	mux := http.NewServeMux()
	err = o.RegisterOpenAPIVersionedService("/openapi/v2", mux)
	if err != nil {
		panic(fmt.Errorf("Failed to register OpenAPI service: %v", err))
	}
	fmt.Println("Listening on :8080")
	panic(http.ListenAndServe(":8080", mux))
}
