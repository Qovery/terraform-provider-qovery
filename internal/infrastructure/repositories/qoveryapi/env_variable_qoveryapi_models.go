package qoveryapi

import (
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func newQoveryEnvVariableRequestFromDomain(request variable.UpsertRequest, isSecret bool, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableRequest {
	return qovery.VariableRequest{
		Key:              request.Key,
		Value:            request.Value,
		MountPath:        qovery.NullableString{},
		IsSecret:         isSecret,
		VariableScope:    parentScope,
		VariableParentId: parentId,
		Description:      *qovery.NewNullableString(&request.Description),
	}
}

func newQoveryEnvVariableCreateAliasRequestFromDomain(request variable.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableAliasRequest {
	return qovery.VariableAliasRequest{
		Key:           request.Key,
		AliasScope:    parentScope,
		AliasParentId: parentId,
		Description:   *qovery.NewNullableString(&request.Description),
	}
}

func newQoveryEnvVariableCreateOverrideRequestFromDomain(request variable.UpsertRequest, parentId string, parentScope qovery.APIVariableScopeEnum) qovery.VariableOverrideRequest {
	return qovery.VariableOverrideRequest{
		Value:            request.Value,
		OverrideScope:    parentScope,
		OverrideParentId: parentId,
		Description:      *qovery.NewNullableString(&request.Description),
	}
}

func newDomainEnvVariablesFromQovery(list *qovery.VariableResponseList) (variable.Variables, error) {
	vars := make(variable.Variables, 0, len(list.GetResults()))
	for _, it := range list.GetResults() {
		v, err := newDomainEnvVariableFromQovery(&it)
		if err != nil {
			return nil, err
		}

		vars = append(vars, *v)
	}

	return vars, nil
}

func newDomainEnvVariableFromQovery(v *qovery.VariableResponse) (*variable.Variable, error) {
	if v == nil {
		return nil, variable.ErrNilVariable
	}

	value := ""
	if v.Value.IsSet() && v.Value.Get() != nil {
		value = *v.Value.Get()
	}

	return variable.NewVariable(variable.NewVariableParams{
		VariableID:  v.GetId(),
		Scope:       string(v.Scope),
		Key:         v.Key,
		Value:       value,
		Type:        string(v.VariableType),
		Description: *v.Description,
	})
}

func newQoveryEnvVariableEditRequestFromDomain(request variable.UpsertRequest) qovery.VariableEditRequest {
	return qovery.VariableEditRequest{
		Key:         request.Key,
		Value:       *qovery.NewNullableString(&request.Value),
		Description: *qovery.NewNullableString(&request.Description),
	}
}
