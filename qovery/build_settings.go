package qovery

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Platform defaults, identical to the API defaults applied to omitted build settings properties.
const (
	buildSettingsDefaultTimeoutMaxSec        = 1800
	buildSettingsDefaultCpuMaxInMilli        = 4000
	buildSettingsDefaultRamMaxInGib          = 8
	buildSettingsDefaultDisableBuildkitCache = false
	buildSettingsDefaultSkipGitSubmodules    = false
)

var buildSettingsAttrTypes = map[string]attr.Type{
	"timeout_max_sec":          types.Int64Type,
	"cpu_max_in_milli":         types.Int64Type,
	"ram_max_in_gib":           types.Int64Type,
	"ephemeral_storage_in_gib": types.Int64Type,
	"disable_buildkit_cache":   types.BoolType,
	"skip_git_submodules":      types.BoolType,
}

func buildSettingsInt64Validators() []validator.Int64 {
	return []validator.Int64{
		validators.Int64MinMaxValidator{Min: 0, Max: math.MaxInt32},
	}
}

func buildSettingsResourceSchemaAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: buildSettingsDescription + buildSettingsManagedNote,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"timeout_max_sec": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(buildSettingsTimeoutMaxSecDescription, buildSettingsDefaultTimeoutMaxSec),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(buildSettingsDefaultTimeoutMaxSec),
				Validators:          buildSettingsInt64Validators(),
			},
			"cpu_max_in_milli": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(buildSettingsCPUMaxInMilliDescription, buildSettingsDefaultCpuMaxInMilli),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(buildSettingsDefaultCpuMaxInMilli),
				Validators:          buildSettingsInt64Validators(),
			},
			"ram_max_in_gib": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(buildSettingsRAMMaxInGibDescription, buildSettingsDefaultRamMaxInGib),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(buildSettingsDefaultRamMaxInGib),
				Validators:          buildSettingsInt64Validators(),
			},
			"ephemeral_storage_in_gib": schema.Int64Attribute{
				MarkdownDescription: buildSettingsEphemeralStorageDescription + " Omitting it uses the platform default.",
				Optional:            true,
				Validators:          buildSettingsInt64Validators(),
			},
			"disable_buildkit_cache": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(buildSettingsDisableBuildkitCacheDescription, buildSettingsDefaultDisableBuildkitCache),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(buildSettingsDefaultDisableBuildkitCache),
			},
			"skip_git_submodules": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(buildSettingsSkipGitSubmodulesDescription, buildSettingsDefaultSkipGitSubmodules),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(buildSettingsDefaultSkipGitSubmodules),
			},
		},
	}
}

func buildSettingsDataSourceSchemaAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: buildSettingsDescription,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"timeout_max_sec": schema.Int64Attribute{
				MarkdownDescription: buildSettingsTimeoutMaxSecDescription,
				Computed:            true,
			},
			"cpu_max_in_milli": schema.Int64Attribute{
				MarkdownDescription: buildSettingsCPUMaxInMilliDescription,
				Computed:            true,
			},
			"ram_max_in_gib": schema.Int64Attribute{
				MarkdownDescription: buildSettingsRAMMaxInGibDescription,
				Computed:            true,
			},
			"ephemeral_storage_in_gib": schema.Int64Attribute{
				MarkdownDescription: buildSettingsEphemeralStorageDescription,
				Computed:            true,
			},
			"disable_buildkit_cache": schema.BoolAttribute{
				MarkdownDescription: buildSettingsDisableBuildkitCacheDescription,
				Computed:            true,
			},
			"skip_git_submodules": schema.BoolAttribute{
				MarkdownDescription: buildSettingsSkipGitSubmodulesDescription,
				Computed:            true,
			},
		},
	}
}

