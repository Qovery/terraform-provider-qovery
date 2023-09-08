package qovery

import (
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

func (ss StorageList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: storageAttrTypes,
		},
	}

	if ss == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		set.Elems = append(set.Elems, v.toTerraformObject())
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
	return types.Object{
		AttrTypes: storageAttrTypes,
		Attrs: map[string]attr.Value{
			"id":          p.ID,
			"type":        p.Type,
			"mount_point": p.MountPoint,
			"size":        p.Size,
		},
	}
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
		ID:         v.Attrs["id"].(types.String),
		Type:       v.Attrs["type"].(types.String),
		MountPoint: v.Attrs["mount_point"].(types.String),
		Size:       v.Attrs["size"].(types.Int64),
	}
}

func toStorageList(vars types.Set) StorageList {
	if vars.Null || vars.Unknown {
		return []Storage{}
	}

	environmentVariables := make([]Storage, 0, len(vars.Elems))
	for _, elem := range vars.Elems {
		environmentVariables = append(environmentVariables, toStorage(elem.(types.Object)))
	}

	return environmentVariables
}
