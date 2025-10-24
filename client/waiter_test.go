//go:build unit && !integration

package client

import (
	"context"
	"io"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// TestRetryOnTransientError_Success verifies successful execution without retry
func TestRetryOnTransientError_Success(t *testing.T) {
	callCount := 0
	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		return true, nil
	}

	ok, err := retryOnTransientError(context.Background(), testFunc)

	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, 1, callCount, "should only call once on success")
}

// TestRetryOnTransientError_NonRetryableError verifies immediate return on non-retryable error
func TestRetryOnTransientError_NonRetryableError(t *testing.T) {
	callCount := 0
	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		return false, &apierrors.APIError{
			// 404 is not retryable
		}
	}

	// Mock the response to make it non-retryable
	notFoundErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 404},
		nil,
	)

	callCount = 0
	testFunc = func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		return false, notFoundErr
	}

	ok, err := retryOnTransientError(context.Background(), testFunc)

	assert.False(t, ok)
	assert.NotNil(t, err)
	assert.Equal(t, 1, callCount, "should not retry non-retryable errors")
}

// TestRetryOnTransientError_TransientErrorWithRetry verifies retry on transient errors
func TestRetryOnTransientError_TransientErrorWithRetry(t *testing.T) {
	callCount := 0
	eofErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 500},
		io.EOF,
	)

	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		if callCount < 2 {
			// Fail first time with transient error
			return false, eofErr
		}
		// Succeed on second attempt
		return true, nil
	}

	startTime := time.Now()
	ok, err := retryOnTransientError(context.Background(), testFunc)
	elapsed := time.Since(startTime)

	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, 2, callCount, "should retry once and succeed")
	// With jitter, should have waited at least half the initial backoff (1s minimum due to equal jitter)
	assert.GreaterOrEqual(t, elapsed, initialBackoff/2)
}

// TestRetryOnTransientError_MaxRetriesExceeded verifies max retry limit
func TestRetryOnTransientError_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 500},
		nil,
	)

	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		// Always return server error (retryable)
		return false, serverErr
	}

	ok, err := retryOnTransientError(context.Background(), testFunc)

	assert.False(t, ok)
	assert.NotNil(t, err)
	assert.Equal(t, maxRetryAttempts, callCount, "should retry up to max attempts")
}

// TestRetryOnTransientError_ContextCancellation verifies context cancellation
func TestRetryOnTransientError_ContextCancellation(t *testing.T) {
	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 500},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())

	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		if callCount == 1 {
			// Cancel context after first call
			go func() {
				time.Sleep(100 * time.Millisecond)
				cancel()
			}()
		}
		return false, serverErr
	}

	ok, err := retryOnTransientError(ctx, testFunc)

	assert.False(t, ok)
	assert.NotNil(t, err)
	// Should stop retrying when context is cancelled
	assert.LessOrEqual(t, callCount, maxRetryAttempts)
}

// TestRetryOnTransientError_ExponentialBackoff verifies exponential backoff timing
func TestRetryOnTransientError_ExponentialBackoff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exponential backoff timing test in short mode")
	}

	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 503},
		nil,
	)

	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		// Always fail to test all retry attempts
		return false, serverErr
	}

	startTime := time.Now()
	retryOnTransientError(context.Background(), testFunc)
	elapsed := time.Since(startTime)

	// With 3 attempts, we should have 2 backoffs with jitter applied:
	// First backoff: 2s with jitter = 1-2s
	// Second backoff: 4s with jitter = 2-4s
	// Minimum total: 1s + 2s = 3s (with jitter at minimum)
	// Maximum total: 2s + 4s = 6s (with jitter at maximum)
	expectedMinDuration := (initialBackoff / 2) + ((initialBackoff * backoffMultiplier) / 2)
	assert.GreaterOrEqual(t, elapsed, expectedMinDuration,
		"total elapsed time should include exponential backoff delays with jitter")
}

// TestRetryOnTransientError_DifferentErrorTypes verifies handling of different error types
func TestRetryOnTransientError_DifferentErrorTypes(t *testing.T) {
	testCases := []struct {
		name          string
		err           *apierrors.APIError
		shouldRetry   bool
		expectedCalls int
	}{
		{
			name: "should retry EOF error",
			err: apierrors.NewReadError(
				apierrors.APIResourceCluster,
				"test-id",
				&http.Response{StatusCode: 500},
				io.EOF,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should retry 502 bad gateway",
			err: apierrors.NewReadError(
				apierrors.APIResourceCluster,
				"test-id",
				&http.Response{StatusCode: 502},
				nil,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should retry connection reset",
			err: apierrors.NewReadError(
				apierrors.APIResourceCluster,
				"test-id",
				&http.Response{StatusCode: 500},
				syscall.ECONNRESET,
			),
			shouldRetry:   true,
			expectedCalls: maxRetryAttempts,
		},
		{
			name: "should not retry 404 not found",
			err: apierrors.NewReadError(
				apierrors.APIResourceCluster,
				"test-id",
				&http.Response{StatusCode: 404},
				nil,
			),
			shouldRetry:   false,
			expectedCalls: 1,
		},
		{
			name: "should not retry 400 bad request",
			err: apierrors.NewReadError(
				apierrors.APIResourceCluster,
				"test-id",
				&http.Response{StatusCode: 400},
				nil,
			),
			shouldRetry:   false,
			expectedCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
				callCount++
				return false, tc.err
			}

			ok, err := retryOnTransientError(context.Background(), testFunc)

			assert.False(t, ok)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedCalls, callCount)
		})
	}
}

