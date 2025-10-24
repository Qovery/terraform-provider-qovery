//go:build unit && !integration

package apierrors

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
