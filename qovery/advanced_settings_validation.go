package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

// advancedSettingsJSONAttr is the attribute name shared by all service resources.
const advancedSettingsJSONAttr = "advanced_settings_json"

// warnUnknownAdvancedSettings adds a plan-time warning for each key in advanced_settings_json
// that is not valid for the given service type.
func warnUnknownAdvancedSettings(
	ctx context.Context,
	svc *advanced_settings.ServiceAdvancedSettingsService,
	serviceType int,
	cfg tfsdk.Config,
	diags *diag.Diagnostics,
) {
	if svc == nil {
		return
	}
	warnUnknownAdvancedSettingsKeys(ctx, func(advancedSettingsJson string) ([]string, error) {
		return svc.UnknownSettingKeys(serviceType, advancedSettingsJson)
	}, cfg, diags)
}

// warnUnknownClusterAdvancedSettings adds a plan-time warning for each key in
// advanced_settings_json that is not a valid cluster advanced setting.
func warnUnknownClusterAdvancedSettings(
	ctx context.Context,
	svc *advanced_settings.ClusterAdvancedSettingsService,
	cfg tfsdk.Config,
	diags *diag.Diagnostics,
) {
	if svc == nil {
		return
	}
	warnUnknownAdvancedSettingsKeys(ctx, svc.UnknownSettingKeys, cfg, diags)
}

// warnUnknownAdvancedSettingsKeys reads advanced_settings_json from the config, resolves the
// unknown keys through lookup, and adds a plan-time warning for each. It never blocks the
// plan: a null/unknown attribute or any error (config read, defaults fetch, or JSON parse)
// degrades silently to "no warning".
func warnUnknownAdvancedSettingsKeys(
	ctx context.Context,
	lookup func(advancedSettingsJson string) ([]string, error),
	cfg tfsdk.Config,
	diags *diag.Diagnostics,
) {
	var raw types.String
	if d := cfg.GetAttribute(ctx, path.Root(advancedSettingsJSONAttr), &raw); d.HasError() {
		// Reading config failed (e.g. during destroy when config is null). Nothing to validate.
		return
	}
	if raw.IsNull() || raw.IsUnknown() || raw.ValueString() == "" {
		return
	}

	unknown, err := lookup(raw.ValueString())
	if err != nil {
		tflog.Warn(ctx, "could not validate advanced settings keys", map[string]any{
			"error": err.Error(),
		})
		return
	}

	for _, key := range unknown {
		diags.AddAttributeWarning(
			path.Root(advancedSettingsJSONAttr),
			fmt.Sprintf("Unknown advanced setting %q", key),
			fmt.Sprintf(
				"The advanced setting %q is not a valid setting for this resource and will "+
					"have no effect. Remove it from advanced_settings_json to clear this warning. "+
					"See the resource's advanced settings documentation for the list of valid keys.",
				key,
			),
		)
	}
}
