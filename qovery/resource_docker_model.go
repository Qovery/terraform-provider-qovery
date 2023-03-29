package qovery

import "github.com/qovery/terraform-provider-qovery/internal/domain/docker"

type Docker struct {
	GitRepository  GitRepository `tfsdk:"git_repository"`
	DockerFilePath *string       `tfsdk:"dockerfile_path"`
}

func (d Docker) toUpsertRequest() *docker.Docker {
	return &docker.Docker{
		GitRepository:  d.GitRepository.toUpsertRequest(),
		DockerFilePath: d.DockerFilePath,
	}
}
