package client

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

const (
	defaultWaitTimeout = 4 * time.Hour
	maxRetryAttempts   = 3
	initialBackoff     = 2 * time.Second
	maxBackoff         = 30 * time.Second
	backoffMultiplier  = 2
	jitterFactor       = 0.5 // ±50% randomization for equal jitter
)

type waitFunc func(ctx context.Context) (bool, *apierrors.APIError)

// applyJitter applies equal jitter to the backoff duration to prevent thundering herd.
// Equal jitter formula: (backoff / 2) + random(0, backoff / 2)
// This provides a balance between predictability and distribution of retry attempts.
func applyJitter(backoff time.Duration) time.Duration {
	if backoff <= 0 {
		return 0
	}

	half := backoff / 2
	if half <= 0 {
		return backoff
	}

	jitter := time.Duration(rand.Int63n(int64(half)))
	return half + jitter
}

func wait(ctx context.Context, f waitFunc, timeout *time.Duration) *apierrors.APIError {
	if timeout == nil {
		timeout = toDurationPointer(defaultWaitTimeout)
	}

	// Run the function once before waiting, with retry logic for transient errors
	ok, apiErr := retryOnTransientError(ctx, f)
	if apiErr != nil {
		return apiErr
	}
	if ok {
		return nil
	}

	ticker := time.NewTicker(10 * time.Second)
	timeoutTicker := time.NewTicker(*timeout)

	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case <-ticker.C:
			ok, apiErr := retryOnTransientError(ctx, f)
			if apiErr != nil {
				return apiErr
			}
			if ok {
				return nil
			}
		}
	}
}

// retryOnTransientError retries a waitFunc with exponential backoff if it encounters transient errors
func retryOnTransientError(ctx context.Context, f waitFunc) (bool, *apierrors.APIError) {
	var lastErr *apierrors.APIError
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		ok, apiErr := f(ctx)

		// Success case
		if apiErr == nil {
			return ok, nil
		}

		// If error is not retryable, return immediately
		if !apierrors.IsRetryable(apiErr) {
			return false, apiErr
		}

		lastErr = apiErr

		// Don't sleep after the last attempt
		if attempt < maxRetryAttempts-1 {
			// Apply jitter to prevent thundering herd problem
			backoffWithJitter := applyJitter(backoff)

			select {
			case <-ctx.Done():
				return false, lastErr
			case <-time.After(backoffWithJitter):
				// Calculate next backoff with exponential growth
				backoff = backoff * backoffMultiplier
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}

	// All retries exhausted
	return false, lastErr
}

func newApplicationStatusCheckerWaitFunc(client *Client, applicationID string, expected qovery.StateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		maxRetry := 5
		var status *qovery.Status
		var apiErr *apierrors.APIError
		for tryCount := 0; tryCount < maxRetry; tryCount++ {
			status, apiErr = client.getApplicationStatus(ctx, applicationID)
			if apiErr != nil {
				if apierrors.IsNotFound(apiErr) && expected == qovery.STATEENUM_DELETED {
					return true, nil
				}
				return false, apiErr
			}
			isExpectedState := status.State == expected
			if !isExpectedState && isEnvFinalState(status.State) {
				time.Sleep(5 * time.Second)
				continue
			}
			return isExpectedState, nil
		}
		return false, apierrors.NewDeployError(apierrors.APIResourceApplication, applicationID, nil, fmt.Errorf("expected status '%s' but got '%s'", expected, status.State))
	}
}

func newApplicationFinalStateCheckerWaitFunc(client *Client, applicationID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getApplicationStatus(ctx, applicationID)
		if apiErr != nil {
			return false, apiErr
		}
		return isEnvFinalState(status.State), nil
	}
}

func newClusterStatusCheckerWaitFunc(client *Client, organizationID string, clusterID string, expected qovery.ClusterStateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getClusterStatus(ctx, organizationID, clusterID)
		if apiErr != nil {
			if (apierrors.IsBadRequest(apiErr) || apierrors.IsNotFound(apiErr)) && expected == qovery.CLUSTERSTATEENUM_DELETED {
				return true, nil
			}
			return false, apiErr
		}
		return status.GetStatus() == expected || isStatusError(status.GetStatus()), nil
	}
}

func newClusterFinalStateCheckerWaitFunc(client *Client, organizationID string, clusterID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getClusterStatus(ctx, organizationID, clusterID)
		if apiErr != nil {
			return false, apiErr
		}
		return isFinalState(status.GetStatus()), nil
	}
}

func newDatabaseStatusCheckerWaitFunc(client *Client, databaseID string, expected qovery.StateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		maxRetry := 5
		var status *qovery.Status
		var apiErr *apierrors.APIError
		for tryCount := 0; tryCount < maxRetry; tryCount++ {
			status, apiErr = client.getDatabaseStatus(ctx, databaseID)
			if apiErr != nil {
				if apierrors.IsNotFound(apiErr) && expected == qovery.STATEENUM_DELETED {
					return true, nil
				}
				return false, apiErr
			}
			isExpectedState := status.State == expected
			if !isExpectedState && isEnvFinalState(status.State) {
				time.Sleep(5 * time.Second)
				continue
			}
			return isExpectedState, nil
		}
		return false, apierrors.NewDeployError(apierrors.APIResourceDatabase, databaseID, nil, fmt.Errorf("expected status '%s' but got '%s'", expected, status.State))
	}
}

func newDatabaseFinalStateCheckerWaitFunc(client *Client, databaseID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getDatabaseStatus(ctx, databaseID)
		if apiErr != nil {
			return false, apiErr
		}
		return isEnvFinalState(status.State), nil
	}
}

func newEnvironmentFinalStateCheckerWaitFunc(client *Client, environmentID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getEnvironmentStatus(ctx, environmentID)
		if apiErr != nil {
			return false, apiErr
		}
		return isEnvFinalState(status.State), nil
	}
}

func isEnvFinalState(state qovery.StateEnum) bool {
	return !isEnvProcessingState(state) &&
		!isEnvWaitingState(state) &&
		!isEnvQueuedState(state)
}

func isEnvProcessingState(state qovery.StateEnum) bool {
	return strings.HasSuffix(string(state), "ING")
}

func isEnvWaitingState(state qovery.StateEnum) bool {
	return strings.Contains(string(state), "_WAITING")
}

func isEnvQueuedState(state qovery.StateEnum) bool {
	return strings.Contains(string(state), "_QUEUED")
}

func isFinalState(state qovery.ClusterStateEnum) bool {
	return !isProcessingState(state) &&
		!isWaitingState(state) &&
		!isQueuedState(state)
}

func isStatusError(state qovery.ClusterStateEnum) bool {
	return strings.HasSuffix(string(state), "_ERROR")
}

func isProcessingState(state qovery.ClusterStateEnum) bool {
	return strings.HasSuffix(string(state), "ING")
}

func isWaitingState(state qovery.ClusterStateEnum) bool {
	return strings.Contains(string(state), "_WAITING")
}

func isQueuedState(state qovery.ClusterStateEnum) bool {
	return strings.Contains(string(state), "_QUEUED")
}

func toDurationPointer(d time.Duration) *time.Duration {
	return &d
}
