package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) deployApplication(ctx context.Context, application *qovery.Application) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case qovery.STATEENUM_DEPLOYED:
		return status, nil
	case qovery.STATEENUM_DEPLOYMENT_ERROR:
		return c.redeployApplication(ctx, application)
	default:
		_, res, err := c.api.ApplicationActionsAPI.
			DeployApplication(ctx, application.Id).
			DeployRequest(qovery.DeployRequest{
				GitCommitId: *application.GitRepository.DeployedCommitId,
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceApplication, application.Id, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, qovery.STATEENUM_DEPLOYED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}

func (c *Client) stopApplication(ctx context.Context, application *qovery.Application) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getApplicationStatus(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case qovery.STATEENUM_STOPPED:
		return status, nil
	default:
		_, res, err := c.api.ApplicationActionsAPI.
			StopApplication(ctx, application.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceApplication, application.Id, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, qovery.STATEENUM_STOPPED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}

func (c *Client) redeployApplication(ctx context.Context, application *qovery.Application) (*qovery.Status, *apierrors.APIError) {
	appFinalStateChecker := newApplicationFinalStateCheckerWaitFunc(c, application.Id)
	if apiErr := wait(ctx, appFinalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	envFinalStateChecker := newEnvironmentFinalStateCheckerWaitFunc(c, application.Environment.Id)
	if apiErr := wait(ctx, envFinalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.ApplicationActionsAPI.
		RedeployApplication(ctx, application.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRedeployError(apierrors.APIResourceApplication, application.Id, res, err)
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, application.Id, qovery.STATEENUM_DEPLOYED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getApplicationStatus(ctx, application.Id)
}
