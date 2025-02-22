package qovery

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/qoveryapi"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
)

type Application struct {
	Id                           types.String              `tfsdk:"id"`
	EnvironmentId                types.String              `tfsdk:"environment_id"`
	Name                         types.String              `tfsdk:"name"`
	IconUri                      types.String              `tfsdk:"icon_uri"`
	GitRepository                *ApplicationGitRepository `tfsdk:"git_repository"`
	BuildMode                    types.String              `tfsdk:"build_mode"`
	DockerfilePath               types.String              `tfsdk:"dockerfile_path"`
	CPU                          types.Int64               `tfsdk:"cpu"`
	Memory                       types.Int64               `tfsdk:"memory"`
	MinRunningInstances          types.Int64               `tfsdk:"min_running_instances"`
	MaxRunningInstances          types.Int64               `tfsdk:"max_running_instances"`
	AutoPreview                  types.Bool                `tfsdk:"auto_preview"`
	Storage                      []ApplicationStorage      `tfsdk:"storage"`
	Ports                        []ApplicationPort         `tfsdk:"ports"`
	CustomDomains                types.Set                 `tfsdk:"custom_domains"`
	BuiltInEnvironmentVariables  types.Set                 `tfsdk:"built_in_environment_variables"`
	EnvironmentVariables         types.Set                 `tfsdk:"environment_variables"`
	EnvironmentVariableAliases   types.Set                 `tfsdk:"environment_variable_aliases"`
	EnvironmentVariableOverrides types.Set                 `tfsdk:"environment_variable_overrides"`
	Secrets                      types.Set                 `tfsdk:"secrets"`
	SecretVariableAliases        types.Set                 `tfsdk:"secret_aliases"`
	SecretVariableOverrides      types.Set                 `tfsdk:"secret_overrides"`
	ExternalHost                 types.String              `tfsdk:"external_host"`
	InternalHost                 types.String              `tfsdk:"internal_host"`
	Entrypoint                   types.String              `tfsdk:"entrypoint"`
	Arguments                    types.List                `tfsdk:"arguments"`
	DeploymentStageId            types.String              `tfsdk:"deployment_stage_id"`
	Healthchecks                 *HealthChecks             `tfsdk:"healthchecks"`
	AdvancedSettingsJson         types.String              `tfsdk:"advanced_settings_json"`
	AutoDeploy                   types.Bool                `tfsdk:"auto_deploy"`
	DeploymentRestrictions       types.Set                 `tfsdk:"deployment_restrictions"`
	AnnotationsGroupIds          types.Set                 `tfsdk:"annotations_group_ids"`
	LabelsGroupIds               types.Set                 `tfsdk:"labels_group_ids"`
	DockerTargetBuildStage       types.String              `tfsdk:"docker_target_build_stage"`
}

func (app Application) EnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(app.EnvironmentVariables)
}

func (app Application) EnvironmentVariableAliasList() EnvironmentVariableList {
	return toEnvironmentVariableList(app.EnvironmentVariableAliases)
}

func (app Application) EnvironmentVariableOverrideList() EnvironmentVariableList {
	return toEnvironmentVariableList(app.EnvironmentVariableOverrides)
}

func (app Application) BuiltInEnvironmentVariableList() EnvironmentVariableList {
	return toEnvironmentVariableList(app.BuiltInEnvironmentVariables)
}

func (app Application) SecretList() SecretList {
	return ToSecretList(app.Secrets)
}

func (app Application) SecretAliasList() SecretList {
	return ToSecretList(app.SecretVariableAliases)
}

func (app Application) SecretOverrideList() SecretList {
	return ToSecretList(app.SecretVariableOverrides)
}

func (app Application) CustomDomainsList() CustomDomainList {
	return toCustomDomainList(app.CustomDomains)
}

