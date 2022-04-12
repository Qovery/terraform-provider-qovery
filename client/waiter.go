package client

import (
	"context"
	"strings"
	"time"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

const defaultWaitTimeout = 30 * time.Minute

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

func newApplicationStatusCheckerWaitFunc(client *Client, applicationID string, expected string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getApplicationStatus(ctx, applicationID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == "DELETED" {
				return true, nil
			}
			return false, apiErr
		}
		return status.State == expected, nil
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

func newClusterStatusCheckerWaitFunc(client *Client, organizationID string, clusterID string, expected string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getClusterStatus(ctx, organizationID, clusterID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == "DELETED" {
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

func newDatabaseStatusCheckerWaitFunc(client *Client, databaseID string, expected string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getDatabaseStatus(ctx, databaseID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == "DELETED" {
				return true, nil
			}
			return false, apiErr
		}
		return status.State == expected, nil
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

func newEnvironmentStatusCheckerWaitFunc(client *Client, environmentID string, expected string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.getEnvironmentStatus(ctx, environmentID)
		if apiErr != nil {
			if apierrors.IsNotFound(apiErr) && expected == "DELETED" {
				return true, nil
			}
			return false, apiErr
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

func isFinalState(state string) bool {
	return state != "DEPLOYING" &&
		state != "DELETING" &&
		state != "STOPPING" &&
		state != "QUEUED" &&
		!isWaitingState(state)
}

func isWaitingState(state string) bool {
	return strings.HasPrefix(state, "WAITING_")
}

func isStatusError(state string) bool {
	return strings.HasSuffix(state, "_ERROR")
}

func toDurationPointer(d time.Duration) *time.Duration {
	return &d
}
