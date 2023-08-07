package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure jobEnvironmentVariablesQoveryAPI defined types fully satisfy the variable.Repository interface.
var _ variable.Repository = jobEnvironmentVariablesQoveryAPI{}

// jobEnvironmentVariablesQoveryAPI implements the interface variable.Repository.
type jobEnvironmentVariablesQoveryAPI struct {
	client *qovery.APIClient
}

// newJobEnvironmentVariablesQoveryAPI return a new instance of a variable.Repository that uses Qovery's API.
func newJobEnvironmentVariablesQoveryAPI(client *qovery.APIClient) (variable.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &jobEnvironmentVariablesQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment variable for a job using the given jobID and request.
func (p jobEnvironmentVariablesQoveryAPI) Create(ctx context.Context, jobID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.JobEnvironmentVariableApi.
		CreateJobEnvironmentVariable(ctx, jobID).
		EnvironmentVariableRequest(newQoveryEnvironmentVariableRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceJobEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// List calls Qovery's API to retrieve an environment variables from a job using the given jobID and variableID.
func (p jobEnvironmentVariablesQoveryAPI) List(ctx context.Context, jobID string) (variable.Variables, error) {
	vars, resp, err := p.client.JobEnvironmentVariableApi.
		ListJobEnvironmentVariable(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceJobEnvironmentVariable, jobID, resp, err)
	}

	return newDomainVariablesFromQovery(vars)
}

// Update calls Qovery's API to update an environment variable from a job using the given jobID, credentialsID and request.
func (p jobEnvironmentVariablesQoveryAPI) Update(ctx context.Context, jobID string, credentialsID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.JobEnvironmentVariableApi.
		EditJobEnvironmentVariable(ctx, jobID, credentialsID).
		EnvironmentVariableEditRequest(newQoveryEnvironmentVariableEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceJobEnvironmentVariable, credentialsID, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// Delete calls Qovery's API to delete an environment variable from a job using the given jobID and credentialsID.
func (p jobEnvironmentVariablesQoveryAPI) Delete(ctx context.Context, jobID string, credentialsID string) error {
	resp, err := p.client.JobEnvironmentVariableApi.
		DeleteJobEnvironmentVariable(ctx, jobID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceJobEnvironmentVariable, credentialsID, resp, err)
	}

	return nil
}

func (p jobEnvironmentVariablesQoveryAPI) CreateAlias(ctx context.Context, scopeResourceID string, request variable.UpsertRequest, aliasedVariableId string) (*variable.Variable, error) {
	return nil, nil
}
func (p jobEnvironmentVariablesQoveryAPI) CreateOverride(ctx context.Context, scopeResourceID string, request variable.UpsertRequest, overriddenVariableId string) (*variable.Variable, error) {
	return nil, nil
}
