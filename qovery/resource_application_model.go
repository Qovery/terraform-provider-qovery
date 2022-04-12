package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

type Application struct {
	Id                   types.String              `tfsdk:"id"`
	EnvironmentId        types.String              `tfsdk:"environment_id"`
	Name                 types.String              `tfsdk:"name"`
	Description          types.String              `tfsdk:"description"`
	GitRepository        *ApplicationGitRepository `tfsdk:"git_repository"`
	BuildMode            types.String              `tfsdk:"build_mode"`
	DockerfilePath       types.String              `tfsdk:"dockerfile_path"`
	BuildpackLanguage    types.String              `tfsdk:"buildpack_language"`
	CPU                  types.Int64               `tfsdk:"cpu"`
	Memory               types.Int64               `tfsdk:"memory"`
	MinRunningInstances  types.Int64               `tfsdk:"min_running_instances"`
	MaxRunningInstances  types.Int64               `tfsdk:"max_running_instances"`
	AutoPreview          types.Bool                `tfsdk:"auto_preview"`
	Storage              []ApplicationStorage      `tfsdk:"storage"`
	Ports                []ApplicationPort         `tfsdk:"ports"`
	EnvironmentVariables []EnvironmentVariable     `tfsdk:"environment_variables"`
	State                types.String              `tfsdk:"state"`
}

func (app Application) toCreateApplicationRequest() client.ApplicationCreateParams {
	storage := make([]qovery.ApplicationStorageRequestStorage, 0, len(app.Storage))
	for _, store := range app.Storage {
		storage = append(storage, store.toCreateRequest())
	}

	ports := make([]qovery.ApplicationPortRequestPorts, 0, len(app.Ports))
	for _, port := range app.Ports {
		ports = append(ports, port.toCreateRequest())
	}

	return client.ApplicationCreateParams{
		ApplicationRequest: qovery.ApplicationRequest{
			Name:                toString(app.Name),
			Description:         toNullableString(app.Description),
			BuildMode:           toStringPointer(app.BuildMode),
			DockerfilePath:      toNullableString(app.DockerfilePath),
			BuildpackLanguage:   toNullableNullableBuildPackLanguageEnum(app.BuildpackLanguage),
			Cpu:                 toInt32Pointer(app.CPU),
			Memory:              toInt32Pointer(app.Memory),
			MinRunningInstances: toInt32Pointer(app.MinRunningInstances),
			MaxRunningInstances: toInt32Pointer(app.MinRunningInstances),
			AutoPreview:         toBoolPointer(app.AutoPreview),
			GitRepository:       app.GitRepository.toCreateRequest(),
			Storage:             storage,
			Ports:               ports,
		},
		DesiredState:             app.State.Value,
		EnvironmentVariablesDiff: diffEnvironmentVariables([]EnvironmentVariable{}, app.EnvironmentVariables),
	}

}

func (app Application) toUpdateApplicationRequest(state Application) client.ApplicationUpdateParams {
	storage := make([]qovery.ApplicationStorageResponseStorage, 0, len(app.Storage))
	for _, store := range app.Storage {
		storage = append(storage, store.toUpdateRequest())
	}

	ports := make([]qovery.ApplicationPortResponsePorts, 0, len(app.Ports))
	for _, port := range app.Ports {
		ports = append(ports, port.toUpdateRequest())
	}

	applicationEditRequest := qovery.ApplicationEditRequest{
		Name:                toStringPointer(app.Name),
		Description:         toStringPointer(app.Description),
		BuildMode:           toStringPointer(app.BuildMode),
		DockerfilePath:      toStringPointer(app.DockerfilePath),
		BuildpackLanguage:   toNullableNullableBuildPackLanguageEnum(app.BuildpackLanguage),
		Cpu:                 toInt32Pointer(app.CPU),
		Memory:              toInt32Pointer(app.Memory),
		MinRunningInstances: toInt32Pointer(app.MinRunningInstances),
		MaxRunningInstances: toInt32Pointer(app.MinRunningInstances),
		AutoPreview:         toBoolPointer(app.AutoPreview),
		GitRepository:       app.GitRepository.toUpdateRequest(),
		Storage:             storage,
		Ports:               ports,
	}
	return client.ApplicationUpdateParams{
		ApplicationEditRequest:   applicationEditRequest,
		EnvironmentVariablesDiff: diffEnvironmentVariables(state.EnvironmentVariables, app.EnvironmentVariables),
		DesiredState:             app.State.Value,
	}

}

