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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSettingsObject(timeoutMaxSec int64, ephemeralStorageInGib attr.Value) types.Object {
	return types.ObjectValueMust(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(timeoutMaxSec),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": ephemeralStorageInGib,
		"disable_buildkit_cache":   types.BoolValue(true),
		"skip_git_submodules":      types.BoolValue(false),
	})
}

func TestBuildSettingsObjectToQovery_Nil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, buildSettingsObjectToQovery(types.ObjectNull(buildSettingsAttrTypes)))
}

func TestBuildSettingsObjectToQovery_Unknown(t *testing.T) {
	t.Parallel()
	assert.Nil(t, buildSettingsObjectToQovery(types.ObjectUnknown(buildSettingsAttrTypes)))
}

func TestBuildSettingsObjectToQovery_WithValues(t *testing.T) {
	t.Parallel()

	result := buildSettingsObjectToQovery(buildSettingsObject(3600, types.Int64Value(10)))
	require.NotNil(t, result)
	assert.Equal(t, int32(3600), result.GetTimeoutMaxSec())
	assert.Equal(t, int32(4000), result.GetCpuMaxInMilli())
	assert.Equal(t, int32(8), result.GetRamMaxInGib())
	assert.Equal(t, int32(10), result.GetEphemeralStorageInGib())
	assert.True(t, result.GetDisableBuildkitCache())
	assert.False(t, result.GetSkipGitSubmodules())
}

func TestBuildSettingsObjectToQovery_NullEphemeralStorage(t *testing.T) {
	t.Parallel()

	result := buildSettingsObjectToQovery(buildSettingsObject(1800, types.Int64Null()))
	require.NotNil(t, result)
	assert.True(t, result.EphemeralStorageInGib.IsSet())
	assert.Nil(t, result.EphemeralStorageInGib.Get())
}

func TestBuildSettingsRequest(t *testing.T) {
	t.Parallel()

	null := types.ObjectNull(buildSettingsAttrTypes)
	managed := buildSettingsObject(3600, types.Int64Value(10))

	testCases := []struct {
		TestName        string
		Plan            types.Object
		State           types.Object
		ExpectNil       bool
		ExpectTimeout   int32
		ExpectEphemeral *int32
	}{
		{
			TestName:  "never_set_sends_nothing",
			Plan:      null,
			State:     null,
			ExpectNil: true,
		},
		{
			TestName:        "set_sends_plan",
			Plan:            managed,
			State:           null,
			ExpectTimeout:   3600,
			ExpectEphemeral: qovery.PtrInt32(10),
		},
		{
			TestName:      "removed_resets_to_defaults",
			Plan:          null,
			State:         managed,
			ExpectTimeout: buildSettingsDefaultTimeoutMaxSec,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			result := buildSettingsRequest(tc.Plan, tc.State)
			if tc.ExpectNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			assert.Equal(t, tc.ExpectTimeout, result.GetTimeoutMaxSec())
			assert.True(t, result.EphemeralStorageInGib.IsSet())
			assert.Equal(t, tc.ExpectEphemeral, result.EphemeralStorageInGib.Get())
		})
	}
}

func TestBuildSettingsRequest_ResetMatchesSchemaDefaults(t *testing.T) {
	t.Parallel()

	result := buildSettingsRequest(types.ObjectNull(buildSettingsAttrTypes), buildSettingsObject(3600, types.Int64Value(10)))
	require.NotNil(t, result)
	assert.Equal(t, int32(buildSettingsDefaultTimeoutMaxSec), result.GetTimeoutMaxSec())
	assert.Equal(t, int32(buildSettingsDefaultCpuMaxInMilli), result.GetCpuMaxInMilli())
	assert.Equal(t, int32(buildSettingsDefaultRamMaxInGib), result.GetRamMaxInGib())
	assert.Nil(t, result.EphemeralStorageInGib.Get())
	assert.Equal(t, buildSettingsDefaultDisableBuildkitCache, result.GetDisableBuildkitCache())
	assert.Equal(t, buildSettingsDefaultSkipGitSubmodules, result.GetSkipGitSubmodules())
}

func TestBuildSettingsFromQovery(t *testing.T) {
	t.Parallel()

	assert.True(t, buildSettingsFromQovery(nil).IsNull())

	withEphemeral := qovery.NewBuildSettings()
	withEphemeral.SetTimeoutMaxSec(3600)
	withEphemeral.SetEphemeralStorageInGib(10)
	withEphemeral.SetDisableBuildkitCache(true)
	assert.Equal(t, buildSettingsObject(3600, types.Int64Value(10)), buildSettingsFromQovery(withEphemeral))

	withoutEphemeral := qovery.NewBuildSettings()
	withoutEphemeral.SetTimeoutMaxSec(3600)
	withoutEphemeral.SetEphemeralStorageInGibNil()
	withoutEphemeral.SetDisableBuildkitCache(true)
	assert.Equal(t, buildSettingsObject(3600, types.Int64Null()), buildSettingsFromQovery(withoutEphemeral))
}

