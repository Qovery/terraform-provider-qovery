package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Container struct {
	ID                          types.String  `tfsdk:"id"`
	EnvironmentID               types.String  `tfsdk:"environment_id"`
	RegistryID                  types.String  `tfsdk:"registry_id"`
	Name                        types.String  `tfsdk:"name"`
	ImageName                   types.String  `tfsdk:"image_name"`
	Tag                         types.String  `tfsdk:"tag"`
	Entrypoint                  types.String  `tfsdk:"entrypoint"`
	CPU                         types.Int64   `tfsdk:"cpu"`
	Memory                      types.Int64   `tfsdk:"memory"`
	MinRunningInstances         types.Int64   `tfsdk:"min_running_instances"`
	MaxRunningInstances         types.Int64   `tfsdk:"max_running_instances"`
	AutoPreview                 types.Bool    `tfsdk:"auto_preview"`
	BuiltInEnvironmentVariables types.Set     `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables        types.Set     `tfsdk:"environment_variables"`
	Secrets                     types.Set     `tfsdk:"secrets"`
	Storages                    types.Set     `tfsdk:"storage"`
	Ports                       types.Set     `tfsdk:"ports"`
	Arguments                   types.List    `tfsdk:"arguments"`
	ExternalHost                types.String  `tfsdk:"external_host"`
	InternalHost                types.String  `tfsdk:"internal_host"`
	DeploymentStageId           types.String  `tfsdk:"deployment_stage_id"`
	Healthchecks                *HealthChecks `tfsdk:"healthchecks"`
	AdvancedSettingsJson        types.String  `tfsdk:"advanced_settings_json"`
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
	return ToStringArray(cont.Arguments)
}

//func (cont Container) CustomDomainsList() CustomDomainList {
//	return toCustomDomainList(cont.CustomDomains)
//}

func (cont Container) toUpsertServiceRequest(state *Container) (*container.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
	}

	var stateSecrets SecretList
	if state != nil {
		stateSecrets = state.SecretList()
	}

	//var stateCustomDomains CustomDomainList
	//if state != nil {
	//	stateCustomDomains = state.CustomDomainsList()
	//}

	return &container.UpsertServiceRequest{
		ContainerUpsertRequest: cont.toUpsertRepositoryRequest(),
		EnvironmentVariables:   cont.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		Secrets:                cont.SecretList().diffRequest(stateSecrets),
		//CustomDomains:          cont.CustomDomainsList().diff(stateCustomDomains),
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
		RegistryID:           ToString(cont.RegistryID),
		Name:                 ToString(cont.Name),
		ImageName:            ToString(cont.ImageName),
		Tag:                  ToString(cont.Tag),
		AutoPreview:          ToBoolPointer(cont.AutoPreview),
		Entrypoint:           ToStringPointer(cont.Entrypoint),
		CPU:                  ToInt32Pointer(cont.CPU),
		Memory:               ToInt32Pointer(cont.Memory),
		MinRunningInstances:  ToInt32Pointer(cont.MinRunningInstances),
		MaxRunningInstances:  ToInt32Pointer(cont.MaxRunningInstances),
		Arguments:            cont.ArgumentList(),
		Storages:             storages,
		Ports:                ports,
		DeploymentStageID:    ToString(cont.DeploymentStageId),
		Healthchecks:         cont.Healthchecks.toHealthchecksRequest(),
		AdvancedSettingsJson: ToString(cont.AdvancedSettingsJson),
	}
}

func convertDomainContainerToContainer(state Container, container *container.Container) Container {
	return Container{
		ID:                          FromString(container.ID.String()),
		EnvironmentID:               FromString(container.EnvironmentID.String()),
		RegistryID:                  FromString(container.RegistryID.String()),
		Name:                        FromString(container.Name),
		ImageName:                   FromString(container.ImageName),
		Tag:                         FromString(container.Tag),
		CPU:                         FromInt32(container.CPU),
		Memory:                      FromInt32(container.Memory),
		MinRunningInstances:         FromInt32(container.MinRunningInstances),
		MaxRunningInstances:         FromInt32(container.MaxRunningInstances),
		AutoPreview:                 FromBool(container.AutoPreview),
		Arguments:                   FromStringArray(container.Arguments),
		Storages:                    convertDomainStoragesToStorageList(container.Storages).toTerraformSet(),
		Ports:                       convertDomainPortsToPortList(container.Ports).toTerraformSet(),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(container.EnvironmentVariables, variable.ScopeContainer).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(container.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		InternalHost:                FromStringPointer(container.InternalHost),
		ExternalHost:                FromStringPointer(container.ExternalHost),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), container.Secrets, variable.ScopeContainer).toTerraformSet(),
		DeploymentStageId:           FromString(container.DeploymentStageID),
		Healthchecks:                convertHealthchecksResponseToDomain(&container.Healthchecks),
		AdvancedSettingsJson:        FromString(container.AdvancedSettingsJson),
	}
}
