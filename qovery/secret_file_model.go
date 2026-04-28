package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var secretFileAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"key":         types.StringType,
	"value":       types.StringType,
	"mount_path":  types.StringType,
	"description": types.StringType,
}

type SecretFileList []SecretFile

func (ss SecretFileList) toTerraformSet(ctx context.Context) types.Set {
	secretFileObjectType := types.ObjectType{
		AttrTypes: secretFileAttrTypes,
	}
	if ss == nil {
		return types.SetNull(secretFileObjectType)
	}

	elements := make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		elements = append(elements, v.toTerraformObject())
	}

	set, diagnostics := types.SetValueFrom(ctx, secretFileObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func (ss SecretFileList) contains(e SecretFile) bool {
	for _, v := range ss {
		if e.Key == v.Key {
			return true
		}
	}
	return false
}

func (ss SecretFileList) find(key string) *SecretFile {
	for _, v := range ss {
		if ToString(v.Key) == key {
			return &v
		}
	}
	return nil
}

func (ss SecretFileList) diffRequest(old SecretFileList) secret.DiffRequest {
	diff := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{},
		Update: []secret.DiffUpdateRequest{},
		Delete: []secret.DiffDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if updatedVar.MountPath != s.MountPath {
				// mount_path changed — delete + recreate
				diff.Delete = append(diff.Delete, s.toDiffDeleteRequest())
				diff.Create = append(diff.Create, updatedVar.toDiffCreateRequest())
			} else if updatedVar.Value != s.Value || updatedVar.Description != s.Description {
				diff.Update = append(diff.Update, s.toDiffUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, s.toDiffDeleteRequest())
		}
	}

	for _, s := range ss {
		if !old.contains(s) {
			diff.Create = append(diff.Create, s.toDiffCreateRequest())
		}
	}

	return diff
}

func (ss SecretFileList) diff(old SecretFileList) client.SecretsDiff {
	diff := client.SecretsDiff{
		Create: []client.SecretCreateRequest{},
		Update: []client.SecretUpdateRequest{},
		Delete: []client.SecretDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if updatedVar.MountPath != s.MountPath {
				// mount_path changed — delete + recreate
				diff.Delete = append(diff.Delete, client.SecretDeleteRequest{Id: ToString(s.Id)})
				diff.Create = append(diff.Create, updatedVar.toCreateRequest())
			} else if updatedVar.Value != s.Value || updatedVar.Description != s.Description {
				diff.Update = append(diff.Update, s.toUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, client.SecretDeleteRequest{Id: ToString(s.Id)})
		}
	}

	for _, s := range ss {
		if !old.contains(s) {
			diff.Create = append(diff.Create, s.toCreateRequest())
		}
	}

	return diff
}

type SecretFile struct {
	Id          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	MountPath   types.String `tfsdk:"mount_path"`
	Description types.String `tfsdk:"description"`
}

func (s SecretFile) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
		"id":          s.Id,
		"key":         s.Key,
		"value":       s.Value,
		"mount_path":  s.MountPath,
		"description": s.Description,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(secretFileAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (s SecretFile) toCreateRequest() client.SecretCreateRequest {
	req := client.SecretCreateRequest{
		SecretRequest: qovery.SecretRequest{
			Key:         ToString(s.Key),
			Value:       ToStringPointer(s.Value),
			Description: *qovery.NewNullableString(ToStringPointer(s.Description)),
		},
	}
	mountPath := ToString(s.MountPath)
	if mountPath != "" {
		req.SecretRequest.MountPath = *qovery.NewNullableString(&mountPath)
	}
	return req
}

func (s SecretFile) toDiffCreateRequest() secret.DiffCreateRequest {
	return secret.DiffCreateRequest{
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       ToString(s.Value),
			Description: ToString(s.Description),
			MountPath:   ToString(s.MountPath),
		},
	}
}