func (app Application) DeploymentRestrictionList(deploymentRestrictionsState *types.Set) (*deploymentrestriction.ServiceDeploymentRestrictionsDiff, error) {
	return deploymentrestriction.ToDeploymentRestrictionDiff(app.DeploymentRestrictions, deploymentRestrictionsState)
}

func (app Application) toCreateApplicationRequest() (*client.ApplicationCreateParams, error) {
	storage := make([]qovery.ServiceStorageRequestStorageInner, 0, len(app.Storage))
	for _, store := range app.Storage {
		s, err := store.toCreateRequest()
		if err != nil {
			return nil, err
		}
		storage = append(storage, *s)
	}

	ports := make([]qovery.ServicePortRequestPortsInner, 0, len(app.Ports))
	for _, port := range app.Ports {
		p, err := port.toCreateRequest()
		if err != nil {
			return nil, err
		}
		ports = append(ports, *p)
	}

	var buildMode *qovery.BuildModeEnum
	if !app.BuildMode.IsNull() && !app.BuildMode.IsUnknown() {
		bm, err := qovery.NewBuildModeEnumFromValue(ToString(app.BuildMode))
		if err != nil {
			return nil, err
		}
		buildMode = bm
	}

	deploymentRestrictions, err := app.DeploymentRestrictionList(nil)
	if err != nil {
		return nil, err
	}

	annotations := make([]string, 0, len(app.AnnotationsGroupIds.Elements()))
	for _, id := range app.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotations = append(annotations, id.ValueString())
	}

	annotationsGroups, err := qoveryapi.NewQoveryServiceAnnotationsGroupRequestFromDomain(annotations)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labels := make([]string, 0, len(app.LabelsGroupIds.Elements()))
	for _, id := range app.LabelsGroupIds.Elements() {
		id := id.(types.String)
		labels = append(labels, id.ValueString())
	}

	labelsGroups, err := qoveryapi.NewQoveryServiceLabelsGroupRequestFromDomain(labels)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	return &client.ApplicationCreateParams{
		ApplicationRequest: qovery.ApplicationRequest{
			Name:                   ToString(app.Name),
			IconUri:                ToStringPointer(app.IconUri),
			BuildMode:              buildMode,
			DockerfilePath:         ToNullableString(app.DockerfilePath),
			Cpu:                    ToInt32Pointer(app.CPU),
			Memory:                 ToInt32Pointer(app.Memory),
			MinRunningInstances:    ToInt32Pointer(app.MinRunningInstances),
			MaxRunningInstances:    ToInt32Pointer(app.MaxRunningInstances),
			AutoPreview:            ToBoolPointer(app.AutoPreview),
			GitRepository:          app.GitRepository.toCreateRequest(),
			Storage:                storage,
			Ports:                  ports,
			Entrypoint:             ToStringPointer(app.Entrypoint),
			Arguments:              ToStringArray(app.Arguments),
			Healthchecks:           app.Healthchecks.toHealthchecksRequest(),
			AutoDeploy:             *qovery.NewNullableBool(ToBoolPointer(app.AutoDeploy)),
			AnnotationsGroups:      annotationsGroups,
			LabelsGroups:           labelsGroups,
			DockerTargetBuildStage: ToNullableString(app.DockerTargetBuildStage),
		},
		EnvironmentVariablesDiff:         app.EnvironmentVariableList().diff(nil),
		EnvironmentVariableAliasesDiff:   app.EnvironmentVariableAliasList().diff(nil),
		EnvironmentVariableOverridesDiff: app.EnvironmentVariableOverrideList().diff(nil),
		SecretsDiff:                      app.SecretList().diff(nil),
		SecretAliasesDiff:                app.SecretAliasList().diff(nil),
		SecretOverridesDiff:              app.SecretOverrideList().diff(nil),
		CustomDomainsDiff:                app.CustomDomainsList().diff(nil),
		DeploymentRestrictionsDiff:       *deploymentRestrictions,
		ApplicationDeploymentStageID:     ToString(app.DeploymentStageId),
		AdvancedSettingsJson:             ToString(app.AdvancedSettingsJson),
	}, nil

}

