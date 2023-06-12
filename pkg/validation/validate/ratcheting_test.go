package validate_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

func ptr[T any](v T) *T {
	return &v
}

var smallIntSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type:    spec.StringOrArray{"number"},
		Maximum: ptr(float64(50)),
	},
}

var mediumIntSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type:    spec.StringOrArray{"number"},
		Minimum: ptr(float64(50)),
	},
}

func TestScalarRatcheting(t *testing.T) {
	validator := validate.NewRatchetingSchemaValidator(mediumIntSchema, nil, "", strfmt.Default)
	require.True(t, validator.ValidateUpdate(1, 1).IsValid())
	require.False(t, validator.ValidateUpdate(1, 2).IsValid())
}

var objectSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type: spec.StringOrArray{"object"},
		Properties: map[string]spec.Schema{
			"a": *smallIntSchema,
			"b": *mediumIntSchema,
		},
	},
}

func TestObjectRatcheting(t *testing.T) {
	validator := validate.NewRatchetingSchemaValidator(objectSchema, nil, "", strfmt.Default)
	require.True(t, validator.ValidateUpdate(map[string]interface{}{
		"a": 500,
	}, map[string]interface{}{
		"a": 500,
	}).IsValid())
	require.True(t, validator.ValidateUpdate(map[string]interface{}{
		"a": 501,
	}, map[string]interface{}{
		"a": 501,
		"b": 500,
	}).IsValid())
	require.False(t, validator.ValidateUpdate(map[string]interface{}{
		"a": 500,
	}, map[string]interface{}{
		"a": 501,
	}).IsValid())
}
