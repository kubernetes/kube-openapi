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

package proto

import (
	"fmt"
	"sort"
	"strings"

	"github.com/googleapis/gnostic/OpenAPIv2"
	"gopkg.in/yaml.v2"
)

func newSchemaError(path *Path, format string, a ...interface{}) error {
	err := fmt.Sprintf(format, a...)
	if path.Len() == 0 {
		return fmt.Errorf("SchemaError: %v", err)
	}
	return fmt.Errorf("SchemaError(%v): %v", path, err)
}

// VendorExtensionToMap converts openapi VendorExtension to a map.
func VendorExtensionToMap(e []*openapi_v2.NamedAny) map[string]interface{} {
	values := map[string]interface{}{}

	for _, na := range e {
		if na.GetName() == "" || na.GetValue() == nil {
			continue
		}
		if na.GetValue().GetYaml() == "" {
			continue
		}
		var value interface{}
		err := yaml.Unmarshal([]byte(na.GetValue().GetYaml()), &value)
		if err != nil {
			continue
		}

		values[na.GetName()] = value
	}

	return values
}

// definitions is an implementation of `Models`. It looks for
// models in an openapi Schema.
type definitions struct {
	models map[string]Schema
	opts   ConversionOptions
}

var _ Models = &definitions{}

// ConversionOptions describe how to convert an openapi_v2.Document to Models.
type ConversionOptions struct {
	// Permissive allows to drop unsupported OpenAPI fields while making the schema weaker, never stronger.
	Permissive bool
}

// NewOpenAPIData creates a new `Models` out of the openapi document.
func NewOpenAPIData(doc *openapi_v2.Document) (Models, error) {
	return NewOpenAPIDataWithOptions(doc, ConversionOptions{})
}

// NewOpenAPIDataWithOptions creates a new `Models` out of the openapi document, with custom options.
func NewOpenAPIDataWithOptions(doc *openapi_v2.Document, opts ConversionOptions) (Models, error) {
	definitions := definitions{
		models: map[string]Schema{},
		opts:   opts,
	}

	// Save the list of all models first. This will allow us to
	// validate that we don't have any dangling reference.
	for _, namedSchema := range doc.GetDefinitions().GetAdditionalProperties() {
		definitions.models[namedSchema.GetName()] = nil
	}

	// Now, parse each model. We can validate that references exists.
	for _, namedSchema := range doc.GetDefinitions().GetAdditionalProperties() {
		path := NewPath(namedSchema.GetName())
		schema, err := definitions.ParseSchema(namedSchema.GetValue(), &path)
		if err != nil {
			return nil, err
		}
		definitions.models[namedSchema.GetName()] = schema
	}

	return &definitions, nil
}

// We believe the schema is a reference, verify that and returns a new
// Schema
func (d *definitions) parseReference(s *openapi_v2.Schema, path *Path) (Schema, error) {
	// According to OpenAPI standard a schema with a $ref can have properties, but if they
	// contradict the referenced schema, the semantics is undefined.
	if !d.opts.Permissive && len(s.GetProperties().GetAdditionalProperties()) > 0 {
		return nil, newSchemaError(path, "unallowed embedded type definition")
	}
	if !d.opts.Permissive && len(s.GetType().GetValue()) > 0 {
		return nil, newSchemaError(path, "definition reference can't have a type")
	}

	// $refs outside of the definitions are valid, but we don't support them in the whole stack,
	// hence, reject them here.
	if !strings.HasPrefix(s.GetXRef(), "#/definitions/") {
		return nil, newSchemaError(path, "unallowed reference to non-definition %q", s.GetXRef())
	}
	reference := strings.TrimPrefix(s.GetXRef(), "#/definitions/")
	if _, ok := d.models[reference]; !ok {
		return nil, newSchemaError(path, "unknown model in reference: %q", reference)
	}
	return &Ref{
		BaseSchema:  d.parseBaseSchema(s, path),
		reference:   reference,
		definitions: d,
	}, nil
}

func (d *definitions) parseBaseSchema(s *openapi_v2.Schema, path *Path) BaseSchema {
	return BaseSchema{
		Description: s.GetDescription(),
		Extensions:  VendorExtensionToMap(s.GetVendorExtension()),
		Path:        *path,
	}
}

// We believe the schema is a map, verify and return a new schema
func (d *definitions) parseMap(s *openapi_v2.Schema, path *Path) (Schema, error) {
	if len(s.GetType().GetValue()) != 0 && s.GetType().GetValue()[0] != object {
		return nil, newSchemaError(path, "invalid object type")
	}
	var sub Schema
	// TODO(incomplete): this misses the boolean case as AdditionalProperties is a bool+schema sum type.
	if s.GetAdditionalProperties().GetSchema() == nil {
		sub = &Arbitrary{
			BaseSchema: d.parseBaseSchema(s, path),
		}
	} else {
		var err error
		sub, err = d.ParseSchema(s.GetAdditionalProperties().GetSchema(), path)
		if err != nil {
			return nil, err
		}
	}
	return &Map{
		BaseSchema: d.parseBaseSchema(s, path),
		SubType:    sub,
	}, nil
}

