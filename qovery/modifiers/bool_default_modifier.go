package modifiers

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//)
//
//func NewBoolDefaultModifier(defaultValue bool) BoolDefaultModifier {
//	return BoolDefaultModifier{
//		Default: defaultValue,
//	}
//}
//
//// BoolDefaultModifier is a plan modifier that sets a default value for a
//// types.BoolType attribute when it is not configured. The attribute must be
//// marked as Optional and Computed. When setting the state during the resource
//// Create, Read, or Update methods, this default value must also be included or
//// the Terraform CLI will generate an error.
//type BoolDefaultModifier struct {
//	Default bool
//}
//
//// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
//func (m BoolDefaultModifier) Description(_ context.Context) string {
//	return fmt.Sprintf("If value is not configured, defaults to %t", m.Default)
//}
//
//// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
//func (m BoolDefaultModifier) MarkdownDescription(_ context.Context) string {
//	return fmt.Sprintf("If value is not configured, defaults to `%t`", m.Default)
//}
//
//// Modify runs the logic of the plan modifier.
//// Access to the configuration, plan, and state is available in `req`, while
//// `resp` contains fields for updating the planned value, triggering resource
//// replacement, and returning diagnostics.
//func (m BoolDefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
//	// types.Bool must be the attr.Value produced by the attr.Type in the schema for this attribute
//	// for generic plan modifiers, use
//	// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ConvertValue
//	// to convert into a known type.
//	var attr types.Bool
//	diags := tfsdk.ValueAs(ctx, req.AttributePlan, &attr)
//	resp.Diagnostics.Append(diags...)
//	if diags.HasError() {
//		return
//	}
//
//	if !attr.Null && !attr.Unknown {
//		return
//	}
//
//	resp.AttributePlan = types.Bool{Value: m.Default}
//}
