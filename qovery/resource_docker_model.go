package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
)

type Docker struct {
	GitRepository  GitRepository `tfsdk:"git_repository"`
	DockerFilePath types.String  `tfsdk:"dockerfile_path"`
	DockerfileRaw  types.String  `tfsdk:"dockerfile_raw"`
}

func (d Docker) toUpsertRequest() *docker.Docker {
	return &docker.Docker{
		GitRepository:  d.GitRepository.toUpsertRequest(),
		DockerFilePath: ToStringPointer(d.DockerFilePath),
		DockerFileRaw:  ToStringPointer(d.DockerfileRaw),
	}
}
