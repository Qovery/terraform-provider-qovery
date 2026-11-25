//go:build unit && !integration

package apierrors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		_, res, err := client.ApplicationCustomDomainAPI.
			CreateApplicationCustomDomain(context.Background(), "application-id").
			CustomDomainRequest(qovery.CustomDomainRequest{Domain: "foo.example.com"}).
			Execute()
		require.Error(t, err)

		apiErr := NewCreateError(APIResourceApplicationCustomDomain, "foo.example.com", res, err)

		assert.Equal(t, "Could not create application custom domain 'foo.example.com', unexpected error: 400 Bad Request - domain foo.example.com is already used", apiErr.Detail())
	})

	t.Run("payment_required_without_title", func(t *testing.T) {
		t.Parallel()

		client := newAPIClientAnswering(t, http.StatusPaymentRequired, `{"status":402,"detail":"Your plan allows 1 cluster."}`)
		_, res, err := client.ClustersAPI.
			CreateCluster(context.Background(), "organization-id").
			ClusterRequest(qovery.ClusterRequest{Name: "my-cluster", Region: "eu-west-3", CloudProvider: qovery.CLOUDVENDORENUM_AWS}).
			Execute()
		require.Error(t, err)

		apiErr := NewCreateError(APIResourceCluster, "my-cluster", res, err)

		assert.Equal(t, "Could not create cluster 'my-cluster', unexpected error: 402 Payment Required - Your plan allows 1 cluster.", apiErr.Detail())
	})
}

func TestNewContextError(t *testing.T) {
	t.Parallel()

	apiErr := NewContextError(context.Canceled)
	assert.ErrorIs(t, apiErr, context.Canceled)
	assert.Equal(t, "Could not read, unexpected error: context canceled", apiErr.Error())
	assert.Equal(t, "Error on read", apiErr.Summary())
}

// TestAPIError_BufferedBody verifies that response body is buffered and can be read multiple times
func TestAPIError_BufferedBody(t *testing.T) {
	testCases := []struct {
		name           string
		payload        *errorPayload
		expectedDetail string
	}{
		{
			name: "should buffer and parse error payload from response body",
			payload: &errorPayload{
				Status:  400,
				Message: "Invalid request",
			},
			expectedDetail: "Invalid request",
		},
		{
			name: "should handle empty error message",
			payload: &errorPayload{
				Status:  500,
				Message: "",
			},
			expectedDetail: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a response with JSON body
			bodyJSON, err := json.Marshal(tc.payload)
			assert.NoError(t, err)

			res := &http.Response{
				StatusCode: tc.payload.Status,
				Body:       io.NopCloser(bytes.NewReader(bodyJSON)),
			}

			// Create the error (this should buffer the body)
			apiErr := NewError(APIActionRead, APIResourceCluster, "test-id", res, errors.New("test error"))

			// Verify body was buffered
			assert.NotNil(t, apiErr.bufferedBody)
			assert.Equal(t, bodyJSON, apiErr.bufferedBody)

			// Call errorPayload multiple times to verify it can be read repeatedly
			payload1 := apiErr.errorPayload()
			payload2 := apiErr.errorPayload()
			payload3 := apiErr.errorPayload()

			// All should return the same result
			assert.NotNil(t, payload1)
			assert.Equal(t, tc.expectedDetail, payload1.Message)
			assert.Equal(t, payload1, payload2)
			assert.Equal(t, payload1, payload3)
		})
	}
}

// TestAPIError_NilResponseBody verifies handling when response body is nil
func TestAPIError_NilResponseBody(t *testing.T) {
	res := &http.Response{
		StatusCode: 500,
		Body:       nil,
	}

	apiErr := NewError(APIActionRead, APIResourceCluster, "test-id", res, errors.New("test error"))

	assert.Nil(t, apiErr.bufferedBody)
	payload := apiErr.errorPayload()
	assert.Nil(t, payload)
}

// TestAPIError_InvalidJSON verifies handling of invalid JSON in response body
func TestAPIError_InvalidJSON(t *testing.T) {
	res := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewReader([]byte("not valid json"))),
	}

	apiErr := NewError(APIActionRead, APIResourceCluster, "test-id", res, errors.New("test error"))

	// Body should be buffered even if it's invalid JSON
	assert.NotNil(t, apiErr.bufferedBody)
	assert.Equal(t, []byte("not valid json"), apiErr.bufferedBody)

	// errorPayload should return nil for invalid JSON
	payload := apiErr.errorPayload()
	assert.Nil(t, payload)
}

// TestAPIError_EmptyBody verifies handling of empty response body
func TestAPIError_EmptyBody(t *testing.T) {
	res := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	apiErr := NewError(APIActionRead, APIResourceCluster, "test-id", res, errors.New("test error"))

	assert.NotNil(t, apiErr.bufferedBody)
	assert.Equal(t, []byte(""), apiErr.bufferedBody)

	payload := apiErr.errorPayload()
	assert.Nil(t, payload)
}

// TestAPIError_DetailMessage verifies that Detail() works correctly with buffered body
func TestAPIError_DetailMessage(t *testing.T) {
	payload := &errorPayload{
		Status:  400,
		Message: "Cluster not found",
	}
	bodyJSON, _ := json.Marshal(payload)

	res := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewReader(bodyJSON)),
	}

	apiErr := NewError(APIActionRead, APIResourceCluster, "cluster-123", res, errors.New("request failed"))

	// First call to Detail()
	detail1 := apiErr.Detail()
	assert.Contains(t, detail1, "Could not read cluster 'cluster-123'")
	assert.Contains(t, detail1, "request failed")
	assert.Contains(t, detail1, "Cluster not found")

	// Second call should work the same (tests that buffered body can be read multiple times)
	detail2 := apiErr.Detail()
	assert.Equal(t, detail1, detail2)
}

// TestAPIError_NoError verifies behavior when error is nil
func TestAPIError_NoError(t *testing.T) {
	res := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"status": 404, "detail": "Not found"}`))),
	}

	apiErr := NewError(APIActionRead, APIResourceCluster, "test-id", res, nil)

	// errorPayload should return nil when err is nil
	payload := apiErr.errorPayload()
	assert.Nil(t, payload)

	// Detail should still work
	detail := apiErr.Detail()
	assert.Contains(t, detail, "unexpected status code: 404")
}
