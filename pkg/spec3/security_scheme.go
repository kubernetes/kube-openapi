package spec3

import (
	"encoding/json"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
)

// SecurityScheme defines reusable Security Scheme Object, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#securitySchemeObject
type SecurityScheme struct {
	spec.Refable
	SecuritySchemeProps
	spec.VendorExtensible
}

// MarshalJSON is a custom marshal function that knows how to encode SecurityScheme as JSON
func (s *SecurityScheme) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.SecuritySchemeProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(s.Refable)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

// SecuritySchemeProps defines a security scheme that can be used by the operations
type SecuritySchemeProps struct {
	// Type of the security scheme
	Type string `json:"type,omitempty"`
	// Description holds a short description for security scheme
	Description string `json:"description,omitempty"`
	// Name holds the name of the header, query or cookie parameter to be used
	Name string `json:"name,omitempty"`
	// In holds the location of the API key
	In string `json:"in,omitempty"`
	// Scheme holds the name of the HTTP Authorization scheme to be used in the Authorization header
	Scheme string `json:"scheme,omitempty"`
	// BearerFormat holds a hint to the client to identify how the bearer token is formatted
	BearerFormat string `json:"bearerFormat,omitempty"`
	// Flows contains configuration information for the flow types supported.
	Flows map[string]*OAuthFlow `json:"flows,omitempty"`
	// OpenIdConnectUrl holds an url to discover OAuth2 configuration values from
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlow contains configuration information for the flow types supported.
type OAuthFlow struct {
	OAuthFlowProps
	spec.VendorExtensible
}

// MarshalJSON is a custom marshal function that knows how to encode OAuthFlow as JSON
func (o *OAuthFlow) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(o.OAuthFlowProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(o.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2), nil
}

// OAuthFlowProps holds configuration details for a supported OAuth Flow
type OAuthFlowProps struct {
	// AuthorizationUrl hold the authorization URL to be used for this flow
	AuthorizationUrl string `json:"authorizationUrl,omitempty"`
	// TokenUrl holds the token URL to be used for this flow
	TokenUrl string `json:"tokenUrl,omitempty"`
	// RefreshUrl holds the URL to be used for obtaining refresh tokens
	RefreshUrl string `json:"refreshUrl,omitempty"`
	// Scopes holds the available scopes for the OAuth2 security scheme
	Scopes map[string]string `json:"scopes,omitempty"`
}
