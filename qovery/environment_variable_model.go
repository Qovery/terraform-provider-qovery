package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

var environmentVariableAttrTypes = map[string]attr.Type{
	"id":    types.StringType,
	"key":   types.StringType,
	"value": types.StringType,
}

type EnvironmentVariableList []EnvironmentVariable

func (vars EnvironmentVariableList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: environmentVariableAttrTypes,
		},
	}

	if vars == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(vars))
	for _, v := range vars {
		set.Elems = append(set.Elems, v.toTerraformObject())
	}
	return set
}

func (vars EnvironmentVariableList) contains(e EnvironmentVariable) bool {
	for _, v := range vars {
		if e.Key == v.Key {
			return true
		}
	}
	return false
}

func (vars EnvironmentVariableList) find(key string) *EnvironmentVariable {
	for _, v := range vars {
		if v.Key.Value == key {
			return &v
		}
	}
	return nil
}

func (vars EnvironmentVariableList) diff(old EnvironmentVariableList) client.EnvironmentVariablesDiff {
	diff := client.EnvironmentVariablesDiff{
		Create: []client.EnvironmentVariableCreateRequest{},
		Update: []client.EnvironmentVariableUpdateRequest{},
		Delete: []client.EnvironmentVariableDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := vars.find(toString(e.Key)); updatedVar != nil {
			if updatedVar.Value != e.Value {
				diff.Update = append(diff.Update, e.toUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, e.toDeleteRequest())
		}
	}

	for _, e := range vars {
		if !old.contains(e) {
			diff.Create = append(diff.Create, e.toCreateRequest())
		}
	}

	return diff
}

type EnvironmentVariable struct {
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (e EnvironmentVariable) toTerraformObject() types.Object {
	return types.Object{
		AttrTypes: environmentVariableAttrTypes,
		Attrs: map[string]attr.Value{
			"id":    e.Id,
			"key":   e.Key,
			"value": e.Value,
		},
	}
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

func fromEnvironmentVariable(v *qovery.EnvironmentVariable) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    fromString(v.Id),
		Key:   fromString(v.Key),
		Value: fromString(v.Value),
	}
}

func fromEnvironmentVariableList(vars []*qovery.EnvironmentVariable, scope qovery.EnvironmentVariableScopeEnum) EnvironmentVariableList {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope {
			continue
		}
		list = append(list, fromEnvironmentVariable(v))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func toEnvironmentVariable(v types.Object) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    v.Attrs["id"].(types.String),
		Key:   v.Attrs["key"].(types.String),
		Value: v.Attrs["value"].(types.String),
	}
}

func toEnvironmentVariableList(vars types.Set) EnvironmentVariableList {
	if vars.Null || vars.Unknown {
		return nil
	}

	environmentVariables := make([]EnvironmentVariable, 0, len(vars.Elems))
	for _, elem := range vars.Elems {
		environmentVariables = append(environmentVariables, toEnvironmentVariable(elem.(types.Object)))
	}

	return environmentVariables
}
