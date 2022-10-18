package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
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
	Arguments    types.Set    `tfsdk:"arguments"`
	ExternalHost types.String `tfsdk:"external_host"`
	InternalHost types.String `tfsdk:"internal_host"`
	State        types.String `tfsdk:"state"`
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

func (cont Container) toUpsertServiceRequest(state *Container) (*container.UpsertServiceRequest, error) {
	desiredState, err := status.NewStateFromString(toString(cont.State))
	if err != nil {
		return nil, err
	}

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
		DesiredState: *desiredState,
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
	}
}

func convertDomainContainerToContainer(state Container, cont *container.Container) Container {
	return Container{
		ID:                          fromString(cont.ID.String()),
		EnvironmentID:               fromString(cont.EnvironmentID.String()),
		RegistryID:                  fromString(cont.RegistryID.String()),
		Name:                        fromString(cont.Name),
		ImageName:                   fromString(cont.ImageName),
		Tag:                         fromString(cont.Tag),
		CPU:                         fromInt32(cont.CPU),
		Memory:                      fromInt32(cont.Memory),
		MinRunningInstances:         fromInt32(cont.MinRunningInstances),
		MaxRunningInstances:         fromInt32(cont.MaxRunningInstances),
		AutoPreview:                 fromBool(cont.AutoPreview),
		Arguments:                   fromStringArray(cont.Arguments),
		Storages:                    convertDomainStoragesToStorageList(cont.Storages).toTerraformSet(),
		Ports:                       convertDomainPortsToPortList(cont.Ports).toTerraformSet(),
		State:                       fromClientEnum(cont.State),
		EnvironmentVariables:        convertDomainVariablesToEnvironmentVariableList(cont.EnvironmentVariables, variable.ScopeContainer).toTerraformSet(),
		BuiltInEnvironmentVariables: convertDomainVariablesToEnvironmentVariableList(cont.BuiltInEnvironmentVariables, variable.ScopeBuiltIn).toTerraformSet(),
		InternalHost:                fromStringPointer(cont.InternalHost),
		ExternalHost:                fromStringPointer(cont.ExternalHost),
		Secrets:                     convertDomainSecretsToSecretList(state.SecretList(), cont.Secrets, variable.ScopeContainer).toTerraformSet(),
	}
}
