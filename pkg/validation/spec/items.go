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

	"github.com/go-openapi/swag"
	"gopkg.in/yaml.v3"
)

const (
	jsonRef = "$ref"
)

// SimpleSchema describe swagger simple schemas for parameters and headers
type SimpleSchema struct {
	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Nullable         bool        `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Format           string      `json:"format,omitempty" yaml:"format,omitempty"`
	Items            *Items      `json:"items,omitempty" yaml:"items,omitempty"`
	CollectionFormat string      `json:"collectionFormat,omitempty" yaml:"collectionFormat,omitempty"`
	Default          interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Example          interface{} `json:"example,omitempty" yaml:"example,omitempty"`
}

// CommonValidations describe common JSON-schema validations
type CommonValidations struct {
	Maximum          *float64      `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength        *int64        `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        *int64        `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         *int64        `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         *int64        `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MultipleOf       *float64      `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Enum             []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
}

// Items a limited subset of JSON-Schema's items object.
// It is used by parameter definitions that are not located in "body".
//
// For more information: http://goo.gl/8us55a#items-object
type Items struct {
	Refable
	CommonValidations
	SimpleSchema
	VendorExtensible
}

func (i *Items) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&i.CommonValidations); err != nil {
		return err
	}

	if err := value.Decode(&i.Refable); err != nil {
		return err
	}

	if err := value.Decode(&i.SimpleSchema); err != nil {
		return err
	}

	if err := i.VendorExtensible.UnmarshalYAML(value); err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (i *Items) UnmarshalJSON(data []byte) error {
	var validations CommonValidations
	if err := json.Unmarshal(data, &validations); err != nil {
		return err
	}
	var ref Refable
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}
	var simpleSchema SimpleSchema
	if err := json.Unmarshal(data, &simpleSchema); err != nil {
		return err
	}
	var vendorExtensible VendorExtensible
	if err := json.Unmarshal(data, &vendorExtensible); err != nil {
		return err
	}
	i.Refable = ref
	i.CommonValidations = validations
	i.SimpleSchema = simpleSchema
	i.VendorExtensible = vendorExtensible
	return nil
}

// MarshalJSON converts this items object to JSON
func (i Items) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(i.CommonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(i.SimpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(i.Refable)
	if err != nil {
		return nil, err
	}
	b4, err := json.Marshal(i.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b4, b3, b1, b2), nil
}
