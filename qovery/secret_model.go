package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// secretAttrTypes is the element type of secret_aliases, secret_overrides, and of secrets on the
// data sources.
var secretAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"key":         types.StringType,
	"value":       types.StringType,
	"description": types.StringType,
}

// secretWithWriteOnlyAttrTypes is the element type of secrets on the resources, which can set the
// value write-only (see write_only.go).
var secretWithWriteOnlyAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"key":              types.StringType,
	"value":            types.StringType,
	"value_wo":         types.StringType,
	"value_wo_version": types.StringType,
	"description":      types.StringType,
}

// secretCollection is the value of a secret attribute: the secrets list, or the secret_aliases
// and secret_overrides sets.
type secretCollection interface {
	IsNull() bool
	IsUnknown() bool
	Elements() []attr.Value
}

// writeOnlySecretValues holds the value_wo of the secrets entries that set it, by key, read from
// the config. The plan and the state never hold them.
type writeOnlySecretValues map[string]types.String

type SecretList []Secret

func (ss SecretList) toTerraformSet(ctx context.Context) types.Set {
	secretObjectType := types.ObjectType{
		AttrTypes: secretAttrTypes,
	}
	if ss == nil {
		return types.SetNull(secretObjectType)
	}

	elements := make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		elements = append(elements, v.toTerraformObject())
	}

	set, diagnostics := types.SetValueFrom(ctx, secretObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
	}
	return set
}

// toTerraformList builds the secrets attribute of a resource, in the order of ss.
func (ss SecretList) toTerraformList() types.List {
	secretObjectType := types.ObjectType{AttrTypes: secretWithWriteOnlyAttrTypes}
	if ss == nil {
		return types.ListNull(secretObjectType)
	}

	elements := make([]attr.Value, 0, len(ss))
	for _, v := range ss {
		elements = append(elements, v.toWriteOnlyTerraformObject())
	}
	return types.ListValueMust(secretObjectType, elements)
}

// secretsForDataSource drops value_wo and value_wo_version from the secrets the resource converters
// build, which the data sources do not have.
func secretsForDataSource(secrets types.List) types.List {
	secretObjectType := types.ObjectType{AttrTypes: secretAttrTypes}
	if secrets.IsNull() {
		return types.ListNull(secretObjectType)
	}
	if secrets.IsUnknown() {
		return types.ListUnknown(secretObjectType)
	}

	elements := make([]attr.Value, 0, len(secrets.Elements()))
	for _, v := range ToSecretList(secrets) {
		elements = append(elements, v.toTerraformObject())
	}
	return types.ListValueMust(secretObjectType, elements)
}

// secretValuesWoFromConfig reads the value_wo of every secrets entry that sets it from the config.
//
// It also fails for a planned entry that the diff with old would create or update while it has no
// value: a null plan value and no value_wo under its key in the config. ignore_changes on secrets
// can plan a key the config no longer has, and sending that entry would blank the secret. old is
// nil on create.
func secretValuesWoFromConfig(ctx context.Context, config tfsdk.Config, planned, old SecretList, diags *diag.Diagnostics) writeOnlySecretValues {
	var secrets types.List
	diags.Append(config.GetAttribute(ctx, path.Root("secrets"), &secrets)...)
	values := writeOnlySecretValues{}
	for _, s := range ToSecretList(secrets) {
		if !s.ValueWo.IsNull() {
			values[ToString(s.Key)] = s.ValueWo
		}
	}

	for _, s := range planned {
		key := ToString(s.Key)
		if !s.Value.IsNull() || isKnownAndSet(values[key]) {
			continue
		}
		if prior := old.find(key); prior != nil && !prior.changedIn(s) {
			continue
		}
		diags.AddAttributeError(
			path.Root("secrets"),
			"Missing secret value",
			fmt.Sprintf("The plan creates or updates the secret %q, but the configuration sets no value or value_wo under that key, so Qovery would store an empty value. "+
				"An ignore_changes rule on secrets can keep a key the configuration no longer has.", key),
		)
	}
	return values
}

