package test_helper

import (
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	git_repository_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/git_repository/test_helper"
)

var (
	DefaultDockerFilePath = "/"

	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultValidNewDockerParams = docker.NewDockerParams{
		GitRepository:  git_repository_test_helper.DefaultValidNewGitRepositoryParams,
		DockerFilePath: &DefaultDockerFilePath,
	}
	DefaultValidDocker = docker.Docker{
		GitRepository:  git_repository_test_helper.DefaultValidGitRepository,
		DockerFilePath: &DefaultDockerFilePath,
	}
	/// Exposed to tests needing to get such object without having to know internal sauce magic
	DefaultInvalidNewDockerParams = docker.NewDockerParams{
		GitRepository:  git_repository_test_helper.DefaultInvalidNewGitRepositoryParams,
		DockerFilePath: nil,
	}
	DefaultInvalidDocker = docker.Docker{
		GitRepository:  git_repository_test_helper.DefaultInvalidGitRepository,
		DockerFilePath: nil,
	}
	DefaultInvalidNewDockerParamsError = errors.Wrap(git_repository_test_helper.DefaultInvalidNewGitRepositoryParamsError, docker.ErrInvalidGitRepositoryParam.Error())
)
