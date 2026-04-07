package qoveryapi

import (
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

func newQoveryEnvSecretVariableRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableRequest {
	req := qovery.VariableRequest{
		Key:              request.Key,
		Value:            request.Value,
		MountPath:        qovery.NullableString{},
		IsSecret:         true,
		VariableScope:    parentScope,
		VariableParentId: parentId,
		Description:      *qovery.NewNullableString(&request.Description),
	}
	if request.MountPath != "" {
		req.MountPath = *qovery.NewNullableString(&request.MountPath)
	}
	return req
}

func newDomainEnvSecretsFromQovery(list *qovery.VariableResponseList) (secret.Secrets, error) {
	vars := make(secret.Secrets, 0, len(list.GetResults()))
	for _, it := range list.GetResults() {
		v, err := newDomainEnvSecretFromQovery(&it)
		if err != nil {
			return nil, err
		}

		vars = append(vars, *v)
	}

	return vars, nil
}

func newDomainEnvSecretFromQovery(v *qovery.VariableResponse) (*secret.Secret, error) {
	if v == nil {
		return nil, secret.ErrNilSecret
	}

	mountPath := ""
	if v.MountPath.IsSet() && v.MountPath.Get() != nil {
		mountPath = *v.MountPath.Get()
	}

	return secret.NewSecret(secret.NewSecretParams{
		SecretID:    v.GetId(),
		Scope:       string(v.Scope),
		Key:         v.GetKey(),
		Type:        string(v.VariableType),
		Description: *v.Description,
		MountPath:   mountPath,
	})
}

func newQoveryEnvSecretEditRequestFromDomain(request secret.UpsertRequest) qovery.VariableEditRequest {
	return qovery.VariableEditRequest{
		Key:         request.Key,
		Value:       *qovery.NewNullableString(&request.Value),
		Description: *qovery.NewNullableString(&request.Description),
	}
}

func newQoveryEnvSecretCreateAliasRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableAliasRequest {
	return qovery.VariableAliasRequest{
		Key:           request.Key,
		AliasScope:    parentScope,
		AliasParentId: parentId,
		Description:   *qovery.NewNullableString(&request.Description),
	}
}

func newQoveryEnvSecretCreateOverrideRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableOverrideRequest {
	return qovery.VariableOverrideRequest{
		Value:            request.Value,
		OverrideScope:    parentScope,
		OverrideParentId: parentId,
		Description:      *qovery.NewNullableString(&request.Description),
	}
}
