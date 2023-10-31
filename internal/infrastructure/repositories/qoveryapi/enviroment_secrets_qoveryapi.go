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
	v, resp, err := p.client.EnvironmentSecretAPI.
		CreateEnvironmentSecret(ctx, environmentID).
		SecretRequest(newQoverySecretRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceEnvironmentSecret, request.Key, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// List calls Qovery's API to retrieve an environment secrets from an environment using the given environmentID and secretID.
func (p environmentSecretsQoveryAPI) List(ctx context.Context, environmentID string) (secret.Secrets, error) {
	vars, resp, err := p.client.EnvironmentSecretAPI.
		ListEnvironmentSecrets(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceEnvironmentSecret, environmentID, resp, err)
	}

	return newDomainSecretsFromQovery(vars)
}

// Update calls Qovery's API to update an environment secret from an environment using the given environmentID, credentialsID and request.
func (p environmentSecretsQoveryAPI) Update(ctx context.Context, environmentID string, credentialsID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.EnvironmentSecretAPI.
		EditEnvironmentSecret(ctx, environmentID, credentialsID).
		SecretEditRequest(newQoverySecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceEnvironmentSecret, credentialsID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// Delete calls Qovery's API to delete an environment secret from an environment using the given environmentID and credentialsID.
func (p environmentSecretsQoveryAPI) Delete(ctx context.Context, environmentID string, credentialsID string) *apierrors.APIError {
	resp, err := p.client.EnvironmentSecretAPI.
		DeleteEnvironmentSecret(ctx, environmentID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceEnvironmentSecret, credentialsID, resp, err)
	}

	return nil
}

func (p environmentSecretsQoveryAPI) CreateAlias(ctx context.Context, environmentId string, request secret.UpsertRequest, aliasedSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.EnvironmentSecretAPI.
		CreateEnvironmentSecretAlias(ctx, environmentId, aliasedSecretId).
		Key(qovery.Key{Key: request.Key}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceEnvironmentSecret, environmentId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

func (p environmentSecretsQoveryAPI) CreateOverride(ctx context.Context, environmentId string, request secret.UpsertRequest, overriddenSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.EnvironmentSecretAPI.
		CreateEnvironmentSecretOverride(ctx, environmentId, overriddenSecretId).
		Value(qovery.Value{Value: &request.Value}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceEnvironmentSecret, environmentId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}
