//go:build unit || !integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// --- helpers ----------------------------------------------------------------

func makeExternalSecretObj(id, key, description, reference, secretManagerAccessId string) types.Object {
	return types.ObjectValueMust(externalSecretAttrTypes, map[string]attr.Value{
		"id":                       strOrNull(id),
		"key":                      types.StringValue(key),
		"description":              strOrNull(description),
		"reference":                types.StringValue(reference),
		"secret_manager_access_id": types.StringValue(secretManagerAccessId),
	})
}

func makeExternalSecretSet(objs ...types.Object) types.Set {
	elems := make([]attr.Value, len(objs))
	for i, o := range objs {
		elems[i] = o
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: externalSecretAttrTypes}, elems)
}

func makeDomainExternalSecret(id, key, description, reference, secretManagerAccessId string, scope variable.Scope, variableType string) variable.ExternalSecret {
	return variable.ExternalSecret{
		ID:                    uuid.MustParse(id),
		Key:                   key,
		Description:           description,
		Reference:             reference,
		SecretManagerAccessId: secretManagerAccessId,
		Scope:                 scope,
		VariableType:          variableType,
	}
}

func esItem(id, key, description, reference, secretManagerAccessId string) ExternalSecretItem {
	return ExternalSecretItem{
		Id:                    strOrNull(id),
		Key:                   types.StringValue(key),
		Description:           strOrNull(description),
		Reference:             types.StringValue(reference),
		SecretManagerAccessId: types.StringValue(secretManagerAccessId),
	}
}

// --- ExternalSecretList.diffRequest -----------------------------------------