// inPriorOrder returns the entries of ss whose key prior holds in the order of prior, then the
// other entries in the order of ss. The secrets list keeps the order of the configuration this
// way, and a key added outside Terraform comes last.
func (ss SecretList) inPriorOrder(prior SecretList) SecretList {
	if ss == nil {
		return nil
	}
	ordered := make(SecretList, 0, len(ss))
	placed := make(map[string]bool, len(ss))
	for _, p := range prior {
		if s := ss.find(ToString(p.Key)); s != nil && !placed[ToString(s.Key)] {
			ordered = append(ordered, *s)
			placed[ToString(s.Key)] = true
		}
	}
	for _, s := range ss {
		if !placed[ToString(s.Key)] {
			ordered = append(ordered, s)
		}
	}
	return ordered
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

// diff compares the planned secrets ss with the state secrets old, by key. valuesWo holds the
// value_wo of the planned entries that set it, nil for the attributes without a write-only form.
func (ss SecretList) diff(old SecretList, valuesWo writeOnlySecretValues) client.SecretsDiff {
	diff := client.SecretsDiff{
		Create: []client.SecretCreateRequest{},
		Update: []client.SecretUpdateRequest{},
		Delete: []client.SecretDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if s.changedIn(*updatedVar) {
				diff.Update = append(diff.Update, s.toUpdateRequest(*updatedVar, valuesWo[ToString(s.Key)]))
			}
		} else {
			diff.Delete = append(diff.Delete, s.toDeleteRequest())
		}
	}

	for _, s := range ss {
		if !old.contains(s) {
			diff.Create = append(diff.Create, s.toCreateRequest(valuesWo[ToString(s.Key)]))
		}
	}

	return diff
}

// diffRequest compares the planned secrets ss with the state secrets old, by key. valuesWo holds
// the value_wo of the planned entries that set it, nil for the attributes without a write-only
// form.
func (ss SecretList) diffRequest(old SecretList, valuesWo writeOnlySecretValues) secret.DiffRequest {
	diff := secret.DiffRequest{
		Create: []secret.DiffCreateRequest{},
		Update: []secret.DiffUpdateRequest{},
		Delete: []secret.DiffDeleteRequest{},
	}

	for _, s := range old {
		if updatedVar := ss.find(ToString(s.Key)); updatedVar != nil {
			if s.changedIn(*updatedVar) {
				diff.Update = append(diff.Update, s.toDiffUpdateRequest(*updatedVar, valuesWo[ToString(s.Key)]))
			}
		} else {
			diff.Delete = append(diff.Delete, s.toDiffDeleteRequest())
		}
	}

	for _, s := range ss {
		if !old.contains(s) {
			diff.Create = append(diff.Create, s.toDiffCreateRequest(valuesWo[ToString(s.Key)]))
		}
	}

	return diff
}

type Secret struct {
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
	// ValueWo is always null here: the framework nulls it in the plan and the state, and Create and
	// Update read it from the config (see write_only.go).
	ValueWo        types.String `tfsdk:"value_wo"`
	ValueWoVersion types.String `tfsdk:"value_wo_version"`
	Description    types.String `tfsdk:"description"`
}

// changedIn reports whether the planned entry differs from the state entry s in what the API
// stores: the value, the version that stands for a write-only value, or the description. An entry
// that only moved in the list is not changed.
func (s Secret) changedIn(planned Secret) bool {
	return planned.Value != s.Value || planned.ValueWoVersion != s.ValueWoVersion || planned.Description != s.Description
}

