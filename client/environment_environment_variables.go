package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

func (c *Client) GetEnvironmentEnvironmentVariables(ctx context.Context, environmentID string) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	vars, res, err := c.api.EnvironmentVariableApi.
		ListEnvironmentEnvironmentVariable(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceEnvironmentEnvironmentVariable, environmentID, res, err)
	}
	return environmentVariableResponseListToArray(vars, EnvironmentVariableScopeEnvironment), nil
}

func (c *Client) updateEnvironmentEnvironmentVariables(ctx context.Context, environmentID string, request EnvironmentVariablesDiff) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	variables := make([]*qovery.EnvironmentVariableResponse, 0, len(request.Create)+len(request.Update))

	for _, variable := range request.Delete {
		res, err := c.api.EnvironmentVariableApi.
			DeleteEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeleteError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		v, res, err := c.api.EnvironmentVariableApi.
			EditEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id).
			EnvironmentVariableEditRequest(variable.EnvironmentVariableEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewUpdateError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Id, res, err)
		}
		variables = append(variables, v)
	}

	for _, variable := range request.Create {
		v, res, err := c.api.EnvironmentVariableApi.
			CreateEnvironmentEnvironmentVariable(ctx, environmentID).
			EnvironmentVariableRequest(variable.EnvironmentVariableRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Key, res, err)
		}
		variables = append(variables, v)
	}
	return variables, nil
}
