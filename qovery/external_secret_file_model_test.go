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

func makeExternalSecretFileObj(id, key, description, mountPath, reference, secretManagerAccessId string) types.Object {
	return types.ObjectValueMust(externalSecretFileAttrTypes, map[string]attr.Value{
		"id":                       strOrNull(id),
		"key":                      types.StringValue(key),
		"description":              strOrNull(description),
		"mount_path":               types.StringValue(mountPath),
		"reference":                types.StringValue(reference),
		"secret_manager_access_id": types.StringValue(secretManagerAccessId),
	})
}

func makeExternalSecretFileSet(objs ...types.Object) types.Set {
	elems := make([]attr.Value, len(objs))
	for i, o := range objs {
		elems[i] = o
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: externalSecretFileAttrTypes}, elems)
}

func makeDomainExternalSecretFile(id, key, description, mountPath, reference, secretManagerAccessId string, scope variable.Scope, variableType string) variable.ExternalSecretFile {
	return variable.ExternalSecretFile{
		ID:                    uuid.MustParse(id),
		Key:                   key,
		Description:           description,
		MountPath:             mountPath,
		Reference:             reference,
		SecretManagerAccessId: secretManagerAccessId,
		Scope:                 scope,
		VariableType:          variableType,
	}
}

func esfItem(id, description, mountPath, reference, secretManagerAccessId string) ExternalSecretFileItem {
	return ExternalSecretFileItem{
		Id:                    strOrNull(id),
		Key:                   types.StringValue("CONFIG_FILE"),
		Description:           strOrNull(description),
		MountPath:             types.StringValue(mountPath),
		Reference:             types.StringValue(reference),
		SecretManagerAccessId: types.StringValue(secretManagerAccessId),
	}
}

// --- ExternalSecretFileList.diffRequest -------------------------------------

