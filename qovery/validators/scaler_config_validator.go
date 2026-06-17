package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScalerConfigExactlyOneValidator validates that every scaler in the set sets
// exactly one of `config_json` / `config_yaml`, and that at least one scaler is
// enabled. It mirrors the API constraints (config_json ⊻ config_yaml, and the
// requirement that a KEDA policy carries at least one enabled scaler) at plan
// time, before any service mutation reaches the backend.
type ScalerConfigExactlyOneValidator struct{}

func (v ScalerConfigExactlyOneValidator) Description(_ context.Context) string {
	return "each scaler must set exactly one of config_json or config_yaml, and at least one scaler must be enabled"
}

func (v ScalerConfigExactlyOneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ScalerConfigExactlyOneValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if len(req.ConfigValue.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Scaler Configuration",
			"At least one scaler is required when an autoscaling block is set.",
		)
		return
	}

	enabledCount := 0
	for i, elem := range req.ConfigValue.Elements() {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			continue
		}

		attrs := obj.Attributes()
		hasJSON := isAttrSet(attrs["config_json"])
		hasYAML := isAttrSet(attrs["config_yaml"])

		if hasJSON == hasYAML {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Scaler Configuration",
				fmt.Sprintf("Scaler at index %d must set exactly one of config_json or config_yaml.", i),
			)
		}

		// `enabled` defaults to true and is only applied at plan time, so a null
		// or unknown value here means the scaler will be enabled. Only an
		// explicit `enabled = false` disables it.
		if !isScalerExplicitlyDisabled(attrs["enabled"]) {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Scaler Configuration",
			"When a KEDA autoscaling block is set, at least one scaler must be enabled. All scalers currently have enabled = false.",
		)
	}
}

// isScalerExplicitlyDisabled reports whether the `enabled` attribute is set to a
// concrete false value. Null/unknown counts as enabled (default true).
func isScalerExplicitlyDisabled(v attr.Value) bool {
	b, ok := v.(types.Bool)
	if !ok || b.IsNull() || b.IsUnknown() {
		return false
	}
	return !b.ValueBool()
}

func isAttrSet(v attr.Value) bool {
	return v != nil && !v.IsNull() && !v.IsUnknown()
}
