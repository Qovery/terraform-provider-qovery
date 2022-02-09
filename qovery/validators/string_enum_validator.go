package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringEnumValidator validates that the value is contained in Enum
type StringEnumValidator struct {
	Enum []string
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v StringEnumValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string value must be one of [%s]", strings.Join(v.Enum, ", "))
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v StringEnumValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("string value must be one of [`%s`]", strings.Join(v.Enum, "`, `"))
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v StringEnumValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// types.String must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic validators, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	for _, e := range v.Enum {
		if e == str.Value {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.AttributePath,
		"Invalid String Value",
		fmt.Sprintf("string value must be one of [%s], got: %s.", strings.Join(v.Enum, ", "), str.Value),
	)
}
