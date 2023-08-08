package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

// Ensure jobSecretsQoveryAPI defined types fully satisfy the secret.Repository interface.
var _ secret.Repository = jobSecretsQoveryAPI{}

// jobSecretsQoveryAPI implements the interface secret.Repository.
type jobSecretsQoveryAPI struct {
	client *qovery.APIClient
}

// newJobSecretsQoveryAPI return a new instance of a secret.Repository that uses Qovery's API.
func newJobSecretsQoveryAPI(client *qovery.APIClient) (secret.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &jobSecretsQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an environment secret for a job using the given jobID and request.
func (p jobSecretsQoveryAPI) Create(ctx context.Context, jobID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.JobSecretApi.
		CreateJobSecret(ctx, jobID).
		SecretRequest(newQoverySecretRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceJobSecret, request.Key, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// List calls Qovery's API to retrieve an environment secrets from a job using the given jobID and secretID.
func (p jobSecretsQoveryAPI) List(ctx context.Context, jobID string) (secret.Secrets, error) {
	vars, resp, err := p.client.JobSecretApi.
		ListJobSecrets(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceJobSecret, jobID, resp, err)
	}

	return newDomainSecretsFromQovery(vars)
}

// Update calls Qovery's API to update an environment secret from a job using the given jobID, credentialsID and request.
func (p jobSecretsQoveryAPI) Update(ctx context.Context, jobID string, credentialsID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := p.client.JobSecretApi.
		EditJobSecret(ctx, jobID, credentialsID).
		SecretEditRequest(newQoverySecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceJobSecret, credentialsID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

// Delete calls Qovery's API to delete an environment secret from a job using the given jobID and credentialsID.
func (p jobSecretsQoveryAPI) Delete(ctx context.Context, jobID string, credentialsID string) *apierrors.ApiError {
	resp, err := p.client.JobSecretApi.
		DeleteJobSecret(ctx, jobID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceJobSecret, credentialsID, resp, err)
	}

	return nil
}

func (p jobSecretsQoveryAPI) CreateAlias(ctx context.Context, jobID string, request secret.UpsertRequest, aliasedSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.JobSecretApi.
		CreateJobSecretAlias(ctx, jobID, aliasedSecretId).
		Key(qovery.Key{Key: request.Key}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceJobSecret, jobID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}

func (p jobSecretsQoveryAPI) CreateOverride(ctx context.Context, jobID string, request secret.UpsertRequest, overriddenSecretId string) (*secret.Secret, error) {
	v, resp, err := p.client.JobSecretApi.
		CreateJobSecretOverride(ctx, jobID, overriddenSecretId).
		Value(qovery.Value{Value: &request.Value}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceJobSecret, jobID, resp, err)
	}

	return newDomainSecretFromQovery(v)
}
