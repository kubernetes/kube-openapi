package conv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
)

func TestDateValue(t *testing.T) {
	assert.Equal(t, strfmt.Date{}, DateValue(nil))
	date := strfmt.Date(time.Now())
	assert.Equal(t, date, DateValue(&date))
}
