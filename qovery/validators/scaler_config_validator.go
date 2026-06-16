package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScalerConfigExactlyOneValidator validates that every scaler in the set sets
// exactly one of `config_json` / `config_yaml`. It mirrors the API's
// config_json ⊻ config_yaml constraint at plan time.
type ScalerConfigExactlyOneValidator struct{}

func (v ScalerConfigExactlyOneValidator) Description(_ context.Context) string {
	return "each scaler must set exactly one of config_json or config_yaml"
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
	}
}

func isAttrSet(v attr.Value) bool {
	return v != nil && !v.IsNull() && !v.IsUnknown()
}
