package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type EnvironmentResponse struct {
	EnvironmentResponse             *qovery.EnvironmentResponse
	EnvironmentEnvironmentVariables []*qovery.EnvironmentVariableResponse
}

type EnvironmentCreateParams struct {
	EnvironmentRequest       qovery.EnvironmentRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
}

type EnvironmentUpdateParams struct {
	EnvironmentEditRequest   qovery.EnvironmentEditRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID string, params EnvironmentCreateParams) (*EnvironmentResponse, *apierrors.APIError) {
	environment, res, err := c.api.EnvironmentsApi.
		CreateEnvironment(ctx, projectID).
		EnvironmentRequest(params.EnvironmentRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceEnvironment, params.EnvironmentRequest.Name, res, err)
	}

	if !params.EnvironmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateEnvironmentEnvironmentVariables(ctx, environment.Id, params.EnvironmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	environmentVariables, apiErr := c.getEnvironmentEnvironmentVariables(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
	}, nil
}

func (c *Client) GetEnvironment(ctx context.Context, environmentID string) (*EnvironmentResponse, *apierrors.APIError) {
	environment, res, err := c.api.EnvironmentMainCallsApi.
		GetEnvironment(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceEnvironment, environmentID, res, err)
	}

	environmentVariables, apiErr := c.getEnvironmentEnvironmentVariables(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
	}, nil
}

func (c *Client) UpdateEnvironment(ctx context.Context, environmentID string, params EnvironmentUpdateParams) (*EnvironmentResponse, *apierrors.APIError) {
	environment, res, err := c.api.EnvironmentMainCallsApi.
		EditEnvironment(ctx, environmentID).
		EnvironmentEditRequest(params.EnvironmentEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceEnvironment, environmentID, res, err)
	}

	if !params.EnvironmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateEnvironmentEnvironmentVariables(ctx, environment.Id, params.EnvironmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	environmentVariables, apiErr := c.getEnvironmentEnvironmentVariables(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	// TODO restart the whole environment if env vars have been changed

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
	}, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, environmentID string) *apierrors.APIError {
	finalStateChecker := newEnvironmentFinalStateCheckerWaitFunc(c, environmentID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err := c.api.EnvironmentMainCallsApi.
		DeleteEnvironment(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceEnvironment, environmentID, res, err)
	}

	checker := newEnvironmentStatusCheckerWaitFunc(c, environmentID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}
