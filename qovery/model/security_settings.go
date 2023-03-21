package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var SecuritySettingsDefault = map[string]AdvSettingAttr{
	"security.service_account_name": {"Allows you to set an existing Kubernetes service account name", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier(""),
	}, types.String{Value: ""}},
}
