package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Container struct {
	ID                          types.String `tfsdk:"id"`
	EnvironmentID               types.String `tfsdk:"environment_id"`
	RegistryID                  types.String `tfsdk:"registry_id"`
	Name                        types.String `tfsdk:"name"`
	ImageName                   types.String `tfsdk:"image_name"`
	Tag                         types.String `tfsdk:"tag"`
	Entrypoint                  types.String `tfsdk:"entrypoint"`
	CPU                         types.Int64  `tfsdk:"cpu"`
	Memory                      types.Int64  `tfsdk:"memory"`
	MinRunningInstances         types.Int64  `tfsdk:"min_running_instances"`
	MaxRunningInstances         types.Int64  `tfsdk:"max_running_instances"`
	AutoPreview                 types.Bool   `tfsdk:"auto_preview"`
	BuiltInEnvironmentVariables types.Set    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	Secrets                     types.Set    `tfsdk:"secrets"`
	Storages                    types.Set    `tfsdk:"storage"`
	Ports                       types.Set    `tfsdk:"ports"`
	//CustomDomains               types.Set    `tfsdk:"custom_domains"`
	Arguments         types.List   `tfsdk:"arguments"`
	ExternalHost      types.String `tfsdk:"external_host"`
	InternalHost      types.String `tfsdk:"internal_host"`
	DeploymentStageId types.String `tfsdk:"deployment_stage_id"`
	AdvancedSettings  types.Object `tfsdk:"advanced_settings"`
}

func (cont Container) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(cont.EnvironmentVariables)
}

func (cont Container) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(cont.BuiltInEnvironmentVariables)
}

func (cont Container) SecretList() SecretList {
	return toSecretList(cont.Secrets)
}

func (cont Container) StorageList() StorageList {
	return toStorageList(cont.Storages)
}

func (cont Container) PortList() PortList {
	return toPortList(cont.Ports)
}

func (cont Container) ArgumentList() []string {
	return toStringArray(cont.Arguments)
}

//func (cont Container) CustomDomainsList() CustomDomainList {
//	return toCustomDomainList(cont.CustomDomains)
//}

func (cont Container) toUpsertServiceRequest(plan *Container) (*container.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	if plan != nil {
		stateEnvironmentVariables = plan.EnvironmentVariableList()
	}

	var stateSecrets SecretList
	if plan != nil {
		stateSecrets = plan.SecretList()
	}

	//var stateCustomDomains CustomDomainList
	//if state != nil {
	//	stateCustomDomains = state.CustomDomainsList()
	//}

	advSettings, err := toMapStringString(cont.AdvancedSettings)
	if err != nil {
		tflog.Warn(context.Background(), "Unable to parse advanced settings, some values will be skipped. It could be related to an outdated version of the provider.", map[string]interface{}{"error": err.Error()})
	}

	return &container.UpsertServiceRequest{
		ContainerUpsertRequest: cont.toUpsertRepositoryRequest(),
		EnvironmentVariables:   cont.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		Secrets:                cont.SecretList().diffRequest(stateSecrets),
		//CustomDomains:          cont.CustomDomainsList().diff(stateCustomDomains),
		AdvancedSettings: advSettings,
	}, nil
}

func (cont Container) toUpsertRepositoryRequest() container.UpsertRepositoryRequest {
	storageList := cont.StorageList()
	storages := make([]storage.UpsertRequest, 0, len(storageList))
	for _, store := range storageList {
		storages = append(storages, store.toUpsertRequest())
	}

	portList := cont.PortList()
	ports := make([]port.UpsertRequest, 0, len(portList))
	for _, prt := range portList {
		ports = append(ports, prt.toUpsertRequest())
	}

	return container.UpsertRepositoryRequest{
		RegistryID:          toString(cont.RegistryID),
		Name:                toString(cont.Name),
		ImageName:           toString(cont.ImageName),
		Tag:                 toString(cont.Tag),
		AutoPreview:         toBoolPointer(cont.AutoPreview),
		Entrypoint:          toStringPointer(cont.Entrypoint),
		CPU:                 toInt32Pointer(cont.CPU),
		Memory:              toInt32Pointer(cont.Memory),
		MinRunningInstances: toInt32Pointer(cont.MinRunningInstances),
		MaxRunningInstances: toInt32Pointer(cont.MaxRunningInstances),
		Arguments:           cont.ArgumentList(),
		Storages:            storages,
		Ports:               ports,
		DeploymentStageID:   toString(cont.DeploymentStageId),
	}
}

func convertDomainContainerToContainer(state Container, container *container.Container) Container {
	return Container{
		ID:                          fromString(container.ID.String()),
		EnvironmentID:               fromString(container.EnvironmentID.String()),
		RegistryID:                  fromString(container.RegistryID.String()),
		Name:                        fromString(container.Name),
		ImageName:                   fromString(container.ImageName),
		Tag:                         fromString(container.Tag),
		CPU:                         fromInt32(container.CPU),
		Memory:                      fromInt32(container.Memory),
		MinRunningInstances:         fromInt32(container.MinRunningInstances),
		MaxRunningInstances:         fromInt32(container.MaxRunningInstances),
		AutoPreview:                 fromBool(container.AutoPreview),
		Arguments:                   fromStringArray(container.Arguments),
		Storages:                    convertDomainStoragesToStorageList(container.Storages).toTerraformSet(),
		Ports:                       convertDomainPortsToPortList(container.Ports).toTerraformSet(),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(container.EnvironmentVariables, variable.ScopeContainer).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(container.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		InternalHost:                fromStringPointer(container.InternalHost),
		ExternalHost:                fromStringPointer(container.ExternalHost),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), container.Secrets, variable.ScopeContainer).toTerraformSet(),
		DeploymentStageId:           fromString(container.DeploymentStageID),
		AdvancedSettings:            fromStringMap(container.ContainerAdvancedSettings, GetContainerSettingsDefault()),
	}
}
