/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	builderv3 "k8s.io/kube-openapi/pkg/builder3"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/test/integration/pkg/generated"
	"k8s.io/kube-openapi/test/integration/testutil"
)

// TODO: Change this to output the generated swagger to stdout.
const defaultSwaggerFile = "generated.v3.json"

func main() {
	// Get the name of the generated swagger file from the args
	// if it exists; otherwise use the default file name.
	swaggerFilename := defaultSwaggerFile
	if len(os.Args) > 1 {
		swaggerFilename = os.Args[1]
	}

	// Generate the definition names from the map keys returned
	// from GetOpenAPIDefinitions. Anonymous function returning empty
	// Ref is not used.
	var defNames []string
	for name, _ := range generated.GetOpenAPIDefinitions(func(name string) spec.Ref {
		return spec.Ref{}
	}) {
		defNames = append(defNames, name)
	}

	// Create a minimal builder config, then call the builder with the definition names.
	config := testutil.CreateOpenAPIBuilderConfig()
	config.GetDefinitions = generated.GetOpenAPIDefinitions
	// Build the Paths using a simple WebService for the final spec
	swagger, serr := builderv3.BuildOpenAPISpec(testutil.CreateWebServices(), config)
	if serr != nil {
		log.Fatalf("ERROR: %s", serr.Error())
	}

	// Marshal the swagger spec into JSON, then write it out.
	specBytes, err := json.MarshalIndent(swagger, " ", " ")
	if err != nil {
		log.Fatalf("json marshal error: %s", err.Error())
	}

	loader := openapi3.NewLoader()
	specForValidator, err := loader.LoadFromData(specBytes)

	if err != nil {
		log.Fatalf("OpenAPI v3 ref resolve error: %s", err.Error())
	}

	err = specForValidator.Validate(loader.Context)

	if err != nil {
		log.Fatalf("OpenAPI v3 validation error: %s", err.Error())
	}

	err = ioutil.WriteFile(swaggerFilename, specBytes, 0644)
	if err != nil {
		log.Fatalf("stdout write error: %s", err.Error())
	}

}
