package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getProjectEnvironmentVariables(ctx context.Context, projectID string) ([]*qovery.EnvironmentVariable, *apierrors.APIError) {
	projectVariables, res, err := c.api.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceProjectEnvironmentVariable, projectID, res, err)
	}

	return environmentVariableResponseListToArray(projectVariables, qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT), nil
}

func (c *Client) updateProjectEnvironmentVariables(ctx context.Context, projectID string, request EnvironmentVariablesDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.ProjectEnvironmentVariableApi.
			DeleteProjectEnvironmentVariable(ctx, projectID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceProjectEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.ProjectEnvironmentVariableApi.
			EditProjectEnvironmentVariable(ctx, projectID, variable.Id).
			EnvironmentVariableEditRequest(variable.EnvironmentVariableEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceProjectEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.ProjectEnvironmentVariableApi.
			CreateProjectEnvironmentVariable(ctx, projectID).
			EnvironmentVariableRequest(variable.EnvironmentVariableRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceProjectEnvironmentVariable, variable.Key, res, err)
		}
	}
	return nil
}
