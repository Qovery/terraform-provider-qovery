package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type EnvironmentResponse struct {
	EnvironmentResponse             *qovery.Environment
	EnvironmentEnvironmentVariables []*qovery.EnvironmentVariable
	EnvironmentSecret               []*qovery.Secret
}

type EnvironmentCreateParams struct {
	EnvironmentRequest       qovery.EnvironmentRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	SecretsDiff              SecretsDiff
}

type EnvironmentUpdateParams struct {
	EnvironmentEditRequest   qovery.EnvironmentEditRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	SecretsDiff              SecretsDiff
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID string, params *EnvironmentCreateParams) (*EnvironmentResponse, *apierrors.APIError) {
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

	if !params.SecretsDiff.IsEmpty() {
		if apiErr := c.updateEnvironmentSecrets(ctx, environment.Id, params.SecretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	environmentVariables, apiErr := c.getEnvironmentEnvironmentVariables(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getEnvironmentSecrets(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
		EnvironmentSecret:               secrets,
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

	secrets, apiErr := c.getEnvironmentSecrets(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
		EnvironmentSecret:               secrets,
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

	if !params.SecretsDiff.IsEmpty() {
		if apiErr := c.updateEnvironmentSecrets(ctx, environment.Id, params.SecretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	// Restart environment if environment variables / secrets has been updated
	if !params.EnvironmentVariablesDiff.IsEmpty() || !params.SecretsDiff.IsEmpty() {
		if _, apiErr := c.restartEnvironment(ctx, environment.Id); apiErr != nil {
			return nil, apiErr
		}
	}

	environmentVariables, apiErr := c.getEnvironmentEnvironmentVariables(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getEnvironmentSecrets(ctx, environment.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &EnvironmentResponse{
		EnvironmentResponse:             environment,
		EnvironmentEnvironmentVariables: environmentVariables,
		EnvironmentSecret:               secrets,
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
