package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainVariablesFromQovery(list *qovery.EnvironmentVariableResponseList) (variable.Variables, error) {
	vars := make(variable.Variables, 0, len(list.GetResults()))
	for _, it := range list.GetResults() {
		v, err := newDomainVariableFromQovery(&it)
		if err != nil {
			return nil, err
		}

		vars = append(vars, *v)
	}

	return vars, nil
}

// newDomainCredentialsFromQovery takes a qovery.EnvironmentVariable returned by the API client and turns it into the domain model variable.Variable.
func newDomainVariableFromQovery(v *qovery.EnvironmentVariable) (*variable.Variable, error) {
	if v == nil {
		return nil, variable.ErrNilVariable
	}

	value := ""
	if v.Value != nil {
		value = *v.Value
	}
	description := ""
	if v.Description.IsSet() {
		description = *v.Description.Get()
	}

	return variable.NewVariable(variable.NewVariableParams{
		VariableID:  v.GetId(),
		Scope:       string(v.Scope),
		Key:         v.Key,
		Value:       value,
		Type:        string(v.VariableType),
		Description: description,
	})
}

// newQoveryEnvironmentVariableRequestFromDomain takes the domain request variable.UpsertRequest and turns it into a qovery.EnvironmentVariableRequest to make the api call.
func newQoveryEnvironmentVariableRequestFromDomain(request variable.UpsertRequest) qovery.EnvironmentVariableRequest {
	return qovery.EnvironmentVariableRequest{
		Key:         request.Key,
		Value:       &request.Value,
		Description: *qovery.NewNullableString(&request.Description),
	}
}

// newQoveryEnvironmentVariableEditRequestFromDomain takes the domain request variable.UpsertRequest and turns it into a qovery.EnvironmentVariableEditRequest to make the api call.
func newQoveryEnvironmentVariableEditRequestFromDomain(request variable.UpsertRequest) qovery.EnvironmentVariableEditRequest {
	return qovery.EnvironmentVariableEditRequest{
		Key:         request.Key,
		Value:       &request.Value,
		Description: *qovery.NewNullableString(&request.Description),
	}
}
