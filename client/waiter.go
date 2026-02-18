package client

import (
	"context"
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
		timeout = new(defaultWaitTimeout)
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
			return apierrors.NewTimeoutError(*timeout)
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

	for attempt := range maxRetryAttempts {
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
				backoff = min(backoff*backoffMultiplier, maxBackoff)
			}
		}
	}

	// All retries exhausted
	return false, lastErr
}

func newApplicationStatusCheckerWaitFunc(client *Client, applicationID string, expected qovery.StateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getApplicationStatus(ctx, applicationID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == qovery.STATEENUM_DELETED {
				return true, nil
			}
			return false, apiErr
		}

		// Check if reached expected state
		if status.State == expected {
			return true, nil
		}

		// Check if in terminal error state
		if isEnvErrorState(status.State) {
			return false, apierrors.NewUnexpectedStateError(
				apierrors.APIResourceApplication,
				applicationID,
				expected,
				status.State,
			)
		}

		// Still in progress, continue waiting
		return false, nil
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

		currentState := status.GetStatus()

		// Check if reached expected state
		if currentState == expected {
			return true, nil
		}

		// Check if in terminal error state (and we're not expecting an error state)
		if isClusterErrorState(currentState) && !isClusterErrorState(expected) {
			return false, apierrors.NewUnexpectedClusterStateError(
				organizationID,
				clusterID,
				expected,
				currentState,
			)
		}

		// Still in progress, continue waiting
		return false, nil
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
		status, apiErr := client.getDatabaseStatus(ctx, databaseID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == qovery.STATEENUM_DELETED {
				return true, nil
			}
			return false, apiErr
		}

		// Check if reached expected state
		if status.State == expected {
			return true, nil
		}

		// Check if in terminal error state
		if isEnvErrorState(status.State) {
			return false, apierrors.NewUnexpectedStateError(
				apierrors.APIResourceDatabase,
				databaseID,
				expected,
				status.State,
			)
		}

		// Still in progress, continue waiting
		return false, nil
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

// isEnvErrorState checks if the state indicates a terminal error
func isEnvErrorState(state qovery.StateEnum) bool {
	stateStr := string(state)
	return strings.HasSuffix(stateStr, "_ERROR") ||
		strings.Contains(stateStr, "FAILED") ||
		strings.Contains(stateStr, "ERROR")
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

// isClusterErrorState checks if cluster state indicates a terminal error
func isClusterErrorState(state qovery.ClusterStateEnum) bool {
	return strings.HasSuffix(string(state), "_ERROR")
}

