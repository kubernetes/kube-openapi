// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-openapi/swag"
	"gopkg.in/yaml.v3"
	"k8s.io/kube-openapi/pkg/util"
)

// Responses is a container for the expected responses of an operation.
// The container maps a HTTP response code to the expected response.
// It is not expected from the documentation to necessarily cover all possible HTTP response codes,
// since they may not be known in advance. However, it is expected from the documentation to cover
// a successful operation response and any known errors.
//
// The `default` can be used a default response object for all HTTP codes that are not covered
// individually by the specification.
//
// The `Responses Object` MUST contain at least one response code, and it SHOULD be the response
// for a successful operation call.
//
// For more information: http://goo.gl/8us55a#responsesObject
type Responses struct {
	VendorExtensible
	ResponsesProps
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (r *Responses) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.ResponsesProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.VendorExtensible); err != nil {
		return err
	}
	if reflect.DeepEqual(ResponsesProps{}, r.ResponsesProps) {
		r.ResponsesProps = ResponsesProps{}
	}
	return nil
}

func (r *Responses) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&r.ResponsesProps); err != nil {
		return err
	}
	if err := value.Decode(&r.VendorExtensible); err != nil {
		return err
	}
	if reflect.DeepEqual(ResponsesProps{}, r.ResponsesProps) {
		r.ResponsesProps = ResponsesProps{}
	}
	return nil
}

// MarshalJSON converts this items object to JSON
func (r Responses) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.ResponsesProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}

// ResponsesProps describes all responses for an operation.
// It tells what is the default response and maps all responses with a
// HTTP status code.
type ResponsesProps struct {
	Default             *Response
	StatusCodeResponses map[int]Response
}

// MarshalJSON marshals responses as JSON
func (r ResponsesProps) MarshalJSON() ([]byte, error) {
	toser := map[string]Response{}
	if r.Default != nil {
		toser["default"] = *r.Default
	}
	for k, v := range r.StatusCodeResponses {
		toser[strconv.Itoa(k)] = v
	}
	return json.Marshal(toser)
}

// UnmarshalJSON unmarshals responses from JSON
func (r *ResponsesProps) UnmarshalJSON(data []byte) error {
	var res map[string]Response
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}
	if v, ok := res["default"]; ok {
		r.Default = &v
		delete(res, "default")
	}
	for k, v := range res {
		if nk, err := strconv.Atoi(k); err == nil {
			if r.StatusCodeResponses == nil {
				r.StatusCodeResponses = map[int]Response{}
			}
			r.StatusCodeResponses[nk] = v
		}
	}
	return nil
}

func (r *ResponsesProps) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node. Got %v", value.Kind)
	} else if len(value.Content)%2 != 0 {
		return fmt.Errorf("expected mapping node to have even number of subchildren. Got %v", len(value.Content))
	}

	for i := 0; i < len(value.Content); i += 2 {
		var keyStr string
		if err := util.DecodeYAMLString(value.Content[i], &keyStr); err != nil {
			return err
		}

		val := value.Content[i+1]
		if keyStr == "default" {
			r.Default = &Response{}
			if err := r.Default.UnmarshalYAML(val); err != nil {
				return err
			}
		} else if nk, err := strconv.Atoi(keyStr); err == nil {
			resp := Response{}
			if r.StatusCodeResponses == nil {
				r.StatusCodeResponses = map[int]Response{}
			}
			if err := resp.UnmarshalYAML(val); err != nil {
				return err
			}

			r.StatusCodeResponses[nk] = resp
		}
	}
	return nil
}
