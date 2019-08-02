package spec3

import "github.com/go-openapi/spec"

// Components holds a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
//
// more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#componentsObject
type Components struct {
	// Schemas holds reusable Schema Objects
	Schemas Schemas `json:"schemas,omitempty"`
	// SecuritySchemes holds reusable Security Scheme Objects, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#securitySchemeObject
	SecuritySchemes SecuritySchemes `json:"securitySchemes,omitempty"`
	// the following fields are missing:
	//
	// responses	Map[string, Response Object | Reference Object]
	// parameters	Map[string, Parameter Object | Reference Object]
	// examples	Map[string, Example Object | Reference Object]
	// requestBodies	Map[string, Request Body Object | Reference Object]
	// headers	Map[string, Header Object | Reference Object]
	// links	Map[string, Link Object | Reference Object]
	// callbacks	Map[string, Callback Object | Reference Object]
	//
	// all fields are defined at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#componentsObject
}

// SecuritySchemes holds reusable Security Scheme Objects, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#securitySchemeObject
type SecuritySchemes map[string]*SecurityScheme

// Schemas holds reusable Schema Objects
type Schemas map[string]*spec.Schema