func convertResponseToApplication(app *client.ApplicationResponse) Application {
	return Application{
		Id:                   fromString(app.ApplicationResponse.Id),
		EnvironmentId:        fromString(app.ApplicationResponse.Environment.Id),
		Name:                 fromStringPointer(app.ApplicationResponse.Name),
		Description:          fromNullableString(app.ApplicationResponse.Description),
		BuildMode:            fromStringPointer(app.ApplicationResponse.BuildMode),
		DockerfilePath:       fromNullableString(app.ApplicationResponse.DockerfilePath),
		BuildpackLanguage:    fromNullableNullableBuildPackLanguageEnum(app.ApplicationResponse.BuildpackLanguage),
		CPU:                  fromInt32Pointer(app.ApplicationResponse.Cpu),
		Memory:               fromInt32Pointer(app.ApplicationResponse.Memory),
		MinRunningInstances:  fromInt32Pointer(app.ApplicationResponse.MinRunningInstances),
		MaxRunningInstances:  fromInt32Pointer(app.ApplicationResponse.MaxRunningInstances),
		AutoPreview:          fromBoolPointer(app.ApplicationResponse.AutoPreview),
		GitRepository:        convertResponseToApplicationGitRepository(app.ApplicationResponse.GitRepository),
		Storage:              convertResponseToApplicationStorage(app.ApplicationResponse.Storage),
		Ports:                convertResponseToApplicationPorts(app.ApplicationResponse.Ports),
		State:                fromString(app.ApplicationStatus.State),
		EnvironmentVariables: convertResponseToEnvironmentVariables(app.ApplicationEnvironmentVariables),
	}
}

type ApplicationGitRepository struct {
	URL      types.String `tfsdk:"url"`
	RootPath types.String `tfsdk:"root_path"`
	Branch   types.String `tfsdk:"branch"`
}

func (repo ApplicationGitRepository) toCreateRequest() qovery.ApplicationGitRepositoryRequest {
	return qovery.ApplicationGitRepositoryRequest{
		Url:      toString(repo.URL),
		RootPath: toStringPointer(repo.RootPath),
		Branch:   toStringPointer(repo.Branch),
	}
}

func (repo ApplicationGitRepository) toUpdateRequest() *qovery.ApplicationGitRepositoryRequest {
	req := repo.toCreateRequest()
	return &req
}

func convertResponseToApplicationGitRepository(gitRepository *qovery.ApplicationGitRepositoryResponse) *ApplicationGitRepository {
	if gitRepository == nil {
		return nil
	}

	return &ApplicationGitRepository{
		URL:      fromStringPointer(gitRepository.Url),
		RootPath: fromStringPointer(gitRepository.RootPath),
		Branch:   fromStringPointer(gitRepository.Branch),
	}
}

type ApplicationStorage struct {
	Id         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	Size       types.Int64  `tfsdk:"size"`
	MountPoint types.String `tfsdk:"mount_point"`
}

func (store ApplicationStorage) toCreateRequest() qovery.ApplicationStorageRequestStorage {
	return qovery.ApplicationStorageRequestStorage{
		Type:       toString(store.Type),
		Size:       toInt32(store.Size),
		MountPoint: toString(store.MountPoint),
	}
}

func (store ApplicationStorage) toUpdateRequest() qovery.ApplicationStorageResponseStorage {
	return qovery.ApplicationStorageResponseStorage{
		Id:         toStringPointer(store.Id),
		Type:       toString(store.Type),
		Size:       toInt32(store.Size),
		MountPoint: toString(store.MountPoint),
	}
}

func convertResponseToApplicationStorage(storage []qovery.ApplicationStorageResponseStorage) []ApplicationStorage {
	if len(storage) == 0 {
		return nil
	}

	list := make([]ApplicationStorage, 0, len(storage))
	for _, s := range storage {
		list = append(list, ApplicationStorage{
			Id:         fromStringPointer(s.Id),
			Type:       fromString(s.Type),
			Size:       fromInt32(s.Size),
			MountPoint: fromString(s.MountPoint),
		})
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
}

func (port ApplicationPort) toCreateRequest() qovery.ApplicationPortRequestPorts {
	return qovery.ApplicationPortRequestPorts{
		Name:               toNullableString(port.Name),
		InternalPort:       toInt32(port.InternalPort),
		ExternalPort:       toInt32Pointer(port.ExternalPort),
		Protocol:           toStringPointer(port.Protocol),
		PubliclyAccessible: toBool(port.PubliclyAccessible),
	}
}

func (port ApplicationPort) toUpdateRequest() qovery.ApplicationPortResponsePorts {
	return qovery.ApplicationPortResponsePorts{
		Id:                 toStringPointer(port.Id),
		Name:               toNullableString(port.Name),
		InternalPort:       toInt32(port.InternalPort),
		ExternalPort:       toInt32Pointer(port.ExternalPort),
		Protocol:           toStringPointer(port.Protocol),
		PubliclyAccessible: toBool(port.PubliclyAccessible),
	}
}

func convertResponseToApplicationPorts(ports []qovery.ApplicationPortResponsePorts) []ApplicationPort {
	if len(ports) == 0 {
		return nil
	}

	list := make([]ApplicationPort, 0, len(ports))
	for _, p := range ports {
		list = append(list, ApplicationPort{
			Id:                 fromStringPointer(p.Id),
			Name:               fromNullableString(p.Name),
			InternalPort:       fromInt32(p.InternalPort),
			ExternalPort:       fromInt32Pointer(p.ExternalPort),
			Protocol:           fromStringPointer(p.Protocol),
			PubliclyAccessible: fromBool(p.PubliclyAccessible),
		})
	}
	return list
}
