package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// Ensure environmentSecretsQoveryAPI defined types fully satisfy the secret.Repository interface.
var _ secret.Repository = environmentSecretsQoveryAPI{}

// environmentSecretsQoveryAPI implements the interface secret.Repository.
type environmentSecretsQoveryAPI struct {
	client *qovery.APIClient
}

// newEnvironmentSecretsQoveryAPI return a new instance of a secret.Repository that uses Qovery's API.
func newEnvironmentSecretsQoveryAPI(client *qovery.APIClient) (secret.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &environmentSecretsQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment secret for an environment using the given environmentID and request.
func (p environmentSecretsQoveryAPI) Create(ctx context.Context, environmentID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.EnvironmentSecretApi.
		CreateEnvironmentSecret(ctx, environmentID).
		SecretRequest(newQoverySecretRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceEnvironmentSecret, request.Key, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// List calls Qovery's API to retrieve an environment secrets from an environment using the given environmentID and secretID.
func (p environmentSecretsQoveryAPI) List(ctx context.Context, environmentID string) (secret.Secrets, error) {
	vars, resp, err := p.client.EnvironmentSecretApi.
		ListEnvironmentSecrets(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceEnvironmentSecret, environmentID, resp, err)
	}

	return newDomainSecretsFromQovery(vars)
}

// Update calls Qovery's API to update an environment secret from an environment using the given environmentID, credentialsID and request.
func (p environmentSecretsQoveryAPI) Update(ctx context.Context, environmentID string, credentialsID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.EnvironmentSecretApi.
		EditEnvironmentSecret(ctx, environmentID, credentialsID).
		SecretEditRequest(newQoverySecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceEnvironmentSecret, credentialsID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// Delete calls Qovery's API to delete an environment secret from an environment using the given environmentID and credentialsID.
func (p environmentSecretsQoveryAPI) Delete(ctx context.Context, environmentID string, credentialsID string) error {
	resp, err := p.client.EnvironmentSecretApi.
		DeleteEnvironmentSecret(ctx, environmentID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceEnvironmentSecret, credentialsID, resp, err)
	}

	return nil
}
