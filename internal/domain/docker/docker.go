package docker

import (
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

var (
	// ErrInvalidGitRepositoryParam is returned if the git repository param is invalid.
	ErrInvalidGitRepositoryParam = errors.New("invalid git repository param")
)

type Docker struct {
	GitRepository  git_repository.GitRepository
	DockerFilePath *string
}

func (d Docker) Validate() error {
	if err := d.GitRepository.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidGitRepositoryParam.Error())
	}

	return nil
}

type NewDockerParams struct {
	GitRepository  git_repository.NewGitRepositoryParams
	DockerFilePath *string
}

func NewDocker(params NewDockerParams) (*Docker, error) {
	gitRepository, err := git_repository.NewGitRepository(params.GitRepository)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidGitRepositoryParam.Error())
	}
	docker := &Docker{
		GitRepository:  *gitRepository,
		DockerFilePath: params.DockerFilePath,
	}

	if err := docker.Validate(); err != nil {
		return nil, err
	}

	return docker, nil
}
