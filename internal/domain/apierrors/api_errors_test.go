//go:build unit && !integration
// +build unit,!integration

package apierrors

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// A response built without a Body (e.g. NewNotFoundAPIError, or tests) must not make
// Detail()/Error() panic when errorPayload looks for it.
func TestAPIError_Detail_NilResponseBody(t *testing.T) {
	t.Parallel()

	apiErr := NewReadAPIError(
		APIResourceOrganization,
		"some-id",
		&http.Response{StatusCode: http.StatusInternalServerError},
		errors.New("boom"),
	)

	assert.Contains(t, apiErr.Detail(), "boom")
}

func TestAPIError_Detail_NotFoundAPIError(t *testing.T) {
	t.Parallel()

	apiErr := NewNotFoundAPIError(APIResourceOrganization, "some-id")

	assert.Contains(t, apiErr.Detail(), "resource not found")
	assert.True(t, apiErr.IsNotFound())
}

// The response body is buffered at construction, so inspecting the error more than once
// (e.g. IsNotFound on the 400 branch, then Error for a diagnostic) must keep the payload.
func TestAPIError_Detail_BodyReadableMultipleTimes(t *testing.T) {
	t.Parallel()

	apiErr := NewReadAPIError(
		APIResourceOrganization,
		"some-id",
		&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`{"status":400,"detail":"organization does not exist"}`)),
		},
		errors.New("boom"),
	)

	assert.True(t, apiErr.IsNotFound())
	assert.Contains(t, apiErr.Error(), "organization does not exist")
}