func (app Application) toUpdateApplicationRequest(state Application) (*client.ApplicationUpdateParams, error) {
	// Create a hashmap containing current terraform state ApplicationStorage by MountPoint
	// MountPoint is unique with an application
	stateApplicationStoragesByMountPoint := make(map[string]ApplicationStorage)
	for _, existingStorage := range state.Storage {
		stateApplicationStoragesByMountPoint[existingStorage.MountPoint.String()] = existingStorage
	}

	applicationStorageRequest := make([]qovery.ServiceStorageRequestStorageInner, 0, len(app.Storage))
	for _, storage := range app.Storage {
		// The storage id can be:
		// - nil if a new storage is declared in the application resource
		// - set if it already exists
		var storageId string
		value, exists := stateApplicationStoragesByMountPoint[storage.MountPoint.String()]
		if exists {
			storageId = ToString(value.Id)
		}

		s, err := storage.toUpdateRequest(storageId)
		if err != nil {
			return nil, err
		}
		applicationStorageRequest = append(applicationStorageRequest, *s)
	}

	ports := make([]qovery.ServicePort, 0, len(app.Ports))
	for _, port := range app.Ports {
		p, err := port.toUpdateRequest()
		if err != nil {
			return nil, err
		}
		ports = append(ports, *p)
	}

	var buildMode *qovery.BuildModeEnum
	if !app.BuildMode.IsNull() && !app.BuildMode.IsUnknown() {
		bm, err := qovery.NewBuildModeEnumFromValue(ToString(app.BuildMode))
		if err != nil {
			return nil, err
		}
		buildMode = bm
	}

	deploymentRestrictions, err := app.DeploymentRestrictionList(&state.DeploymentRestrictions)
	if err != nil {
		return nil, err
	}

	annotations := make([]string, 0, len(app.AnnotationsGroupIds.Elements()))
	for _, id := range app.AnnotationsGroupIds.Elements() {
		id := id.(types.String)
		annotations = append(annotations, id.ValueString())
	}

	annotationsGroups, err := qoveryapi.NewQoveryServiceAnnotationsGroupRequestFromDomain(annotations)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	labels := make([]string, 0, len(app.LabelsGroupIds.Elements()))
	for _, id := range app.LabelsGroupIds.Elements() {
		id := id.(types.String)
		labels = append(labels, id.ValueString())
	}

	labelsGroups, err := qoveryapi.NewQoveryServiceLabelsGroupRequestFromDomain(labels)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	applicationEditRequest := qovery.ApplicationEditRequest{
		Name:                   ToStringPointer(app.Name),
		IconUri:                ToStringPointer(app.IconUri),
		BuildMode:              buildMode,
		DockerfilePath:         ToNullableString(app.DockerfilePath),
		Cpu:                    ToInt32Pointer(app.CPU),
		Memory:                 ToInt32Pointer(app.Memory),
		MinRunningInstances:    ToInt32Pointer(app.MinRunningInstances),
		MaxRunningInstances:    ToInt32Pointer(app.MaxRunningInstances),
		AutoPreview:            ToBoolPointer(app.AutoPreview),
		GitRepository:          app.GitRepository.toUpdateRequest(),
		Storage:                applicationStorageRequest,
		Ports:                  ports,
		Entrypoint:             ToStringPointer(app.Entrypoint),
		Arguments:              ToStringArray(app.Arguments),
		Healthchecks:           app.Healthchecks.toHealthchecksRequest(),
		AutoDeploy:             *qovery.NewNullableBool(ToBoolPointer(app.AutoDeploy)),
		AnnotationsGroups:      annotationsGroups,
		LabelsGroups:           labelsGroups,
		DockerTargetBuildStage: ToNullableString(app.DockerTargetBuildStage),
	}
	return &client.ApplicationUpdateParams{
		ApplicationEditRequest:           applicationEditRequest,
		EnvironmentVariablesDiff:         app.EnvironmentVariableList().diff(state.EnvironmentVariableList()),
		EnvironmentVariableAliasesDiff:   app.EnvironmentVariableAliasList().diff(state.EnvironmentVariableAliasList()),
		EnvironmentVariableOverridesDiff: app.EnvironmentVariableOverrideList().diff(state.EnvironmentVariableOverrideList()),
		SecretsDiff:                      app.SecretList().diff(state.SecretList()),
		SecretAliasesDiff:                app.SecretAliasList().diff(state.SecretAliasList()),
		SecretOverridesDiff:              app.SecretOverrideList().diff(state.SecretOverrideList()),
		CustomDomainsDiff:                app.CustomDomainsList().diff(state.CustomDomainsList()),
		DeploymentRestrictionsDiff:       *deploymentRestrictions,
		ApplicationDeploymentStageID:     ToString(app.DeploymentStageId),
		AdvancedSettingsJson:             ToString(app.AdvancedSettingsJson),
	}, nil

}

