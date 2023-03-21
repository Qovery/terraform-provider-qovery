package model

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type AdvSettingAttr struct {
	Description   string
	Type          attr.Type
	PlanModifiers tfsdk.AttributePlanModifiers
	DefaultValue  attr.Value
}

func GetApplicationSettingsDefault() map[string]AdvSettingAttr {
	applicationSettingsDefault := make(map[string]AdvSettingAttr)

	for k, v := range AppSettingsDefault {
		applicationSettingsDefault[k] = v
	}

	for k, v := range DeploymentSettingsDefault {
		applicationSettingsDefault[k] = v
	}

	for k, v := range ProbesSettingsDefault {
		applicationSettingsDefault[k] = v
	}

	for k, v := range NetworkSettingsDefault {
		applicationSettingsDefault[k] = v
	}

	for k, v := range AutoScalingSettingsDefault {
		applicationSettingsDefault[k] = v
	}

	for k, v := range SecuritySettingsDefault {
		applicationSettingsDefault[k] = v
	}

	return applicationSettingsDefault
}

func GetContainerSettingsDefault() map[string]AdvSettingAttr {
	containerSettingsDefault := make(map[string]AdvSettingAttr)

	for k, v := range DeploymentSettingsDefault {
		containerSettingsDefault[k] = v
	}

	for k, v := range ProbesSettingsDefault {
		containerSettingsDefault[k] = v
	}

	for k, v := range NetworkSettingsDefault {
		containerSettingsDefault[k] = v
	}

	for k, v := range AutoScalingSettingsDefault {
		containerSettingsDefault[k] = v
	}

	for k, v := range SecuritySettingsDefault {
		containerSettingsDefault[k] = v
	}

	return containerSettingsDefault
}

func GetCronJobSettingsDefault() map[string]AdvSettingAttr {
	cronJobSettingsDefault := make(map[string]AdvSettingAttr)

	for k, v := range DeploymentSettingsDefault {
		cronJobSettingsDefault[k] = v
	}

	for k, v := range ProbesSettingsDefault {
		cronJobSettingsDefault[k] = v
	}

	for k, v := range JobSettingsDefault {
		cronJobSettingsDefault[k] = v
	}

	for k, v := range CronJobSettingsDefault {
		cronJobSettingsDefault[k] = v
	}

	for k, v := range SecuritySettingsDefault {
		cronJobSettingsDefault[k] = v
	}

	return cronJobSettingsDefault
}

func GetLifecycleJobSettingsDefault() map[string]AdvSettingAttr {
	lifecycleJobSettingsDefault := make(map[string]AdvSettingAttr)

	for k, v := range DeploymentSettingsDefault {
		lifecycleJobSettingsDefault[k] = v
	}

	for k, v := range ProbesSettingsDefault {
		lifecycleJobSettingsDefault[k] = v
	}

	for k, v := range JobSettingsDefault {
		lifecycleJobSettingsDefault[k] = v
	}

	for k, v := range SecuritySettingsDefault {
		lifecycleJobSettingsDefault[k] = v
	}

	return lifecycleJobSettingsDefault
}

func GetClusterSettingsDefault() map[string]AdvSettingAttr {
	return ClusterSettingsDefault
}
