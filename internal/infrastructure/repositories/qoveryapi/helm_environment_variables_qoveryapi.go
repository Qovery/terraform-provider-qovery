package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var _ variable.Repository = helmEnvironmentVariablesQoveryAPI{}

type helmEnvironmentVariablesQoveryAPI struct {
	client *qovery.APIClient
}

func newHelmEnvironmentVariablesQoveryAPI(client *qovery.APIClient) (variable.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &helmEnvironmentVariablesQoveryAPI{
		client: client,
	}, nil
}

func (p helmEnvironmentVariablesQoveryAPI) Create(ctx context.Context, helmID string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.VariableMainCallsAPI.
		CreateVariable(ctx).
		VariableRequest(newQoveryEnvVariableRequestFromDomain(request, false, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmEnvironmentVariable, request.Key, resp, err)
	}

	return newDomainEnvVariableFromQovery(v)
}

func (p helmEnvironmentVariablesQoveryAPI) List(ctx context.Context, helmID string) (variable.Variables, error) {
	vars, resp, err := p.client.VariableMainCallsAPI.
		ListVariables(ctx).
		ParentId(helmID).
		Scope(qovery.APIVARIABLESCOPEENUM_HELM).
		IsSecret(false).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelmEnvironmentVariable, helmID, resp, err)
	}

	return newDomainEnvVariablesFromQovery(vars)
}

func (p helmEnvironmentVariablesQoveryAPI) Update(ctx context.Context, helmID string, variableId string, request variable.UpsertRequest) (*variable.Variable, error) {
	v, resp, err := p.client.VariableMainCallsAPI.
		EditVariable(ctx, variableId).
		VariableEditRequest(newQoveryEnvVariableEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelmEnvironmentVariable, variableId, resp, err)
	}

	return newDomainEnvVariableFromQovery(v)
}

func (p helmEnvironmentVariablesQoveryAPI) Delete(ctx context.Context, helmID string, variableId string) *apierrors.APIError {
	resp, err := p.client.VariableMainCallsAPI.
		DeleteVariable(ctx, variableId).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceHelmEnvironmentVariable, variableId, resp, err)
	}

	return nil
}

func (p helmEnvironmentVariablesQoveryAPI) CreateAlias(ctx context.Context, helmID string, request variable.UpsertRequest, aliasedVariableId string) (*variable.Variable, error) {
	v, resp, err := p.client.VariableMainCallsAPI.
		CreateVariableAlias(ctx, aliasedVariableId).
		VariableAliasRequest(newQoveryEnvVariableCreateAliasRequestFromDomain(request, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmEnvironmentVariable, helmID, resp, err)
	}

	return newDomainEnvVariableFromQovery(v)
}

func (p helmEnvironmentVariablesQoveryAPI) CreateOverride(ctx context.Context, helmID string, request variable.UpsertRequest, overriddenVariableId string) (*variable.Variable, error) {
	v, resp, err := p.client.VariableMainCallsAPI.
		CreateVariableOverride(ctx, overriddenVariableId).
		VariableOverrideRequest(newQoveryEnvVariableCreateOverrideRequestFromDomain(request, helmID, qovery.APIVARIABLESCOPEENUM_HELM)).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceHelmEnvironmentVariable, helmID, resp, err)
	}

	return newDomainEnvVariableFromQovery(v)
}
