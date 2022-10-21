package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure projectEnvironmentVariablesQoveryAPI defined types fully satisfy the variable.Repository interface.
var _ variable.Repository = projectEnvironmentVariablesQoveryAPI{}

// projectEnvironmentVariablesQoveryAPI implements the interface variable.Repository.
type projectEnvironmentVariablesQoveryAPI struct {
	client *qovery.APIClient
}

// newProjectEnvironmentVariablesQoveryAPI return a new instance of a variable.Repository that uses Qovery's API.
func newProjectEnvironmentVariablesQoveryAPI(client *qovery.APIClient) (variable.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &projectEnvironmentVariablesQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment variable for a project using the given projectID and request.
func (p projectEnvironmentVariablesQoveryAPI) Create(ctx context.Context, projectID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.ProjectEnvironmentVariableApi.
		CreateProjectEnvironmentVariable(ctx, projectID).
		EnvironmentVariableRequest(newQoveryEnvironmentVariableRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceProjectEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// List calls Qovery's API to retrieve an environment variables from a project using the given projectID and variableID.
func (p projectEnvironmentVariablesQoveryAPI) List(ctx context.Context, projectID string) (variable.Variables, error) {
	vars, resp, err := p.client.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, projectID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceProjectEnvironmentVariable, projectID, resp, err)
	}

	return newDomainVariablesFromQovery(vars)
}

// Update calls Qovery's API to update an environment variable from a project using the given projectID, credentialsID and request.
func (p projectEnvironmentVariablesQoveryAPI) Update(ctx context.Context, projectID string, credentialsID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.ProjectEnvironmentVariableApi.
		EditProjectEnvironmentVariable(ctx, projectID, credentialsID).
		EnvironmentVariableEditRequest(newQoveryEnvironmentVariableEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceProjectEnvironmentVariable, credentialsID, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// Delete calls Qovery's API to delete an environment variable from a project using the given projectID and credentialsID.
func (p projectEnvironmentVariablesQoveryAPI) Delete(ctx context.Context, projectID string, credentialsID string) error {
	resp, err := p.client.ProjectEnvironmentVariableApi.
		DeleteProjectEnvironmentVariable(ctx, projectID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceProjectEnvironmentVariable, credentialsID, resp, err)
	}

	return nil
}
