package client

import (
	"context"
	"time"

	"terraform-provider-qovery/client/apierrors"
)

const defaultWaitTimeout = 30 * time.Minute

func waitForStatus(ctx context.Context, checker StatusChecker, timeout *time.Duration) *apierrors.APIError {
	if timeout == nil {
		timeout = toDurationPointer(defaultWaitTimeout)
	}

	ticker := time.NewTicker(10 * time.Second)
	timeoutTicker := time.NewTicker(*timeout)

	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case <-ticker.C:
			ok, apiErr := checker.Exec(ctx)
			if apiErr != nil {
				return apiErr
			}
			if ok {
				return nil
			}
		}
	}
}

type StatusChecker interface {
	Exec(ctx context.Context) (bool, *apierrors.APIError)
}

type ApplicationStatusChecker struct {
	client        *Client
	applicationID string
	expected      string
}

func NewApplicationStatusChecker(client *Client, applicationID string, expected string) *ApplicationStatusChecker {
	return &ApplicationStatusChecker{
		client:        client,
		applicationID: applicationID,
		expected:      expected,
	}
}

func (c ApplicationStatusChecker) Exec(ctx context.Context) (bool, *apierrors.APIError) {
	status, err := c.client.GetApplicationStatus(ctx, c.applicationID)
	if err != nil {
		return false, err
	}
	return status.State == c.expected, nil
}

type ApplicationFinalStateChecker struct {
	client        *Client
	applicationID string
}

func NewApplicationFinalStateChecker(client *Client, applicationID string) *ApplicationFinalStateChecker {
	return &ApplicationFinalStateChecker{
		client:        client,
		applicationID: applicationID,
	}
}

func (c ApplicationFinalStateChecker) Exec(ctx context.Context) (bool, *apierrors.APIError) {
	status, err := c.client.GetApplicationStatus(ctx, c.applicationID)
	if err != nil {
		return false, err
	}
	return isFinalState(status.State), nil
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
