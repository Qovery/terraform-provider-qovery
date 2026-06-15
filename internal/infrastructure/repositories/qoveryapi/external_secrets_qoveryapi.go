package qoveryapi

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var _ variable.ExternalSecretRepository = externalSecretsQoveryAPI{}

type externalSecretsQoveryAPI struct {
	client      *qovery.APIClient
	scope       qovery.APIVariableScopeEnum
	apiResource apierrors.APIResource
}

func newExternalSecretsQoveryAPI(client *qovery.APIClient, scope qovery.APIVariableScopeEnum, apiResource apierrors.APIResource) (variable.ExternalSecretRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &externalSecretsQoveryAPI{
		client:      client,
		scope:       scope,
		apiResource: apiResource,
	}, nil
}

func (p externalSecretsQoveryAPI) Create(ctx context.Context, serviceID string, request variable.ExternalSecretUpsertRequest) (*variable.ExternalSecret, error) {
	smAccessID := request.SecretManagerAccessId
	v, resp, err := p.client.VariableMainCallsAPI.
		CreateVariable(ctx).
		VariableRequest(qovery.VariableRequest{
			Key:                   request.Key,
			Value:                 request.Reference,
			IsSecret:              false,
			VariableScope:         p.scope,
			VariableParentId:      serviceID,
			SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
			Description:           *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(p.apiResource, request.Key, resp, err)
	}

	return newDomainExternalSecretFromQovery(v)
}

func (p externalSecretsQoveryAPI) Update(ctx context.Context, variableID string, request variable.ExternalSecretUpsertRequest) (*variable.ExternalSecret, error) {
	smAccessID := request.SecretManagerAccessId
	v, resp, err := p.client.VariableMainCallsAPI.
		EditVariable(ctx, variableID).
		VariableEditRequest(qovery.VariableEditRequest{
			Key:                   request.Key,
			Value:                 *qovery.NewNullableString(&request.Reference),
			SecretManagerAccessId: *qovery.NewNullableString(&smAccessID),
			Description:           *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(p.apiResource, variableID, resp, err)
	}

	return newDomainExternalSecretFromQovery(v)
}

func (p externalSecretsQoveryAPI) Delete(ctx context.Context, variableID string) error {
	resp, err := p.client.VariableMainCallsAPI.
		DeleteVariable(ctx, variableID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		apiErr := apierrors.NewDeleteAPIError(p.apiResource, variableID, resp, err)
		return apiErr
	}

	return nil
}

func (p externalSecretsQoveryAPI) List(ctx context.Context, serviceID string) (variable.ExternalSecrets, error) {
	vars, resp, err := p.client.VariableMainCallsAPI.
		ListVariables(ctx).
		ParentId(serviceID).
		Scope(p.scope).
		IsSecret(false).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(p.apiResource, serviceID, resp, err)
	}

	return newDomainExternalSecretsFromQovery(vars)
}

func newDomainExternalSecretsFromQovery(list *qovery.VariableResponseList) (variable.ExternalSecrets, error) {
	secrets := make(variable.ExternalSecrets, 0)
	for _, v := range list.GetResults() {
		if !strings.EqualFold(string(v.VariableType), "EXTERNAL_SECRET") {
			continue
		}
		s, err := newDomainExternalSecretFromQovery(&v)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, *s)
	}
	return secrets, nil
}

func newDomainExternalSecretFromQovery(v *qovery.VariableResponse) (*variable.ExternalSecret, error) {
	reference := ""
	if v.Value.IsSet() && v.Value.Get() != nil {
		reference = *v.Value.Get()
	}

	smAccessID := ""
	if v.SecretManagerAccessId.IsSet() && v.SecretManagerAccessId.Get() != nil {
		smAccessID = *v.SecretManagerAccessId.Get()
	}

	description := ""
	if v.Description != nil {
		description = *v.Description
	}

	scope, err := variable.NewScopeFromString(string(v.Scope))
	if err != nil {
		return nil, errors.Wrap(err, variable.ErrInvalidScopeParam.Error())
	}
	return &variable.ExternalSecret{
		ID:                    uuid.MustParse(v.GetId()),
		Key:                   v.Key,
		Description:           description,
		Reference:             reference,
		SecretManagerAccessId: smAccessID,
		Scope:                 *scope,
		VariableType:          string(v.VariableType),
	}, nil
}
