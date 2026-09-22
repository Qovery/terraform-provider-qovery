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

// advancedSettingsRefreshSemantics is appended to the MarkdownDescription of
// advanced_settings_json on every resource that exposes advanced settings. It
// documents the refresh behaviour implemented by computeOverriddenSettings: the
// attribute is a partial override map, the API has no ownership flag, so only
// keys already tracked in state are reconciled (QOV-2028).
const advancedSettingsRefreshSemantics = "\n\n" +
	"  Refresh semantics — this attribute is desired state, not a mirror of the remote configuration:\n\n" +
	"  - Refresh only reconciles keys already tracked in the Terraform state. A setting overridden only in the Qovery Console is not pulled into state, so declaring it afterwards plans as an addition even if the remote value already matches.\n" +
	"  - Changes made in the Console to a tracked key, including a reset to its default value, are reflected on refresh and planned back to the configured value.\n" +
	"  - Removing a key from the JSON does not reset it remotely: omitted keys keep their current value. To reset a setting, set it to its default value explicitly. Omitting the attribute entirely leaves the previously applied settings untouched.\n" +
	"  - `terraform import` records every setting whose value differs from the default."

// advancedSettingsRefreshSemanticsPlain is the plain-text counterpart used for
// the schema Description.
const advancedSettingsRefreshSemanticsPlain = " Refresh only reconciles keys already tracked in state: " +
	"settings overridden only in the Qovery Console are not pulled into state, while Console changes to tracked keys " +
	"(including a reset to the default) are. Removing a key does not reset it remotely; set it to its default value " +
	"explicitly. Import records every non-default setting."

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
