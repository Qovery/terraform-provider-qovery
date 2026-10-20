//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateCustomRoleConfig(t *testing.T, model CustomRole) *resource.ValidateConfigResponse {
	t.Helper()
	ctx := context.Background()
	r := newCustomRoleResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	diags := state.Set(ctx, &model)
	require.False(t, diags.HasError(), diags)

	resp := &resource.ValidateConfigResponse{}
	r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw},
	}, resp)
	return resp
}

func customRoleWithProduction(projectID types.String, production types.String, shell, portForward types.Bool) CustomRole {
	env := func(envType, permission string, shell, portForward types.Bool) attr.Value {
		return types.ObjectValueMust(customRoleEnvPermissionAttrTypes, map[string]attr.Value{
			"environment_type": types.StringValue(envType),
			"permission":       types.StringValue(permission),
			"shell":            shell,
			"port_forward":     portForward,
		})
	}
	null := types.BoolNull()
	return CustomRole{
		Id:                 types.StringUnknown(),
		OrganizationId:     types.StringValue("00000000-0000-0000-0000-000000000001"),
		Name:               types.StringValue("my-role"),
		Description:        types.StringNull(),
		ClusterPermissions: types.SetNull(customRoleClusterObjectType),
		ProjectPermissions: types.SetValueMust(customRoleProjectObjectType, []attr.Value{
			types.ObjectValueMust(customRoleProjectPermissionAttrTypes, map[string]attr.Value{
				"project_id": projectID,
				"is_admin":   types.BoolValue(false),
				"permissions": types.SetValueMust(types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes}, []attr.Value{
					env("DEVELOPMENT", "MANAGER", null, null),
					env("PREVIEW", "MANAGER", null, null),
					env("STAGING", "DEPLOYER", null, null),
					types.ObjectValueMust(customRoleEnvPermissionAttrTypes, map[string]attr.Value{
						"environment_type": types.StringValue("PRODUCTION"),
						"permission":       production,
						"shell":            shell,
						"port_forward":     portForward,
					}),
				}),
			}),
		}),
	}
}

func TestCustomRoleValidateConfig_RemoteAccessWithNoAccess(t *testing.T) {
	noAccess := types.StringValue("NO_ACCESS")
	on, off, null := types.BoolValue(true), types.BoolValue(false), types.BoolNull()

	tests := []struct {
		name      string
		project   types.String
		level     types.String
		shell     types.Bool
		wantError bool
		portFwd   types.Bool
	}{
		{name: "shell_with_no_access", project: types.StringValue("p"), level: noAccess, shell: on, portFwd: null, wantError: true},
		{name: "port_forward_with_no_access", project: types.StringValue("p"), level: noAccess, shell: null, portFwd: on, wantError: true},
		// the ids are not known yet at plan time when they come from resources created in the same run
		{name: "shell_with_no_access_unknown_project", project: types.StringUnknown(), level: noAccess, shell: on, portFwd: null, wantError: true},
		{name: "port_forward_with_no_access_unknown_project", project: types.StringUnknown(), level: noAccess, shell: null, portFwd: on, wantError: true},
		{name: "false_flags_with_no_access", project: types.StringUnknown(), level: noAccess, shell: off, portFwd: off, wantError: false},
		{name: "unset_flags_with_no_access", project: types.StringUnknown(), level: noAccess, shell: null, portFwd: null, wantError: false},
		{name: "flags_with_viewer", project: types.StringUnknown(), level: types.StringValue("VIEWER"), shell: on, portFwd: on, wantError: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validateCustomRoleConfig(t, customRoleWithProduction(tt.project, tt.level, tt.shell, tt.portFwd))
			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
