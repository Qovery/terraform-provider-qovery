package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

type Container struct {
	ID                           types.String  `tfsdk:"id"`
	EnvironmentID                types.String  `tfsdk:"environment_id"`
	RegistryID                   types.String  `tfsdk:"registry_id"`
	Name                         types.String  `tfsdk:"name"`
	IconUri                      types.String  `tfsdk:"icon_uri"`
	ImageName                    types.String  `tfsdk:"image_name"`
	Tag                          types.String  `tfsdk:"tag"`
	Entrypoint                   types.String  `tfsdk:"entrypoint"`
	CPU                          types.Int64   `tfsdk:"cpu"`
	Memory                       types.Int64   `tfsdk:"memory"`
	MinRunningInstances          types.Int64   `tfsdk:"min_running_instances"`
	MaxRunningInstances          types.Int64   `tfsdk:"max_running_instances"`
	AutoPreview                  types.Bool    `tfsdk:"auto_preview"`
	BuiltInEnvironmentVariables  types.List    `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set     `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set     `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set     `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set     `tfsdk:"secrets"`
	SecretAliases                types.Set     `tfsdk:"secret_aliases"`
	SecretOverrides              types.Set     `tfsdk:"secret_overrides"`
	Storages                     types.Set     `tfsdk:"storage"`
	Ports                        types.List    `tfsdk:"ports"`
	Arguments                    types.List    `tfsdk:"arguments"`
	CustomDomains                types.Set     `tfsdk:"custom_domains"`
	ExternalHost                 types.String  `tfsdk:"external_host"`
	InternalHost                 types.String  `tfsdk:"internal_host"`
	DeploymentStageId            types.String  `tfsdk:"deployment_stage_id"`
	IsSkipped                    types.Bool    `tfsdk:"is_skipped"`
	Healthchecks                 *HealthChecks `tfsdk:"healthchecks"`
	AdvancedSettingsJson         types.String  `tfsdk:"advanced_settings_json"`
	AutoDeploy                   types.Bool    `tfsdk:"auto_deploy"`
	AnnotationsGroupIds          types.Set     `tfsdk:"annotations_group_ids"`
	LabelsGroupIds               types.Set     `tfsdk:"labels_group_ids"`
}

func (cont Container) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(cont.EnvironmentVariables)
}

func (cont Container) EnvironmentVariableAliasList() EnvironmentVariableList {
	return toEnvironmentVariableList(cont.EnvironmentVariableAliases)
}

func (cont Container) EnvironmentVariableOverrideList() EnvironmentVariableList {
	return toEnvironmentVariableList(cont.EnvironmentVariableOverrides)
}

func (cont Container) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableListFromTerraformList(cont.BuiltInEnvironmentVariables)
}

func (cont Container) SecretList() SecretList {
	return ToSecretList(cont.Secrets)
}

func (cont Container) SecretAliasesList() SecretList {
	return ToSecretList(cont.SecretAliases)
}

func (cont Container) SecretOverridesList() SecretList {
	return ToSecretList(cont.SecretOverrides)
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

func (cont Container) CustomDomainsList() CustomDomainList {
	return toCustomDomainList(cont.CustomDomains)
}

func (cont Container) toUpsertServiceRequest(state *Container) (*container.UpsertServiceRequest, error) {
	var stateEnvironmentVariables EnvironmentVariableList
	var stateEnvironmentVariableAliases EnvironmentVariableList
	var stateEnvironmentVariableOverrides EnvironmentVariableList
	var stateSecrets SecretList
	var stateSecretAliases SecretList
	var stateSecretOverrides SecretList
	var stateCustomDomains CustomDomainList

	if state != nil {
		stateEnvironmentVariables = state.EnvironmentVariableList()
		stateEnvironmentVariableAliases = state.EnvironmentVariableAliasList()
		stateEnvironmentVariableOverrides = state.EnvironmentVariableOverrideList()
		stateSecrets = state.SecretList()
		stateSecretAliases = state.SecretAliasesList()
		stateSecretOverrides = state.SecretOverridesList()
		stateCustomDomains = state.CustomDomainsList()
	}

	return &container.UpsertServiceRequest{
		ContainerUpsertRequest:       cont.toUpsertRepositoryRequest(cont.CustomDomainsList().diff(stateCustomDomains)),
		EnvironmentVariables:         cont.EnvironmentVariableList().diffRequest(stateEnvironmentVariables),
		EnvironmentVariableAliases:   cont.EnvironmentVariableAliasList().diffRequest(stateEnvironmentVariableAliases),
		EnvironmentVariableOverrides: cont.EnvironmentVariableOverrideList().diffRequest(stateEnvironmentVariableOverrides),
		Secrets:                      cont.SecretList().diffRequest(stateSecrets),
		SecretAliases:                cont.SecretAliasesList().diffRequest(stateSecretAliases),
		SecretOverrides:              cont.SecretOverridesList().diffRequest(stateSecretOverrides),
	}, nil
}

func (cont Container) toUpsertRepositoryRequest(customDomainsDiff client.CustomDomainsDiff) container.UpsertRepositoryRequest {
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

	annotationsGroupIds := make([]string, 0, len(cont.AnnotationsGroupIds.Elements()))
	for _, id := range cont.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotationsGroupIds = append(annotationsGroupIds, id.ValueString())
	}

	labelsGroupIds := make([]string, 0, len(cont.LabelsGroupIds.Elements()))
	for _, id := range cont.LabelsGroupIds.Elements() {
		id := id.(types.String)
		labelsGroupIds = append(labelsGroupIds, id.ValueString())
	}

	return container.UpsertRepositoryRequest{
		RegistryID:           ToString(cont.RegistryID),
		Name:                 ToString(cont.Name),
		IconUri:              ToStringPointer(cont.IconUri),
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
		IsSkipped:            ToBool(cont.IsSkipped),
		Healthchecks:         cont.Healthchecks.toHealthchecksRequest(),
		AdvancedSettingsJson: ToString(cont.AdvancedSettingsJson),
		CustomDomains:        customDomainsDiff,
		AutoDeploy:           *qovery.NewNullableBool(ToBoolPointer(cont.AutoDeploy)),
		AnnotationsGroupIds:  annotationsGroupIds,
		LabelsGroupIds:       labelsGroupIds,
	}
}