func (s Secret) toTerraformObject() types.Object {
	attributes := map[string]attr.Value{
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

func (s Secret) toWriteOnlyTerraformObject() types.Object {
	return types.ObjectValueMust(secretWithWriteOnlyAttrTypes, map[string]attr.Value{
		"id":               s.Id,
		"key":              s.Key,
		"value":            s.Value,
		"value_wo":         s.ValueWo,
		"value_wo_version": s.ValueWoVersion,
		"description":      s.Description,
	})
}

// toCreateRequest builds the request that creates s. valueWo is its value_wo from the config.
func (s Secret) toCreateRequest(valueWo types.String) client.SecretCreateRequest {
	return client.SecretCreateRequest{
		SecretRequest: qovery.SecretRequest{
			Key:         ToString(s.Key),
			Value:       writeOnlyOrPlain(s.Value, valueWo),
			Description: *qovery.NewNullableString(ToStringPointer(s.Description)),
		},
	}
}

// toDiffCreateRequest builds the request that creates s. valueWo is its value_wo from the config.
func (s Secret) toDiffCreateRequest(valueWo types.String) secret.DiffCreateRequest {
	return secret.DiffCreateRequest{
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       secretValueToSend(s.Value, valueWo),
			Description: ToString(s.Description),
		},
	}
}

// toUpdateRequest builds the request that updates the state entry s to new. valueWo is the
// value_wo of new from the config.
func (s Secret) toUpdateRequest(new Secret, valueWo types.String) client.SecretUpdateRequest {
	return client.SecretUpdateRequest{
		Id: ToString(s.Id),
		SecretEditRequest: qovery.SecretEditRequest{
			Key:         ToString(s.Key),
			Value:       writeOnlyOrPlain(new.Value, valueWo),
			Description: *qovery.NewNullableString(ToStringPointer(new.Description)),
		},
	}
}

// toDiffUpdateRequest builds the request that updates the state entry s to new. valueWo is the
// value_wo of new from the config.
func (s Secret) toDiffUpdateRequest(new Secret, valueWo types.String) secret.DiffUpdateRequest {
	return secret.DiffUpdateRequest{
		SecretID: ToString(s.Id),
		UpsertRequest: secret.UpsertRequest{
			Key:         ToString(s.Key),
			Value:       secretValueToSend(new.Value, valueWo),
			Description: ToString(new.Description),
		},
	}
}

// secretValueToSend returns the value a domain request sends: the write-only value when config
// sets it, else the plain one, empty when neither is set.
func secretValueToSend(plain, valueWo types.String) string {
	if value := writeOnlyOrPlain(plain, valueWo); value != nil {
		return *value
	}
	return ""
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

// fromSecret converts an API secret. state is the prior entry for this key, nil when there is none.
func fromSecret(v *qovery.Secret, state *Secret) Secret {
	var prior Secret
	if state != nil {
		prior = *state
	}
	return Secret{
		Id:  FromString(v.Id),
		Key: FromString(v.Key),
		// The API never returns secret values: keep the prior one and its write-only version
		// (null on import).
		Value:          prior.Value,
		ValueWo:        types.StringNull(),
		ValueWoVersion: prior.ValueWoVersion,
		// The API description always wins, so a description set in the Console shows in the plan.
		Description: planAwareOptionalString(v.GetDescription(), prior.Description),
	}
}

// fromSecretList converts the API secrets of the given scope and type, in the order of
// initialState, the plan on apply and the state on refresh.
func fromSecretList(initialState secretCollection, secrets []*qovery.Secret, scope qovery.APIVariableScopeEnum, secretType string) SecretList {
	state := ToSecretList(initialState)

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
	return SecretList(list).inPriorOrder(state)
}

// convertDomainSecretsToSecretList converts the domain secrets of the given scope and type, in the
// order of initialState, the plan on apply and the state on refresh.
func convertDomainSecretsToSecretList(initialState secretCollection, secrets secret.Secrets, scope variable.Scope, variableType string) SecretList {
	state := ToSecretList(initialState)

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
	return SecretList(list).inPriorOrder(state)
}

// convertDomainSecretToSecret converts a domain secret. state is the prior entry for this key, nil
// when there is none.
func convertDomainSecretToSecret(s secret.Secret, state *Secret) Secret {
	var prior Secret
	if state != nil {
		prior = *state
	}
	return Secret{
		Id:  FromString(s.ID.String()),
		Key: FromString(s.Key),
		// The API never returns secret values: keep the prior one and its write-only version
		// (null on import).
		Value:          prior.Value,
		ValueWo:        types.StringNull(),
		ValueWoVersion: prior.ValueWoVersion,
		// The API description always wins, so a description set in the Console shows in the plan.
		Description: planAwareOptionalString(s.Description, prior.Description),
	}
}

// toSecret reads an element of a secret attribute. The elements of secret_aliases,
// secret_overrides and the data source secrets have no value_wo and value_wo_version, which read
// as null.
func toSecret(v types.Object) Secret {
	attributes := v.Attributes()
	stringAttribute := func(name string) types.String {
		if value, ok := attributes[name].(types.String); ok {
			return value
		}
		return types.StringNull()
	}
	return Secret{
		Id:             stringAttribute("id"),
		Key:            stringAttribute("key"),
		Value:          stringAttribute("value"),
		ValueWo:        stringAttribute("value_wo"),
		ValueWoVersion: stringAttribute("value_wo_version"),
		Description:    stringAttribute("description"),
	}
}

func ToSecretList(vars secretCollection) SecretList {
	if vars.IsNull() || vars.IsUnknown() {
		return []Secret{}
	}

	secrets := make([]Secret, 0, len(vars.Elements()))
	for _, elem := range vars.Elements() {
		secrets = append(secrets, toSecret(elem.(types.Object)))
	}

	return secrets
}
