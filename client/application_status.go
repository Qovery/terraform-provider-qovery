package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationStatus(ctx context.Context, applicationID string) (*qovery.Status, *apierrors.APIError) {
	status, res, err := c.api.ApplicationMainCallsApi.
		GetApplicationStatus(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationStatus, applicationID, res, err)
	}

	// Handle READY as STOPPED state
	if status.State == qovery.STATEENUM_READY {
		status.State = qovery.STATEENUM_STOPPED
	}
	return status, nil
}

func (c *Client) updateApplicationStatus(ctx context.Context, application *qovery.Application, desiredState qovery.StateEnum, forceRestart bool) (*qovery.Status, *apierrors.APIError) {
	// wait until we can stop the application - otherwise it will fail
	checker := newApplicationFinalStateCheckerWaitFunc(c, application.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	status, apiErr := c.getApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	envChecker := newEnvironmentFinalStateCheckerWaitFunc(c, application.Environment.Id)
	if apiErr := wait(ctx, envChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	if status.State != desiredState {
		switch desiredState {
		case qovery.STATEENUM_RUNNING:
			return c.deployApplication(ctx, application)
		case qovery.STATEENUM_STOPPED:
			return c.stopApplication(ctx, application)
		}
	}

	deploymentStatus := status.ServiceDeploymentStatus.Get()
	if (deploymentStatus != nil && *deploymentStatus == qovery.SERVICEDEPLOYMENTSTATUSENUM_OUT_OF_DATE) || forceRestart {
		return c.restartApplication(ctx, application)
	}

	return status, nil
}
