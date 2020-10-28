package conv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
)

func TestDateTimeValue(t *testing.T) {
	assert.Equal(t, strfmt.DateTime{}, DateTimeValue(nil))
	time := strfmt.DateTime(time.Now())
	assert.Equal(t, time, DateTimeValue(&time))
}