func convertResponseToApplication(ctx context.Context, state Application, app *client.ApplicationResponse) Application {
	healthchecks := convertHealthchecksResponseToDomain(app.ApplicationResponse.Healthchecks)
	return Application{
		Id:                           FromString(app.ApplicationResponse.Id),
		EnvironmentId:                FromString(app.ApplicationResponse.Environment.Id),
		Name:                         FromString(app.ApplicationResponse.Name),
		IconUri:                      FromString(app.ApplicationResponse.IconUri),
		BuildMode:                    fromClientEnumPointer(app.ApplicationResponse.BuildMode),
		DockerfilePath:               FromNullableString(app.ApplicationResponse.DockerfilePath),
		CPU:                          FromInt32Pointer(app.ApplicationResponse.Cpu),
		Memory:                       FromInt32Pointer(app.ApplicationResponse.Memory),
		MinRunningInstances:          FromInt32Pointer(app.ApplicationResponse.MinRunningInstances),
		MaxRunningInstances:          FromInt32Pointer(app.ApplicationResponse.MaxRunningInstances),
		AutoPreview:                  FromBoolPointer(app.ApplicationResponse.AutoPreview),
		GitRepository:                convertResponseToApplicationGitRepository(app.ApplicationResponse.GitRepository),
		Storage:                      convertResponseToApplicationStorage(state.Storage, app.ApplicationResponse.Storage),
		Ports:                        convertResponseToApplicationPorts(state.Ports, app.ApplicationResponse.Ports),
		BuiltInEnvironmentVariables:  fromEnvironmentVariableList(app.ApplicationEnvironmentVariables, qovery.APIVARIABLESCOPEENUM_BUILT_IN, "BUILT_IN").toTerraformSet(ctx),
		EnvironmentVariables:         fromEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariables, app.ApplicationEnvironmentVariables, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE").toTerraformSet(ctx),
		EnvironmentVariableAliases:   fromEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableAliases, app.ApplicationEnvironmentVariableAliases, qovery.APIVARIABLESCOPEENUM_APPLICATION, "ALIAS").toTerraformSet(ctx),
		EnvironmentVariableOverrides: fromEnvironmentVariableListWithNullableInitialState(ctx, state.EnvironmentVariableOverrides, app.ApplicationEnvironmentVariableOverrides, qovery.APIVARIABLESCOPEENUM_APPLICATION, "OVERRIDE").toTerraformSet(ctx),
		Secrets:                      fromSecretList(state.Secrets, app.ApplicationSecrets, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE").toTerraformSet(ctx),
		SecretVariableAliases:        fromSecretList(state.SecretVariableAliases, app.ApplicationSecretAliases, qovery.APIVARIABLESCOPEENUM_APPLICATION, "ALIAS").toTerraformSet(ctx),
		SecretVariableOverrides:      fromSecretList(state.SecretVariableOverrides, app.ApplicationSecretOverrides, qovery.APIVARIABLESCOPEENUM_APPLICATION, "OVERRIDE").toTerraformSet(ctx),
		CustomDomains:                fromCustomDomainList(state.CustomDomains, app.ApplicationCustomDomains).toTerraformSet(ctx),
		InternalHost:                 FromString(app.ApplicationInternalHost),
		ExternalHost:                 FromStringPointer(app.ApplicationExternalHost),
		Entrypoint:                   FromStringPointer(app.ApplicationResponse.Entrypoint),
		Arguments:                    FromStringArray(app.ApplicationResponse.Arguments),
		DeploymentStageId:            FromString(app.ApplicationDeploymentStageID),
		Healthchecks:                 &healthchecks,
		AdvancedSettingsJson:         FromString(app.AdvancedSettingsJson),
		AutoDeploy:                   FromBoolPointer(app.ApplicationResponse.AutoDeploy),
		DeploymentRestrictions:       FromDeploymentRestrictionList(state.DeploymentRestrictions, app.ApplicationDeploymentRestrictions),
		AnnotationsGroupIds:          fromAnnotationsGroupResponseList(ctx, state.AnnotationsGroupIds, app.ApplicationResponse.AnnotationsGroups),
		LabelsGroupIds:               fromLabelsGroupResponseList(ctx, state.LabelsGroupIds, app.ApplicationResponse.LabelsGroups),
		DockerTargetBuildStage:       FromNullableString(app.ApplicationResponse.DockerTargetBuildStage),
	}
}

