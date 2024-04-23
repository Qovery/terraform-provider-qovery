package qoveryapi

import (
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
)

func newQoveryEnvSecretVariableRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableRequest {
	return qovery.VariableRequest{
		Key:              request.Key,
		Value:            request.Value,
		MountPath:        qovery.NullableString{},
		IsSecret:         true,
		VariableScope:    parentScope,
		VariableParentId: parentId,
	}
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

	return secret.NewSecret(secret.NewSecretParams{
		SecretID: v.GetId(),
		Scope:    string(v.Scope),
		Key:      v.GetKey(),
		Type:     string(v.VariableType),
	})
}

func newQoveryEnvSecretEditRequestFromDomain(request secret.UpsertRequest) qovery.VariableEditRequest {
	return qovery.VariableEditRequest{
		Key:   request.Key,
		Value: request.Value,
	}
}

func newQoveryEnvSecretCreateAliasRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableAliasRequest {
	return qovery.VariableAliasRequest{
		Key:           request.Key,
		AliasScope:    parentScope,
		AliasParentId: parentId,
	}
}

func newQoveryEnvSecretCreateOverrideRequestFromDomain(request secret.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableOverrideRequest {
	return qovery.VariableOverrideRequest{
		Value:            request.Value,
		OverrideScope:    parentScope,
		OverrideParentId: parentId,
	}
}
