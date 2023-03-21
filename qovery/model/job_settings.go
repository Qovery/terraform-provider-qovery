package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var JobSettingsDefault = map[string]AdvSettingAttr{
	"job.delete_ttl_seconds_after_finished": {"Kubernetes will automatically cleanup completed jobs after the ttl", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(0),
	}, types.Int64{Value: 0}},
}

var CronJobSettingsDefault = map[string]AdvSettingAttr{
	"cronjob.concurrency_policy": {"Define if it is allowed to start another instance of the same job if the previous execution didn't finish yet", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("Forbidden"),
	}, types.String{Value: "Forbidden"}},
	"cronjob.failed_job_history_limit": {"Define the maximum number of failed job executions that should be returned in the job execution history", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("1"),
	}, types.String{Value: "1"}},
	"cronjob.success_job_history_limit": {"Define the maximum number of succeeded job executions that should be returned in the job execution history", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("1"),
	}, types.String{Value: "1"}},
}
