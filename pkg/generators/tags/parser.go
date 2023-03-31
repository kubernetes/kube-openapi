/*
Copyright 2023 The Kubernetes Authors.

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

package tags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// FieldValue represents a IDL tag value with multiple fields.
type FieldValue map[string]interface{}

// ParseFieldValues parses an IDL tag value with multiple fields.
// The value must be a comma separated list of fields. Each
// field must be a ':' separated name and value pair.
// The name must be a valid go identifier. The value may be
// a string, number (integer of float), boolean or valid go
// identifier. Strings may be quoted with double quotes or backticks
// and adhere to go string escaping rules. Whitespace between tokens
// is insignificant and allowed.
//
// The IDL tag value must be of the form:
//
//	<tag-field-values> ::= <field> | <tag-field-values> ',' <field>
//	<field>            ::= <key> ':' <value>
//	<key>              ::= IDENT
//	<value>            ::= STRING | RAW_STRING | FLOAT | INT | BOOL | IDENT
//
// Examples:
//
//	rule:"a < b",message:"a must be less than b" // FieldValue{"rule": "a < b", "message": "a must be less than b"}
//	arg1:100,arg2:"xyz",arg3:true, arg4:-1.5,arg5:Blue // FieldValue{"arg1": int64(100), "arg2": "xyz", "arg3": true, "arg4": float64(-1.5), "arg5": "Blue"}
//
// If an error is encountered during parsing, it is returned.
func ParseFieldValues(in string) (FieldValue, error) {
	fields := FieldValue{}
	var s scanner.Scanner
	s.Mode ^= scanner.SkipComments // disallow comments
	s.Init(strings.NewReader(in))

	var errs []error
	s.Error = func(scanner *scanner.Scanner, msg string) {
		errs = append(errs, fmt.Errorf("error parsing '%s' at %v: %s", in, scanner.Position, msg))
	}
	unexpectedTokenError := func(expected string, token string) (FieldValue, error) {
		s.Error(&s, fmt.Sprintf("expected %s but got (%q)", expected, token))
		return nil, errors.Join(errs...)
	}

	for {
		var name string
		var value interface{}
		var err error

		if s.Scan() != scanner.Ident {
			return unexpectedTokenError("field name", s.TokenText())
		}
		name = s.TokenText()

		if s.Scan() != ':' {
			return unexpectedTokenError("','", s.TokenText())
		}

		token := s.Scan()
		switch token {
		case scanner.String, scanner.RawString:
			value, err = strconv.Unquote(s.TokenText())
			if err != nil {
				return unexpectedTokenError(fmt.Sprintf("error parsing string: %s", err), s.TokenText())
			}
		case '-', scanner.Int, scanner.Float:
			value, err = parseNumber(&s, token, unexpectedTokenError)
			if err != nil {
				return nil, err
			}
		case scanner.Ident:
			switch s.TokenText() {
			case "true":
				value = true
			case "false":
				value = false
			default:
				value = s.TokenText()
			}
		default:
			return unexpectedTokenError("field value", s.TokenText())
		}

		fields[name] = value

		switch s.Scan() {
		case ',':
		case scanner.EOF:
			return fields, nil
		default:
			return unexpectedTokenError("',' or end of tag", s.TokenText())
		}
	}
}

func parseNumber(s *scanner.Scanner, token rune, unexpectedTokenError func(expected string, token string) (FieldValue, error)) (interface{}, error) {
	positive := true
	if token == '-' {
		positive = false
		token = s.Scan()
	}

	switch token {
	case scanner.Int:
		i, err := strconv.ParseInt(s.TokenText(), 10, 64)
		if err != nil {
			return unexpectedTokenError(fmt.Sprintf("error parsing string: %s", err), s.TokenText())
		}
		if !positive {
			i = -i
		}
		return i, nil
	case scanner.Float:
		f, err := strconv.ParseFloat(s.TokenText(), 64)
		if err != nil {
			return unexpectedTokenError(fmt.Sprintf("error parsing string: %s", err), s.TokenText())
		}
		if !positive {
			f = -f
		}
		return f, nil
	default:
		return unexpectedTokenError("number", s.TokenText())
	}
}
