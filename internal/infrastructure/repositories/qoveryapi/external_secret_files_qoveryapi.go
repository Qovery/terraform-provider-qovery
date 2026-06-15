package qoveryapi

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var _ variable.ExternalSecretFileRepository = externalSecretFilesQoveryAPI{}

type externalSecretFilesQoveryAPI struct {
	client      *qovery.APIClient
	scope       qovery.APIVariableScopeEnum
	apiResource apierrors.APIResource
}

func newExternalSecretFilesQoveryAPI(client *qovery.APIClient, scope qovery.APIVariableScopeEnum, apiResource apierrors.APIResource) (variable.ExternalSecretFileRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &externalSecretFilesQoveryAPI{
		client:      client,
		scope:       scope,
		apiResource: apiResource,
	}, nil
}

func (p externalSecretFilesQoveryAPI) Create(ctx context.Context, serviceID string, request variable.ExternalSecretFileUpsertRequest) (*variable.ExternalSecretFile, error) {
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
			MountPath:             *qovery.NewNullableString(&request.MountPath),
			Description:           *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(p.apiResource, request.Key, resp, err)
	}

	return newDomainExternalSecretFileFromQovery(v)
}

func (p externalSecretFilesQoveryAPI) Update(ctx context.Context, variableID string, request variable.ExternalSecretFileUpsertRequest) (*variable.ExternalSecretFile, error) {
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

	return newDomainExternalSecretFileFromQovery(v)
}

func (p externalSecretFilesQoveryAPI) Delete(ctx context.Context, variableID string) error {
	resp, err := p.client.VariableMainCallsAPI.
		DeleteVariable(ctx, variableID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(p.apiResource, variableID, resp, err)
	}

	return nil
}

func (p externalSecretFilesQoveryAPI) List(ctx context.Context, serviceID string) (variable.ExternalSecretFiles, error) {
	vars, resp, err := p.client.VariableMainCallsAPI.
		ListVariables(ctx).
		ParentId(serviceID).
		Scope(p.scope).
		IsSecret(false).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(p.apiResource, serviceID, resp, err)
	}

	return newDomainExternalSecretFilesFromQovery(vars)
}

func newDomainExternalSecretFilesFromQovery(list *qovery.VariableResponseList) (variable.ExternalSecretFiles, error) {
	files := make(variable.ExternalSecretFiles, 0)
	for _, v := range list.GetResults() {
		if !strings.EqualFold(string(v.VariableType), "FILE_EXTERNAL_SECRET") {
			continue
		}
		f, err := newDomainExternalSecretFileFromQovery(&v)
		if err != nil {
			return nil, err
		}
		files = append(files, *f)
	}
	return files, nil
}

func newDomainExternalSecretFileFromQovery(v *qovery.VariableResponse) (*variable.ExternalSecretFile, error) {
	reference := ""
	if v.Value.IsSet() && v.Value.Get() != nil {
		reference = *v.Value.Get()
	}

	smAccessID := ""
	if v.SecretManagerAccessId.IsSet() && v.SecretManagerAccessId.Get() != nil {
		smAccessID = *v.SecretManagerAccessId.Get()
	}

	mountPath := ""
	if v.MountPath.IsSet() && v.MountPath.Get() != nil {
		mountPath = *v.MountPath.Get()
	}

	description := ""
	if v.Description != nil {
		description = *v.Description
	}

	return &variable.ExternalSecretFile{
		ID:                    uuid.MustParse(v.GetId()),
		Key:                   v.Key,
		Description:           description,
		MountPath:             mountPath,
		Reference:             reference,
		SecretManagerAccessId: smAccessID,
	}, nil
}
