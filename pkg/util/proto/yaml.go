/*
Copyright 2025 The Kubernetes Authors.

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

	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"sigs.k8s.io/yaml/goyaml.v3"
)

// toRawInfo returns a description of DefaultType suitable for JSON or YAML export.
func toRawInfo(m *openapi_v3.DefaultType) *yaml.Node {
	// ONE OF WRAPPER
	// DefaultType
	// {Name:number Type:float StringEnumValues:[] MapType: Repeated:false Pattern: Implicit:false Description:}
	if v0, ok := m.GetOneof().(*openapi_v3.DefaultType_Number); ok {
		return newScalarNodeForFloat(v0.Number)
	}
	// {Name:boolean Type:bool StringEnumValues:[] MapType: Repeated:false Pattern: Implicit:false Description:}
	if v1, ok := m.GetOneof().(*openapi_v3.DefaultType_Boolean); ok {
		return newScalarNodeForBool(v1.Boolean)
	}
	// {Name:string Type:string StringEnumValues:[] MapType: Repeated:false Pattern: Implicit:false Description:}
	if v2, ok := m.GetOneof().(*openapi_v3.DefaultType_String_); ok {
		return newScalarNodeForString(v2.String_)
	}
	return newNullNode()
}

// newNullNode creates a new Null node.
func newNullNode() *yaml.Node {
	node := &yaml.Node{
		Kind: yaml.ScalarNode,
		Tag:  "!!null",
	}
	return node
}

// newScalarNodeForString creates a new node to hold a string.
func newScalarNodeForString(s string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: s,
	}
}

// newScalarNodeForFloat creates a new node to hold a float.
func newScalarNodeForFloat(f float64) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!float",
		Value: fmt.Sprintf("%g", f),
	}
}

// newScalarNodeForBool creates a new node to hold a bool.
func newScalarNodeForBool(b bool) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!bool",
		Value: fmt.Sprintf("%t", b),
	}
}
