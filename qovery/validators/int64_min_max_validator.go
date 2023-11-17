package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Int64MinMaxValidator struct {
	Min int64
	Max int64
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v Int64MinMaxValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("number value must be greater than %d", v.Min)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v Int64MinMaxValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("number value must be greater than `%d`", v.Min)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v Int64MinMaxValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// types.Int64 must be the attr.Value produced by the attr.Type in the schema for this attribute
	// for generic validators, use
	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
	// to convert into a known type.
	var number types.Int64
	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.ConfigValue, &number)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if number.IsUnknown() || number.IsNull() {
		return
	}

	if number.ValueInt64() < v.Min || number.ValueInt64() > v.Max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Number Value",
			fmt.Sprintf("Number value must be between (inclusive) %d and %d, got: %d.", v.Min, v.Max, number.ValueInt64()),
		)
		return
	}
}
