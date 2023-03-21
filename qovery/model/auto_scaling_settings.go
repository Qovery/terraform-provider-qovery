package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var AutoScalingSettingsDefault = map[string]AdvSettingAttr{
	"hpa.cpu.average_utilization_percent": {"CPU usage autoscaling trigger value", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
}
