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
// that is not valid for the given service type. It never blocks the plan: a null/unknown
// attribute or any fetch error degrades silently to "no warning".
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

	var raw types.String
	if d := cfg.GetAttribute(ctx, path.Root(advancedSettingsJSONAttr), &raw); d.HasError() {
		// Reading config failed (e.g. during destroy when config is null). Nothing to validate.
		return
	}
	if raw.IsNull() || raw.IsUnknown() || raw.ValueString() == "" {
		return
	}

	unknown, err := svc.UnknownSettingKeys(serviceType, raw.ValueString())
	if err != nil {
		tflog.Warn(ctx, "could not validate advanced settings keys", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, key := range unknown {
		diags.AddAttributeWarning(
			path.Root(advancedSettingsJSONAttr),
			"Unknown advanced setting",
			fmt.Sprintf(
				"The advanced setting %q is not a valid setting for this service type and will "+
					"have no effect. Remove it from advanced_settings_json to clear this warning. "+
					"See the service's advanced settings documentation for the list of valid keys.",
				key,
			),
		)
	}
}