func TestExternalSecretList_DiffRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		desired     ExternalSecretList
		old         ExternalSecretList
		wantCreates int
		wantUpdates int
		wantDeletes int
		checkCreate func(t *testing.T, creates []variable.ExternalSecretDiffCreateRequest)
		checkUpdate func(t *testing.T, updates []variable.ExternalSecretDiffUpdateRequest)
		checkDelete func(t *testing.T, deletes []variable.ExternalSecretDiffDeleteRequest)
	}{
		{
			name:        "both_empty",
			desired:     ExternalSecretList{},
			old:         ExternalSecretList{},
			wantCreates: 0,
			wantUpdates: 0,
			wantDeletes: 0,
		},
		{
			name:        "new_item_creates",
			desired:     ExternalSecretList{esItem("", "NEW_KEY", "desc", "ref-value", "sm-access-id")},
			old:         ExternalSecretList{},
			wantCreates: 1,
			wantUpdates: 0,
			wantDeletes: 0,
			checkCreate: func(t *testing.T, creates []variable.ExternalSecretDiffCreateRequest) {
				c := creates[0]
				assert.Equal(t, "NEW_KEY", c.Key)
				assert.Equal(t, "desc", c.Description)
				assert.Equal(t, "ref-value", c.Reference)
				assert.Equal(t, "sm-access-id", c.SecretManagerAccessId)
			},
		},
		{
			name:        "removed_item_deletes",
			desired:     ExternalSecretList{},
			old:         ExternalSecretList{esItem("var-id-1", "OLD_KEY", "", "ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 0,
			wantDeletes: 1,
			checkDelete: func(t *testing.T, deletes []variable.ExternalSecretDiffDeleteRequest) {
				assert.Equal(t, "var-id-1", deletes[0].VariableID)
			},
		},
		{
			name:        "reference_changed_updates",
			desired:     ExternalSecretList{esItem("", "MY_KEY", "", "new-ref", "sm-id")},
			old:         ExternalSecretList{esItem("var-id-1", "MY_KEY", "", "old-ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretDiffUpdateRequest) {
				u := updates[0]
				assert.Equal(t, "var-id-1", u.VariableID)
				assert.Equal(t, "MY_KEY", u.Key)
				assert.Equal(t, "new-ref", u.Reference)
			},
		},
		{
			name:        "secret_manager_access_id_changed_updates",
			desired:     ExternalSecretList{esItem("", "MY_KEY", "", "ref", "new-sm-id")},
			old:         ExternalSecretList{esItem("var-id-2", "MY_KEY", "", "ref", "old-sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretDiffUpdateRequest) {
				u := updates[0]
				assert.Equal(t, "var-id-2", u.VariableID)
				assert.Equal(t, "new-sm-id", u.SecretManagerAccessId)
			},
		},
		{
			name:        "description_changed_updates",
			desired:     ExternalSecretList{esItem("", "MY_KEY", "new-desc", "ref", "sm-id")},
			old:         ExternalSecretList{esItem("var-id-3", "MY_KEY", "old-desc", "ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretDiffUpdateRequest) {
				assert.Equal(t, "new-desc", updates[0].Description)
			},
		},
		{
			name:        "unchanged_item_no_operations",
			desired:     ExternalSecretList{esItem("var-id-1", "MY_KEY", "desc", "ref", "sm-id")},
			old:         ExternalSecretList{esItem("var-id-1", "MY_KEY", "desc", "ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 0,
			wantDeletes: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := tc.desired.diffRequest(tc.old)

			assert.Len(t, diff.Create, tc.wantCreates, "creates count")
			assert.Len(t, diff.Update, tc.wantUpdates, "updates count")
			assert.Len(t, diff.Delete, tc.wantDeletes, "deletes count")

			if tc.checkCreate != nil {
				tc.checkCreate(t, diff.Create)
			}
			if tc.checkUpdate != nil {
				tc.checkUpdate(t, diff.Update)
			}
			if tc.checkDelete != nil {
				tc.checkDelete(t, diff.Delete)
			}
		})
	}
}

// --- convertDomainExternalSecretsToExternalSecretList -----------------------

func TestConvertDomainExternalSecretsToExternalSecretList(t *testing.T) {
	t.Parallel()

	const fixedID = "00000000-0000-0000-0000-000000000001"
	ctx := context.Background()
	nullPlan := types.SetNull(types.ObjectType{AttrTypes: externalSecretAttrTypes})
	emptyPlan := makeExternalSecretSet()

	tests := []struct {
		name    string
		domain  variable.ExternalSecrets
		plan    types.Set
		scope   variable.Scope
		check   func(t *testing.T, list ExternalSecretList)
		wantNil bool
	}{
		{
			name:    "empty_domain_null_plan_returns_nil",
			domain:  variable.ExternalSecrets{},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name:   "empty_domain_non_null_plan_returns_empty_list",
			domain: variable.ExternalSecrets{},
			plan:   emptyPlan,
			scope:  variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.NotNil(t, list)
				assert.Len(t, list, 0)
			},
		},
		{
			name: "items_with_wrong_scope_filtered",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "", "ref", "sm-id", variable.ScopeContainer, "EXTERNAL_SECRET"),
			},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name: "items_with_wrong_variable_type_filtered",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "", "ref", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name: "matching_item_included",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "", "my-ref", "sm-access-1", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.Len(t, list, 1)
				item := list[0]
				assert.Equal(t, fixedID, item.Id.ValueString())
				assert.Equal(t, "MY_KEY", item.Key.ValueString())
				assert.Equal(t, "my-ref", item.Reference.ValueString())
				assert.Equal(t, "sm-access-1", item.SecretManagerAccessId.ValueString())
			},
		},
		{
			name: "description_non_null_in_plan_returns_api_value",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "api-desc", "ref", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan: makeExternalSecretSet(
				makeExternalSecretObj(fixedID, "MY_KEY", "plan-desc", "ref", "sm-id"),
			),
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.Len(t, list, 1)
				desc := list[0].Description
				assert.False(t, desc.IsNull())
				assert.Equal(t, "api-desc", desc.ValueString())
			},
		},
		{
			name: "description_null_in_plan_empty_in_api_returns_null",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "", "ref", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.Len(t, list, 1)
				assert.True(t, list[0].Description.IsNull())
			},
		},
		{
			name: "description_null_in_plan_non_empty_in_api_returns_api_value",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret(fixedID, "MY_KEY", "api-desc", "ref", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.Len(t, list, 1)
				desc := list[0].Description
				assert.False(t, desc.IsNull())
				assert.Equal(t, "api-desc", desc.ValueString())
			},
		},
		{
			name: "only_matching_scope_and_type_included",
			domain: variable.ExternalSecrets{
				makeDomainExternalSecret("00000000-0000-0000-0000-000000000001", "KEEP", "", "ref1", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
				makeDomainExternalSecret("00000000-0000-0000-0000-000000000002", "WRONG_SCOPE", "", "ref2", "sm-id", variable.ScopeContainer, "EXTERNAL_SECRET"),
				makeDomainExternalSecret("00000000-0000-0000-0000-000000000003", "WRONG_TYPE", "", "ref3", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretList) {
				require.Len(t, list, 1)
				assert.Equal(t, "KEEP", list[0].Key.ValueString())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := convertDomainExternalSecretsToExternalSecretList(tc.domain, tc.plan, tc.scope)

			if tc.wantNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			if tc.check != nil {
				tc.check(t, result)
			}

			// Ensure the result can round-trip to a Terraform set without panicking.
			_ = result.toTerraformSet(ctx)
		})
	}
}
