package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var secretAttrTypes = map[string]attr.Type{
	"id":    types.StringType,
	"key":   types.StringType,
	"value": types.StringType,
}

type SecretList []Secret

func (ss SecretList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: secretAttrTypes,
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
		if toString(v.Key) == key {
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
		if updatedVar := ss.find(toString(s.Key)); updatedVar != nil {
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
		if updatedVar := ss.find(toString(s.Key)); updatedVar != nil {
			if updatedVar.Value != s.Value {
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
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (s Secret) toTerraformObject() types.Object {
	return types.Object{
		AttrTypes: secretAttrTypes,
		Attrs: map[string]attr.Value{
			"id":    s.Id,
			"key":   s.Key,
			"value": s.Value,
		},
	}
}

func (s Secret) toCreateRequest() client.SecretCreateRequest {
	return client.SecretCreateRequest{
		SecretRequest: qovery.SecretRequest{
			Key:   toString(s.Key),
			Value: toString(s.Value),
		},
	}
}

func (s Secret) toDiffCreateRequest() secret.DiffCreateRequest {
	return secret.DiffCreateRequest{
		UpsertRequest: secret.UpsertRequest{
			Key:   toString(s.Key),
			Value: toString(s.Value),
		},
	}
}

func (s Secret) toUpdateRequest(new Secret) client.SecretUpdateRequest {
	return client.SecretUpdateRequest{
		Id: toString(s.Id),
		SecretEditRequest: qovery.SecretEditRequest{
			Key:   toString(s.Key),
			Value: toString(new.Value),
		},
	}
}

func (s Secret) toDiffUpdateRequest(new Secret) secret.DiffUpdateRequest {
	return secret.DiffUpdateRequest{
		SecretID: toString(s.Id),
		UpsertRequest: secret.UpsertRequest{
			Key:   toString(s.Key),
			Value: toString(new.Value),
		},
	}
}

func (s Secret) toDeleteRequest() client.SecretDeleteRequest {
	return client.SecretDeleteRequest{
		Id: toString(s.Id),
	}
}

func (s Secret) toDiffDeleteRequest() secret.DiffDeleteRequest {
	return secret.DiffDeleteRequest{
		SecretID: toString(s.Id),
	}
}

func fromSecret(v *qovery.Secret, state *Secret) Secret {
	sec := Secret{
		Id:  fromString(v.Id),
		Key: fromString(v.Key),
	}
	if state != nil {
		sec.Value = state.Value
	}
	return sec
}

func fromSecretList(state SecretList, secrets []*qovery.Secret, scope qovery.APIVariableScopeEnum) SecretList {
	stateByKey := make(map[string]Secret)
	for _, s := range state {
		stateByKey[toString(s.Key)] = s
	}

	list := make([]Secret, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope {
			continue
		}
		list = append(list, fromSecret(s, state.find(s.GetKey())))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func convertDomainSecretsToSecretList(state SecretList, secrets secret.Secrets, scope variable.Scope) SecretList {
	stateByKey := make(map[string]Secret)
	for _, s := range state {
		stateByKey[toString(s.Key)] = s
	}

	list := make([]Secret, 0, len(secrets))
	for _, s := range secrets {
		if s.Scope != scope {
			continue
		}
		list = append(list, convertDomainSecretToSecret(s, state.find(s.Key)))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func convertDomainSecretToSecret(s secret.Secret, state *Secret) Secret {
	sec := Secret{
		Id:  fromString(s.ID.String()),
		Key: fromString(s.Key),
	}
	if state != nil {
		sec.Value = state.Value
	}
	return sec
}

func toSecret(v types.Object) Secret {
	return Secret{
		Id:    v.Attrs["id"].(types.String),
		Key:   v.Attrs["key"].(types.String),
		Value: v.Attrs["value"].(types.String),
	}
}

func toSecretList(vars types.Set) SecretList {
	if vars.Null || vars.Unknown {
		return []Secret{}
	}

	secrets := make([]Secret, 0, len(vars.Elems))
	for _, elem := range vars.Elems {
		secrets = append(secrets, toSecret(elem.(types.Object)))
	}

	return secrets
}
