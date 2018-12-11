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

package aggregator

import (
	"strings"

	"github.com/go-openapi/spec"
)

const (
	definitionPrefix = "#/definitions/"
)

// Run a walkRefCallback method on all references of an OpenAPI spec
type referenceWalker struct {
	// walkRefCallback will be called on each reference and the return value
	// will replace that reference. This will allow the callers to change
	// all/some references of an spec (e.g. useful in renaming definitions).
	walkRefCallback func(ref spec.Ref) spec.Ref

	// The spec to walk through.
	root *spec.Swagger

	// Keep track of visited references
	alreadyVisited map[string]bool
}

func walkOnAllReferences(walkRef func(ref spec.Ref) spec.Ref, sp *spec.Swagger) {
	walker := &referenceWalker{walkRefCallback: walkRef, root: sp, alreadyVisited: map[string]bool{}}
	walker.Start()
}

func (s *referenceWalker) walkRef(ref spec.Ref) spec.Ref {
	refStr := ref.String()
	// References that start with #/definitions/ has a definition
	// inside the same spec file. If that is the case, walk through
	// those definitions too.
	// We do not support external references yet.
	if !s.alreadyVisited[refStr] && strings.HasPrefix(refStr, definitionPrefix) {
		s.alreadyVisited[refStr] = true
		k := refStr[len(definitionPrefix):]
		def := s.root.Definitions[k]
		s.walkSchema(&def)
		// Make sure we don't assign to nil map
		if s.root.Definitions == nil {
			s.root.Definitions = spec.Definitions{}
		}
		s.root.Definitions[k] = def
	}
	return s.walkRefCallback(ref)
}

func (s *referenceWalker) walkSchema(schema *spec.Schema) {
	if schema == nil {
		return
	}
	schema.Ref = s.walkRef(schema.Ref)
	for k, v := range schema.Definitions {
		s.walkSchema(&v)
		schema.Definitions[k] = v
	}
	for k, v := range schema.Properties {
		s.walkSchema(&v)
		schema.Properties[k] = v
	}
	for k, v := range schema.PatternProperties {
		s.walkSchema(&v)
		schema.PatternProperties[k] = v
	}
	for i := range schema.AllOf {
		s.walkSchema(&schema.AllOf[i])
	}
	for i := range schema.AnyOf {
		s.walkSchema(&schema.AnyOf[i])
	}
	for i := range schema.OneOf {
		s.walkSchema(&schema.OneOf[i])
	}
	if schema.Not != nil {
		s.walkSchema(schema.Not)
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		s.walkSchema(schema.AdditionalProperties.Schema)
	}
	if schema.AdditionalItems != nil && schema.AdditionalItems.Schema != nil {
		s.walkSchema(schema.AdditionalItems.Schema)
	}
	if schema.Items != nil {
		if schema.Items.Schema != nil {
			s.walkSchema(schema.Items.Schema)
		}
		for i := range schema.Items.Schemas {
			s.walkSchema(&schema.Items.Schemas[i])
		}
	}
}

func (s *referenceWalker) walkParams(params []spec.Parameter) {
	if params == nil {
		return
	}
	for _, param := range params {
		param.Ref = s.walkRef(param.Ref)
		s.walkSchema(param.Schema)
		if param.Items != nil {
			param.Items.Ref = s.walkRef(param.Items.Ref)
		}
	}
}

func (s *referenceWalker) walkResponse(resp *spec.Response) {
	if resp == nil {
		return
	}
	resp.Ref = s.walkRef(resp.Ref)
	s.walkSchema(resp.Schema)
}

func (s *referenceWalker) walkOperation(op *spec.Operation) {
	if op == nil {
		return
	}
	s.walkParams(op.Parameters)
	if op.Responses == nil {
		return
	}
	s.walkResponse(op.Responses.Default)
	for _, r := range op.Responses.StatusCodeResponses {
		s.walkResponse(&r)
	}
}

func (s *referenceWalker) Start() {
	if s.root.Paths == nil {
		return
	}
	for _, pathItem := range s.root.Paths.Paths {
		s.walkParams(pathItem.Parameters)
		s.walkOperation(pathItem.Delete)
		s.walkOperation(pathItem.Get)
		s.walkOperation(pathItem.Head)
		s.walkOperation(pathItem.Options)
		s.walkOperation(pathItem.Patch)
		s.walkOperation(pathItem.Post)
		s.walkOperation(pathItem.Put)
	}
}
