package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var externalSecretFileAttrTypes = map[string]attr.Type{
	"id":                       types.StringType,
	"key":                      types.StringType,
	"description":              types.StringType,
	"mount_path":               types.StringType,
	"reference":                types.StringType,
	"secret_manager_access_id": types.StringType,
}

type ExternalSecretFileList []ExternalSecretFileItem

type ExternalSecretFileItem struct {
	Id                    types.String `tfsdk:"id"`
	Key                   types.String `tfsdk:"key"`
	Description           types.String `tfsdk:"description"`
	MountPath             types.String `tfsdk:"mount_path"`
	Reference             types.String `tfsdk:"reference"`
	SecretManagerAccessId types.String `tfsdk:"secret_manager_access_id"`
}

func (list ExternalSecretFileList) toTerraformSet(ctx context.Context) types.Set {
	objectType := types.ObjectType{AttrTypes: externalSecretFileAttrTypes}
	if list == nil {
		return types.SetNull(objectType)
	}

	elements := make([]attr.Value, 0, len(list))
	for _, item := range list {
		elements = append(elements, item.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, objectType, elements)
	if diagnostics.HasError() {
		panic("should never happen")
	}
	return set
}

func (list ExternalSecretFileList) contains(key string) bool {
	for _, item := range list {
		if item.Key.ValueString() == key {
			return true
		}
	}
	return false
}

func (list ExternalSecretFileList) find(key string) *ExternalSecretFileItem {
	for _, item := range list {
		if item.Key.ValueString() == key {
			return &item
		}
	}
	return nil
}

// diffRequest computes Create/Update/Delete ops; mount_path changes trigger delete+recreate.
func (list ExternalSecretFileList) diffRequest(old ExternalSecretFileList) variable.ExternalSecretFileDiffRequest {
	diff := variable.ExternalSecretFileDiffRequest{
		Create: []variable.ExternalSecretFileDiffCreateRequest{},
		Update: []variable.ExternalSecretFileDiffUpdateRequest{},
		Delete: []variable.ExternalSecretFileDiffDeleteRequest{},
	}

	for _, e := range old {
		if updated := list.find(e.Key.ValueString()); updated != nil {
			if updated.MountPath != e.MountPath {
				// mount_path cannot be updated via the API — delete and recreate.
				diff.Delete = append(diff.Delete, variable.ExternalSecretFileDiffDeleteRequest{
					VariableID: e.Id.ValueString(),
				})
				diff.Create = append(diff.Create, variable.ExternalSecretFileDiffCreateRequest{
					ExternalSecretFileUpsertRequest: variable.ExternalSecretFileUpsertRequest{
						Key:                   updated.Key.ValueString(),
						Description:           updated.Description.ValueString(),
						MountPath:             updated.MountPath.ValueString(),
						Reference:             updated.Reference.ValueString(),
						SecretManagerAccessId: updated.SecretManagerAccessId.ValueString(),
					},
				})
			} else if updated.Reference != e.Reference || updated.SecretManagerAccessId != e.SecretManagerAccessId || updated.Description != e.Description {
				diff.Update = append(diff.Update, variable.ExternalSecretFileDiffUpdateRequest{
					VariableID: e.Id.ValueString(),
					ExternalSecretFileUpsertRequest: variable.ExternalSecretFileUpsertRequest{
						Key:                   e.Key.ValueString(),
						Description:           updated.Description.ValueString(),
						MountPath:             e.MountPath.ValueString(),
						Reference:             updated.Reference.ValueString(),
						SecretManagerAccessId: updated.SecretManagerAccessId.ValueString(),
					},
				})
			}
		} else {
			diff.Delete = append(diff.Delete, variable.ExternalSecretFileDiffDeleteRequest{
				VariableID: e.Id.ValueString(),
			})
		}
	}

	for _, e := range list {
		if !old.contains(e.Key.ValueString()) {
			diff.Create = append(diff.Create, variable.ExternalSecretFileDiffCreateRequest{
				ExternalSecretFileUpsertRequest: variable.ExternalSecretFileUpsertRequest{
					Key:                   e.Key.ValueString(),
					Description:           e.Description.ValueString(),
					MountPath:             e.MountPath.ValueString(),
					Reference:             e.Reference.ValueString(),
					SecretManagerAccessId: e.SecretManagerAccessId.ValueString(),
				},
			})
		}
	}

	return diff
}

func (item ExternalSecretFileItem) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":                       item.Id,
		"key":                      item.Key,
		"description":              item.Description,
		"mount_path":               item.MountPath,
		"reference":                item.Reference,
		"secret_manager_access_id": item.SecretManagerAccessId,
	}
	obj, diagnostics := types.ObjectValue(externalSecretFileAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("should never happen")
	}
	return obj
}

func toExternalSecretFileItem(v types.Object) ExternalSecretFileItem {
	return ExternalSecretFileItem{
		Id:                    v.Attributes()["id"].(types.String),
		Key:                   v.Attributes()["key"].(types.String),
		Description:           v.Attributes()["description"].(types.String),
		MountPath:             v.Attributes()["mount_path"].(types.String),
		Reference:             v.Attributes()["reference"].(types.String),
		SecretManagerAccessId: v.Attributes()["secret_manager_access_id"].(types.String),
	}
}

func toExternalSecretFileList(s types.Set) ExternalSecretFileList {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}

	list := make(ExternalSecretFileList, 0, len(s.Elements()))
	for _, elem := range s.Elements() {
		list = append(list, toExternalSecretFileItem(elem.(types.Object)))
	}
	return list
}

func convertDomainExternalSecretFilesToExternalSecretFileList(externalSecretFiles variable.ExternalSecretFiles, planValue types.Set, scope variable.Scope) ExternalSecretFileList {
	planMap := buildExternalSecretFilePlanMap(planValue)
	list := make(ExternalSecretFileList, 0, len(externalSecretFiles))
	for _, externalSecretFile := range externalSecretFiles {
		if externalSecretFile.Scope != scope || externalSecretFile.VariableType != "FILE_EXTERNAL_SECRET" {
			continue
		}
		planItem := planMap[externalSecretFile.Key]
		list = append(list, ExternalSecretFileItem{
			Id:                    FromString(externalSecretFile.ID.String()),
			Key:                   FromString(externalSecretFile.Key),
			Description:           planAwareOptionalString(externalSecretFile.Description, planItem.Description),
			MountPath:             FromString(externalSecretFile.MountPath),
			Reference:             FromString(externalSecretFile.Reference),
			SecretManagerAccessId: FromString(externalSecretFile.SecretManagerAccessId),
		})
	}

	// We need to re-do the check in case the external secrets come from a higher scope
	if len(list) == 0 {
		if planValue.IsNull() {
			return nil
		}
		return ExternalSecretFileList{}
	}

	return list
}

// buildExternalSecretFilePlanMap returns a map keyed by secret file key from the plan set.
func buildExternalSecretFilePlanMap(planValue types.Set) map[string]ExternalSecretFileItem {
	list := toExternalSecretFileList(planValue)
	m := make(map[string]ExternalSecretFileItem, len(list))
	for _, item := range list {
		m[item.Key.ValueString()] = item
	}
	return m
}