func TestExternalSecretFileList_DiffRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		desired     ExternalSecretFileList
		old         ExternalSecretFileList
		wantCreates int
		wantUpdates int
		wantDeletes int
		checkCreate func(t *testing.T, creates []variable.ExternalSecretFileDiffCreateRequest)
		checkUpdate func(t *testing.T, updates []variable.ExternalSecretFileDiffUpdateRequest)
		checkDelete func(t *testing.T, deletes []variable.ExternalSecretFileDiffDeleteRequest)
	}{
		{
			name:        "both_empty",
			desired:     ExternalSecretFileList{},
			old:         ExternalSecretFileList{},
			wantCreates: 0,
			wantUpdates: 0,
			wantDeletes: 0,
		},
		{
			name:        "new_item_creates",
			desired:     ExternalSecretFileList{esfItem("", "desc", "/etc/config", "ref-value", "sm-access-id")},
			old:         ExternalSecretFileList{},
			wantCreates: 1,
			wantUpdates: 0,
			wantDeletes: 0,
			checkCreate: func(t *testing.T, creates []variable.ExternalSecretFileDiffCreateRequest) {
				c := creates[0]
				assert.Equal(t, "CONFIG_FILE", c.Key)
				assert.Equal(t, "desc", c.Description)
				assert.Equal(t, "/etc/config", c.MountPath)
				assert.Equal(t, "ref-value", c.Reference)
				assert.Equal(t, "sm-access-id", c.SecretManagerAccessId)
			},
		},
		{
			name:        "removed_item_deletes",
			desired:     ExternalSecretFileList{},
			old:         ExternalSecretFileList{esfItem("var-id-1", "", "/etc/config", "ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 0,
			wantDeletes: 1,
			checkDelete: func(t *testing.T, deletes []variable.ExternalSecretFileDiffDeleteRequest) {
				assert.Equal(t, "var-id-1", deletes[0].VariableID)
			},
		},
		{
			name:        "reference_changed_updates",
			desired:     ExternalSecretFileList{esfItem("", "", "/etc/config", "new-ref", "sm-id")},
			old:         ExternalSecretFileList{esfItem("var-id-1", "", "/etc/config", "old-ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretFileDiffUpdateRequest) {
				u := updates[0]
				assert.Equal(t, "var-id-1", u.VariableID)
				assert.Equal(t, "CONFIG_FILE", u.Key)
				assert.Equal(t, "new-ref", u.Reference)
				assert.Equal(t, "/etc/config", u.MountPath)
			},
		},
		{
			name:        "secret_manager_access_id_changed_updates",
			desired:     ExternalSecretFileList{esfItem("", "", "/etc/config", "ref", "new-sm-id")},
			old:         ExternalSecretFileList{esfItem("var-id-2", "", "/etc/config", "ref", "old-sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretFileDiffUpdateRequest) {
				assert.Equal(t, "var-id-2", updates[0].VariableID)
				assert.Equal(t, "new-sm-id", updates[0].SecretManagerAccessId)
			},
		},
		{
			name:        "description_changed_updates",
			desired:     ExternalSecretFileList{esfItem("", "new-desc", "/etc/config", "ref", "sm-id")},
			old:         ExternalSecretFileList{esfItem("var-id-3", "old-desc", "/etc/config", "ref", "sm-id")},
			wantCreates: 0,
			wantUpdates: 1,
			wantDeletes: 0,
			checkUpdate: func(t *testing.T, updates []variable.ExternalSecretFileDiffUpdateRequest) {
				assert.Equal(t, "new-desc", updates[0].Description)
			},
		},
		{
			// mount_path cannot be updated via the API — the provider deletes and recreates.
			name:        "mount_path_changed_deletes_and_recreates",
			desired:     ExternalSecretFileList{esfItem("", "", "/new/path", "ref", "sm-id")},
			old:         ExternalSecretFileList{esfItem("var-id-4", "", "/old/path", "ref", "sm-id")},
			wantCreates: 1,
			wantUpdates: 0,
			wantDeletes: 1,
			checkDelete: func(t *testing.T, deletes []variable.ExternalSecretFileDiffDeleteRequest) {
				assert.Equal(t, "var-id-4", deletes[0].VariableID)
			},
			checkCreate: func(t *testing.T, creates []variable.ExternalSecretFileDiffCreateRequest) {
				assert.Equal(t, "CONFIG_FILE", creates[0].Key)
				assert.Equal(t, "/new/path", creates[0].MountPath)
			},
		},
		{
			name:        "unchanged_item_no_operations",
			desired:     ExternalSecretFileList{esfItem("var-id-1", "desc", "/etc/config", "ref", "sm-id")},
			old:         ExternalSecretFileList{esfItem("var-id-1", "desc", "/etc/config", "ref", "sm-id")},
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

// --- convertDomainExternalSecretFilesToExternalSecretFileList ---------------

func TestConvertDomainExternalSecretFilesToExternalSecretFileList(t *testing.T) {
	t.Parallel()

	const fixedID = "00000000-0000-0000-0000-000000000001"
	ctx := context.Background()
	nullPlan := types.SetNull(types.ObjectType{AttrTypes: externalSecretFileAttrTypes})
	emptyPlan := makeExternalSecretFileSet()

	tests := []struct {
		name    string
		domain  variable.ExternalSecretFiles
		plan    types.Set
		scope   variable.Scope
		check   func(t *testing.T, list ExternalSecretFileList)
		wantNil bool
	}{
		{
			name:    "empty_domain_null_plan_returns_nil",
			domain:  variable.ExternalSecretFiles{},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name:   "empty_domain_non_null_plan_returns_empty_list",
			domain: variable.ExternalSecretFiles{},
			plan:   emptyPlan,
			scope:  variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.NotNil(t, list)
				assert.Len(t, list, 0)
			},
		},
		{
			name: "items_with_wrong_scope_filtered",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "", "/etc/config", "ref", "sm-id", variable.ScopeContainer, "FILE_EXTERNAL_SECRET"),
			},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name: "items_with_wrong_variable_type_filtered",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "", "/etc/config", "ref", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan:    nullPlan,
			scope:   variable.ScopeApplication,
			wantNil: true,
		},
		{
			name: "matching_item_included",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "", "/etc/config", "my-ref", "sm-access-1", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.Len(t, list, 1)
				item := list[0]
				assert.Equal(t, fixedID, item.Id.ValueString())
				assert.Equal(t, "CONFIG_FILE", item.Key.ValueString())
				assert.Equal(t, "/etc/config", item.MountPath.ValueString())
				assert.Equal(t, "my-ref", item.Reference.ValueString())
				assert.Equal(t, "sm-access-1", item.SecretManagerAccessId.ValueString())
			},
		},
		{
			name: "description_non_null_in_plan_returns_api_value",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "api-desc", "/etc/config", "ref", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan: makeExternalSecretFileSet(
				makeExternalSecretFileObj(fixedID, "CONFIG_FILE", "plan-desc", "/etc/config", "ref", "sm-id"),
			),
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.Len(t, list, 1)
				desc := list[0].Description
				assert.False(t, desc.IsNull())
				assert.Equal(t, "api-desc", desc.ValueString())
			},
		},
		{
			name: "description_null_in_plan_empty_in_api_returns_null",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "", "/etc/config", "ref", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.Len(t, list, 1)
				assert.True(t, list[0].Description.IsNull())
			},
		},
		{
			name: "description_null_in_plan_non_empty_in_api_returns_api_value",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile(fixedID, "CONFIG_FILE", "api-desc", "/etc/config", "ref", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.Len(t, list, 1)
				desc := list[0].Description
				assert.False(t, desc.IsNull())
				assert.Equal(t, "api-desc", desc.ValueString())
			},
		},
		{
			name: "only_matching_scope_and_type_included",
			domain: variable.ExternalSecretFiles{
				makeDomainExternalSecretFile("00000000-0000-0000-0000-000000000001", "KEEP", "", "/keep", "ref1", "sm-id", variable.ScopeApplication, "FILE_EXTERNAL_SECRET"),
				makeDomainExternalSecretFile("00000000-0000-0000-0000-000000000002", "WRONG_SCOPE", "", "/x", "ref2", "sm-id", variable.ScopeContainer, "FILE_EXTERNAL_SECRET"),
				makeDomainExternalSecretFile("00000000-0000-0000-0000-000000000003", "WRONG_TYPE", "", "/y", "ref3", "sm-id", variable.ScopeApplication, "EXTERNAL_SECRET"),
			},
			plan:  nullPlan,
			scope: variable.ScopeApplication,
			check: func(t *testing.T, list ExternalSecretFileList) {
				require.Len(t, list, 1)
				assert.Equal(t, "KEEP", list[0].Key.ValueString())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := convertDomainExternalSecretFilesToExternalSecretFileList(tc.domain, tc.plan, tc.scope)

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
