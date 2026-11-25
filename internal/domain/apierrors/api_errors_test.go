//go:build unit && !integration
// +build unit,!integration

package apierrors

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAPIClientAnswering returns a client whose every call gets status and body.
func newAPIClientAnswering(t *testing.T, status int, body string) *qovery.APIClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	return qovery.NewAPIClient(cfg)
}

// TestAPIError_DetailOfAnErrorStatus guards the text of an error status against the client
// formatting typed problem bodies into its error.
func TestAPIError_DetailOfAnErrorStatus(t *testing.T) {
	t.Parallel()

	t.Run("custom_domain_create_rejected", func(t *testing.T) {
		t.Parallel()

		client := newAPIClientAnswering(t, http.StatusBadRequest, `{"status":400,"detail":"domain foo.example.com is already used"}`)
		_, resp, err := client.ContainerCustomDomainAPI.
			CreateContainerCustomDomain(context.Background(), "container-id").
			CustomDomainRequest(qovery.CustomDomainRequest{Domain: "foo.example.com"}).
			Execute()
		require.Error(t, err)

		apiErr := NewCreateAPIError(APIResourceContainerCustomDomain, "my-container", resp, err)

		assert.Equal(t, "Could not create container custom domain 'my-container', unexpected error: 400 Bad Request - domain foo.example.com is already used", apiErr.Detail())
	})

	t.Run("payment_required_without_title", func(t *testing.T) {
		t.Parallel()

		client := newAPIClientAnswering(t, http.StatusPaymentRequired, `{"status":402,"detail":"Your plan allows 1 environment."}`)
		_, resp, err := client.EnvironmentsAPI.
			CreateEnvironment(context.Background(), "project-id").
			CreateEnvironmentRequest(qovery.CreateEnvironmentRequest{Name: "my-environment"}).
			Execute()
		require.Error(t, err)

		apiErr := NewCreateAPIError(APIResourceEnvironment, "my-environment", resp, err)

		assert.Equal(t, "Could not create environment 'my-environment', unexpected error: 402 Payment Required - Your plan allows 1 environment.", apiErr.Detail())
	})
}

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
