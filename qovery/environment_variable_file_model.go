package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var environmentVariableFileAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"key":         types.StringType,
	"value":       types.StringType,
	"mount_path":  types.StringType,
	"description": types.StringType,
}

type EnvironmentVariableFileList []EnvironmentVariableFile

func (vars EnvironmentVariableFileList) toTerraformSet(ctx context.Context) types.Set {
	environmentVariableFileObjectType := types.ObjectType{
		AttrTypes: environmentVariableFileAttrTypes,
	}
	if vars == nil {
		return types.SetNull(environmentVariableFileObjectType)
	}

	elements := make([]attr.Value, 0, len(vars))
	for _, v := range vars {
		elements = append(elements, v.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, environmentVariableFileObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func (vars EnvironmentVariableFileList) contains(e EnvironmentVariableFile) bool {
	for _, v := range vars {
		if e.Key == v.Key {
			return true
		}
	}
	return false
}

func (vars EnvironmentVariableFileList) find(key string) *EnvironmentVariableFile {
	for _, v := range vars {
		if v.Key.ValueString() == key {
			return &v
		}
	}
	return nil
}

func (vars EnvironmentVariableFileList) diffRequest(old EnvironmentVariableFileList) variable.DiffRequest {
	diff := variable.DiffRequest{
		Create: []variable.DiffCreateRequest{},
		Update: []variable.DiffUpdateRequest{},
		Delete: []variable.DiffDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := vars.find(ToString(e.Key)); updatedVar != nil {
			if updatedVar.MountPath != e.MountPath {
				// mount_path changed — delete + recreate
				diff.Delete = append(diff.Delete, e.toDiffDeleteRequest())
				diff.Create = append(diff.Create, updatedVar.toDiffCreateRequest())
			} else if updatedVar.Value != e.Value || updatedVar.Description != e.Description {
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

func (vars EnvironmentVariableFileList) diff(old EnvironmentVariableFileList) client.EnvironmentVariablesDiff {
	diff := client.EnvironmentVariablesDiff{
		Create: []client.EnvironmentVariableCreateRequest{},
		Update: []client.EnvironmentVariableUpdateRequest{},
		Delete: []client.EnvironmentVariableDeleteRequest{},
	}

	for _, e := range old {
		if updatedVar := vars.find(ToString(e.Key)); updatedVar != nil {
			if updatedVar.MountPath != e.MountPath {
				// mount_path changed — delete + recreate
				diff.Delete = append(diff.Delete, client.EnvironmentVariableDeleteRequest{Id: ToString(e.Id)})
				diff.Create = append(diff.Create, updatedVar.toCreateRequest())
			} else if updatedVar.Value != e.Value || updatedVar.Description != e.Description {
				diff.Update = append(diff.Update, e.toUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, client.EnvironmentVariableDeleteRequest{Id: ToString(e.Id)})
		}
	}

	for _, e := range vars {
		if !old.contains(e) {
			diff.Create = append(diff.Create, e.toCreateRequest())
		}
	}

	return diff
}

type EnvironmentVariableFile struct {
	Id          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	MountPath   types.String `tfsdk:"mount_path"`
	Description types.String `tfsdk:"description"`
}

func (e EnvironmentVariableFile) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":          e.Id,
		"key":         e.Key,
		"value":       e.Value,
		"mount_path":  e.MountPath,
		"description": e.Description,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(environmentVariableFileAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (e EnvironmentVariableFile) toCreateRequest() client.EnvironmentVariableCreateRequest {
	req := client.EnvironmentVariableCreateRequest{
		EnvironmentVariableRequest: qovery.EnvironmentVariableRequest{
			Key:         ToString(e.Key),
			Value:       ToStringPointer(e.Value),
			Description: *qovery.NewNullableString(ToStringPointer(e.Description)),
		},
	}
	mountPath := ToString(e.MountPath)
	if mountPath != "" {
		req.EnvironmentVariableRequest.MountPath = *qovery.NewNullableString(&mountPath)
	}
	return req
}

func (e EnvironmentVariableFile) toDiffCreateRequest() variable.DiffCreateRequest {
	return variable.DiffCreateRequest{
		UpsertRequest: variable.UpsertRequest{
			Key:         ToString(e.Key),
			Value:       ToString(e.Value),
			Description: ToString(e.Description),
			MountPath:   ToString(e.MountPath),
		},
	}
}

func (e EnvironmentVariableFile) toUpdateRequest(new EnvironmentVariableFile) client.EnvironmentVariableUpdateRequest {
	req := client.EnvironmentVariableUpdateRequest{
		Id: ToString(e.Id),
		EnvironmentVariableEditRequest: qovery.EnvironmentVariableEditRequest{
			Key:         ToString(e.Key),
			Value:       ToStringPointer(new.Value),
			Description: *qovery.NewNullableString(ToStringPointer(new.Description)),
		},
	}
	mountPath := ToString(e.MountPath)
	if mountPath != "" {
		req.EnvironmentVariableEditRequest.MountPath = *qovery.NewNullableString(&mountPath)
	}
	return req
}

func (e EnvironmentVariableFile) toDiffUpdateRequest(new EnvironmentVariableFile) variable.DiffUpdateRequest {
	return variable.DiffUpdateRequest{
		VariableID: ToString(e.Id),
		UpsertRequest: variable.UpsertRequest{
			Key:         ToString(e.Key),
			Value:       ToString(new.Value),
			Description: ToString(new.Description),
			MountPath:   ToString(e.MountPath),
		},
	}
}

func (e EnvironmentVariableFile) toDiffDeleteRequest() variable.DiffDeleteRequest {
	return variable.DiffDeleteRequest{
		VariableID: ToString(e.Id),
	}
}

func toEnvironmentVariableFile(v types.Object) EnvironmentVariableFile {
	return EnvironmentVariableFile{
		Id:          v.Attributes()["id"].(types.String),
		Key:         v.Attributes()["key"].(types.String),
		Value:       v.Attributes()["value"].(types.String),
		MountPath:   v.Attributes()["mount_path"].(types.String),
		Description: v.Attributes()["description"].(types.String),
	}
}

func toEnvironmentVariableFileList(vars types.Set) EnvironmentVariableFileList {
	if vars.IsNull() || vars.IsUnknown() {
		return nil
	}

	environmentVariableFiles := make([]EnvironmentVariableFile, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		environmentVariableFiles = append(environmentVariableFiles, toEnvironmentVariableFile(elem.(types.Object)))
	}

	return environmentVariableFiles
}

func convertDomainVariablesToEnvironmentVariableFileListWithNullableInitialState(ctx context.Context, initialState types.Set, vars variable.Variables, scope variable.Scope, variableType string) EnvironmentVariableFileList {
	list := make([]EnvironmentVariableFile, 0, len(vars))
	variableMapByKey := buildVariableFileMap(ctx, initialState)

	for _, v := range vars {
		if v.Scope != scope || v.Type != variableType {
			continue
		}
		currentVariable := variableMapByKey[v.Key]
		list = append(list, convertDomainVariableToEnvironmentVariableFile(v, &currentVariable))
	}

	// Return nil only if list is empty and original state is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func buildVariableFileMap(ctx context.Context, initialState types.Set) map[string]EnvironmentVariableFile {
	initialVariables := make([]EnvironmentVariableFile, 0, len(initialState.Elements()))
	initialState.ElementsAs(ctx, &initialVariables, false)
	variableMapByKey := make(map[string]EnvironmentVariableFile, len(initialVariables))
	for _, currentVariable := range initialVariables {
		variableMapByKey[currentVariable.Key.ValueString()] = currentVariable
	}
	return variableMapByKey
}

func convertDomainVariableToEnvironmentVariableFile(v variable.Variable, variableInState *EnvironmentVariableFile) EnvironmentVariableFile {
	description := FromString(v.Description)
	if variableInState != nil && variableInState.Description.IsNull() {
		description = basetypes.NewStringNull()
	}
	return EnvironmentVariableFile{
		Id:          FromString(v.ID.String()),
		Key:         FromString(v.Key),
		Value:       FromString(v.Value),
		MountPath:   FromString(v.MountPath),
		Description: description,
	}
}

func fromEnvironmentVariableFileList(ctx context.Context, initialState types.Set, vars []*qovery.EnvironmentVariable, scope qovery.APIVariableScopeEnum, variableType string) EnvironmentVariableFileList {
	variableMap := buildVariableFileMap(ctx, initialState)

	list := make([]EnvironmentVariableFile, 0, len(vars))
	for _, v := range vars {
		if v.Scope != scope || string(v.VariableType) != variableType {
			continue
		}
		currentVariable := variableMap[v.Key]
		description := FromNullableString(v.Description)
		if currentVariable.Description.IsNull() && !initialState.IsNull() {
			description = basetypes.NewStringNull()
		}
		mountPath := ""
		if v.MountPath.IsSet() && v.MountPath.Get() != nil {
			mountPath = *v.MountPath.Get()
		}
		list = append(list, EnvironmentVariableFile{
			Id:          FromString(v.Id),
			Key:         FromString(v.Key),
			Value:       FromStringPointer(v.Value),
			MountPath:   FromString(mountPath),
			Description: description,
		})
	}

	// Return nil only if list is empty and original state is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}
