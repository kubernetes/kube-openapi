package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

func ptr[T any](v T) *T {
	return &v
}

var zeroIntSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type:    spec.StringOrArray{"number"},
		Minimum: ptr(float64(0)),
		Maximum: ptr(float64(0)),
	},
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
		Maximum: ptr(float64(10000)),
	},
}

var largeIntSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type:    spec.StringOrArray{"number"},
		Minimum: ptr(float64(10000)),
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
			"zero":   *zeroIntSchema,
			"small":  *smallIntSchema,
			"medium": *mediumIntSchema,
			"large":  *largeIntSchema,
		},
	},
}

var objectObjectSchema *spec.Schema = &spec.Schema{
	SchemaProps: spec.SchemaProps{
		Type: spec.StringOrArray{"object"},
		Properties: map[string]spec.Schema{
			"nested": *objectSchema,
		},
	},
}

// Shows scalar fields of objects can be ratcheted
func TestObjectScalarFieldsRatcheting(t *testing.T) {
	validator := validate.NewRatchetingSchemaValidator(objectSchema, nil, "", strfmt.Default)
	assert.True(t, validator.ValidateUpdate(map[string]interface{}{
		"small": 500,
	}, map[string]interface{}{
		"small": 500,
	}).IsValid())
	assert.True(t, validator.ValidateUpdate(map[string]interface{}{
		"small": 501,
	}, map[string]interface{}{
		"small":  501,
		"medium": 500,
	}).IsValid())
	assert.False(t, validator.ValidateUpdate(map[string]interface{}{
		"small": 500,
	}, map[string]interface{}{
		"small": 501,
	}).IsValid())
}

// Shows schemas with object fields which themselves are ratcheted can be ratcheted
func TestObjectObjectFieldsRatcheting(t *testing.T) {
	validator := validate.NewRatchetingSchemaValidator(objectObjectSchema, nil, "", strfmt.Default)
	assert.True(t, validator.ValidateUpdate(map[string]interface{}{
		"nested": map[string]interface{}{
			"small": 500,
		}}, map[string]interface{}{
		"nested": map[string]interface{}{
			"small": 500,
		}}).IsValid())
	assert.True(t, validator.ValidateUpdate(map[string]interface{}{
		"nested": map[string]interface{}{
			"small": 501,
		}}, map[string]interface{}{
		"nested": map[string]interface{}{
			"small":  501,
			"medium": 500,
		}}).IsValid())
	assert.False(t, validator.ValidateUpdate(map[string]interface{}{
		"nested": map[string]interface{}{
			"small": 500,
		}}, map[string]interface{}{
		"nested": map[string]interface{}{
			"small": 501,
		}}).IsValid())
}

