package client

import (
	"context"
	"time"

	"terraform-provider-qovery/client/apierrors"
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
		status, apiErr := client.GetApplicationStatus(ctx, applicationID)
		if apiErr != nil {
			return false, apiErr
		}
		return status.State == expected, nil
	}
}

func newApplicationFinalStateCheckerWaitFunc(client *Client, applicationID string) waitFunc {
	return func(ctx context.Context) (bool, *apierrors.APIError) {
		status, apiErr := client.GetApplicationStatus(ctx, applicationID)
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
		state != "QUEUED"
}

func toDurationPointer(d time.Duration) *time.Duration {
	return &d
}
