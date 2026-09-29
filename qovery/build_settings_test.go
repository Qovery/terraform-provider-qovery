//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildSettingsObjectToQovery_Nil(t *testing.T) {
	result := buildSettingsObjectToQovery(types.ObjectNull(buildSettingsAttrTypes))
	assert.Nil(t, result)
}

func TestBuildSettingsObjectToQovery_Unknown(t *testing.T) {
	result := buildSettingsObjectToQovery(types.ObjectUnknown(buildSettingsAttrTypes))
	assert.Nil(t, result)
}

func TestBuildSettingsObjectToQovery_WithValues(t *testing.T) {
	obj, diags := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(3600),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Value(10),
		"disable_buildkit_cache":   types.BoolValue(true),
		"skip_git_submodules":      types.BoolValue(false),
	})
	assert.False(t, diags.HasError())

	result := buildSettingsObjectToQovery(obj)
	assert.NotNil(t, result)
	assert.Equal(t, int32(3600), result.GetTimeoutMaxSec())
	assert.Equal(t, int32(4000), result.GetCpuMaxInMilli())
	assert.Equal(t, int32(8), result.GetRamMaxInGib())
	assert.Equal(t, int32(10), result.GetEphemeralStorageInGib())
	assert.True(t, result.GetDisableBuildkitCache())
	assert.False(t, result.GetSkipGitSubmodules())
}

func TestBuildSettingsObjectToQovery_NullEphemeralStorage(t *testing.T) {
	obj, diags := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})
	assert.False(t, diags.HasError())

	result := buildSettingsObjectToQovery(obj)
	assert.NotNil(t, result)
	assert.True(t, result.EphemeralStorageInGib.IsSet())
	assert.Nil(t, result.EphemeralStorageInGib.Get())
}

func TestBuildSettingsPreservePriorState_Null(t *testing.T) {
	result := buildSettingsPreservePriorState(types.ObjectNull(buildSettingsAttrTypes))
	assert.True(t, result.IsNull())
}

func TestBuildSettingsPreservePriorState_WithValue(t *testing.T) {
	obj, diags := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(3600),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})
	assert.False(t, diags.HasError())

	result := buildSettingsPreservePriorState(obj)
	assert.False(t, result.IsNull())
	assert.Equal(t, obj, result)
}

func TestValidateBuildSettingsConflict_NoBuildSettings(t *testing.T) {
	diags := validateBuildSettingsConflict(
		types.ObjectNull(buildSettingsAttrTypes),
		types.StringValue(`{"build.timeout_max_sec": 3600}`),
		types.StringNull(),
	)
	assert.False(t, diags.HasError())
}

func TestValidateBuildSettingsConflict_NoAdvancedSettings(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	diags := validateBuildSettingsConflict(obj, types.StringNull(), types.StringNull())
	assert.False(t, diags.HasError())
}

func TestValidateBuildSettingsConflict_NoConflict(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	diags := validateBuildSettingsConflict(obj, types.StringValue(`{"some.other.key": "value"}`), types.StringNull())
	assert.False(t, diags.HasError())
}

func TestValidateBuildSettingsConflict_Conflict(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	diags := validateBuildSettingsConflict(obj, types.StringValue(`{"build.timeout_max_sec": 3600}`), types.StringNull())
	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "build.timeout_max_sec")
}

func TestValidateBuildSettingsConflict_EmptyJson(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	diags := validateBuildSettingsConflict(obj, types.StringValue(`{}`), types.StringNull())
	assert.False(t, diags.HasError())
}

func TestValidateBuildSettingsConflict_BackendSyncedKeys(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	syncedJson := `{"build.timeout_max_sec": 1800, "build.cpu_max_in_milli": 4000}`

	diags := validateBuildSettingsConflict(obj, types.StringValue(syncedJson), types.StringValue(syncedJson))
	assert.False(t, diags.HasError())
}

func TestValidateBuildSettingsConflict_UserChangedAdvancedSettings(t *testing.T) {
	obj, _ := types.ObjectValue(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(1800),
		"cpu_max_in_milli":         types.Int64Value(4000),
		"ram_max_in_gib":           types.Int64Value(8),
		"ephemeral_storage_in_gib": types.Int64Null(),
		"disable_buildkit_cache":   types.BoolValue(false),
		"skip_git_submodules":      types.BoolValue(false),
	})

	stateJson := `{"some.other.key": "value"}`
	planJson := `{"some.other.key": "value", "build.timeout_max_sec": 3600}`

	diags := validateBuildSettingsConflict(obj, types.StringValue(planJson), types.StringValue(stateJson))
	assert.True(t, diags.HasError())
}
