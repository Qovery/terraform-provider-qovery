package qovery

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/qovery/validators"
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
		Description: "Build configuration settings for the service. Mutually exclusive with build.* keys in advanced_settings_json — Terraform will reject a plan that uses both.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"timeout_max_sec": schema.Int64Attribute{
				Description: "Maximum build timeout in seconds. Default: 1800.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1800),
				Validators:  buildSettingsInt64Validators(),
			},
			"cpu_max_in_milli": schema.Int64Attribute{
				Description: "Maximum CPU resources for the build in millicores. Default: 4000.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(4000),
				Validators:  buildSettingsInt64Validators(),
			},
			"ram_max_in_gib": schema.Int64Attribute{
				Description: "Maximum RAM resources for the build in GiB. Default: 8.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8),
				Validators:  buildSettingsInt64Validators(),
			},
			"ephemeral_storage_in_gib": schema.Int64Attribute{
				Description: "Ephemeral storage for the build in GiB. When not set, the platform default is used.",
				Optional:    true,
				Validators:  buildSettingsInt64Validators(),
			},
			"disable_buildkit_cache": schema.BoolAttribute{
				Description: "Disable buildkit registry cache during build. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"skip_git_submodules": schema.BoolAttribute{
				Description: "Skip git submodules update when cloning the repository. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func buildSettingsDataSourceSchemaAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Build configuration settings for the service.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"timeout_max_sec": schema.Int64Attribute{
				Description: "Maximum build timeout in seconds. Default: 1800.",
				Computed:    true,
			},
			"cpu_max_in_milli": schema.Int64Attribute{
				Description: "Maximum CPU resources for the build in millicores. Default: 4000.",
				Computed:    true,
			},
			"ram_max_in_gib": schema.Int64Attribute{
				Description: "Maximum RAM resources for the build in GiB. Default: 8.",
				Computed:    true,
			},
			"ephemeral_storage_in_gib": schema.Int64Attribute{
				Description: "Ephemeral storage for the build in GiB.",
				Computed:    true,
			},
			"disable_buildkit_cache": schema.BoolAttribute{
				Description: "Disable buildkit registry cache during build. Default: false.",
				Computed:    true,
			},
			"skip_git_submodules": schema.BoolAttribute{
				Description: "Skip git submodules update when cloning the repository. Default: false.",
				Computed:    true,
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

func buildSettingsPreservePriorState(prior types.Object) types.Object {
	if prior.IsNull() || prior.IsUnknown() {
		return types.ObjectNull(buildSettingsAttrTypes)
	}
	return prior
}

func validateBuildSettingsConflict(buildSettings types.Object, planAdvancedSettingsJson types.String, stateAdvancedSettingsJson types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if buildSettings.IsNull() || buildSettings.IsUnknown() {
		return diags
	}

	if planAdvancedSettingsJson.IsNull() || planAdvancedSettingsJson.IsUnknown() {
		return diags
	}

	if !stateAdvancedSettingsJson.IsNull() && !stateAdvancedSettingsJson.IsUnknown() &&
		planAdvancedSettingsJson.ValueString() == stateAdvancedSettingsJson.ValueString() {
		return diags
	}

	jsonStr := planAdvancedSettingsJson.ValueString()
	if jsonStr == "" || jsonStr == "{}" {
		return diags
	}

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return diags
	}

	buildKeys := []string{
		"build.timeout_max_sec",
		"build.cpu_max_in_milli",
		"build.ram_max_in_gib",
		"build.ephemeral_storage_in_gib",
		"build.disable_buildkit_cache",
		"build.skip_git_submodules",
	}

	for _, key := range buildKeys {
		if _, ok := settings[key]; ok {
			diags.AddError(
				"Conflicting build configuration",
				fmt.Sprintf(
					"Cannot set both build_settings and %q in advanced_settings_json. "+
						"Remove the build.* keys from advanced_settings_json and use the build_settings block instead.",
					key,
				),
			)
			return diags
		}
	}

	return diags
}
