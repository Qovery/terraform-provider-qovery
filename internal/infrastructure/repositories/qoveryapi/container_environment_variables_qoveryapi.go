package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Ensure containerEnvironmentVariablesQoveryAPI defined types fully satisfy the variable.Repository interface.
var _ variable.Repository = containerEnvironmentVariablesQoveryAPI{}

// containerEnvironmentVariablesQoveryAPI implements the interface variable.Repository.
type containerEnvironmentVariablesQoveryAPI struct {
	client *qovery.APIClient
}

// newContainerEnvironmentVariablesQoveryAPI return a new instance of a variable.Repository that uses Qovery's API.
func newContainerEnvironmentVariablesQoveryAPI(client *qovery.APIClient) (variable.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &containerEnvironmentVariablesQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment variable for a container using the given containerID and request.
func (p containerEnvironmentVariablesQoveryAPI) Create(ctx context.Context, containerID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.ContainerEnvironmentVariableApi.
		CreateContainerEnvironmentVariable(ctx, containerID).
		EnvironmentVariableRequest(newQoveryEnvironmentVariableRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainerEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// List calls Qovery's API to retrieve an environment variables from a container using the given containerID and variableID.
func (p containerEnvironmentVariablesQoveryAPI) List(ctx context.Context, containerID string) (variable.Variables, error) {
	vars, resp, err := p.client.ContainerEnvironmentVariableApi.
		ListContainerEnvironmentVariable(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceContainerEnvironmentVariable, containerID, resp, err)
	}

	return newDomainVariablesFromQovery(vars)
}

// Update calls Qovery's API to update an environment variable from a container using the given containerID, credentialsID and request.
func (p containerEnvironmentVariablesQoveryAPI) Update(ctx context.Context, containerID string, credentialsID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.ContainerEnvironmentVariableApi.
		EditContainerEnvironmentVariable(ctx, containerID, credentialsID).
		EnvironmentVariableEditRequest(newQoveryEnvironmentVariableEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceContainerEnvironmentVariable, credentialsID, resp, err)
	}

	return newDomainVariableFromQovery(v)
}

// Delete calls Qovery's API to delete an environment variable from a container using the given containerID and credentialsID.
func (p containerEnvironmentVariablesQoveryAPI) Delete(ctx context.Context, containerID string, credentialsID string) error {
	resp, err := p.client.ContainerEnvironmentVariableApi.
		DeleteContainerEnvironmentVariable(ctx, containerID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceContainerEnvironmentVariable, credentialsID, resp, err)
	}

	return nil
}

func (p containerEnvironmentVariablesQoveryAPI) CreateAlias(ctx context.Context, containerID string, request variable.UpsertRequest, aliasedVariableId string) (*variable.Variable, error) {
	v, resp, err := p.client.ContainerEnvironmentVariableApi.
		CreateContainerEnvironmentVariableAlias(ctx, containerID, aliasedVariableId).
		Key(qovery.Key{Key: request.Key}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainerEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainVariableFromQovery(v)
}
func (p containerEnvironmentVariablesQoveryAPI) CreateOverride(ctx context.Context, containerID string, request variable.UpsertRequest, overriddenVariableId string) (*variable.Variable, error) {
	v, resp, err := p.client.ContainerEnvironmentVariableApi.
		CreateContainerEnvironmentVariableOverride(ctx, containerID, overriddenVariableId).
		Value(qovery.Value{Value: &request.Value}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainerEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainVariableFromQovery(v)
}
