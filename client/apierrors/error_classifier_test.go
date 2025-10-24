//go:build unit && !integration

package apierrors

import (
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsTransientError verifies transient error detection
func TestIsTransientError(t *testing.T) {
	testCases := []struct {
		name        string
		apiErr      *APIError
		shouldRetry bool
	}{
		{
			name:        "should return false for nil error",
			apiErr:      nil,
			shouldRetry: false,
		},
		{
			name: "should detect 500 server error as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 500},
			},
			shouldRetry: true,
		},
		{
			name: "should detect 502 bad gateway as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 502},
			},
			shouldRetry: true,
		},
		{
			name: "should detect 503 service unavailable as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 503},
			},
			shouldRetry: true,
		},
		{
			name: "should detect 429 rate limit as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 429},
			},
			shouldRetry: true,
		},
		{
			name: "should not treat 404 as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 404},
			},
			shouldRetry: false,
		},
		{
			name: "should not treat 400 as transient",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 400},
			},
			shouldRetry: false,
		},
		{
			name: "should detect EOF as transient",
			apiErr: &APIError{
				err: io.EOF,
			},
			shouldRetry: true,
		},
		{
			name: "should detect unexpected EOF as transient",
			apiErr: &APIError{
				err: io.ErrUnexpectedEOF,
			},
			shouldRetry: true,
		},
		{
			name: "should detect connection refused as transient",
			apiErr: &APIError{
				err: syscall.ECONNREFUSED,
			},
			shouldRetry: true,
		},
		{
			name: "should detect connection reset as transient",
			apiErr: &APIError{
				err: syscall.ECONNRESET,
			},
			shouldRetry: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsTransientError(tc.apiErr)
			assert.Equal(t, tc.shouldRetry, result)
		})
	}
}

// TestIsRetryable verifies retryable error detection
func TestIsRetryable(t *testing.T) {
	testCases := []struct {
		name        string
		apiErr      *APIError
		shouldRetry bool
	}{
		{
			name:        "should return false for nil error",
			apiErr:      nil,
			shouldRetry: false,
		},
		{
			name: "should retry 500 errors",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 500},
			},
			shouldRetry: true,
		},
		{
			name: "should retry 429 rate limit",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 429},
			},
			shouldRetry: true,
		},
		{
			name: "should not retry 404 not found",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 404},
			},
			shouldRetry: false,
		},
		{
			name: "should not retry 400 bad request",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 400},
			},
			shouldRetry: false,
		},
		{
			name: "should not retry 403 forbidden",
			apiErr: &APIError{
				res: &http.Response{StatusCode: 403},
			},
			shouldRetry: false,
		},
		{
			name: "should retry EOF errors",
			apiErr: &APIError{
				err: io.EOF,
			},
			shouldRetry: true,
		},
		{
			name: "should retry connection errors",
			apiErr: &APIError{
				err: syscall.ECONNRESET,
			},
			shouldRetry: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRetryable(tc.apiErr)
			assert.Equal(t, tc.shouldRetry, result)
		})
	}
}

// TestIsNetworkError verifies network error detection
func TestIsNetworkError(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		isNetwork bool
	}{
		{
			name:      "should return false for nil error",
			err:       nil,
			isNetwork: false,
		},
		{
			name:      "should detect EOF as network error",
			err:       io.EOF,
			isNetwork: true,
		},
		{
			name:      "should detect unexpected EOF as network error",
			err:       io.ErrUnexpectedEOF,
			isNetwork: true,
		},
		{
			name:      "should detect connection refused as network error",
			err:       syscall.ECONNREFUSED,
			isNetwork: true,
		},
		{
			name:      "should detect connection reset as network error",
			err:       syscall.ECONNRESET,
			isNetwork: true,
		},
		{
			name:      "should detect broken pipe as network error",
			err:       syscall.EPIPE,
			isNetwork: true,
		},
		{
			name: "should detect net.OpError as network error",
			err: &net.OpError{
				Op:  "dial",
				Err: errors.New("connection refused"),
			},
			isNetwork: true,
		},
		{
			name:      "should not detect generic errors as network error",
			err:       errors.New("some other error"),
			isNetwork: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNetworkError(tc.err)
			assert.Equal(t, tc.isNetwork, result)
		})
	}
}
