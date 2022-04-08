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
	application, res, err := c.API.ApplicationsApi.
		CreateApplication(ctx, environmentID).
		ApplicationRequest(params.ApplicationRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceApplication, params.ApplicationRequest.Name, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.DesiredState)
}

func (c *Client) GetApplication(ctx context.Context, applicationID string) (*ApplicationResponse, *apierrors.APIError) {
	application, res, err := c.API.ApplicationMainCallsApi.
		GetApplication(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplication, applicationID, res, err)
	}

	status, apiErr := c.getApplicationStatus(ctx, applicationID)
	if apiErr != nil {
		return nil, apiErr
	}

	environmentVariables, apiErr := c.getApplicationEnvironmentVariables(ctx, application.Id)
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
	application, res, err := c.API.ApplicationMainCallsApi.
		EditApplication(ctx, applicationID).
		ApplicationEditRequest(params.ApplicationEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceApplication, applicationID, res, err)
	}
	return c.updateApplication(ctx, application, params.EnvironmentVariablesDiff, params.DesiredState)
}

func (c *Client) DeleteApplication(ctx context.Context, applicationID string) *apierrors.APIError {
	finalStateChecker := newApplicationFinalStateCheckerWaitFunc(c, applicationID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.API.ApplicationMainCallsApi.
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

	environmentVariables, apiErr := c.getApplicationEnvironmentVariables(ctx, application.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ApplicationResponse{
		ApplicationResponse:             application,
		ApplicationStatus:               status,
		ApplicationEnvironmentVariables: environmentVariables,
	}, nil
}
