package modifiers

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/hashicorp/terraform-plugin-framework/attr"
//	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//)
//
//func NewStringMapDefaultModifier(defaultValue map[string]string) StringMapDefaultModifier {
//	return StringMapDefaultModifier{
//		Default: defaultValue,
//	}
//}
//
//// StringMapDefaultModifier is a plan modifier that sets a default value for a
//// types.MapType attribute with types.StringType elements when it is not configured. The attribute must be
//// marked as Optional and Computed. When setting the state during the resource
//// Create, Read, or Update methods, this default value must also be included or
//// the Terraform CLI will generate an error.
//type StringMapDefaultModifier struct {
//	Default map[string]string
//}
//
//// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
//func (m StringMapDefaultModifier) Description(_ context.Context) string {
//	return fmt.Sprintf("If value is not configured, defaults to %s", m.Default)
//}
//
//// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
//func (m StringMapDefaultModifier) MarkdownDescription(_ context.Context) string {
//	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Default)
//}
//
//// Modify runs the logic of the plan modifier.
//// Access to the configuration, plan, and state is available in `req`, while
//// `resp` contains fields for updating the planned value, triggering resource
//// replacement, and returning diagnostics.
//func (m StringMapDefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
//	var attribute types.Map
//	resp.Diagnostics.Append(tfsdk.ValueAs(ctx, req.AttributePlan, &attribute)...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	if !attribute.Null && !attribute.Unknown {
//		return
//	}
//
//	_map := types.Map{ElemType: types.StringType}
//	_map.Elems = make(map[string]attr.Value)
//
//	for k, v := range m.Default {
//		_map.Elems[k] = types.String{Value: v}
//	}
//
//	resp.AttributePlan = _map
//}
