package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getProjectEnvironmentVariables(ctx context.Context, projectID string) ([]*qovery.EnvironmentVariableResponse, *apierrors.APIError) {
	projectVariables, res, err := c.api.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceProjectEnvironmentVariable, projectID, res, err)
	}

	return environmentVariableResponseListToArray(projectVariables, EnvironmentVariableScopeProject), nil
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

func environmentVariableResponseListToArray(list *qovery.EnvironmentVariableResponseList, scope EnvironmentVariableScope) []*qovery.EnvironmentVariableResponse {
	vars := make([]*qovery.EnvironmentVariableResponse, 0, len(list.GetResults()))
	for _, v := range list.GetResults() {
		if v.Scope != scope.String() && v.Scope != EnvironmentVariableScopeBuiltIn.String() {
			continue
		}
		cpy := v
		vars = append(vars, &cpy)
	}
	return vars
}