// TestApplyJitter verifies that jitter is applied correctly
func TestApplyJitter(t *testing.T) {
	testCases := []struct {
		name    string
		backoff time.Duration
	}{
		{
			name:    "should apply jitter to 2 second backoff",
			backoff: 2 * time.Second,
		},
		{
			name:    "should apply jitter to 4 second backoff",
			backoff: 4 * time.Second,
		},
		{
			name:    "should apply jitter to 30 second backoff",
			backoff: 30 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jittered := applyJitter(tc.backoff)

			// Jittered value should be between backoff/2 and backoff (equal jitter)
			minExpected := tc.backoff / 2
			maxExpected := tc.backoff

			assert.GreaterOrEqual(t, jittered, minExpected,
				"jittered backoff should be at least half the original")
			assert.LessOrEqual(t, jittered, maxExpected,
				"jittered backoff should not exceed original backoff")
		})
	}
}

// TestApplyJitter_Distribution verifies that jitter provides good distribution
func TestApplyJitter_Distribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distribution test in short mode")
	}

	backoff := 10 * time.Second
	samples := 1000
	results := make([]time.Duration, samples)

	// Collect many samples
	for i := 0; i < samples; i++ {
		results[i] = applyJitter(backoff)
	}

	// Calculate statistics
	var sum time.Duration
	min := results[0]
	max := results[0]

	for _, r := range results {
		sum += r
		if r < min {
			min = r
		}
		if r > max {
			max = r
		}
	}

	avg := sum / time.Duration(samples)

	// With equal jitter, average should be around 3/4 of backoff (midpoint between backoff/2 and backoff)
	expectedAvg := (backoff / 2) + (backoff / 4) // 75% of backoff
	tolerance := backoff / 10                     // Allow 10% variance

	assert.InDelta(t, int64(expectedAvg), int64(avg), float64(tolerance),
		"average jittered backoff should be around 75%% of original backoff")

	// Verify we're getting good distribution
	assert.GreaterOrEqual(t, min, backoff/2,
		"minimum should be at least half the backoff")
	assert.LessOrEqual(t, max, backoff,
		"maximum should not exceed backoff")

	// Ensure we're not getting the same value repeatedly (would indicate broken randomness)
	assert.Greater(t, max-min, backoff/4,
		"range of jittered values should be substantial")
}

// TestRetryOnTransientError_JitterIsApplied verifies that jitter varies backoff between retries
func TestRetryOnTransientError_JitterIsApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping jitter timing test in short mode")
	}

	callCount := 0
	serverErr := apierrors.NewReadError(
		apierrors.APIResourceCluster,
		"test-id",
		&http.Response{StatusCode: 503},
		nil,
	)

	testFunc := func(ctx context.Context) (bool, *apierrors.APIError) {
		callCount++
		// Always fail to test all retry attempts
		return false, serverErr
	}

	// Run multiple times to verify jitter causes timing variance
	runs := 5
	durations := make([]time.Duration, runs)

	for i := 0; i < runs; i++ {
		callCount = 0
		startTime := time.Now()
		retryOnTransientError(context.Background(), testFunc)
		durations[i] = time.Since(startTime)
	}

	// Check that we don't get the exact same duration every time (which would indicate no jitter)
	allSame := true
	firstDuration := durations[0]
	for _, d := range durations[1:] {
		// Allow 100ms tolerance for timing precision
		if d < firstDuration-100*time.Millisecond || d > firstDuration+100*time.Millisecond {
			allSame = false
			break
		}
	}

	assert.False(t, allSame,
		"retry durations should vary due to jitter (not all identical)")
}

// TestApplyJitter_ZeroBackoff verifies handling of edge case with zero backoff
func TestApplyJitter_ZeroBackoff(t *testing.T) {
	jittered := applyJitter(0)
	assert.Equal(t, time.Duration(0), jittered,
		"jitter on zero backoff should return zero")
}

// TestApplyJitter_SmallBackoff verifies handling of very small backoff values
func TestApplyJitter_SmallBackoff(t *testing.T) {
	backoff := 10 * time.Millisecond
	jittered := applyJitter(backoff)

	// Should still respect the equal jitter bounds
	assert.GreaterOrEqual(t, jittered, backoff/2)
	assert.LessOrEqual(t, jittered, backoff)
}
