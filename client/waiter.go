package client

import (
	"context"
	"strings"
	"time"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

const (
	defaultWaitTimeout      = 4 * time.Hour
	defaultWaitPollInterval = 10 * time.Second
)

type waitFunc func(ctx context.Context) (bool, *apierrors.APIError)

func wait(ctx context.Context, f waitFunc) *apierrors.APIError {
	return waitWithTimeout(ctx, f, defaultWaitTimeout, defaultWaitPollInterval)
}

// waitWithTimeout polls f every pollInterval until it reports ok, returns an error, the context
// is done, or timeout elapses. Running out of time is a failure, never a silent success.
// f is called once per poll: the HTTP client already retries transient API failures.
func waitWithTimeout(ctx context.Context, f waitFunc, timeout, pollInterval time.Duration) *apierrors.APIError {
	// Run the function once before waiting
	ok, apiErr := f(ctx)
	if apiErr != nil {
		return apiErr
	}
	if ok {
		return nil
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return apierrors.NewContextError(ctx.Err())
		case <-timeoutTimer.C:
			return apierrors.NewTimeoutError(timeout)
		case <-ticker.C:
			ok, apiErr := f(ctx)
			if apiErr != nil {
				return apiErr
			}
			if ok {
				return nil
			}
		}
	}
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
