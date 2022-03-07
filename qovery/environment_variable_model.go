package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client"
)

type EnvironmentVariable struct {
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (e EnvironmentVariable) toCreateRequest() client.EnvironmentVariableCreateRequest {
	return client.EnvironmentVariableCreateRequest{
		EnvironmentVariableRequest: qovery.EnvironmentVariableRequest{
			Key:   toString(e.Key),
			Value: toString(e.Value),
		},
	}
}

func (e EnvironmentVariable) toUpdateRequest(new EnvironmentVariable) client.EnvironmentVariableUpdateRequest {
	return client.EnvironmentVariableUpdateRequest{
		Id: toString(e.Id),
		EnvironmentVariableEditRequest: qovery.EnvironmentVariableEditRequest{
			Key:   toString(e.Key),
			Value: toString(new.Value),
		},
	}
}

func (e EnvironmentVariable) toDeleteRequest() client.EnvironmentVariableDeleteRequest {
	return client.EnvironmentVariableDeleteRequest{
		Id: toString(e.Id),
	}
}

func findEnvironmentVariables(env []EnvironmentVariable, key string) *EnvironmentVariable {
	for _, e := range env {
		if e.Key.Value == key {
			return &e
		}
	}
	return nil
}

func containsEnvironmentVariables(env []EnvironmentVariable, v EnvironmentVariable) bool {
	for _, e := range env {
		if e.Key == v.Key {
			return true
		}
	}
	return false
}

func diffEnvironmentVariables(old, new []EnvironmentVariable) client.EnvironmentVariablesDiff {
	diff := client.EnvironmentVariablesDiff{
		Create: []client.EnvironmentVariableCreateRequest{},
		Update: []client.EnvironmentVariableUpdateRequest{},
		Delete: []client.EnvironmentVariableDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := findEnvironmentVariables(new, e.Key.Value); updatedVar != nil {
			if updatedVar.Value != e.Value {
				diff.Update = append(diff.Update, e.toUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, e.toDeleteRequest())
		}
	}

	for _, e := range new {
		if !containsEnvironmentVariables(old, e) {
			diff.Create = append(diff.Create, e.toCreateRequest())
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

func convertResponseToEnvironmentVariables(vars []*qovery.EnvironmentVariableResponse) []EnvironmentVariable {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		list = append(list, convertResponseToEnvironmentVariable(v))
	}
	return list
}
