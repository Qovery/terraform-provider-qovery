package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

const (
	applicationStateRunning = "RUNNING"
	applicationStateStopped = "STOPPED"
)

func (c *Client) GetApplicationStatus(ctx context.Context, applicationID string) (*qovery.Status, *apierrors.APIError) {
	status, res, err := c.api.ApplicationMainCallsApi.
		GetApplicationStatus(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationStatus, applicationID, res, err)
	}
	return status, nil
}

func (c *Client) updateApplicationStatus(ctx context.Context, application *qovery.ApplicationResponse, desiredState string, forceRestart bool) (*qovery.Status, *apierrors.APIError) {
	// wait until we can stop the application - otherwise it will fail
	checker := newApplicationFinalStateCheckerWaitFunc(c, application.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	status, apiErr := c.GetApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	if status.State != desiredState {
		switch desiredState {
		case applicationStateRunning:
			return c.deployApplication(ctx, application.Id, *application.GitRepository.DeployedCommitId)
		case applicationStateStopped:
			return c.stopApplication(ctx, application.Id)
		}
	}

	if forceRestart {
		return c.restartApplication(ctx, application.Id)
	}

	return status, nil
}
