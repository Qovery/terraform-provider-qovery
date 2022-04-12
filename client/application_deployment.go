package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) deployApplication(ctx context.Context, application *qovery.ApplicationResponse) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case applicationStateRunning:
		return status, nil
	case "DEPLOYMENT_ERROR":
		return c.restartApplication(ctx, application)
	default:
		_, res, err := c.api.ApplicationActionsApi.
			DeployApplication(ctx, application.Id).
			DeployRequest(qovery.DeployRequest{
				GitCommitId: *application.GitRepository.DeployedCommitId,
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceApplication, application.Id, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, applicationStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}

func (c *Client) stopApplication(ctx context.Context, application *qovery.ApplicationResponse) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case "STOPPED":
		return status, nil
	default:
		_, res, err := c.api.ApplicationActionsApi.
			StopApplication(ctx, application.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceApplication, application.Id, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, applicationStateStopped)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}

func (c *Client) restartApplication(ctx context.Context, application *qovery.ApplicationResponse) (*qovery.Status, *apierrors.APIError) {
	appFinalStateChecker := newApplicationFinalStateCheckerWaitFunc(c, application.Id)
	if apiErr := wait(ctx, appFinalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	envFinalStateChecker := newEnvironmentFinalStateCheckerWaitFunc(c, application.Environment.Id)
	if apiErr := wait(ctx, envFinalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.ApplicationActionsApi.
		RestartApplication(ctx, application.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRestartError(apierrors.APIResourceApplication, application.Id, res, err)
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, applicationStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}
