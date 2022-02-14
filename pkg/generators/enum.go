/*
Copyright 2021 The Kubernetes Authors.

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

package generators

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"
)

const tagEnumType = "enum"
const enumTypeDescriptionHeader = "Possible enum values:"

type enumValue struct {
	Name    string
	Value   string
	Comment string
}

type enumType struct {
	Name   types.Name
	Values []*enumValue

	// indexes map constant value to its index in Values
	// so that we can de-dup constant values in case of an alias
	indexes map[string]int
}

// enumMap is a map from the name to the matching enum type.
type enumMap map[types.Name]*enumType

type enumContext struct {
	enumTypes enumMap
}

func newEnumType(name types.Name) *enumType {
	return &enumType{
		Name:    name,
		indexes: make(map[string]int),
	}
}

func newEnumContext(c *generator.Context) (*enumContext, error) {
	et, err := parseEnums(c)
	if err != nil {
		return nil, err
	}
	return &enumContext{enumTypes: et}, nil
}

// EnumType checks and finds the enumType for a given type.
// If the given type is a known enum type, returns the enumType, true
// Otherwise, returns nil, false
func (ec *enumContext) EnumType(t *types.Type) (enum *enumType, isEnum bool) {
	enum, ok := ec.enumTypes[t.Name]
	return enum, ok
}

// ValueStrings returns all possible values of the enum type as strings
// the results are sorted and quoted as Go literals.
func (et *enumType) ValueStrings() []string {
	var values []string
	for _, value := range et.Values {
		// use "%q" format to generate a Go literal of the string const value
		values = append(values, fmt.Sprintf("%q", value.Value))
	}
	sort.Strings(values)
	return values
}

// DescriptionLines returns a description of the enum in this format:
//
// Possible enum values:
//  - `"value1"` description 1
//  - `"value2"` description 2
func (et *enumType) DescriptionLines() []string {
	var lines []string
	for _, value := range et.Values {
		lines = append(lines, value.Description())
	}
	sort.Strings(lines)
	// Prepend an empty string to initiate a new paragraph.
	return append([]string{"", enumTypeDescriptionHeader}, lines...)
}

func parseEnums(c *generator.Context) (enumMap, error) {
	// First, find the builtin "string" type
	stringType := c.Universe.Type(types.Name{Name: "string"})

	enumTypes := make(enumMap)
	for _, p := range c.Universe {
		// find all enum types.
		for _, t := range p.Types {
			if isEnumType(stringType, t) {
				if _, ok := enumTypes[t.Name]; !ok {
					enumTypes[t.Name] = newEnumType(t.Name)
				}
			}
		}
		// find all enum values from constants, and try to match each with its type.
		for _, c := range p.Constants {
			enumType := c.Underlying
			if _, ok := enumTypes[enumType.Name]; ok {
				value := &enumValue{
					Name:    c.Name.Name,
					Value:   *c.ConstValue,
					Comment: strings.Join(c.CommentLines, " "),
				}
				if err := enumTypes[enumType.Name].appendValue(value); err != nil {
					return nil, err
				}
			}
		}
	}

	return enumTypes, nil
}

func (et *enumType) appendValue(value *enumValue) error {
	if idx, ok := et.indexes[value.Value]; ok {
		if value.Comment != "" {
			existing := et.Values[idx]
			if existing.Comment != "" {
				return fmt.Errorf("duplicated comment for %v", value.Value)
			}
			existing.Comment = value.Comment
		}
		return nil
	}
	et.Values = append(et.Values, value)
	et.indexes[value.Value] = len(et.Values) - 1
	return nil
}

// Description returns the description line for the enumValue
// with the format:
//  - `"FooValue"` is the Foo value
func (ev *enumValue) Description() string {
	comment := strings.TrimSpace(ev.Comment)
	// The comment should start with the type name, trim it first.
	// TODO(jiahuif) gengo needs a way to normalize aliases of constants
	// // Foo is foo
	// const Foo = "foo"
	// const FooAlias = Foo
	// will result Name = "FooAlias", Comment = "// Foo if foo",
	// split the string to workaround it
	parts := strings.SplitN(comment, " ", 2)
	if len(parts) == 2 {
		comment = parts[1]
	}
	// Trim the possible space after previous step.
	comment = strings.TrimSpace(comment)
	// The comment may be multiline, cascade all consecutive whitespaces.
	comment = whitespaceRegex.ReplaceAllString(comment, " ")
	return fmt.Sprintf(" - `%q` %s", ev.Value, comment)
}

// isEnumType checks if a given type is an enum by the definition
// An enum type should be an alias of string and has tag '+enum' in its comment.
// Additionally, pass the type of builtin 'string' to check against.
func isEnumType(stringType *types.Type, t *types.Type) bool {
	return t.Kind == types.Alias && t.Underlying == stringType && hasEnumTag(t)
}

func hasEnumTag(t *types.Type) bool {
	return types.ExtractCommentTags("+", t.CommentLines)[tagEnumType] != nil
}

// whitespaceRegex is the regex for consecutive whitespaces.
var whitespaceRegex = regexp.MustCompile(`\s+`)
