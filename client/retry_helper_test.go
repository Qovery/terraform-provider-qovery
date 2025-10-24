//go:build unit && !integration

package client

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// TestRetryAPICall_Success verifies successful API call without retry
func TestRetryAPICall_Success(t *testing.T) {
	callCount := 0
	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		return nil
	}

	err := retryAPICall(context.Background(), testFunc)

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount, "should only call once on success")
}

// TestRetryAPICall_DNSError verifies retry on DNS lookup failure
func TestRetryAPICall_DNSError(t *testing.T) {
	callCount := 0
	dnsErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{
			Err:        "no such host",
			Name:       "api.qovery.com",
			IsNotFound: true,
		},
	}

	apiErr := apierrors.NewReadError(
		apierrors.APIResourceClusterStatus,
		"test-cluster-id",
		&http.Response{StatusCode: 0},
		dnsErr,
	)

	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		if callCount < 2 {
			// Fail first time with DNS error
			return apiErr
		}
		// Succeed on second attempt
		return nil
	}

	startTime := time.Now()
	err := retryAPICall(context.Background(), testFunc)
	elapsed := time.Since(startTime)

	assert.Nil(t, err, "should succeed after retry")
	assert.Equal(t, 2, callCount, "should retry once after DNS error")
	// Should have waited at least half the initial backoff due to jitter
	assert.GreaterOrEqual(t, elapsed, initialBackoff/2,
		"should have applied backoff between retries")
}

// TestRetryAPICall_NonRetryableError verifies immediate return on non-retryable error
func TestRetryAPICall_NonRetryableError(t *testing.T) {
	callCount := 0
	notFoundErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 404},
		nil,
	)

	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		return notFoundErr
	}

	err := retryAPICall(context.Background(), testFunc)

	assert.NotNil(t, err)
	assert.Equal(t, 1, callCount, "should not retry non-retryable errors like 404")
}

// TestRetryAPICall_TransientErrorExhaustsRetries verifies max retry limit
func TestRetryAPICall_TransientErrorExhaustsRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry exhaustion test in short mode")
	}

	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceClusterStatus,
		"test-id",
		&http.Response{StatusCode: 503},
		nil,
	)

	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		// Always return server error (retryable)
		return serverErr
	}

	err := retryAPICall(context.Background(), testFunc)

	assert.NotNil(t, err, "should return error after exhausting retries")
	assert.Equal(t, maxRetryAttempts, callCount,
		"should retry up to max attempts")
}

// TestRetryAPICall_EOFError verifies retry on EOF error
func TestRetryAPICall_EOFError(t *testing.T) {
	callCount := 0
	eofErr := apierrors.NewReadError(
		apierrors.APIResourceClusterStatus,
		"test-id",
		&http.Response{StatusCode: 500},
		io.EOF,
	)

	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		if callCount < 3 {
			// Fail first two times with EOF
			return eofErr
		}
		// Succeed on third attempt
		return nil
	}

	err := retryAPICall(context.Background(), testFunc)

	assert.Nil(t, err, "should succeed after retries")
	assert.Equal(t, 3, callCount, "should retry twice and succeed on third attempt")
}

// TestRetryAPICall_ContextCancellation verifies context cancellation stops retries
func TestRetryAPICall_ContextCancellation(t *testing.T) {
	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceClusterStatus,
		"test-id",
		&http.Response{StatusCode: 500},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())

	testFunc := func(ctx context.Context) *apierrors.APIError {
		callCount++
		if callCount == 1 {
			// Cancel context after first call
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
		}
		return serverErr
	}

	err := retryAPICall(ctx, testFunc)

	assert.NotNil(t, err)
	// Should stop retrying when context is cancelled
	assert.LessOrEqual(t, callCount, maxRetryAttempts,
		"should stop retrying after context cancellation")
}

// TestRetryAPICall_MultipleErrorTypes verifies handling of various transient errors
func TestRetryAPICall_MultipleErrorTypes(t *testing.T) {
	testCases := []struct {
		name          string
		err           *apierrors.APIError
		shouldRetry   bool
		expectedCalls int
	}{
		{
			name: "should retry 500 internal server error",
			err: apierrors.NewReadError(
				apierrors.APIResourceClusterStatus,
				"test-id",
				&http.Response{StatusCode: 500},
				nil,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should retry 502 bad gateway",
			err: apierrors.NewReadError(
				apierrors.APIResourceClusterStatus,
				"test-id",
				&http.Response{StatusCode: 502},
				nil,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should retry 429 rate limit",
			err: apierrors.NewReadError(
				apierrors.APIResourceClusterStatus,
				"test-id",
				&http.Response{StatusCode: 429},
				nil,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should not retry 400 bad request",
			err: apierrors.NewReadError(
				apierrors.APIResourceClusterStatus,
				"test-id",
				&http.Response{StatusCode: 400},
				nil,
			),
			shouldRetry:   false,
			expectedCalls: 1,
		},
		{
			name: "should not retry 403 forbidden",
			err: apierrors.NewReadError(
				apierrors.APIResourceClusterStatus,
				"test-id",
				&http.Response{StatusCode: 403},
				nil,
			),
			shouldRetry:   false,
			expectedCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			testFunc := func(ctx context.Context) *apierrors.APIError {
				callCount++
				return tc.err
			}

			err := retryAPICall(context.Background(), testFunc)

			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedCalls, callCount)
		})
	}
}
