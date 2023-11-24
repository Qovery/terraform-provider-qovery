package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var environmentVariableAttrTypes = map[string]attr.Type{
	"id":    types.StringType,
	"key":   types.StringType,
	"value": types.StringType,
}

type EnvironmentVariableList []EnvironmentVariable

func (vars EnvironmentVariableList) toTerraformSet(ctx context.Context) types.Set {
	var environmentVariableObjectType = types.ObjectType{
		AttrTypes: environmentVariableAttrTypes,
	}
	if vars == nil {
		return types.SetNull(environmentVariableObjectType)
	}

	var elements = make([]attr.Value, 0, len(vars))
	for _, v := range vars {
		elements = append(elements, v.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, environmentVariableObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
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
		if v.Key.ValueString() == key {
			return &v
		}
	}
	return nil
}

func (vars EnvironmentVariableList) diffRequest(old EnvironmentVariableList) variable.DiffRequest {
	diff := variable.DiffRequest{
		Create: []variable.DiffCreateRequest{},
		Update: []variable.DiffUpdateRequest{},
		Delete: []variable.DiffDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := vars.find(ToString(e.Key)); updatedVar != nil {
			if updatedVar.Value != e.Value {
				diff.Update = append(diff.Update, e.toDiffUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, e.toDiffDeleteRequest())
		}
	}

	for _, e := range vars {
		if !old.contains(e) {
			diff.Create = append(diff.Create, e.toDiffCreateRequest())
		}
	}

	return diff
}

func (vars EnvironmentVariableList) diff(old EnvironmentVariableList) client.EnvironmentVariablesDiff {
	diff := client.EnvironmentVariablesDiff{
		Create: []client.EnvironmentVariableCreateRequest{},
		Update: []client.EnvironmentVariableUpdateRequest{},
		Delete: []client.EnvironmentVariableDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := vars.find(ToString(e.Key)); updatedVar != nil {
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
	var attributes = map[string]attr.Value{
		"id":    e.Id,
		"key":   e.Key,
		"value": e.Value,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(environmentVariableAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (e EnvironmentVariable) toCreateRequest() client.EnvironmentVariableCreateRequest {
	return client.EnvironmentVariableCreateRequest{
		EnvironmentVariableRequest: qovery.EnvironmentVariableRequest{
			Key:   ToString(e.Key),
			Value: ToStringPointer(e.Value),
		},
	}
}

func (e EnvironmentVariable) toDiffCreateRequest() variable.DiffCreateRequest {
	return variable.DiffCreateRequest{
		UpsertRequest: variable.UpsertRequest{
			Key:   ToString(e.Key),
			Value: ToString(e.Value),
		},
	}
}

func (e EnvironmentVariable) toUpdateRequest(new EnvironmentVariable) client.EnvironmentVariableUpdateRequest {
	return client.EnvironmentVariableUpdateRequest{
		Id: ToString(e.Id),
		EnvironmentVariableEditRequest: qovery.EnvironmentVariableEditRequest{
			Key:   ToString(e.Key),
			Value: ToStringPointer(new.Value),
		},
	}
}

func (e EnvironmentVariable) toDiffUpdateRequest(new EnvironmentVariable) variable.DiffUpdateRequest {
	return variable.DiffUpdateRequest{
		VariableID: ToString(e.Id),
		UpsertRequest: variable.UpsertRequest{
			Key:   ToString(e.Key),
			Value: ToString(new.Value),
		},
	}
}

func (e EnvironmentVariable) toDeleteRequest() client.EnvironmentVariableDeleteRequest {
	return client.EnvironmentVariableDeleteRequest{
		Id: ToString(e.Id),
	}
}

func (e EnvironmentVariable) toDiffDeleteRequest() variable.DiffDeleteRequest {
	return variable.DiffDeleteRequest{
		VariableID: ToString(e.Id),
	}
}

func fromEnvironmentVariable(v *qovery.EnvironmentVariable) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    FromString(v.Id),
		Key:   FromString(v.Key),
		Value: FromStringPointer(v.Value),
	}
}

func fromEnvironmentVariableList(vars []*qovery.EnvironmentVariable, scope qovery.APIVariableScopeEnum, variableType string) EnvironmentVariableList {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope || string(*v.VariableType) != variableType {
			continue
		}
		list = append(list, fromEnvironmentVariable(v))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func fromEnvironmentVariableListWithNullableInitialState(initialState types.Set, vars []*qovery.EnvironmentVariable, scope qovery.APIVariableScopeEnum, variableType string) EnvironmentVariableList {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope || string(*v.VariableType) != variableType {
			continue
		}
		list = append(list, fromEnvironmentVariable(v))
	}

	// Return nil only if list is empty and original state is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func toEnvironmentVariable(v types.Object) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    v.Attributes()["id"].(types.String),
		Key:   v.Attributes()["key"].(types.String),
		Value: v.Attributes()["value"].(types.String),
	}
}

func toEnvironmentVariableList(vars types.Set) EnvironmentVariableList {
	if vars.IsNull() || vars.IsUnknown() {
		return nil
	}

	environmentVariables := make([]EnvironmentVariable, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		environmentVariables = append(environmentVariables, toEnvironmentVariable(elem.(types.Object)))
	}

	return environmentVariables
}

func convertDomainVariablesToEnvironmentVariableList(vars variable.Variables, scope variable.Scope, variableType string) EnvironmentVariableList {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope || v.Type != variableType {
			continue
		}
		list = append(list, convertDomainVariableToEnvironmentVariable(v))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(initialState types.Set, vars variable.Variables, scope variable.Scope, variableType string) EnvironmentVariableList {
	list := make([]EnvironmentVariable, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope || v.Type != variableType {
			continue
		}
		list = append(list, convertDomainVariableToEnvironmentVariable(v))
	}

	// Return nil only if list is empty and original state is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func convertDomainVariableToEnvironmentVariable(v variable.Variable) EnvironmentVariable {
	return EnvironmentVariable{
		Id:    FromString(v.ID.String()),
		Key:   FromString(v.Key),
		Value: FromString(v.Value),
	}
}