func TestBuildSettingsToState(t *testing.T) {
	t.Parallel()

	null := types.ObjectNull(buildSettingsAttrTypes)
	prior := buildSettingsObject(3600, types.Int64Null())
	remote := qovery.NewBuildSettings()
	remote.SetTimeoutMaxSec(7200)
	remote.SetEphemeralStorageInGibNil()
	remote.SetDisableBuildkitCache(true)

	testCases := []struct {
		TestName string
		Prior    types.Object
		FromAPI  *qovery.BuildSettings
		Expected types.Object
	}{
		{
			TestName: "unmanaged_stays_null_even_when_api_returns_settings",
			Prior:    null,
			FromAPI:  remote,
			Expected: null,
		},
		{
			TestName: "managed_keeps_prior_without_api_value",
			Prior:    prior,
			FromAPI:  nil,
			Expected: prior,
		},
		{
			TestName: "managed_takes_api_value_to_surface_drift",
			Prior:    prior,
			FromAPI:  remote,
			Expected: buildSettingsObject(7200, types.Int64Null()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, buildSettingsToState(tc.Prior, tc.FromAPI))
		})
	}
}

func TestValidateBuildSettingsConflict(t *testing.T) {
	t.Parallel()

	managed := buildSettingsObject(1800, types.Int64Null())

	testCases := []struct {
		TestName         string
		BuildSettings    types.Object
		AdvancedSettings types.String
		ExpectError      bool
	}{
		{
			TestName:         "no_build_settings",
			BuildSettings:    types.ObjectNull(buildSettingsAttrTypes),
			AdvancedSettings: types.StringValue(`{"build.timeout_max_sec": 3600}`),
		},
		{
			TestName:         "no_advanced_settings",
			BuildSettings:    managed,
			AdvancedSettings: types.StringNull(),
		},
		{
			TestName:         "unknown_advanced_settings",
			BuildSettings:    managed,
			AdvancedSettings: types.StringUnknown(),
		},
		{
			TestName:         "empty_json",
			BuildSettings:    managed,
			AdvancedSettings: types.StringValue(`{}`),
		},
		{
			TestName:         "no_conflict",
			BuildSettings:    managed,
			AdvancedSettings: types.StringValue(`{"some.other.key": "value"}`),
		},
		{
			TestName:         "conflict",
			BuildSettings:    managed,
			AdvancedSettings: types.StringValue(`{"build.timeout_max_sec": 3600}`),
			ExpectError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			diags := validateBuildSettingsConflict(tc.BuildSettings, tc.AdvancedSettings)
			assert.Equal(t, tc.ExpectError, diags.HasError())
			if tc.ExpectError {
				assert.Contains(t, diags.Errors()[0].Detail(), "build.timeout_max_sec")
			}
		})
	}
}

func TestStripBuildSettingsKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Input         string
		Expected      string
		ExpectChanged bool
	}{
		{
			TestName: "no_build_keys",
			Input:    `{"network.ingress.proxy_body_size_mb":100}`,
			Expected: `{"network.ingress.proxy_body_size_mb":100}`,
		},
		{
			TestName:      "strips_only_build_keys",
			Input:         `{"build.ephemeral_storage_in_gib":null,"build.timeout_max_sec":3600,"network.ingress.proxy_body_size_mb":100}`,
			Expected:      `{"network.ingress.proxy_body_size_mb":100}`,
			ExpectChanged: true,
		},
		{
			TestName: "invalid_json_untouched",
			Input:    `not json`,
			Expected: `not json`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			result, changed := stripBuildSettingsKeys(tc.Input)
			assert.Equal(t, tc.ExpectChanged, changed)
			assert.Equal(t, tc.Expected, result)
		})
	}
}

// buildSettingsPlanTestSchema holds only the two attributes modifyBuildSettingsPlan reads and writes.
func buildSettingsPlanTestSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"build_settings":         buildSettingsResourceSchemaAttributes(),
			"advanced_settings_json": schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func buildSettingsPlanTestValue(t *testing.T, buildSettings tftypes.Value, advancedSettings tftypes.Value) tftypes.Value {
	t.Helper()
	objectType := buildSettingsPlanTestSchema().Type().TerraformType(context.Background()).(tftypes.Object)
	return tftypes.NewValue(objectType, map[string]tftypes.Value{
		"build_settings":         buildSettings,
		"advanced_settings_json": advancedSettings,
	})
}

func TestModifyBuildSettingsPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := buildSettingsPlanTestSchema()
	buildSettingsType := s.Type().TerraformType(ctx).(tftypes.Object).AttributeTypes["build_settings"]

	nullBlock := tftypes.NewValue(buildSettingsType, nil)
	unknownBlock := tftypes.NewValue(buildSettingsType, tftypes.UnknownValue)
	setBlock := tftypes.NewValue(buildSettingsType, map[string]tftypes.Value{
		"timeout_max_sec":          tftypes.NewValue(tftypes.Number, 3600),
		"cpu_max_in_milli":         tftypes.NewValue(tftypes.Number, 4000),
		"ram_max_in_gib":           tftypes.NewValue(tftypes.Number, 8),
		"ephemeral_storage_in_gib": tftypes.NewValue(tftypes.Number, nil),
		"disable_buildkit_cache":   tftypes.NewValue(tftypes.Bool, false),
		"skip_git_submodules":      tftypes.NewValue(tftypes.Bool, false),
	})
	nullJSON := tftypes.NewValue(tftypes.String, nil)
	unknownJSON := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	inheritedJSON := tftypes.NewValue(tftypes.String, `{"build.timeout_max_sec":1800,"network.ingress.proxy_body_size_mb":100}`)

	testCases := []struct {
		TestName         string
		ConfigBlock      tftypes.Value
		ConfigJSON       tftypes.Value
		PlanJSON         tftypes.Value
		ExpectError      bool
		ExpectedPlanJSON types.String
	}{
		{
			TestName:         "legacy_build_keys_without_block_are_kept",
			ConfigBlock:      nullBlock,
			ConfigJSON:       inheritedJSON,
			PlanJSON:         inheritedJSON,
			ExpectedPlanJSON: types.StringValue(`{"build.timeout_max_sec":1800,"network.ingress.proxy_body_size_mb":100}`),
		},
		{
			TestName:         "inherited_build_keys_kept_without_block",
			ConfigBlock:      nullBlock,
			ConfigJSON:       nullJSON,
			PlanJSON:         inheritedJSON,
			ExpectedPlanJSON: types.StringValue(`{"build.timeout_max_sec":1800,"network.ingress.proxy_body_size_mb":100}`),
		},
		{
			TestName:    "configured_build_keys_conflict_with_block",
			ConfigBlock: setBlock,
			ConfigJSON:  tftypes.NewValue(tftypes.String, `{"build.timeout_max_sec":1800}`),
			PlanJSON:    tftypes.NewValue(tftypes.String, `{"build.timeout_max_sec":1800}`),
			ExpectError: true,
		},
		{
			TestName:    "build_key_added_next_to_other_settings_conflicts_with_block",
			ConfigBlock: setBlock,
			ConfigJSON:  tftypes.NewValue(tftypes.String, `{"network.ingress.proxy_body_size_mb":100,"build.timeout_max_sec":3600}`),
			PlanJSON:    tftypes.NewValue(tftypes.String, `{"network.ingress.proxy_body_size_mb":100,"build.timeout_max_sec":3600}`),
			ExpectError: true,
		},
		{
			TestName:         "configured_json_without_build_keys_is_kept",
			ConfigBlock:      setBlock,
			ConfigJSON:       tftypes.NewValue(tftypes.String, `{"network.ingress.proxy_body_size_mb":100}`),
			PlanJSON:         tftypes.NewValue(tftypes.String, `{"network.ingress.proxy_body_size_mb":100}`),
			ExpectedPlanJSON: types.StringValue(`{"network.ingress.proxy_body_size_mb":100}`),
		},
		{
			TestName:         "inherited_build_keys_are_stripped_when_block_is_set",
			ConfigBlock:      setBlock,
			ConfigJSON:       nullJSON,
			PlanJSON:         inheritedJSON,
			ExpectedPlanJSON: types.StringValue(`{"network.ingress.proxy_body_size_mb":100}`),
		},
		{
			TestName:         "unknown_plan_json_is_left_unknown",
			ConfigBlock:      setBlock,
			ConfigJSON:       nullJSON,
			PlanJSON:         unknownJSON,
			ExpectedPlanJSON: types.StringUnknown(),
		},
		{
			TestName:         "unknown_block_leaves_inherited_json_open",
			ConfigBlock:      unknownBlock,
			ConfigJSON:       nullJSON,
			PlanJSON:         inheritedJSON,
			ExpectedPlanJSON: types.StringUnknown(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			plan := tfsdk.Plan{Schema: s, Raw: buildSettingsPlanTestValue(t, tc.ConfigBlock, tc.PlanJSON)}
			req := resource.ModifyPlanRequest{
				Config: tfsdk.Config{Schema: s, Raw: buildSettingsPlanTestValue(t, tc.ConfigBlock, tc.ConfigJSON)},
				Plan:   plan,
			}
			resp := &resource.ModifyPlanResponse{Plan: plan}

			modifyBuildSettingsPlan(ctx, req, resp)

			assert.Equal(t, tc.ExpectError, resp.Diagnostics.HasError())
			if tc.ExpectError {
				return
			}
			var planJSON types.String
			require.False(t, resp.Plan.GetAttribute(ctx, path.Root("advanced_settings_json"), &planJSON).HasError())
			assert.Equal(t, tc.ExpectedPlanJSON, planJSON)
		})
	}
}

func TestModifyBuildSettingsPlan_Destroy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := buildSettingsPlanTestSchema()
	nullPlan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	resp := &resource.ModifyPlanResponse{Plan: nullPlan}

	modifyBuildSettingsPlan(ctx, resource.ModifyPlanRequest{Plan: nullPlan}, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.Plan.Raw.IsNull())
}
