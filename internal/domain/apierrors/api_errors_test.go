//go:build unit && !integration
// +build unit,!integration

package apierrors

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// A response built without a Body (e.g. NewNotFoundAPIError, or tests) must not make
// Detail()/Error() panic when errorPayload tries to read it.
func TestAPIError_Detail_NilResponseBody(t *testing.T) {
	t.Parallel()

	apiErr := NewReadAPIError(
		APIResourceOrganization,
		"some-id",
		&http.Response{StatusCode: http.StatusInternalServerError},
		errors.New("boom"),
	)

	assert.NotPanics(t, func() {
		assert.Contains(t, apiErr.Detail(), "boom")
	})
}

func TestAPIError_Detail_NotFoundAPIError(t *testing.T) {
	t.Parallel()

	apiErr := NewNotFoundAPIError(APIResourceOrganization, "some-id")

	assert.NotPanics(t, func() {
		assert.Contains(t, apiErr.Detail(), "resource not found")
	})
	assert.True(t, apiErr.IsNotFound())
}
