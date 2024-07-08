package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// Ensure containerSecretsQoveryAPI defined types fully satisfy the secret.Repository interface.
var _ secret.Repository = containerSecretsQoveryAPI{}

// containerSecretsQoveryAPI implements the interface secret.Repository.
type containerSecretsQoveryAPI struct {
	client *qovery.APIClient
}

// newContainerSecretsQoveryAPI return a new instance of a secret.Repository that uses Qovery's API.
func newContainerSecretsQoveryAPI(client *qovery.APIClient) (secret.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &containerSecretsQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment secret for a container using the given containerID and request.
func (p containerSecretsQoveryAPI) Create(ctx context.Context, containerID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.ContainerSecretAPI.
		CreateContainerSecret(ctx, containerID).
		SecretRequest(newQoverySecretRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerSecret, request.Key, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// List calls Qovery's API to retrieve an environment secrets from a container using the given containerID and secretID.
func (p containerSecretsQoveryAPI) List(ctx context.Context, containerID string) (secret.Secrets, error) {
	vars, resp, err := p.client.ContainerSecretAPI.
		ListContainerSecrets(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceContainerSecret, containerID, resp, err)
	}

	return newDomainSecretsFromQovery(vars)
}

// Update calls Qovery's API to update an environment secret from a container using the given containerID, credentialsID and request.
func (p containerSecretsQoveryAPI) Update(ctx context.Context, containerID string, credentialsID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.ContainerSecretAPI.
		EditContainerSecret(ctx, containerID, credentialsID).
		SecretEditRequest(newQoverySecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceContainerSecret, credentialsID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// Delete calls Qovery's API to delete an environment secret from a container using the given containerID and credentialsID.
func (p containerSecretsQoveryAPI) Delete(ctx context.Context, containerID string, credentialsID string) *apierrors.APIError {
	resp, err := p.client.ContainerSecretAPI.
		DeleteContainerSecret(ctx, containerID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceContainerSecret, credentialsID, resp, err)
	}

	return nil
}

func (p containerSecretsQoveryAPI) CreateAlias(ctx context.Context, containerId string, request secret.UpsertRequest, aliasedSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.ContainerSecretAPI.
		CreateContainerSecretAlias(ctx, containerId, aliasedSecretId).
		Key(qovery.Key{
			Key:         request.Key,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerSecret, containerId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

func (p containerSecretsQoveryAPI) CreateOverride(ctx context.Context, containerId string, request secret.UpsertRequest, overriddenSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.ContainerSecretAPI.
		CreateContainerSecretOverride(ctx, containerId, overriddenSecretId).
		Value(qovery.Value{
			Value:       &request.Value,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerSecret, containerId, resp, err)
	}

	return newDomainSecretFromQovery(v)
}