func buildSettingsObjectToQovery(obj types.Object) *qovery.BuildSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	attrs := obj.Attributes()

	result := qovery.NewBuildSettings()
	result.SetTimeoutMaxSec(int32(attrs["timeout_max_sec"].(types.Int64).ValueInt64()))
	result.SetCpuMaxInMilli(int32(attrs["cpu_max_in_milli"].(types.Int64).ValueInt64()))
	result.SetRamMaxInGib(int32(attrs["ram_max_in_gib"].(types.Int64).ValueInt64()))

	ephemeral := attrs["ephemeral_storage_in_gib"].(types.Int64)
	if !ephemeral.IsNull() && !ephemeral.IsUnknown() {
		result.SetEphemeralStorageInGib(int32(ephemeral.ValueInt64()))
	} else {
		result.SetEphemeralStorageInGibNil()
	}

	result.SetDisableBuildkitCache(attrs["disable_buildkit_cache"].(types.Bool).ValueBool())
	result.SetSkipGitSubmodules(attrs["skip_git_submodules"].(types.Bool).ValueBool())

	return result
}

// buildSettingsRequest returns the build settings to send when creating or updating a service, given the
// prior state (null on create). Removing a block that was in state resets the service to the defaults.
// A block that was never set sends nothing, so services configured through build.* keys in
// advanced_settings_json are left untouched.
func buildSettingsRequest(plan types.Object, state types.Object) *qovery.BuildSettings {
	if !plan.IsNull() {
		return buildSettingsObjectToQovery(plan)
	}
	if state.IsNull() || state.IsUnknown() {
		return nil
	}

	defaults := qovery.NewBuildSettings()
	defaults.SetTimeoutMaxSec(buildSettingsDefaultTimeoutMaxSec)
	defaults.SetCpuMaxInMilli(buildSettingsDefaultCpuMaxInMilli)
	defaults.SetRamMaxInGib(buildSettingsDefaultRamMaxInGib)
	defaults.SetEphemeralStorageInGibNil()
	defaults.SetDisableBuildkitCache(buildSettingsDefaultDisableBuildkitCache)
	defaults.SetSkipGitSubmodules(buildSettingsDefaultSkipGitSubmodules)
	return defaults
}

func buildSettingsFromQovery(buildSettings *qovery.BuildSettings) types.Object {
	if buildSettings == nil {
		return types.ObjectNull(buildSettingsAttrTypes)
	}

	ephemeral := types.Int64Null()
	if value, ok := buildSettings.GetEphemeralStorageInGibOk(); ok && value != nil {
		ephemeral = types.Int64Value(int64(*value))
	}

	return types.ObjectValueMust(buildSettingsAttrTypes, map[string]attr.Value{
		"timeout_max_sec":          types.Int64Value(int64(buildSettings.GetTimeoutMaxSec())),
		"cpu_max_in_milli":         types.Int64Value(int64(buildSettings.GetCpuMaxInMilli())),
		"ram_max_in_gib":           types.Int64Value(int64(buildSettings.GetRamMaxInGib())),
		"ephemeral_storage_in_gib": ephemeral,
		"disable_buildkit_cache":   types.BoolValue(buildSettings.GetDisableBuildkitCache()),
		"skip_git_submodules":      types.BoolValue(buildSettings.GetSkipGitSubmodules()),
	})
}

// buildSettingsToState returns the build_settings value to store after a create, read or update.
// The block stays null unless it is already managed, so services configured through build.* keys in
// advanced_settings_json (or just imported) do not start tracking it. When the API returned build
// settings (reads only), they replace the prior value so out-of-band changes show up as drift.
func buildSettingsToState(prior types.Object, fromAPI *qovery.BuildSettings) types.Object {
	if prior.IsNull() || prior.IsUnknown() {
		return types.ObjectNull(buildSettingsAttrTypes)
	}
	if fromAPI == nil {
		return prior
	}
	return buildSettingsFromQovery(fromAPI)
}

// validateBuildSettingsConflict rejects build.* keys written in advanced_settings_json next to build_settings.
// It is given configuration values, so keys advanced_settings_json only carries over from state never conflict.
func validateBuildSettingsConflict(buildSettings types.Object, advancedSettingsJson types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if buildSettings.IsNull() || buildSettings.IsUnknown() {
		return diags
	}

	if keys := buildSettingsKeysIn(advancedSettingsJson); len(keys) > 0 {
		diags.AddError(
			"Conflicting build configuration",
			fmt.Sprintf(
				"Cannot set both build_settings and %q in advanced_settings_json. "+
					"Remove the build.* keys from advanced_settings_json and use the build_settings block instead.",
				keys[0],
			),
		)
	}

	return diags
}