type ApplicationGitRepository struct {
	URL        types.String `tfsdk:"url"`
	RootPath   types.String `tfsdk:"root_path"`
	Branch     types.String `tfsdk:"branch"`
	GitTokenId types.String `tfsdk:"git_token_id"`
}

func (repo ApplicationGitRepository) toCreateRequest() qovery.ApplicationGitRepositoryRequest {
	return qovery.ApplicationGitRepositoryRequest{
		Url:        ToString(repo.URL),
		RootPath:   ToStringPointer(repo.RootPath),
		Branch:     ToStringPointer(repo.Branch),
		GitTokenId: *qovery.NewNullableString(ToStringPointer(repo.GitTokenId)),
	}
}

func (repo ApplicationGitRepository) toUpdateRequest() *qovery.ApplicationGitRepositoryRequest {
	req := repo.toCreateRequest()
	return &req
}

func convertResponseToApplicationGitRepository(gitRepository *qovery.ApplicationGitRepository) *ApplicationGitRepository {
	if gitRepository == nil {
		return nil
	}

	return &ApplicationGitRepository{
		URL:        FromString(gitRepository.Url),
		RootPath:   FromStringPointer(gitRepository.RootPath),
		Branch:     FromStringPointer(gitRepository.Branch),
		GitTokenId: FromStringPointer(gitRepository.GitTokenId.Get()),
	}
}

type ApplicationStorage struct {
	Id         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	Size       types.Int64  `tfsdk:"size"`
	MountPoint types.String `tfsdk:"mount_point"`
}

func (store ApplicationStorage) toCreateRequest() (*qovery.ServiceStorageRequestStorageInner, error) {
	storageType, err := qovery.NewStorageTypeEnumFromValue(ToString(store.Type))
	if err != nil {
		return nil, err
	}

	return &qovery.ServiceStorageRequestStorageInner{
		Type:       *storageType,
		Size:       ToInt32(store.Size),
		MountPoint: ToString(store.MountPoint),
	}, nil
}

func (store ApplicationStorage) toUpdateRequest(id string) (*qovery.ServiceStorageRequestStorageInner, error) {
	storageType, err := qovery.NewStorageTypeEnumFromValue(ToString(store.Type))
	if err != nil {
		return nil, err
	}

	return &qovery.ServiceStorageRequestStorageInner{
		Id:         &id,
		Type:       *storageType,
		Size:       ToInt32(store.Size),
		MountPoint: ToString(store.MountPoint),
	}, nil
}

