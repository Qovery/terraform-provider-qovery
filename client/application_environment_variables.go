package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

func (c *Client) GetApplicationEnvironmentVariables(ctx context.Context, applicationID string) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	applicationVariables, res, err := c.api.ApplicationEnvironmentVariableApi.
		ListApplicationEnvironmentVariable(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationEnvironmentVariable, applicationID, res, err)
	}
	return environmentVariableResponseListToArray(applicationVariables, EnvironmentVariableScopeApplication), nil
}

func (c *Client) updateApplicationEnvironmentVariables(ctx context.Context, applicationID string, request EnvironmentVariablesDiff) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	variables := make([]*qovery.EnvironmentVariableResponse, 0, len(request.Create)+len(request.Update))

	for _, variable := range request.Delete {
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		v, res, err := c.api.ApplicationEnvironmentVariableApi.
			EditApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			EnvironmentVariableEditRequest(variable.EnvironmentVariableEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
		variables = append(variables, v)
	}

	for _, variable := range request.Create {
		v, res, err := c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariable(ctx, applicationID).
			EnvironmentVariableRequest(variable.EnvironmentVariableRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Key, res, err)
		}
		variables = append(variables, v)
	}
	return variables, nil
}
