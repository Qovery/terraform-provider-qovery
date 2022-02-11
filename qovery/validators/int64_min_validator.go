package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Int64MinValidator struct {
	Min int64
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v Int64MinValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("number value must be greater than %d", v.Min)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v Int64MinValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("number value must be greater than `%d`", v.Min)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v Int64MinValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	// types.Int64 must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic validators, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var number types.Int64
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributeConfig, &number)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if number.Unknown || number.Null {
		return
	}

	if number.Value < v.Min {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Number Value",
			fmt.Sprintf("Number value must be greater than %d, got: %d.", v.Min, number.Value),
		)
		return
	}
}
