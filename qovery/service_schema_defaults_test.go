//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schemaDefault returns the planned value of the schema Default of the attribute at p, or nil
// when the attribute has none.
func schemaDefault(t *testing.T, r resource.Resource, p path.Path) attr.Value {
	t.Helper()
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	attribute, diags := schemaResp.Schema.AttributeAtPath(ctx, p)
	require.False(t, diags.HasError(), "%v", diags)

	switch a := attribute.(type) {
	case schema.StringAttribute:
		if a.Default == nil {
			return nil
		}
		resp := &defaults.StringResponse{}
		a.Default.DefaultString(ctx, defaults.StringRequest{Path: p}, resp)
		return resp.PlanValue
	case schema.BoolAttribute:
		if a.Default == nil {
			return nil
		}
		resp := &defaults.BoolResponse{}
		a.Default.DefaultBool(ctx, defaults.BoolRequest{Path: p}, resp)
		return resp.PlanValue
	case schema.Int64Attribute:
		if a.Default == nil {
			return nil
		}
		resp := &defaults.Int64Response{}
		a.Default.DefaultInt64(ctx, defaults.Int64Request{Path: p}, resp)
		return resp.PlanValue
	}
	t.Fatalf("unhandled attribute type %T at %s", attribute, p)
	return nil
}

// TestServiceSchemaDefaults pins the q-core defaults the service schemas plan when an attribute
// is omitted, so that removing it from the configuration plans the reset.
func TestServiceSchemaDefaults(t *testing.T) {
	t.Parallel()

	application := applicationResource{}
	container := containerResource{}
	job := jobResource{}
	helm := helmResource{}
	database := databaseResource{}

	testCases := []struct {
		TestName string
		Resource resource.Resource
		Path     path.Path
		Expect   attr.Value
	}{
		{TestName: "application_icon_uri", Resource: application, Path: path.Root("icon_uri"), Expect: types.StringValue("app://qovery-console/application")},
		{TestName: "application_auto_deploy", Resource: application, Path: path.Root("auto_deploy"), Expect: types.BoolValue(true)},
		{TestName: "application_auto_preview", Resource: application, Path: path.Root("auto_preview"), Expect: types.BoolValue(false)},
		{TestName: "application_ephemeral_storage", Resource: application, Path: path.Root("ephemeral_storage"), Expect: types.Int64Value(0)},
		{TestName: "container_icon_uri", Resource: container, Path: path.Root("icon_uri"), Expect: types.StringValue("app://qovery-console/container")},
		{TestName: "container_auto_deploy", Resource: container, Path: path.Root("auto_deploy"), Expect: types.BoolValue(true)},
		{TestName: "container_auto_preview", Resource: container, Path: path.Root("auto_preview"), Expect: types.BoolValue(false)},
		{TestName: "container_ephemeral_storage", Resource: container, Path: path.Root("ephemeral_storage"), Expect: types.Int64Value(0)},
		{TestName: "container_ports_protocol", Resource: container, Path: path.Root("ports").AtListIndex(0).AtName("protocol"), Expect: types.StringValue("HTTP")},
		{TestName: "job_auto_deploy", Resource: job, Path: path.Root("auto_deploy"), Expect: types.BoolValue(true)},
		{TestName: "job_auto_preview", Resource: job, Path: path.Root("auto_preview"), Expect: types.BoolValue(false)},
		{TestName: "job_ephemeral_storage", Resource: job, Path: path.Root("ephemeral_storage"), Expect: types.Int64Value(0)},
		{TestName: "job_root_path", Resource: job, Path: path.Root("source").AtName("docker").AtName("git_repository").AtName("root_path"), Expect: types.StringValue("/")},
		{TestName: "helm_icon_uri", Resource: helm, Path: path.Root("icon_uri"), Expect: types.StringValue("app://qovery-console/helm")},
		{TestName: "helm_auto_deploy_is_the_0x_wire_value", Resource: helm, Path: path.Root("auto_deploy"), Expect: types.BoolValue(false)},
		{TestName: "helm_auto_preview", Resource: helm, Path: path.Root("auto_preview"), Expect: types.BoolValue(false)},
		{TestName: "helm_ports_protocol", Resource: helm, Path: path.Root("ports").AtMapKey("web").AtName("protocol"), Expect: types.StringValue("HTTP")},
		{TestName: "database_icon_uri", Resource: database, Path: path.Root("icon_uri"), Expect: types.StringValue("app://qovery-console/database")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := schemaDefault(t, tc.Resource, tc.Path)
			require.NotNil(t, got, "no schema Default at %s", tc.Path)
			assert.True(t, tc.Expect.Equal(got), "got %s, want %s", got, tc.Expect)
		})
	}
}

// TestServiceSchemaExceptionsKeepNoDefault pins the attributes that stay Optional + Computed
// without a Default: the documented exceptions and the values planned by a plan modifier.
func TestServiceSchemaExceptionsKeepNoDefault(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Resource resource.Resource
		Path     path.Path
	}{
		{TestName: "job_icon_uri_depends_on_the_schedule", Resource: jobResource{}, Path: path.Root("icon_uri")},
		{TestName: "application_ports_name_derives_from_the_port", Resource: applicationResource{}, Path: path.Root("ports").AtListIndex(0).AtName("name")},
		{TestName: "application_deployment_stage_id", Resource: applicationResource{}, Path: path.Root("deployment_stage_id")},
		{TestName: "application_git_repository_branch", Resource: applicationResource{}, Path: path.Root("git_repository").AtName("branch")},
		{TestName: "helm_blueprint_id", Resource: helmResource{}, Path: path.Root("blueprint_id")},
		{TestName: "terraform_service_blueprint_id", Resource: terraformServiceResource{}, Path: path.Root("blueprint_id")},
		{TestName: "database_instance_type", Resource: databaseResource{}, Path: path.Root("instance_type")},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Nil(t, schemaDefault(t, tc.Resource, tc.Path))
		})
	}
}

