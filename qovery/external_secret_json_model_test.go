//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

func TestExternalSecretList_JSONTargetKeyLifecycle(t *testing.T) {
	t.Parallel()
	const id = "00000000-0000-0000-0000-000000000001"
	for _, tc := range []struct {
		name           string
		old, desired   types.String
		create, update bool
	}{
		{"create_json", types.StringNull(), types.StringValue("accounts.0.password"), true, false},
		{"enable", types.StringNull(), types.StringValue("credentials.password"), false, true},
		{"change", types.StringValue("credentials.password"), types.StringValue(`database\.password`), false, true},
		{"remove", types.StringValue("credentials.password"), types.StringNull(), false, true},
		{"unchanged", types.StringValue("credentials.password"), types.StringValue("credentials.password"), false, false},
		{"legacy_whole_secret", types.StringNull(), types.StringNull(), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			desired := esItem(id, "PASSWORD", "", "/prod/db", "manager")
			desired.JSONTargetKey = tc.desired
			old := desired
			old.JSONTargetKey = tc.old
			previous := ExternalSecretList{old}
			if tc.create {
				previous = nil
			}
			diff := (ExternalSecretList{desired}).diffRequest(previous)
			require.Empty(t, diff.Delete)
			switch {
			case tc.create:
				require.Len(t, diff.Create, 1)
				assert.Equal(t, tc.desired.ValueStringPointer(), diff.Create[0].JSONTargetKey)
			case tc.update:
				require.Len(t, diff.Update, 1)
				assert.Equal(t, id, diff.Update[0].VariableID)
				assert.Equal(t, tc.desired.ValueStringPointer(), diff.Update[0].JSONTargetKey)
			default:
				assert.True(t, diff.IsEmpty())
			}
			set := (ExternalSecretList{desired}).toTerraformSet(context.Background())
			assert.Equal(t, tc.desired, toExternalSecretList(set)[0].JSONTargetKey)
		})
	}
}

func TestExternalSecretRead_JSONTargetKeyRecordsAPIDriftAndImport(t *testing.T) {
	t.Parallel()
	const id = "00000000-0000-0000-0000-000000000001"
	for _, tc := range []struct {
		name    string
		api     *string
		planned types.String
	}{
		{"import", new("accounts.0.password"), types.StringNull()},
		{"outside_change", new("credentials.new_password"), types.StringValue("credentials.password")},
		{"outside_removal", nil, types.StringValue("credentials.password")},
		{"whole_secret", nil, types.StringNull()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			domain := makeDomainExternalSecret(id, "PASSWORD", "", "/prod/db", "manager", variable.ScopeApplication, "EXTERNAL_SECRET")
			domain.JSONTargetKey = tc.api
			plan := esItem(id, "PASSWORD", "", "/prod/db", "manager")
			plan.JSONTargetKey = tc.planned
			actual := convertDomainExternalSecretsToExternalSecretList(variable.ExternalSecrets{domain}, (ExternalSecretList{plan}).toTerraformSet(context.Background()), variable.ScopeApplication)
			require.Len(t, actual, 1)
			assert.Equal(t, FromStringPointer(tc.api), actual[0].JSONTargetKey)
		})
	}
}
