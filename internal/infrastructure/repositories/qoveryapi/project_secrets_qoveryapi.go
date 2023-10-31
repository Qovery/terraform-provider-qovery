package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// Ensure projectSecretsQoveryAPI defined types fully satisfy the secret.Repository interface.
var _ secret.Repository = projectSecretsQoveryAPI{}

// projectSecretsQoveryAPI implements the interface secret.Repository.
type projectSecretsQoveryAPI struct {
	client *qovery.APIClient
}

// newProjectSecretsQoveryAPI return a new instance of a secret.Repository that uses Qovery's API.
func newProjectSecretsQoveryAPI(client *qovery.APIClient) (secret.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &projectSecretsQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment secret for a project using the given projectID and request.
func (p projectSecretsQoveryAPI) Create(ctx context.Context, projectID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.ProjectSecretAPI.
		CreateProjectSecret(ctx, projectID).
		SecretRequest(newQoverySecretRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceProjectSecret, request.Key, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// List calls Qovery's API to retrieve an environment secrets from a project using the given projectID and secretID.
func (p projectSecretsQoveryAPI) List(ctx context.Context, projectID string) (secret.Secrets, error) {
	vars, resp, err := p.client.ProjectSecretAPI.
		ListProjectSecrets(ctx, projectID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceProjectSecret, projectID, resp, err)
	}

	return newDomainSecretsFromQovery(vars)
}

// Update calls Qovery's API to update an environment secret from a project using the given projectID, credentialsID and request.
func (p projectSecretsQoveryAPI) Update(ctx context.Context, projectID string, credentialsID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.ProjectSecretAPI.
		EditProjectSecret(ctx, projectID, credentialsID).
		SecretEditRequest(newQoverySecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceProjectSecret, credentialsID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// Delete calls Qovery's API to delete an environment secret from a project using the given projectID and credentialsID.
func (p projectSecretsQoveryAPI) Delete(ctx context.Context, projectID string, credentialsID string) *apierrors.APIError {
	resp, err := p.client.ProjectSecretAPI.
		DeleteProjectSecret(ctx, projectID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceProjectSecret, credentialsID, resp, err)
	}

	return nil
}

func (p projectSecretsQoveryAPI) CreateAlias(ctx context.Context, projectId string, request secret.UpsertRequest, aliasedSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.ProjectSecretAPI.
		CreateProjectSecretAlias(ctx, projectId, aliasedSecretId).
		Key(qovery.Key{Key: request.Key}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceProjectSecret, projectId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}
func (p projectSecretsQoveryAPI) CreateOverride(ctx context.Context, projectId string, request secret.UpsertRequest, overriddenSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.ProjectSecretAPI.
		CreateProjectSecretOverride(ctx, projectId, overriddenSecretId).
		Value(qovery.Value{Value: &request.Value}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceProjectSecret, projectId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}
