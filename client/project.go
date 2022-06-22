package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type ProjectResponse struct {
	ProjectResponse             *qovery.Project
	ProjectEnvironmentVariables []*qovery.EnvironmentVariable
	ProjectSecret               []*qovery.Secret
}

type ProjectUpsertParams struct {
	ProjectRequest           qovery.ProjectRequest
	EnvironmentVariablesDiff EnvironmentVariablesDiff
	SecretsDiff              SecretsDiff
}

func (c *Client) CreateProject(ctx context.Context, organizationID string, params ProjectUpsertParams) (*ProjectResponse, *apierrors.APIError) {
	project, res, err := c.api.ProjectsApi.
		CreateProject(ctx, organizationID).
		ProjectRequest(params.ProjectRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceProject, params.ProjectRequest.Name, res, err)
	}

	if !params.EnvironmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateProjectEnvironmentVariables(ctx, project.Id, params.EnvironmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !params.SecretsDiff.IsEmpty() {
		if apiErr := c.updateProjectSecrets(ctx, project.Id, params.SecretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	projectVariables, apiErr := c.getProjectEnvironmentVariables(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getProjectSecrets(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ProjectResponse{
		ProjectResponse:             project,
		ProjectEnvironmentVariables: projectVariables,
		ProjectSecret:               secrets,
	}, nil
}

func (c *Client) getProject(ctx context.Context, projectID string) (*qovery.Project, *apierrors.APIError) {
	project, res, err := c.api.ProjectMainCallsApi.
		GetProject(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceProject, projectID, res, err)
	}
	return project, nil
}

func (c *Client) GetProject(ctx context.Context, projectID string) (*ProjectResponse, *apierrors.APIError) {
	project, err := c.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	projectVariables, apiErr := c.getProjectEnvironmentVariables(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getProjectSecrets(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ProjectResponse{
		ProjectResponse:             project,
		ProjectEnvironmentVariables: projectVariables,
		ProjectSecret:               secrets,
	}, nil
}

func (c *Client) UpdateProject(ctx context.Context, projectID string, params ProjectUpsertParams) (*ProjectResponse, *apierrors.APIError) {
	project, res, err := c.api.ProjectMainCallsApi.
		EditProject(ctx, projectID).
		ProjectRequest(params.ProjectRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceProject, projectID, res, err)
	}

	if !params.EnvironmentVariablesDiff.IsEmpty() {
		if apiErr := c.updateProjectEnvironmentVariables(ctx, project.Id, params.EnvironmentVariablesDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	if !params.SecretsDiff.IsEmpty() {
		if apiErr := c.updateProjectSecrets(ctx, project.Id, params.SecretsDiff); apiErr != nil {
			return nil, apiErr
		}
	}

	projectVariables, apiErr := c.getProjectEnvironmentVariables(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	secrets, apiErr := c.getProjectSecrets(ctx, project.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	return &ProjectResponse{
		ProjectResponse:             project,
		ProjectEnvironmentVariables: projectVariables,
		ProjectSecret:               secrets,
	}, nil
}

func (c *Client) DeleteProject(ctx context.Context, projectID string) *apierrors.APIError {
	res, err := c.api.ProjectMainCallsApi.
		DeleteProject(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceProject, projectID, res, err)
	}
	return nil
}
