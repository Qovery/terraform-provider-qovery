//go:build unit || !integration

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// minInstSchema mirrors the relevant slice of a service resource schema: the
// validated min_running_instances attribute plus the autoscaling block whose
// presence the validator reads via Config.GetAttribute.
var minInstSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"min_running_instances": schema.Int64Attribute{},
		"autoscaling": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"polling_interval_seconds": schema.Int64Attribute{},
			},
		},
	},
}

// buildMinInstConfig constructs a tfsdk.Config carrying min_running_instances
// and an autoscaling block that is either present (empty object) or null.
func buildMinInstConfig(minVal int64, autoscalingPresent bool) tfsdk.Config {
	ctx := context.Background()
	objType := minInstSchema.Type().TerraformType(ctx).(tftypes.Object)
	autoscalingType := objType.AttributeTypes["autoscaling"]

	autoscalingVal := tftypes.NewValue(autoscalingType, nil)
	if autoscalingPresent {
		autoscalingVal = tftypes.NewValue(autoscalingType, map[string]tftypes.Value{
			"polling_interval_seconds": tftypes.NewValue(tftypes.Number, nil),
		})
	}

	return tfsdk.Config{
		Schema: minInstSchema,
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"min_running_instances": tftypes.NewValue(tftypes.Number, minVal),
			"autoscaling":           autoscalingVal,
		}),
	}
}

func TestMinRunningInstancesAutoscalingValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName           string
		Min                int64
		AutoscalingPresent bool
		ExpectError        bool
	}{
		{TestName: "one_without_autoscaling_ok", Min: 1, AutoscalingPresent: false, ExpectError: false},
		{TestName: "zero_without_autoscaling_errors", Min: 0, AutoscalingPresent: false, ExpectError: true},
		{TestName: "zero_with_autoscaling_ok", Min: 0, AutoscalingPresent: true, ExpectError: false},
		{TestName: "one_with_autoscaling_ok", Min: 1, AutoscalingPresent: true, ExpectError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := validator.Int64Request{
				Path:        path.Root("min_running_instances"),
				ConfigValue: types.Int64Value(tc.Min),
				Config:      buildMinInstConfig(tc.Min, tc.AutoscalingPresent),
			}
			resp := &validator.Int64Response{}
			MinRunningInstancesAutoscalingValidator{AutoscalingAttributePath: "autoscaling"}.
				ValidateInt64(context.Background(), req, resp)
			if tc.ExpectError {
				assert.True(t, resp.Diagnostics.HasError(), "expected validation error")
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "expected no error, got: %s", resp.Diagnostics.Errors())
			}
		})
	}
}
