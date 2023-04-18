package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

const defaultWaitTimeout = 4 * time.Hour

type waitFunc func(ctx context.Context) (bool, *apierrors.APIError)

func wait(ctx context.Context, f waitFunc, timeout *time.Duration) *apierrors.APIError {
	if timeout == nil {
		timeout = toDurationPointer(defaultWaitTimeout)
	}

	// Run the function once before waiting
	ok, apiErr := f(ctx)
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
			if !isExpectedState && isFinalState(status.State) {
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
		return isFinalState(status.State), nil
	}
}

func newClusterStatusCheckerWaitFunc(client *Client, organizationID string, clusterID string, expected qovery.StateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getClusterStatus(ctx, organizationID, clusterID)
		if apiErr != nil {
			if (apierrors.IsBadRequest(apiErr) || apierrors.IsNotFound(apiErr)) && expected == qovery.STATEENUM_DELETED {
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
			if !isExpectedState && isFinalState(status.State) {
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
		return isFinalState(status.State), nil
	}
}

func newEnvironmentStatusCheckerWaitFunc(client *Client, environmentID string, expected qovery.StateEnum) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getEnvironmentStatus(ctx, environmentID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == qovery.STATEENUM_DELETED {
				return true, nil
			}
			return false, apiErr
		}
		isExpectedState := status.State == expected
		if !isExpectedState && isFinalState(status.State) {
			return false, apierrors.NewDeployError(apierrors.APIResourceEnvironment, environmentID, nil, fmt.Errorf("expected status '%s' but got '%s'", expected, status.State))
		}
		return status.State == expected, nil
	}
}

func newEnvironmentFinalStateCheckerWaitFunc(client *Client, environmentID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getEnvironmentStatus(ctx, environmentID)
		if apiErr != nil {
			return false, apiErr
		}
		return isFinalState(status.State), nil
	}
}

func isFinalState(state qovery.StateEnum) bool {
	return !isProcessingState(state) &&
		!isWaitingState(state) &&
		!isQueuedState(state)
}

func isStatusError(state qovery.StateEnum) bool {
	return strings.HasSuffix(string(state), "_ERROR")
}

func isProcessingState(state qovery.StateEnum) bool {
	return strings.HasSuffix(string(state), "ING")
}

func isWaitingState(state qovery.StateEnum) bool {
	return strings.Contains(string(state), "_WAITING")
}

func isQueuedState(state qovery.StateEnum) bool {
	return strings.Contains(string(state), "_QUEUED")
}

func toDurationPointer(d time.Duration) *time.Duration {
	return &d
}
