package proto

import (
	"testing"

	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"sigs.k8s.io/yaml/goyaml.v3"
)

func TestToRawInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    *openapi_v3.DefaultType
		expected *yaml.Node
	}{
		{
			name: "Number type",
			input: &openapi_v3.DefaultType{
				Oneof: &openapi_v3.DefaultType_Number{Number: 42.5},
			},
			expected: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: "42.5",
			},
		},
		{
			name: "Boolean type",
			input: &openapi_v3.DefaultType{
				Oneof: &openapi_v3.DefaultType_Boolean{Boolean: true},
			},
			expected: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: "true",
			},
		},
		{
			name: "String type",
			input: &openapi_v3.DefaultType{
				Oneof: &openapi_v3.DefaultType_String_{String_: "test"},
			},
			expected: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "test",
			},
		},
		{
			name:     "Null type",
			input:    &openapi_v3.DefaultType{},
			expected: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toRawInfo(tt.input)
			if result.Kind != tt.expected.Kind || result.Tag != tt.expected.Tag || result.Value != tt.expected.Value {
				t.Errorf("toRawInfo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewNullNode(t *testing.T) {
	result := newNullNode()
	expected := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"}
	if result.Kind != expected.Kind || result.Tag != expected.Tag {
		t.Errorf("NewNullNode() = %v, want %v", result, expected)
	}
}

func TestNewScalarNodeForString(t *testing.T) {
	result := newScalarNodeForString("test")
	expected := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "test"}
	if result.Kind != expected.Kind || result.Tag != expected.Tag || result.Value != expected.Value {
		t.Errorf("NewScalarNodeForString() = %v, want %v", result, expected)
	}
}

func TestNewScalarNodeForFloat(t *testing.T) {
	result := newScalarNodeForFloat(42.5)
	expected := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: "42.5"}
	if result.Kind != expected.Kind || result.Tag != expected.Tag || result.Value != expected.Value {
		t.Errorf("NewScalarNodeForFloat() = %v, want %v", result, expected)
	}
}

func TestNewScalarNodeForBool(t *testing.T) {
	result := newScalarNodeForBool(true)
	expected := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
	if result.Kind != expected.Kind || result.Tag != expected.Tag || result.Value != expected.Value {
		t.Errorf("NewScalarNodeForBool() = %v, want %v", result, expected)
	}
}
