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

	return environmentVariableResponseListToArray(projectVariables), nil
}

func (c *Client) updateProjectEnvironmentVariables(ctx context.Context, projectID string, request EnvironmentVariablesDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		if err := c.deleteProjectEnvironmentVariable(ctx, projectID, variable); err != nil {
			return err
		}
	}

	for _, variable := range request.Update {
		if err := c.editProjectEnvironmentVariable(ctx, projectID, variable); err != nil {
			return err
		}
	}

	for _, variable := range request.Create {
		if err := c.createProjectEnvironmentVariable(ctx, projectID, variable); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) createProjectEnvironmentVariable(ctx context.Context, projectID string, variable EnvironmentVariableCreateRequest) *apierrors.APIError {
	_, res, err := c.api.ProjectEnvironmentVariableApi.
		CreateProjectEnvironmentVariable(ctx, projectID).
		EnvironmentVariableRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceProjectEnvironmentVariable, variable.Key, res, err)
	}
	return nil
}

func (c *Client) editProjectEnvironmentVariable(ctx context.Context, projectID string, variable EnvironmentVariableUpdateRequest) *apierrors.APIError {
	_, res, err := c.api.ProjectEnvironmentVariableApi.
		EditProjectEnvironmentVariable(ctx, projectID, variable.Id).
		EnvironmentVariableEditRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceProjectEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}

func (c *Client) deleteProjectEnvironmentVariable(ctx context.Context, projectID string, variable EnvironmentVariableDeleteRequest) *apierrors.APIError {
	res, err := c.api.ProjectEnvironmentVariableApi.
		DeleteProjectEnvironmentVariable(ctx, projectID, variable.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewDeleteError(apierrors.APIResourceProjectEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}