func TestJobRootPath_RejectsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	jobResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	rootPath := path.Root("source").AtName("docker").AtName("git_repository").AtName("root_path")
	attribute, diags := schemaResp.Schema.AttributeAtPath(ctx, rootPath)
	require.False(t, diags.HasError(), "%v", diags)
	validators := attribute.(schema.StringAttribute).Validators
	require.NotEmpty(t, validators)

	validate := func(value types.String) bool {
		hasError := false
		for _, v := range validators {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{Path: rootPath, ConfigValue: value}, resp)
			hasError = hasError || resp.Diagnostics.HasError()
		}
		return hasError
	}

	assert.True(t, validate(types.StringValue("")), "\"\" reads back as \"/\", so the configuration must not hold it")
	assert.False(t, validate(types.StringValue("/")))
	assert.False(t, validate(types.StringValue("/jobs/backup")))
	assert.False(t, validate(types.StringNull()))
}

func TestUpgradeArgumentsFrom0x(t *testing.T) {
	t.Parallel()

	arguments := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("--flag")})

	assert.True(t, upgradeArgumentsFrom0x(types.ListValueMust(types.StringType, []attr.Value{})).IsNull(), "the [] 0.x stored becomes null")
	assert.True(t, upgradeArgumentsFrom0x(types.ListNull(types.StringType)).IsNull())
	assert.True(t, arguments.Equal(upgradeArgumentsFrom0x(arguments)), "declared arguments are kept")
}

// TestServiceResource_UpgradeStateV0ToV1 runs the V0 upgrader of application and container on a
// 0.x state.
func TestServiceResource_UpgradeStateV0ToV1(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyArguments := types.ListValueMust(types.StringType, []attr.Value{})
	declaredArguments := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("--flag")})

	resources := map[string]interface {
		resource.Resource
		resource.ResourceWithUpgradeState
	}{
		"application": applicationResource{},
		"container":   containerResource{},
	}

	testCases := []struct {
		TestName       string
		PriorArguments types.List
		Expect         types.List
	}{
		{TestName: "empty_arguments_written_by_0x_become_null", PriorArguments: emptyArguments, Expect: types.ListNull(types.StringType)},
		{TestName: "declared_arguments_are_kept", PriorArguments: declaredArguments, Expect: declaredArguments},
		{TestName: "null_arguments_stay_null", PriorArguments: types.ListNull(types.StringType), Expect: types.ListNull(types.StringType)},
	}

	for name, r := range resources {
		for _, tc := range testCases {
			t.Run(name+"/"+tc.TestName, func(t *testing.T) {
				t.Parallel()

				var schemaResp resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
				assert.Equal(t, int64(1), schemaResp.Schema.Version)

				upgrader, ok := r.UpgradeState(ctx)[0]
				require.True(t, ok)
				require.NotNil(t, upgrader.PriorSchema)

				priorState := tfsdk.State{
					Schema: *upgrader.PriorSchema,
					Raw:    tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), nil),
				}
				require.False(t, priorState.SetAttribute(ctx, path.Root("id"), "service-123").HasError())
				require.False(t, priorState.SetAttribute(ctx, path.Root("arguments"), tc.PriorArguments).HasError())

				resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
				upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

				var id types.String
				require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
				assert.Equal(t, "service-123", id.ValueString())

				var arguments types.List
				require.False(t, resp.State.GetAttribute(ctx, path.Root("arguments"), &arguments).HasError())
				assert.True(t, tc.Expect.Equal(arguments), "got %s, want %s", arguments, tc.Expect)
			})
		}
	}
}

func TestValidateDatabaseInstanceType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Mode          types.String
		InstanceType  types.String
		ExpectError   bool
		ExpectWarning bool
	}{
		{TestName: "managed_with_instance_type", Mode: types.StringValue("MANAGED"), InstanceType: types.StringValue("db.t3.micro")},
		{TestName: "managed_without_instance_type_is_an_error", Mode: types.StringValue("MANAGED"), InstanceType: types.StringNull(), ExpectError: true},
		{TestName: "managed_with_unknown_instance_type", Mode: types.StringValue("MANAGED"), InstanceType: types.StringUnknown()},
		{TestName: "container_without_instance_type", Mode: types.StringValue("CONTAINER"), InstanceType: types.StringNull()},
		{TestName: "container_with_instance_type_is_a_warning", Mode: types.StringValue("CONTAINER"), InstanceType: types.StringValue("db.t3.micro"), ExpectWarning: true},
		{TestName: "container_with_unknown_instance_type_is_a_warning", Mode: types.StringValue("CONTAINER"), InstanceType: types.StringUnknown(), ExpectWarning: true},
		{TestName: "unknown_mode_is_not_checked", Mode: types.StringUnknown(), InstanceType: types.StringNull()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			diags := validateDatabaseInstanceType(tc.Mode, tc.InstanceType)
			assert.Equal(t, tc.ExpectError, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.ExpectWarning, diags.WarningsCount() > 0, "%v", diags)
		})
	}
}

// TestDatabaseTypeAndModeRejectChangeAfterCreate: the update request cannot carry type or
// mode, so a change must fail the plan instead of planning an update that never applies.
func TestDatabaseTypeAndModeRejectChangeAfterCreate(t *testing.T) {
	t.Parallel()
	var resp resource.SchemaResponse
	databaseResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	for _, name := range []string{"type", "mode"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			attribute, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
			require.True(t, ok)
			assert.Contains(t, attribute.PlanModifiers, RejectChangeAfterCreate(databaseTypeModeChangeReason))
		})
	}
}
