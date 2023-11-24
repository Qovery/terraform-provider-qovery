package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = stringEnumValidator{}

// stringEnumValidator validates that the value is contained in enum
type stringEnumValidator struct {
	enum []string
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v stringEnumValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string value must be one of [%s]", strings.Join(v.enum, ", "))
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v stringEnumValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("string value must be one of [`%s`]", strings.Join(v.enum, "`, `"))
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v stringEnumValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// types.String must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic validators, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.ConfigValue, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.IsUnknown() || str.IsNull() {
		return
	}

	for _, e := range v.enum {
		if e == str.ValueString() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid String Value",
		fmt.Sprintf("string value must be one of [%s], got: %s.", strings.Join(v.enum, ", "), str.ValueString()),
	)
}

func NewStringEnumValidator(enum []string) validator.String {
	return stringEnumValidator{
		enum: enum,
	}
}
