package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var secretAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"key":         types.StringType,
	"value":       types.StringType,
	"description": types.StringType,
}

type SecretList []Secret

func (ss SecretList) toTerraformSet(ctx context.Context) types.Set {
	var secretObjectType = types.ObjectType{
		AttrTypes: secretAttrTypes,
	}
	if ss == nil {
		return types.SetNull(secretObjectType)
	}

	var elements = make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		elements = append(elements, v.toTerraformObject())
	}

	set, diagnostics := types.SetValueFrom(ctx, secretObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

func (ss SecretList) contains(e Secret) bool {
	for _, v := range ss {
		if e.Key == v.Key {
			return true
		}
	}
	return false
}

func (ss SecretList) find(key string) *Secret {
	for _, v := range ss {
		if ToString(v.Key) == key {
			return &v
		}
	}
	return nil
}

func (ss SecretList) diff(old SecretList) client.SecretsDiff {
	diff := client.SecretsDiff{
		Create: []client.SecretCreateRequest{},
		Update: []client.SecretUpdateRequest{},
		Delete: []client.SecretDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if updatedVar.Value != s.Value {
				diff.Update = append(diff.Update, s.toUpdateRequest(*updatedVar))
			}
		} else {
			diff.Delete = append(diff.Delete, s.toDeleteRequest())
		}
	}

	for _, s := range ss {
		if !old.contains(s) {
			diff.Create = append(diff.Create, s.toCreateRequest())
		}
	}

	return diff
}

func (ss SecretList) diffRequest(old SecretList) secret.DiffRequest {
	diff := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{},
		Update: []secret.DiffUpdateRequest{},
		Delete: []secret.DiffDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if updatedVar.Value != s.Value || updatedVar.Description != s.Description {
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

type Secret struct {
	Id          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

func (s Secret) toTerraformObject() types.Object {
	var attributes = map[string]attr.Value{
		"id":          s.Id,
		"key":         s.Key,
		"value":       s.Value,
		"description": s.Description,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(secretAttrTypes, attributes)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return terraformObjectValue
}

func (s Secret) toCreateRequest() client.SecretCreateRequest {
	return client.SecretCreateRequest{
		SecretRequest: qovery.SecretRequest{
			Key:         ToString(s.Key),
			Value:       ToStringPointer(s.Value),
			Description: *qovery.NewNullableString(ToStringPointer(s.Description)),
		},
	}
}

func (s Secret) toDiffCreateRequest() secret.DiffCreateRequest {
	return secret.DiffCreateRequest{
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       ToString(s.Value),
			Description: ToString(s.Description),
		},
	}
}

func (s Secret) toUpdateRequest(new Secret) client.SecretUpdateRequest {
	return client.SecretUpdateRequest{
		Id: ToString(s.Id),
		SecretEditRequest: qovery.SecretEditRequest{
			Key:         ToString(s.Key),
			Value:       ToStringPointer(new.Value),
			Description: *qovery.NewNullableString(ToStringPointer(s.Description)),
		},
	}
}

func (s Secret) toDiffUpdateRequest(new Secret) secret.DiffUpdateRequest {
	return secret.DiffUpdateRequest{
		SecretID: ToString(s.Id),
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       ToString(new.Value),
			Description: ToString(new.Description),
		},
	}
}

func (s Secret) toDeleteRequest() client.SecretDeleteRequest {
	return client.SecretDeleteRequest{
		Id: ToString(s.Id),
	}
}

func (s Secret) toDiffDeleteRequest() secret.DiffDeleteRequest {
	return secret.DiffDeleteRequest{
		SecretID: ToString(s.Id),
	}
}

func fromSecret(v *qovery.Secret, state *Secret) Secret {
	sec := Secret{
		Id:          FromString(v.Id),
		Key:         FromString(v.Key),
		Description: FromNullableString(v.Description),
	}
	if state != nil {
		sec.Value = state.Value
		if state.Description.IsNull() {
			sec.Description = basetypes.NewStringNull()
		}
	}
	return sec
}

func fromSecretList(initialState types.Set, secrets []*qovery.Secret, scope qovery.APIVariableScopeEnum, secretType string) SecretList {
	stateByKey := make(map[string]Secret)
	state := ToSecretList(initialState)

	for _, s := range state {
		stateByKey[ToString(s.Key)] = s
	}

	list := make([]Secret, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope || string(*s.VariableType) != secretType {
			continue
		}
		list = append(list, fromSecret(s, state.find(s.GetKey())))
	}

	// Return nil only if list is empty and original state list is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func convertDomainSecretsToSecretList(initialState types.Set, secrets secret.Secrets, scope variable.Scope, variableType string) SecretList {
	stateByKey := make(map[string]Secret)
	state := ToSecretList(initialState)

	for _, s := range state {
		stateByKey[ToString(s.Key)] = s
	}

	list := make([]Secret, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope || s.Type != variableType {
			continue
		}
		list = append(list, convertDomainSecretToSecret(s, state.find(s.Key)))
	}

	// Return nil only if list is empty and original state list is nil
	if len(list) == 0 && initialState.IsNull() {
		return nil
	}
	// Otherwise return the list, even empty (`[]` in the terraform file)
	return list
}

func convertDomainSecretToSecret(s secret.Secret, state *Secret) Secret {
	sec := Secret{
		Id:          FromString(s.ID.String()),
		Key:         FromString(s.Key),
		Description: FromString(s.Description),
	}
	if state != nil {
		sec.Value = state.Value
		if state.Description.IsNull() {
			sec.Description = basetypes.NewStringNull()
		}
	}
	return sec
}

func toSecret(v types.Object) Secret {
	return Secret{
		Id:          v.Attributes()["id"].(types.String),
		Key:         v.Attributes()["key"].(types.String),
		Value:       v.Attributes()["value"].(types.String),
		Description: v.Attributes()["description"].(types.String),
	}
}

func ToSecretList(vars types.Set) SecretList {
	if vars.IsNull() || vars.IsUnknown() {
		return []Secret{}
	}

	secrets := make([]Secret, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		secrets = append(secrets, toSecret(elem.(types.Object)))
	}

	return secrets
}
