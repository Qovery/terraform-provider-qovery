package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var AppSettingsDefault = map[string]AdvSettingAttr{
	"build.timeout_max_sec": {"Interval in seconds after which the application build times out", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(1800),
	}, types.Int64{Value: 1800}},
	"deployment.delay_start_time_sec": {"Interval in seconds after which the application build times out", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(30),
	}, types.Int64{Value: 30}},
}
var DeploymentSettingsDefault = map[string]AdvSettingAttr{
	"deployment.custom_domain_check_enabled": {"Allows you to specify the IAM group name associated to the Qovery user", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(true),
	}, types.Bool{Value: true}},
	"deployment.termination_grace_period_seconds": {"Time in seconds the application is supposed to stop at maximum", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
}