func convertDomainContainerToContainer(ctx context.Context, state Container, container *container.Container) Container {
	healthchecks := convertHealthchecksResponseToDomain(container.Healthchecks)
	return Container{
		ID:                           FromString(container.ID.String()),
		EnvironmentID:                FromString(container.EnvironmentID.String()),
		RegistryID:                   FromString(container.RegistryID.String()),
		Name:                         FromString(container.Name),
		IconUri:                      FromString(container.IconUri),
		ImageName:                    FromString(container.ImageName),
		Tag:                          FromString(container.Tag),
		CPU:                          FromInt32(container.CPU),
		Memory:                       FromInt32(container.Memory),
		MinRunningInstances:          FromInt32(container.MinRunningInstances),
		MaxRunningInstances:          FromInt32(container.MaxRunningInstances),
		AutoPreview:                  FromBool(container.AutoPreview),
		Entrypoint:                   FromStringPointer(container.Entrypoint),
		Arguments:                    FromStringArray(container.Arguments),
		Storages:                     convertDomainStoragesToStorageList(state.Storages, container.Storages).toTerraformSet(ctx),
		Ports:                        convertDomainPortsToPortList(ctx, state.Ports, container.Ports).toTerraformList(ctx),
		BuiltInEnvironmentVariables:  convertDomainVariablesToEnvironmentVariableList(container.BuiltInEnvironmentVariables, variable.ScopeBuiltIn, "BUILT_IN").toTerraformList(ctx),
		EnvironmentVariables:         convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, container.EnvironmentVariables, variable.ScopeContainer, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, container.EnvironmentVariables, variable.ScopeContainer, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: convertDomainVariablesToEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableOverrides, container.EnvironmentVariables, variable.ScopeContainer, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      convertDomainSecretsToSecretList(state.Secrets, container.Secrets, variable.ScopeContainer, "VALUE").toTerraformSet(ctx),
		SecretAliases:                convertDomainSecretsToSecretList(state.SecretAliases, container.Secrets, variable.ScopeContainer, "ALIAS").toTerraformSet(ctx),
		SecretOverrides:              convertDomainSecretsToSecretList(state.SecretOverrides, container.Secrets, variable.ScopeContainer, "OVERRIDE").toTerraformSet(ctx),
		InternalHost:                 FromStringPointer(container.InternalHost),
		ExternalHost:                 FromStringPointer(container.ExternalHost),
		DeploymentStageId:            FromString(container.DeploymentStageID),
		IsSkipped:                    FromBool(container.IsSkipped),
		Healthchecks:                 &healthchecks,
		AdvancedSettingsJson:         FromString(container.AdvancedSettingsJson),
		CustomDomains:                fromCustomDomainList(state.CustomDomains, container.CustomDomains).toTerraformSet(ctx),
		AutoDeploy:                   FromBoolPointer(container.AutoDeploy),
		AnnotationsGroupIds:          fromAnnotationsGroupList(ctx, state.AnnotationsGroupIds, container.AnnotationsGroupIds),
		LabelsGroupIds:               fromLabelsGroupList(ctx, state.LabelsGroupIds, container.LabelsGroupIds),
	}
}
