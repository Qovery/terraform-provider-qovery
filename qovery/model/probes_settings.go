package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var ProbesSettingsDefault = map[string]AdvSettingAttr{
	"liveness_probe.type": {"Specify the type of liveness probe: TCP, HTTP or NONE", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("TCP"),
	}, types.String{Value: "TCP"}},
	"liveness_probe.http_get.path": {"Path to access on the HTTP/HTTPS server to perform the health check", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("/"),
	}, types.String{Value: "/"}},
	"liveness_probe.initial_delay_seconds": {"Interval in seconds between the container start and the first liveness check", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(30),
	}, types.Int64{Value: 30}},
	"liveness_probe.period_seconds": {"Interval in seconds between each liveness check", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(10),
	}, types.Int64{Value: 10}},
	"liveness_probe.timeout_seconds": {"Interval in seconds after the liveness probe times out", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(5),
	}, types.Int64{Value: 5}},
	"liveness_probe.success_threshold": {"Specify how many consecutive successes are needed to be considered successful after having failed", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(1),
	}, types.Int64{Value: 1}},
	"liveness_probe.failure_threshold": {"Specify how many consecutive failures are needed to be considered failed after having succeeded", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(3),
	}, types.Int64{Value: 3}},
	"readiness_probe.type": {"Specify the type of readiness probe: TCP, HTTP or NONE", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("TCP"),
	}, types.String{Value: "TCP"}},
	"readiness_probe.http_get.path": {"Path to access on the HTTP/HTTPS server to perform the health check", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("/"),
	}, types.String{Value: "/"}},
	"readiness_probe.initial_delay_seconds": {"Interval in seconds between the container start and the first readiness check", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(30),
	}, types.Int64{Value: 30}},
	"readiness_probe.period_seconds": {"Interval in seconds between each readiness check", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(10),
	}, types.Int64{Value: 10}},
	"readiness_probe.timeout_seconds": {"Interval in seconds after the readiness probe times out", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(1),
	}, types.Int64{Value: 1}},
	"readiness_probe.success_threshold": {"Specify how many consecutive successes are needed to be considered successful after having failed", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(1),
	}, types.Int64{Value: 1}},
	"readiness_probe.failure_threshold": {"Specify how many consecutive failures are needed to be considered failed after having succeeded", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(3),
	}, types.Int64{Value: 3}},
}