func (d *definitions) parsePrimitive(s *openapi_v2.Schema, path *Path) (Schema, error) {
	var t string
	if len(s.GetType().GetValue()) > 1 {
		return nil, newSchemaError(path, "primitive can't have more than 1 type")
	}
	if len(s.GetType().GetValue()) == 1 {
		t = s.GetType().GetValue()[0]
	}
	switch t {
	case String: // do nothing
	case Number: // do nothing
	case Integer: // do nothing
	case Boolean: // do nothing
	case Null: // do nothing
	default:
		return nil, newSchemaError(path, "Unknown primitive type: %q", t)
	}
	return &Primitive{
		BaseSchema: d.parseBaseSchema(s, path),
		Type:       t,
		Format:     s.GetFormat(),
	}, nil
}

func (d *definitions) parseArray(s *openapi_v2.Schema, path *Path) (Schema, error) {
	if len(s.GetType().GetValue()) != 1 {
		return nil, newSchemaError(path, "array should have exactly one type")
	}
	if s.GetType().GetValue()[0] != array {
		return nil, newSchemaError(path, `array should have type "array"`)
	}

	ret := &Array{
		BaseSchema: d.parseBaseSchema(s, path),
	}

	// TODO(incomplete): support multiple item schemas
	if !d.opts.Permissive && len(s.GetItems().GetSchema()) != 1 {
		return nil, newSchemaError(path, "array should have exactly one sub-item")
	}
	if len(s.GetItems().GetSchema()) == 1 {
		sub, err := d.ParseSchema(s.GetItems().GetSchema()[0], path)
		if err != nil {
			return nil, err
		}
		ret.SubType = sub
	}

	return ret, nil
}

func (d *definitions) parseKind(s *openapi_v2.Schema, path *Path) (Schema, error) {
	if len(s.GetType().GetValue()) != 0 && s.GetType().GetValue()[0] != object {
		return nil, newSchemaError(path, "invalid object type")
	}
	if s.GetProperties() == nil {
		return nil, newSchemaError(path, "object doesn't have properties")
	}

	fields := map[string]Schema{}
	fieldOrder := []string{}

	for _, namedSchema := range s.GetProperties().GetAdditionalProperties() {
		var err error
		name := namedSchema.GetName()
		path := path.FieldPath(name)
		fields[name], err = d.ParseSchema(namedSchema.GetValue(), &path)
		if err != nil {
			return nil, err
		}
		fieldOrder = append(fieldOrder, name)
	}

	return &Kind{
		BaseSchema:     d.parseBaseSchema(s, path),
		RequiredFields: s.GetRequired(),
		Fields:         fields,
		FieldOrder:     fieldOrder,
	}, nil
}

func (d *definitions) parseArbitrary(s *openapi_v2.Schema, path *Path) (Schema, error) {
	return &Arbitrary{
		BaseSchema: d.parseBaseSchema(s, path),
	}, nil
}

// ParseSchema creates a walkable Schema from an openapi schema. While
// this function is public, it doesn't leak through the interface.
func (d *definitions) ParseSchema(s *openapi_v2.Schema, path *Path) (Schema, error) {
	if s.GetXRef() != "" {
		// TODO(incomplete): ignoring the rest of s is wrong. As long as there are no conflict, everything from s must be considered
		// Reference: https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#path-item-object
		return d.parseReference(s, path)
	}
	objectTypes := s.GetType().GetValue()
	switch len(objectTypes) {
	case 0:
		// Definition has no type. Treat it as an arbitrary value
		// TODO(incomplete): this ignores many fields, e.g. properties
		return d.parseArbitrary(s, path)
	case 1:
		t := objectTypes[0]
		switch t {
		case object:
			if s.GetProperties() != nil {
				return d.parseKind(s, path)
			} else {
				return d.parseMap(s, path)
			}
		case array:
			return d.parseArray(s, path)
		}
		return d.parsePrimitive(s, path)
	default:
		if !d.opts.Permissive {
			return nil, newSchemaError(path, "definitions with multiple types aren't supported")
		}

		// Definition has many types. We cannot cope with that. Ignore all of them.
		// TODO(incomplete): add multi-type support
		return d.parseArbitrary(s, path)
	}
}

// LookupModel is public through the interface of Models. It
// returns a visitable schema from the given model name.
func (d *definitions) LookupModel(model string) Schema {
	return d.models[model]
}

func (d *definitions) ListModels() []string {
	models := []string{}

	for model := range d.models {
		models = append(models, model)
	}

	sort.Strings(models)
	return models
}

type Ref struct {
	BaseSchema

	reference   string
	definitions *definitions
}

var _ Reference = &Ref{}

func (r *Ref) Reference() string {
	return r.reference
}

func (r *Ref) SubSchema() Schema {
	return r.definitions.models[r.reference]
}

func (r *Ref) Accept(v SchemaVisitor) {
	v.VisitReference(r)
}

func (r *Ref) GetName() string {
	return fmt.Sprintf("Reference to %q", r.reference)
}
