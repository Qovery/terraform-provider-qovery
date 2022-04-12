package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationEnvironmentVariables(ctx context.Context, applicationID string) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	applicationVariables, res, err := c.api.ApplicationEnvironmentVariableApi.
		ListApplicationEnvironmentVariable(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationEnvironmentVariable, applicationID, res, err)
	}
	return environmentVariableResponseListToArray(applicationVariables, EnvironmentVariableScopeApplication), nil
}

func (c *Client) updateApplicationEnvironmentVariables(ctx context.Context, applicationID string, request EnvironmentVariablesDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			EditApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			EnvironmentVariableEditRequest(variable.EnvironmentVariableEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariable(ctx, applicationID).
			EnvironmentVariableRequest(variable.EnvironmentVariableRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Key, res, err)
		}
	}
	return nil
}
