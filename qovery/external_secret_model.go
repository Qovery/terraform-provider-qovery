package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var externalSecretAttrTypes = map[string]attr.Type{
	"id":                       types.StringType,
	"key":                      types.StringType,
	"reference":                types.StringType,
	"secret_manager_access_id": types.StringType,
}

type ExternalSecretList []ExternalSecretItem

type ExternalSecretItem struct {
	Id                    types.String `tfsdk:"id"`
	Key                   types.String `tfsdk:"key"`
	Reference             types.String `tfsdk:"reference"`
	SecretManagerAccessId types.String `tfsdk:"secret_manager_access_id"`
}

func (list ExternalSecretList) toTerraformSet(ctx context.Context) types.Set {
	objectType := types.ObjectType{AttrTypes: externalSecretAttrTypes}
	if list == nil {
		return types.SetNull(objectType)
	}

	elements := make([]attr.Value, 0, len(list))
	for _, item := range list {
		elements = append(elements, item.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, objectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func (list ExternalSecretList) contains(key string) bool {
	for _, item := range list {
		if item.Key.ValueString() == key {
			return true
		}
	}
	return false
}

func (list ExternalSecretList) find(key string) *ExternalSecretItem {
	for _, item := range list {
		if item.Key.ValueString() == key {
			return &item
		}
	}
	return nil
}

func (list ExternalSecretList) diffRequest(old ExternalSecretList) variable.ExternalSecretDiffRequest {
	diff := variable.ExternalSecretDiffRequest{
		Create: []variable.ExternalSecretDiffCreateRequest{},
		Update: []variable.ExternalSecretDiffUpdateRequest{},
		Delete: []variable.ExternalSecretDiffDeleteRequest{},
	}

	for _, e := range old {
		if updated := list.find(e.Key.ValueString()); updated != nil {
			if updated.Reference != e.Reference || updated.SecretManagerAccessId != e.SecretManagerAccessId {
				diff.Update = append(diff.Update, variable.ExternalSecretDiffUpdateRequest{
					VariableID: e.Id.ValueString(),
					ExternalSecretUpsertRequest: variable.ExternalSecretUpsertRequest{
						Key:                   e.Key.ValueString(),
						Reference:             updated.Reference.ValueString(),
						SecretManagerAccessId: updated.SecretManagerAccessId.ValueString(),
					},
				})
			}
		} else {
			diff.Delete = append(diff.Delete, variable.ExternalSecretDiffDeleteRequest{
				VariableID: e.Id.ValueString(),
			})
		}
	}

	for _, e := range list {
		if !old.contains(e.Key.ValueString()) {
			diff.Create = append(diff.Create, variable.ExternalSecretDiffCreateRequest{
				ExternalSecretUpsertRequest: variable.ExternalSecretUpsertRequest{
					Key:                   e.Key.ValueString(),
					Reference:             e.Reference.ValueString(),
					SecretManagerAccessId: e.SecretManagerAccessId.ValueString(),
				},
			})
		}
	}

	return diff
}

func (item ExternalSecretItem) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":                       item.Id,
		"key":                      item.Key,
		"reference":                item.Reference,
		"secret_manager_access_id": item.SecretManagerAccessId,
	}
	obj, diagnostics := types.ObjectValue(externalSecretAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return obj
}

func toExternalSecretItem(v types.Object) ExternalSecretItem {
	return ExternalSecretItem{
		Id:                    v.Attributes()["id"].(types.String),
		Key:                   v.Attributes()["key"].(types.String),
		Reference:             v.Attributes()["reference"].(types.String),
		SecretManagerAccessId: v.Attributes()["secret_manager_access_id"].(types.String),
	}
}

func toExternalSecretList(s types.Set) ExternalSecretList {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}

	list := make(ExternalSecretList, 0, len(s.Elements()))
	for _, elem := range s.Elements() {
		list = append(list, toExternalSecretItem(elem.(types.Object)))
	}
	return list
}

// applyApplicationExternalSecretsDiff applies external secret create/update/delete operations for the application resource.
// This is used in the Terraform resource layer (not the service layer) for resources that use the old client pattern.
func applyApplicationExternalSecretsDiff(ctx context.Context, repo variable.ExternalSecretRepository, serviceID string, diff variable.ExternalSecretDiffRequest) error {
	for _, d := range diff.Delete {
		if err := repo.Delete(ctx, d.VariableID); err != nil {
			return fmt.Errorf("failed to delete external secret: %w", err)
		}
	}

	for _, c := range diff.Create {
		if _, err := repo.Create(ctx, serviceID, c.ExternalSecretUpsertRequest); err != nil {
			return fmt.Errorf("failed to create external secret: %w", err)
		}
	}

	for _, u := range diff.Update {
		if _, err := repo.Update(ctx, u.VariableID, u.ExternalSecretUpsertRequest); err != nil {
			return fmt.Errorf("failed to update external secret: %w", err)
		}
	}

	return nil
}

func convertDomainExternalSecretsToExternalSecretList(secrets variable.ExternalSecrets) ExternalSecretList {
	if len(secrets) == 0 {
		return nil
	}
	list := make(ExternalSecretList, 0, len(secrets))
	for _, s := range secrets {
		list = append(list, ExternalSecretItem{
			Id:                    FromString(s.ID.String()),
			Key:                   FromString(s.Key),
			Reference:             FromString(s.Reference),
			SecretManagerAccessId: FromString(s.SecretManagerAccessId),
		})
	}
	return list
}
