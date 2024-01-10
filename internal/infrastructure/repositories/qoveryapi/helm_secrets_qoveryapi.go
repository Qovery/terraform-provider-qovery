package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

var _ secret.Repository = helmSecretsQoveryAPI{}

// helmSecretsQoveryAPI implements the interface secret.Repository.
type helmSecretsQoveryAPI struct {
	client *qovery.APIClient
}

func newHelmSecretsQoveryAPI(client *qovery.APIClient) (secret.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &helmSecretsQoveryAPI{
		client: client,
	}, nil
}

func (s helmSecretsQoveryAPI) Create(ctx context.Context, helmID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := s.client.VariableMainCallsAPI.
		CreateVariable(ctx).
		VariableRequest(newQoveryEnvSecretVariableRequestFromDomain(request, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmSecret, request.Key, resp, err)
	}

	return newDomainEnvSecretFromQovery(v)
}

func (s helmSecretsQoveryAPI) List(ctx context.Context, helmID string) (secret.Secrets, error) {
	vars, resp, err := s.client.VariableMainCallsAPI.
		ListVariables(ctx).
		ParentId(helmID).
		Scope(qovery.APIVARIABLESCOPEENUM_HELM).
		IsSecret(true).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelmSecret, helmID, resp, err)
	}

	return newDomainEnvSecretsFromQovery(vars)
}

func (s helmSecretsQoveryAPI) Update(ctx context.Context, helmID string, variableID string, request secret.UpsertRequest) (*secret.Secret, error) {
	v, resp, err := s.client.VariableMainCallsAPI.
		EditVariable(ctx, variableID).
		VariableEditRequest(newQoveryEnvSecretEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelmSecret, variableID, resp, err)
	}

	return newDomainEnvSecretFromQovery(v)
}

func (s helmSecretsQoveryAPI) Delete(ctx context.Context, helmID string, variableID string) *apierrors.APIError {
	resp, err := s.client.VariableMainCallsAPI.
		DeleteVariable(ctx, variableID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceHelmSecret, variableID, resp, err)
	}

	return nil
}

func (s helmSecretsQoveryAPI) CreateAlias(ctx context.Context, helmID string, request secret.UpsertRequest, aliasedSecretId string) (*secret.Secret, error) {
	v, resp, err := s.client.VariableMainCallsAPI.
		CreateVariableAlias(ctx, aliasedSecretId).
		VariableAliasRequest(newQoveryEnvSecretCreateAliasRequestFromDomain(request, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmSecret, helmID, resp, err)
	}

	return newDomainEnvSecretFromQovery(v)
}

func (s helmSecretsQoveryAPI) CreateOverride(ctx context.Context, helmID string, request secret.UpsertRequest, overriddenSecretId string) (*secret.Secret, error) {
	v, resp, err := s.client.VariableMainCallsAPI.
		CreateVariableOverride(ctx, overriddenSecretId).
		VariableOverrideRequest(newQoveryEnvSecretCreateOverrideRequestFromDomain(request, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmSecret, helmID, resp, err)
	}

	return newDomainEnvSecretFromQovery(v)
}