func (s SecretFile) toUpdateRequest(new SecretFile) client.SecretUpdateRequest {
	// SecretEditRequest does NOT have MountPath — only set key/value/description
	return client.SecretUpdateRequest{
		Id: ToString(s.Id),
		SecretEditRequest: qovery.SecretEditRequest{
			Key:         ToString(s.Key),
			Value:       ToStringPointer(new.Value),
			Description: *qovery.NewNullableString(ToStringPointer(new.Description)),
		},
	}
}

func (s SecretFile) toDiffUpdateRequest(new SecretFile) secret.DiffUpdateRequest {
	return secret.DiffUpdateRequest{
		SecretID: ToString(s.Id),
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       ToString(new.Value),
			Description: ToString(new.Description),
			MountPath:   ToString(s.MountPath),
		},
	}
}

func (s SecretFile) toDiffDeleteRequest() secret.DiffDeleteRequest {
	return secret.DiffDeleteRequest{
		SecretID: ToString(s.Id),
	}
}

func toSecretFile(v types.Object) SecretFile {
	return SecretFile{
		Id:          v.Attributes()["id"].(types.String),
		Key:         v.Attributes()["key"].(types.String),
		Value:       v.Attributes()["value"].(types.String),
		MountPath:   v.Attributes()["mount_path"].(types.String),
		Description: v.Attributes()["description"].(types.String),
	}
}

func toSecretFileList(vars types.Set) SecretFileList {
	if vars.IsNull() || vars.IsUnknown() {
		return nil
	}

	secretFiles := make([]SecretFile, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		secretFiles = append(secretFiles, toSecretFile(elem.(types.Object)))
	}

	return secretFiles
}

func convertDomainSecretsToSecretFileList(initialState types.Set, secrets secret.Secrets, scope variable.Scope, variableType string) SecretFileList {
	stateList := toSecretFileList(initialState)

	list := make([]SecretFile, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope || s.Type != variableType {
			continue
		}
		list = append(list, convertDomainSecretToSecretFile(s, stateList.find(s.Key)))
	}

	// Return nil only if list is empty and original state list is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func convertDomainSecretToSecretFile(s secret.Secret, state *SecretFile) SecretFile {
	sec := SecretFile{
		Id:          FromString(s.ID.String()),
		Key:         FromString(s.Key),
		Description: FromString(s.Description),
	}
	if state != nil {
		// Preserve Value from state (Secret API doesn't return values)
		sec.Value = state.Value
		if state.Description.IsNull() {
			sec.Description = basetypes.NewStringNull()
		}
		// CRITICAL: Preserve MountPath from state when domain model has empty MountPath
		// The legacy Secret API used by Container/Job/Environment/Project doesn't return mount_path
		if s.MountPath != "" {
			sec.MountPath = FromString(s.MountPath)
		} else {
			sec.MountPath = state.MountPath
		}
	} else {
		sec.MountPath = FromString(s.MountPath)
	}
	return sec
}

func fromSecretFileList(initialState types.Set, secrets []*qovery.Secret, scope qovery.APIVariableScopeEnum, secretType string) SecretFileList {
	stateList := toSecretFileList(initialState)

	list := make([]SecretFile, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope || string(*s.VariableType) != secretType {
			continue
		}
		state := stateList.find(s.GetKey())
		sec := SecretFile{
			Id:          FromString(s.Id),
			Key:         FromString(s.Key),
			Description: FromNullableString(s.Description),
		}
		if state != nil {
			// Preserve Value from state (Secret API doesn't return values)
			sec.Value = state.Value
			if state.Description.IsNull() && !initialState.IsNull() {
				sec.Description = basetypes.NewStringNull()
			}
			// CRITICAL: qovery.Secret does NOT have MountPath — must come from state
			sec.MountPath = state.MountPath
		} else {
			// qovery.Secret has no MountPath field — on import, mount_path will need to be
			// re-specified in config. This is a known limitation of the Secret API.
			sec.MountPath = FromString("")
		}

		list = append(list, sec)
	}

	// Return nil only if list is empty and original state list is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}
