/*
Copyright 2022 The Kubernetes Authors.

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

package openapiconv

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	builderv2 "k8s.io/kube-openapi/pkg/builder"
	builderv3 "k8s.io/kube-openapi/pkg/builder3"
	"k8s.io/kube-openapi/pkg/openapiconv"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/test/integration/pkg/generated"
	"k8s.io/kube-openapi/test/integration/testutil"
)

func TestConvertGolden(t *testing.T) {
	// Generate the definition names from the map keys returned
	// from GetOpenAPIDefinitions. Anonymous function returning empty
	// Ref is not used.
	var defNames []string
	for name := range generated.GetOpenAPIDefinitions(func(name string) spec.Ref {
		return spec.Ref{}
	}) {
		defNames = append(defNames, name)
	}

	// Create a minimal builder config, then call the builder with the definition names.
	config := testutil.CreateOpenAPIBuilderConfig()
	config3 := testutil.CreateOpenAPIV3BuilderConfig()
	config.GetDefinitions = generated.GetOpenAPIDefinitions
	config3.GetDefinitions = generated.GetOpenAPIDefinitions
	// Build the Paths using a simple WebService for the final spec
	openapiv2, serr := builderv2.BuildOpenAPISpec(testutil.CreateWebServices(false), config)
	if serr != nil {
		log.Fatalf("ERROR: %s", serr.Error())
	}

	openAPIV2JSONBeforeConversion, err := json.Marshal(openapiv2)
	if err != nil {
		t.Fatal(err)
	}
	openapiv3, serr := builderv3.BuildOpenAPISpec(testutil.CreateWebServices(false), config3)
	if serr != nil {
		log.Fatalf("ERROR: %s", serr.Error())
	}

	convertedOpenAPIV3 := openapiconv.ConvertV2ToV3(openapiv2)
	if err != nil {
		t.Fatal(err)
	}
	openAPIV2JSONAfterConversion, err := json.Marshal(openapiv2)
	if !reflect.DeepEqual(openAPIV2JSONBeforeConversion, openAPIV2JSONAfterConversion) {
		t.Errorf("Expected OpenAPI V2 to be untouched before and after conversion")
	}

	if !reflect.DeepEqual(openapiv3, convertedOpenAPIV3) {
		t.Errorf("Expected converted OpenAPI to be equal, %v, %v", openapiv3, convertedOpenAPIV3)
	}
}
