package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

type ApplicationResponse struct {
	ApplicationResponse             *qovery.ApplicationResponse
	ApplicationStatus               *qovery.Status
	ApplicationEnvironmentVariables []*qovery.EnvironmentVariableResponse
}

type ApplicationCreateParams struct {
	ApplicationRequest       qovery.ApplicationRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	DesiredState             string
}

type ApplicationUpdateParams struct {
	ApplicationEditRequest   qovery.ApplicationEditRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	DesiredState             string
}

func (c *Client) CreateApplication(ctx context.Context, environmentID string, params ApplicationCreateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationsApi.
		CreateApplication(ctx, environmentID).
		ApplicationRequest(params.ApplicationRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, params.ApplicationRequest.Name, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.DesiredState)
}

func (c *Client) GetApplication(ctx context.Context, applicationID string) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsApi.
		GetApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	status, apiErr := c.GetApplicationStatus(ctx, applicationID)
	if apiErr != nil {
		return nil, apiErr
	}

	environmentVariables, apiErr := c.GetApplicationEnvironmentVariables(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
	}, nil
}

func (c *Client) UpdateApplication(ctx context.Context, applicationID string, params ApplicationUpdateParams) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.api.ApplicationMainCallsApi.
		EditApplication(ctx, applicationID).
		ApplicationEditRequest(params.ApplicationEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, applicationID, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.DesiredState)
}

func (c *Client) DeleteApplication(ctx context.Context, applicationID string) *apierrors.APIError {
	res, err := c.api.ApplicationMainCallsApi.
		DeleteApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	checker := newApplicationStatusCheckerWaitFunc(c, applicationID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) updateApplication(ctx context.Context, application *qovery.ApplicationResponse, environmentVariablesDiff EnvironmentVariablesDiff, desiredState string) (*ApplicationResponse, *apierrors.APIError) {
	forceRestart := !environmentVariablesDiff.IsEmpty()
	if !environmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateApplicationEnvironmentVariables(ctx, application.Id, environmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	status, apiErr := c.updateApplicationStatus(ctx, application, desiredState, forceRestart)
	if apiErr != nil {
		return nil, apiErr
	}

	environmentVariables, apiErr := c.GetApplicationEnvironmentVariables(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
	}, nil
}

func (c *Client) deployApplication(ctx context.Context, applicationID string, deployedCommitID string) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.GetApplicationStatus(ctx, applicationID)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case applicationStateRunning:
		return status, nil
	case "DEPLOYMENT_ERROR":
		return c.restartApplication(ctx, applicationID)
	default:
		_, res, err := c.api.ApplicationActionsApi.
			DeployApplication(ctx, applicationID).
			DeployRequest(qovery.DeployRequest{
				GitCommitId: deployedCommitID,
			}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceApplication, applicationID, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, applicationID, applicationStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.GetApplicationStatus(ctx, applicationID)
}

func (c *Client) stopApplication(ctx context.Context, applicationID string) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.GetApplicationStatus(ctx, applicationID)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case "STOPPED":
		return status, nil
	default:
		_, res, err := c.api.ApplicationActionsApi.
			StopApplication(ctx, applicationID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceApplication, applicationID, res, err)
		}
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, applicationID, applicationStateStopped)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.GetApplicationStatus(ctx, applicationID)
}

func (c *Client) restartApplication(ctx context.Context, applicationID string) (*qovery.Status, *apierrors.APIError) {
	finalStateChecker := newApplicationFinalStateCheckerWaitFunc(c, applicationID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.ApplicationActionsApi.
		RestartApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRestartError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	statusChecker := newApplicationStatusCheckerWaitFunc(c, applicationID, applicationStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.GetApplicationStatus(ctx, applicationID)
}
