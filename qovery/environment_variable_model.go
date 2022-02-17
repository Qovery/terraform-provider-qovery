package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type EnvironmentVariableScope string

const (
	EnvironmentVariableScopeApplication EnvironmentVariableScope = "APPLICATION"
	EnvironmentVariableScopeEnvironment EnvironmentVariableScope = "ENVIRONMENT"
	EnvironmentVariableScopeProject     EnvironmentVariableScope = "PROJECT"
)

type EnvironmentVariable struct {
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type EnvironmentVariableDiff struct {
	ToCreate []EnvironmentVariable
	ToUpdate []EnvironmentVariable
	ToRemove []EnvironmentVariable
}

func (e EnvironmentVariable) toCreateRequest() qovery.EnvironmentVariableRequest {
	return qovery.EnvironmentVariableRequest{
		Key:   toString(e.Key),
		Value: toString(e.Value),
	}
}

func (e EnvironmentVariable) toUpdateRequest() qovery.EnvironmentVariableEditRequest {
	return qovery.EnvironmentVariableEditRequest{
		Key:   toString(e.Key),
		Value: toString(e.Value),
	}
}

func containsEnvironmentVariables(env []EnvironmentVariable, v EnvironmentVariable) bool {
	for _, e := range env {
		if e.Key == v.Key && e.Value == v.Value {
			return true
		}
	}
	return false
}

func diffEnvironmentVariables(old, new []EnvironmentVariable) EnvironmentVariableDiff {
	diff := EnvironmentVariableDiff{
		ToCreate: []EnvironmentVariable{},
		ToUpdate: []EnvironmentVariable{},
		ToRemove: []EnvironmentVariable{},
	}

	for _, e := range old {
		if containsEnvironmentVariables(new, e) {
			diff.ToUpdate = append(diff.ToUpdate, e)
		} else {
			diff.ToRemove = append(diff.ToRemove, e)
		}
	}

	for _, e := range new {
		if !containsEnvironmentVariables(old, e) {
			diff.ToCreate = append(diff.ToCreate, e)
		}
	}

	return diff
}

func convertResponseToEnvironmentVariable(v *qovery.EnvironmentVariableResponse) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    fromString(v.Id),
		Key:   fromString(v.Key),
		Value: fromString(v.Value),
	}
}

func convertResponseToEnvironmentVariables(vars *qovery.EnvironmentVariableResponseList, scope EnvironmentVariableScope) []EnvironmentVariable {
	list := make([]EnvironmentVariable, 0, len(vars.GetResults()))
	for _, v := range vars.GetResults() {
		if v.Scope != string(scope) {
			continue
		}
		list = append(list, convertResponseToEnvironmentVariable(&v))
	}
	return list
}
