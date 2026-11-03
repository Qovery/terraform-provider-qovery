package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var _ validator.String = JSONTargetKeyValidator{}

// JSONTargetKeyValidator validates optional external-secret JSON member/index selectors.
type JSONTargetKeyValidator struct{}

// Description describes the accepted JSON selector syntax.
func (JSONTargetKeyValidator) Description(context.Context) string {
	return "must be a JSON member/index path of at most 256 characters, such as accounts.0.password"
}

// MarkdownDescription describes the accepted JSON selector syntax.
func (v JSONTargetKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString skips null and unknown values and reports invalid selectors at their attribute path.
func (JSONTargetKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := variable.ValidateJSONTargetKey(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON property path", err.Error())
	}
}
