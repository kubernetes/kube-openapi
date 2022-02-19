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
	"errors"
	"strings"

	"github.com/go-openapi/swag"
	"gopkg.in/yaml.v3"
	"k8s.io/kube-openapi/pkg/util"
)

// Paths holds the relative paths to the individual endpoints.
// The path is appended to the [`basePath`](http://goo.gl/8us55a#swaggerBasePath) in order
// to construct the full URL.
// The Paths may be empty, due to [ACL constraints](http://goo.gl/8us55a#securityFiltering).
//
// For more information: http://goo.gl/8us55a#pathsObject
type Paths struct {
	VendorExtensible
	Paths map[string]PathItem `json:"-" yaml:"-"` // custom serializer to flatten this, each entry must start with "/"
}

func (p *Paths) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("invalid yaml node provided. Expected key-value map")
	} else if len(value.Content)%2 != 0 {
		return errors.New("invalid mapping node provided. Expected even number of children")
	}

	for i := 0; i < len(value.Content); i += 2 {
		var keyStr string
		if err := util.DecodeYAMLString(value.Content[i], &keyStr); err != nil {
			return err
		}

		val := value.Content[i+1]

		if strings.HasPrefix(keyStr, "x-") || strings.HasPrefix(keyStr, "X-") {
			if p.Extensions == nil {
				p.Extensions = make(map[string]interface{})
			}
			var d interface{}
			if err := val.Decode(&d); err != nil {
				return err
			}
			p.Extensions[strings.ToLower(keyStr)] = d
		} else if strings.HasPrefix(keyStr, "/") {
			if p.Paths == nil {
				p.Paths = make(map[string]PathItem)
			}
			pi := PathItem{}
			if err := pi.UnmarshalYAML(val); err != nil {
				return err
			}
			p.Paths[strings.ToLower(keyStr)] = pi
		}
	}
	return nil
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (p *Paths) UnmarshalJSON(data []byte) error {
	var res map[string]json.RawMessage
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	for k, v := range res {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			if p.Extensions == nil {
				p.Extensions = make(map[string]interface{})
			}
			var d interface{}
			if err := json.Unmarshal(v, &d); err != nil {
				return err
			}
			p.Extensions[k] = d
		}
		if strings.HasPrefix(k, "/") {
			if p.Paths == nil {
				p.Paths = make(map[string]PathItem)
			}
			var pi PathItem
			if err := json.Unmarshal(v, &pi); err != nil {
				return err
			}
			p.Paths[k] = pi
		}
	}
	return nil
}

// MarshalJSON converts this items object to JSON
func (p Paths) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(p.VendorExtensible)
	if err != nil {
		return nil, err
	}

	pths := make(map[string]PathItem)
	for k, v := range p.Paths {
		if strings.HasPrefix(k, "/") {
			pths[k] = v
		}
	}
	b2, err := json.Marshal(pths)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}