// warnDeprecatedBuildSettingsKeys warns about build.* keys written in advanced_settings_json.
// It is given configuration values, so keys advanced_settings_json only carries over from state never warn.
func warnDeprecatedBuildSettingsKeys(advancedSettingsJson types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	keys := buildSettingsKeysIn(advancedSettingsJson)
	if len(keys) == 0 {
		return diags
	}

	attributes := make([]string, 0, len(keys))
	for _, key := range keys {
		attributes = append(attributes, strings.TrimPrefix(key, "build."))
	}
	diags.AddAttributeWarning(
		path.Root("advanced_settings_json"),
		"Deprecated build.* advanced settings",
		fmt.Sprintf(
			"The build.* keys of advanced_settings_json are deprecated and a future major release will reject them: %s. "+
				"Remove them and set the build_settings block instead (%s).",
			strings.Join(keys, ", "),
			strings.Join(attributes, ", "),
		),
	)
	return diags
}

// buildSettingsKeysIn returns the build.* keys set in an advanced settings JSON document, in BuildSettingsKeys order.
// A null, unknown or invalid document has none.
func buildSettingsKeysIn(advancedSettingsJson types.String) []string {
	if advancedSettingsJson.IsNull() || advancedSettingsJson.IsUnknown() {
		return nil
	}

	var settings map[string]any
	if err := json.Unmarshal([]byte(advancedSettingsJson.ValueString()), &settings); err != nil {
		return nil
	}

	var keys []string
	for _, key := range advanced_settings.BuildSettingsKeys {
		if _, ok := settings[key]; ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// stripBuildSettingsKeys removes the build.* keys from an advanced settings JSON document.
// It reports false, leaving the document untouched, when there is nothing to remove or the JSON is invalid.
func stripBuildSettingsKeys(advancedSettingsJson string) (string, bool) {
	var settings map[string]any
	if err := json.Unmarshal([]byte(advancedSettingsJson), &settings); err != nil {
		return advancedSettingsJson, false
	}

	changed := false
	for _, key := range advanced_settings.BuildSettingsKeys {
		if _, ok := settings[key]; ok {
			delete(settings, key)
			changed = true
		}
	}
	if !changed {
		return advancedSettingsJson, false
	}

	stripped, err := json.Marshal(settings)
	if err != nil {
		return advancedSettingsJson, false
	}
	return string(stripped), true
}

// modifyBuildSettingsPlan warns about build.* keys written in advanced_settings_json when build_settings is
// not set, and lets build_settings own the build configuration when it is set:
//   - build.* keys written in advanced_settings_json are rejected at plan time;
//   - build.* keys that advanced_settings_json only carries over from state (e.g. after an import, or when
//     migrating away from them) are removed from the plan, otherwise the advanced settings update, which
//     runs after the service update, would write the old values back.
func modifyBuildSettingsPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var configBuildSettings types.Object
	var configAdvancedSettings types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("build_settings"), &configBuildSettings)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("advanced_settings_json"), &configAdvancedSettings)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if configBuildSettings.IsNull() {
		resp.Diagnostics.Append(warnDeprecatedBuildSettingsKeys(configAdvancedSettings)...)
		return
	}

	if !configAdvancedSettings.IsNull() {
		resp.Diagnostics.Append(validateBuildSettingsConflict(configBuildSettings, configAdvancedSettings)...)
		return
	}

	var planAdvancedSettings types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("advanced_settings_json"), &planAdvancedSettings)...)
	if resp.Diagnostics.HasError() || planAdvancedSettings.IsNull() || planAdvancedSettings.IsUnknown() {
		return
	}

	stripped, changed := stripBuildSettingsKeys(planAdvancedSettings.ValueString())
	if !changed {
		return
	}
	if configBuildSettings.IsUnknown() {
		// Whether the block ends up set is only known at apply time: leave the value open rather than
		// planning one that the apply-time plan might have to change.
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("advanced_settings_json"), types.StringUnknown())...)
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("advanced_settings_json"), types.StringValue(stripped))...)
}