func convertResponseToApplicationStorage(initialState []ApplicationStorage, storage []qovery.ServiceStorageStorageInner) []ApplicationStorage {
	list := make([]ApplicationStorage, 0, len(storage))
	for _, s := range storage {
		list = append(list, ApplicationStorage{
			Id:         FromString(s.Id),
			Type:       fromClientEnum(s.Type),
			Size:       FromInt32(s.Size),
			MountPoint: FromString(s.MountPoint),
		})
	}

	if len(storage) == 0 && initialState == nil {
		return nil
	}

	return list
}

type ApplicationPort struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	InternalPort       types.Int64  `tfsdk:"internal_port"`
	ExternalPort       types.Int64  `tfsdk:"external_port"`
	PubliclyAccessible types.Bool   `tfsdk:"publicly_accessible"`
	Protocol           types.String `tfsdk:"protocol"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
}

func (port ApplicationPort) toCreateRequest() (*qovery.ServicePortRequestPortsInner, error) {
	protocol, err := qovery.NewPortProtocolEnumFromValue(ToString(port.Protocol))
	if err != nil {
		return nil, err
	}

	return &qovery.ServicePortRequestPortsInner{
		Name:               ToStringPointer(port.Name),
		InternalPort:       ToInt32(port.InternalPort),
		ExternalPort:       ToInt32Pointer(port.ExternalPort),
		Protocol:           protocol,
		PubliclyAccessible: ToBool(port.PubliclyAccessible),
		IsDefault:          ToBoolPointer(port.IsDefault),
	}, nil
}

func (port ApplicationPort) toUpdateRequest() (*qovery.ServicePort, error) {
	protocol, err := qovery.NewPortProtocolEnumFromValue(ToString(port.Protocol))
	if err != nil {
		return nil, err
	}

	return &qovery.ServicePort{
		Id:                 ToString(port.Id),
		Name:               ToStringPointer(port.Name),
		InternalPort:       ToInt32(port.InternalPort),
		ExternalPort:       ToInt32Pointer(port.ExternalPort),
		Protocol:           *protocol,
		PubliclyAccessible: ToBool(port.PubliclyAccessible),
		IsDefault:          ToBoolPointer(port.IsDefault),
	}, nil
}

func convertResponseToApplicationPorts(initialState []ApplicationPort, ports []qovery.ServicePort) []ApplicationPort {
	// Try to sort ports as similarly as possible to the initialState.
	portsByName := make(map[string]qovery.ServicePort, len(ports))
	for _, p := range ports {
		portsByName[*p.Name] = p
	}

	list := make([]ApplicationPort, 0, len(ports))
	for _, state := range initialState {
		if p, ok := portsByName[state.Name.ValueString()]; ok {
			list = append(list, ApplicationPort{
				Id:                 FromString(p.Id),
				Name:               FromStringPointer(p.Name),
				InternalPort:       FromInt32(p.InternalPort),
				ExternalPort:       FromInt32Pointer(p.ExternalPort),
				Protocol:           fromClientEnum(p.Protocol),
				PubliclyAccessible: FromBool(p.PubliclyAccessible),
				IsDefault:          FromBoolPointer(p.IsDefault),
			})
			delete(portsByName, state.Name.ValueString())
		}
	}

	for _, p := range portsByName {
		list = append(list, ApplicationPort{
			Id:                 FromString(p.Id),
			Name:               FromStringPointer(p.Name),
			InternalPort:       FromInt32(p.InternalPort),
			ExternalPort:       FromInt32Pointer(p.ExternalPort),
			Protocol:           fromClientEnum(p.Protocol),
			PubliclyAccessible: FromBool(p.PubliclyAccessible),
			IsDefault:          FromBoolPointer(p.IsDefault),
		})
	}

	if len(ports) == 0 && initialState == nil {
		return nil
	}
	return list
}
