package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

var storageAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"type":        types.StringType,
	"mount_point": types.StringType,
	"size":        types.Int64Type,
}

type StorageList []Storage

func (ss StorageList) toTerraformSet(ctx context.Context) types.Set {
	var storageObjectType = types.ObjectType{
		AttrTypes: storageAttrTypes,
	}
	if ss == nil {
		return types.SetNull(storageObjectType)
	}

	var elements = make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		elements = append(elements, v.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, storageObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

type Storage struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	MountPoint types.String `tfsdk:"mount_point"`
	Size       types.Int64  `tfsdk:"size"`
}

func (p Storage) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"id":          p.ID,
		"type":        p.Type,
		"mount_point": p.MountPoint,
		"size":        p.Size,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(storageAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (p Storage) toUpsertRequest() storage.UpsertRequest {
	return storage.UpsertRequest{
		ID:         ToStringPointer(p.ID),
		Type:       ToString(p.Type),
		MountPoint: ToString(p.MountPoint),
		Size:       ToInt32(p.Size),
	}
}

func fromStorage(p storage.Storage) Storage {
	return Storage{
		ID:         FromString(p.ID.String()),
		Type:       FromString(p.Type.String()),
		MountPoint: FromString(p.MountPoint),
		Size:       FromInt32(p.Size),
	}
}

func fromStorageList(state StorageList, storages storage.Storages) StorageList {
	list := make([]Storage, 0, len(storages))
	for _, s := range storages {
		list = append(list, fromStorage(s))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func convertDomainStoragesToStorageList(initialState types.Set, storages storage.Storages) StorageList {
	list := make([]Storage, 0, len(storages))
	for _, s := range storages {
		list = append(list, convertDomainStorageToStorage(s))
	}

	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	return list
}

func convertDomainStorageToStorage(s storage.Storage) Storage {
	return Storage{
		ID:         FromString(s.ID.String()),
		Type:       FromString(s.Type.String()),
		MountPoint: FromString(s.MountPoint),
		Size:       FromInt32(s.Size),
	}
}

func toStorage(v types.Object) Storage {
	return Storage{
		ID:         v.Attributes()["id"].(types.String),
		Type:       v.Attributes()["type"].(types.String),
		MountPoint: v.Attributes()["mount_point"].(types.String),
		Size:       v.Attributes()["size"].(types.Int64),
	}
}

func toStorageList(vars types.Set) StorageList {
	if vars.IsNull() || vars.IsUnknown() {
		return []Storage{}
	}

	environmentVariables := make([]Storage, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		environmentVariables = append(environmentVariables, toStorage(elem.(types.Object)))
	}

	return environmentVariables
}
