package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinRunningInstancesAutoscalingValidator validates the minimum number of
// running instances against the presence of a KEDA `autoscaling` block.
//
// KEDA unlocks scale-to-zero: when `autoscaling` is set, `min_running_instances`
// may be 0; otherwise it must be at least 1. This mirrors the q-core 400 error
// "minimum of 0 is only allowed with KEDA autoscaling" at plan time.
type MinRunningInstancesAutoscalingValidator struct {
	// AutoscalingAttributePath is the root attribute name of the autoscaling block (e.g. "autoscaling").
	AutoscalingAttributePath string
}

func (v MinRunningInstancesAutoscalingValidator) Description(_ context.Context) string {
	return "minimum number of instances must be >= 1, or >= 0 when an autoscaling block is set"
}

func (v MinRunningInstancesAutoscalingValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v MinRunningInstancesAutoscalingValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var autoscalingObj types.Object
	diags := req.Config.GetAttribute(ctx, path.Root(v.AutoscalingAttributePath), &autoscalingObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasAutoscaling := !autoscalingObj.IsNull() && !autoscalingObj.IsUnknown()

	floor := int64(1)
	if hasAutoscaling {
		floor = 0
	}

	if req.ConfigValue.ValueInt64() < floor {
		detail := fmt.Sprintf("Number value must be greater than or equal to %d, got: %d.", floor, req.ConfigValue.ValueInt64())
		if !hasAutoscaling && req.ConfigValue.ValueInt64() == 0 {
			detail = "Number value of 0 (scale-to-zero) is only allowed when a KEDA `autoscaling` block is set."
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Minimum Running Instances", detail)
	}
}
