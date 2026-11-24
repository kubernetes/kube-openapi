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

package util

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// hasSiblingElements reports whether s has any field set other than s.Ref.
// It is equivalent to !reflect.DeepEqual(*s, spec.Schema{SchemaProps: spec.SchemaProps{Ref: s.Ref}})
// without boxing spec.Schema into interface{} or allocating a reflect visited map.
func hasSiblingElements(s *spec.Schema) bool {
	return s.Description != "" ||
		s.Nullable ||
		s.Extensions != nil ||
		s.Default != nil ||
		s.Type != nil ||
		s.Format != "" ||
		s.Properties != nil ||
		s.Items != nil ||
		s.AllOf != nil ||
		s.OneOf != nil ||
		s.AnyOf != nil ||
		s.Not != nil ||
		s.AdditionalProperties != nil ||
		s.Required != nil ||
		s.Enum != nil ||
		s.Title != "" ||
		s.ID != "" ||
		s.Schema != "" ||
		s.Maximum != nil ||
		s.ExclusiveMaximum ||
		s.Minimum != nil ||
		s.ExclusiveMinimum ||
		s.MaxLength != nil ||
		s.MinLength != nil ||
		s.Pattern != "" ||
		s.MaxItems != nil ||
		s.MinItems != nil ||
		s.UniqueItems ||
		s.MultipleOf != nil ||
		s.MaxProperties != nil ||
		s.MinProperties != nil ||
		s.PatternProperties != nil ||
		s.Dependencies != nil ||
		s.AdditionalItems != nil ||
		s.Definitions != nil ||
		s.Discriminator != "" ||
		s.ReadOnly ||
		s.ExternalDocs != nil ||
		s.Example != nil ||
		s.ExtraProps != nil
}

func hasNonEmptyRef(r *spec.Ref) bool {
	if u := r.GetURL(); u != nil {
		return u.Fragment != "" || u.Path != "" || u.Host != "" || u.Scheme != "" || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.User != nil
	}
	return r.HasFragmentOnly || len(r.GetPointer().DecodedTokens()) > 0
}

// WrapRefs wraps OpenAPI V3 Schema refs that contain sibling elements.
// AllOf is used to wrap the Ref to prevent references from having sibling elements.
// Please see https://github.com/kubernetes/kubernetes/issues/106387#issuecomment-967640388
//
// WrapRefs uses copy-on-write semantics: unmodified subtrees may share structure
// with schema, so callers must not mutate schema after calling WrapRefs.
func WrapRefs(schema *spec.Schema) *spec.Schema {
	if schema == nil {
		return nil
	}
	v := *schema
	if !wrapSchemaInPlace(&v) {
		return schema
	}
	out := new(spec.Schema)
	*out = v
	return out
}

func wrapSchemaMap(orig map[string]spec.Schema) (map[string]spec.Schema, bool) {
	var out map[string]spec.Schema
	for k, v := range orig {
		if wrapSchemaInPlace(&v) {
			if out == nil {
				out = make(map[string]spec.Schema, len(orig))
				for k2, v2 := range orig {
					out[k2] = v2
				}
			}
			out[k] = v
		}
	}
	if out != nil {
		return out, true
	}
	return orig, false
}

func wrapSchemaSlice(orig []spec.Schema) ([]spec.Schema, bool) {
	var out []spec.Schema
	for i := range orig {
		v := orig[i]
		if wrapSchemaInPlace(&v) {
			if out == nil {
				out = make([]spec.Schema, len(orig))
				copy(out, orig)
			}
			out[i] = v
		}
	}
	if out != nil {
		return out, true
	}
	return orig, false
}

func wrapSchemaInPlace(schema *spec.Schema) bool {
	changed := false
	if hasNonEmptyRef(&schema.Ref) && hasSiblingElements(schema) {
		schema.AllOf = []spec.Schema{{
			SchemaProps: spec.SchemaProps{
				Ref: schema.Ref,
			},
		}}
		schema.Ref = spec.Ref{}
		changed = true
	}
	if m, ok := wrapSchemaMap(schema.Definitions); ok {
		schema.Definitions = m
		changed = true
	}
	if m, ok := wrapSchemaMap(schema.Properties); ok {
		schema.Properties = m
		changed = true
	}
	if m, ok := wrapSchemaMap(schema.PatternProperties); ok {
		schema.PatternProperties = m
		changed = true
	}
	if s, ok := wrapSchemaSlice(schema.AllOf); ok {
		schema.AllOf = s
		changed = true
	}
	if s, ok := wrapSchemaSlice(schema.AnyOf); ok {
		schema.AnyOf = s
		changed = true
	}
	if s, ok := wrapSchemaSlice(schema.OneOf); ok {
		schema.OneOf = s
		changed = true
	}
	if schema.Not != nil {
		if s := WrapRefs(schema.Not); s != schema.Not {
			changed = true
			schema.Not = s
		}
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		if s := WrapRefs(schema.AdditionalProperties.Schema); s != schema.AdditionalProperties.Schema {
			changed = true
			schema.AdditionalProperties = &spec.SchemaOrBool{Schema: s, Allows: schema.AdditionalProperties.Allows}
		}
	}
	if schema.AdditionalItems != nil && schema.AdditionalItems.Schema != nil {
		if s := WrapRefs(schema.AdditionalItems.Schema); s != schema.AdditionalItems.Schema {
			changed = true
			schema.AdditionalItems = &spec.SchemaOrBool{Schema: s, Allows: schema.AdditionalItems.Allows}
		}
	}
	if schema.Items != nil {
		if schema.Items.Schema != nil {
			if s := WrapRefs(schema.Items.Schema); s != schema.Items.Schema {
				changed = true
				schema.Items = &spec.SchemaOrArray{Schema: s}
			}
		} else if s, ok := wrapSchemaSlice(schema.Items.Schemas); ok {
			schema.Items = &spec.SchemaOrArray{Schemas: s}
			changed = true
		}
	}
	return changed
}
