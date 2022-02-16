package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Application struct {
	Id                  types.String              `tfsdk:"id"`
	EnvironmentId       types.String              `tfsdk:"environment_id"`
	Name                types.String              `tfsdk:"name"`
	Description         types.String              `tfsdk:"description"`
	GitRepository       *ApplicationGitRepository `tfsdk:"git_repository"`
	BuildMode           types.String              `tfsdk:"build_mode"`
	DockerfilePath      types.String              `tfsdk:"dockerfile_path"`
	BuildpackLanguage   types.String              `tfsdk:"buildpack_language"`
	CPU                 types.Int64               `tfsdk:"cpu"`
	Memory              types.Int64               `tfsdk:"memory"`
	MinRunningInstances types.Int64               `tfsdk:"min_running_instances"`
	MaxRunningInstances types.Int64               `tfsdk:"max_running_instances"`
	AutoPreview         types.Bool                `tfsdk:"auto_preview"`
	Storage             []ApplicationStorage      `tfsdk:"storage"`
	Ports               []ApplicationPort         `tfsdk:"ports"`
	State               types.String              `tfsdk:"state"`
}

func (app Application) toCreateApplicationRequest() qovery.ApplicationRequest {
	storage := make([]qovery.ApplicationStorageRequestStorage, 0, len(app.Storage))
	for _, store := range app.Storage {
		storage = append(storage, store.toCreateRequest())
	}

	ports := make([]qovery.ApplicationPortRequestPorts, 0, len(app.Ports))
	for _, port := range app.Ports {
		ports = append(ports, port.toCreateRequest())
	}

	return qovery.ApplicationRequest{
		Name:                toString(app.Name),
		Description:         toStringPointer(app.Description),
		BuildMode:           toStringPointer(app.BuildMode),
		DockerfilePath:      toStringPointer(app.DockerfilePath),
		BuildpackLanguage:   toStringPointer(app.BuildpackLanguage),
		Cpu:                 toInt32Pointer(app.CPU),
		Memory:              toInt32Pointer(app.Memory),
		MinRunningInstances: toInt32Pointer(app.MinRunningInstances),
		MaxRunningInstances: toInt32Pointer(app.MinRunningInstances),
		AutoPreview:         toBoolPointer(app.AutoPreview),
		GitRepository:       app.GitRepository.toCreateRequest(),
		Storage:             storage,
		Ports:               ports,
	}
}

func (app Application) toUpdateApplicationRequest() qovery.ApplicationEditRequest {
	storage := make([]qovery.ApplicationStorageResponseStorage, 0, len(app.Storage))
	for _, store := range app.Storage {
		storage = append(storage, store.toUpdateRequest())
	}

	ports := make([]qovery.ApplicationPortResponsePorts, 0, len(app.Ports))
	for _, port := range app.Ports {
		ports = append(ports, port.toUpdateRequest())
	}

	return qovery.ApplicationEditRequest{
		Name:                toStringPointer(app.Name),
		Description:         toStringPointer(app.Description),
		BuildMode:           toStringPointer(app.BuildMode),
		DockerfilePath:      toStringPointer(app.DockerfilePath),
		BuildpackLanguage:   toStringPointer(app.BuildpackLanguage),
		Cpu:                 toInt32Pointer(app.CPU),
		Memory:              toInt32Pointer(app.Memory),
		MinRunningInstances: toInt32Pointer(app.MinRunningInstances),
		MaxRunningInstances: toInt32Pointer(app.MinRunningInstances),
		AutoPreview:         toBoolPointer(app.AutoPreview),
		GitRepository:       app.GitRepository.toUpdateRequest(),
		Storage:             storage,
		Ports:               ports,
	}
}

func convertResponseToApplication(application *qovery.ApplicationResponse, status *qovery.Status) Application {
	return Application{
		Id:                  fromString(application.Id),
		EnvironmentId:       fromString(application.Environment.Id),
		Name:                fromStringPointer(application.Name),
		Description:         fromStringPointer(application.Description),
		BuildMode:           fromStringPointer(application.BuildMode),
		DockerfilePath:      fromStringPointer(application.DockerfilePath),
		BuildpackLanguage:   fromStringPointer(application.BuildpackLanguage),
		CPU:                 fromInt32Pointer(application.Cpu),
		Memory:              fromInt32Pointer(application.Memory),
		MinRunningInstances: fromInt32Pointer(application.MinRunningInstances),
		MaxRunningInstances: fromInt32Pointer(application.MaxRunningInstances),
		AutoPreview:         fromBoolPointer(application.AutoPreview),
		GitRepository:       convertResponseToApplicationGitRepository(application.GitRepository),
		Storage:             convertResponseToApplicationStorage(application.Storage),
		Ports:               convertResponseToApplicationPorts(application.Ports),
		State:               fromString(status.State),
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
		Name:               toStringPointer(port.Name),
		InternalPort:       toInt32(port.InternalPort),
		ExternalPort:       toInt32Pointer(port.ExternalPort),
		Protocol:           toStringPointer(port.Protocol),
		PubliclyAccessible: toBool(port.PubliclyAccessible),
	}
}

func (port ApplicationPort) toUpdateRequest() qovery.ApplicationPortResponsePorts {
	return qovery.ApplicationPortResponsePorts{
		Id:                 toStringPointer(port.Id),
		Name:               toStringPointer(port.Name),
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
			Name:               fromStringPointer(p.Name),
			InternalPort:       fromInt32(p.InternalPort),
			ExternalPort:       fromInt32Pointer(p.ExternalPort),
			Protocol:           fromStringPointer(p.Protocol),
			PubliclyAccessible: fromBool(p.PubliclyAccessible),
		})
	}
	return list
}
