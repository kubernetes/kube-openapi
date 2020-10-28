package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
)

func TestDurationValue(t *testing.T) {
	assert.Equal(t, strfmt.Duration(0), DurationValue(nil))
	duration := strfmt.Duration(42)
	assert.Equal(t, duration, DurationValue(&duration))
}
