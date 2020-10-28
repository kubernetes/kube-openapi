package validate

import (
	"reflect"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"github.com/stretchr/testify/assert"
)

// Validator for string formats
func TestFormatValidator_EdgeCases(t *testing.T) {
	// Apply
	v := formatValidator{
		KnownFormats: strfmt.Default,
	}

	// formatValidator applies to: Items, Parameter,Schema

	p := spec.Parameter{}
	p.Typed(stringType, "email")
	s := spec.Schema{}
	s.Typed(stringType, "uuid")
	i := spec.Items{}
	i.Typed(stringType, "datetime")

	sources := []interface{}{&p, &s, &i}

	for _, source := range sources {
		// Default formats for strings
		assert.True(t, v.Applies(source, reflect.String))
		// Do not apply for number formats
		assert.False(t, v.Applies(source, reflect.Int))
	}

	assert.False(t, v.Applies("A string", reflect.String))
	assert.False(t, v.Applies(nil, reflect.String))
}
