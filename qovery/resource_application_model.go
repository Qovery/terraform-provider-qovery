package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Application struct {
	Id                           types.String              `tfsdk:"id"`
	EnvironmentId                types.String              `tfsdk:"environment_id"`
	Name                         types.String              `tfsdk:"name"`
	GitRepository                *ApplicationGitRepository `tfsdk:"git_repository"`
	BuildMode                    types.String              `tfsdk:"build_mode"`
	DockerfilePath               types.String              `tfsdk:"dockerfile_path"`
	BuildpackLanguage            types.String              `tfsdk:"buildpack_language"`
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
	if !app.BuildMode.Null && !app.BuildMode.Unknown {
		bm, err := qovery.NewBuildModeEnumFromValue(ToString(app.BuildMode))
		if err != nil {
			return nil, err
		}
		buildMode = bm
	}

	return &client.ApplicationCreateParams{
		ApplicationRequest: qovery.ApplicationRequest{
			Name:                ToString(app.Name),
			BuildMode:           buildMode,
			DockerfilePath:      ToNullableString(app.DockerfilePath),
			BuildpackLanguage:   ToNullableNullableBuildPackLanguageEnum(app.BuildpackLanguage),
			Cpu:                 ToInt32Pointer(app.CPU),
			Memory:              ToInt32Pointer(app.Memory),
			MinRunningInstances: ToInt32Pointer(app.MinRunningInstances),
			MaxRunningInstances: ToInt32Pointer(app.MaxRunningInstances),
			AutoPreview:         ToBoolPointer(app.AutoPreview),
			GitRepository:       app.GitRepository.toCreateRequest(),
			Storage:             storage,
			Ports:               ports,
			Entrypoint:          ToStringPointer(app.Entrypoint),
			Arguments:           ToStringArray(app.Arguments),
			Healthchecks:        app.Healthchecks.toHealthchecksRequest(),
		},
		EnvironmentVariablesDiff:         app.EnvironmentVariableList().diff(nil),
		EnvironmentVariableAliasesDiff:   app.EnvironmentVariableAliasList().diff(nil),
		EnvironmentVariableOverridesDiff: app.EnvironmentVariableOverrideList().diff(nil),
		SecretsDiff:                      app.SecretList().diff(nil),
		SecretAliasesDiff:                app.SecretAliasList().diff(nil),
		SecretOverridesDiff:              app.SecretOverrideList().diff(nil),
		CustomDomainsDiff:                app.CustomDomainsList().diff(nil),
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
	if !app.BuildMode.Null && !app.BuildMode.Unknown {
		bm, err := qovery.NewBuildModeEnumFromValue(ToString(app.BuildMode))
		if err != nil {
			return nil, err
		}
		buildMode = bm
	}

	applicationEditRequest := qovery.ApplicationEditRequest{
		Name:                ToStringPointer(app.Name),
		BuildMode:           buildMode,
		DockerfilePath:      ToStringPointer(app.DockerfilePath),
		BuildpackLanguage:   ToNullableNullableBuildPackLanguageEnum(app.BuildpackLanguage),
		Cpu:                 ToInt32Pointer(app.CPU),
		Memory:              ToInt32Pointer(app.Memory),
		MinRunningInstances: ToInt32Pointer(app.MinRunningInstances),
		MaxRunningInstances: ToInt32Pointer(app.MaxRunningInstances),
		AutoPreview:         ToBoolPointer(app.AutoPreview),
		GitRepository:       app.GitRepository.toUpdateRequest(),
		Storage:             applicationStorageRequest,
		Ports:               ports,
		Entrypoint:          ToStringPointer(app.Entrypoint),
		Arguments:           ToStringArray(app.Arguments),
		Healthchecks:        app.Healthchecks.toHealthchecksRequest(),
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
		ApplicationDeploymentStageID:     ToString(app.DeploymentStageId),
		AdvancedSettingsJson:             ToString(app.AdvancedSettingsJson),
	}, nil

}

func convertResponseToApplication(state Application, app *client.ApplicationResponse) Application {
	return Application{
		Id:                           FromString(app.ApplicationResponse.Id),
		EnvironmentId:                FromString(app.ApplicationResponse.Environment.Id),
		Name:                         FromStringPointer(app.ApplicationResponse.Name),
		BuildMode:                    fromClientEnumPointer(app.ApplicationResponse.BuildMode),
		DockerfilePath:               FromNullableString(app.ApplicationResponse.DockerfilePath),
		BuildpackLanguage:            FromNullableNullableBuildPackLanguageEnum(app.ApplicationResponse.BuildpackLanguage),
		CPU:                          FromInt32Pointer(app.ApplicationResponse.Cpu),
		Memory:                       FromInt32Pointer(app.ApplicationResponse.Memory),
		MinRunningInstances:          FromInt32Pointer(app.ApplicationResponse.MinRunningInstances),
		MaxRunningInstances:          FromInt32Pointer(app.ApplicationResponse.MaxRunningInstances),
		AutoPreview:                  FromBoolPointer(app.ApplicationResponse.AutoPreview),
		GitRepository:                convertResponseToApplicationGitRepository(app.ApplicationResponse.GitRepository),
		Storage:                      convertResponseToApplicationStorage(state.Storage, app.ApplicationResponse.Storage),
		Ports:                        convertResponseToApplicationPorts(state.Ports, app.ApplicationResponse.Ports),
		BuiltInEnvironmentVariables:  fromEnvironmentVariableList(app.ApplicationEnvironmentVariables, qovery.APIVARIABLESCOPEENUM_BUILT_IN, "BUILT_IN").toTerraformSet(),
		EnvironmentVariables:         fromEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariables, app.ApplicationEnvironmentVariables, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE").toTerraformSet(),
		EnvironmentVariableAliases:   fromEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableAliases, app.ApplicationEnvironmentVariableAliases, qovery.APIVARIABLESCOPEENUM_APPLICATION, "ALIAS").toTerraformSet(),
		EnvironmentVariableOverrides: fromEnvironmentVariableListWithNullableInitialState(state.EnvironmentVariableOverrides, app.ApplicationEnvironmentVariableOverrides, qovery.APIVARIABLESCOPEENUM_APPLICATION, "OVERRIDE").toTerraformSet(),
		Secrets:                      fromSecretList(state.Secrets, app.ApplicationSecrets, qovery.APIVARIABLESCOPEENUM_APPLICATION, "VALUE").toTerraformSet(),
		SecretVariableAliases:        fromSecretList(state.SecretVariableAliases, app.ApplicationSecretAliases, qovery.APIVARIABLESCOPEENUM_APPLICATION, "ALIAS").toTerraformSet(),
		SecretVariableOverrides:      fromSecretList(state.SecretVariableOverrides, app.ApplicationSecretOverrides, qovery.APIVARIABLESCOPEENUM_APPLICATION, "OVERRIDE").toTerraformSet(),
		CustomDomains:                fromCustomDomainList(state.CustomDomains, app.ApplicationCustomDomains).toTerraformSet(),
		InternalHost:                 FromString(app.ApplicationInternalHost),
		ExternalHost:                 FromStringPointer(app.ApplicationExternalHost),
		Entrypoint:                   FromStringPointer(app.ApplicationResponse.Entrypoint),
		Arguments:                    FromStringArray(app.ApplicationResponse.Arguments),
		DeploymentStageId:            FromString(app.ApplicationDeploymentStageID),
		Healthchecks:                 convertHealthchecksResponseToDomain(app.ApplicationResponse.Healthchecks),
		AdvancedSettingsJson:         FromString(app.AdvancedSettingsJson),
	}
}

type ApplicationGitRepository struct {
	URL      types.String `tfsdk:"url"`
	RootPath types.String `tfsdk:"root_path"`
	Branch   types.String `tfsdk:"branch"`
}

func (repo ApplicationGitRepository) toCreateRequest() qovery.ApplicationGitRepositoryRequest {
	return qovery.ApplicationGitRepositoryRequest{
		Url:      ToString(repo.URL),
		RootPath: ToStringPointer(repo.RootPath),
		Branch:   ToStringPointer(repo.Branch),
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
		URL:      FromStringPointer(gitRepository.Url),
		RootPath: FromStringPointer(gitRepository.RootPath),
		Branch:   FromStringPointer(gitRepository.Branch),
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
	list := make([]ApplicationPort, 0, len(ports))
	for _, p := range ports {
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
