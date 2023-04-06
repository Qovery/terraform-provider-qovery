package docker_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	docker_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/docker/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	git_repository_test_helper "github.com/qovery/terraform-provider-qovery/internal/domain/git_repository/test_helper"
)

func TestDockerValidate(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		gitRepository  git_repository.GitRepository
		dockerFilePath *string
		expectedError  error
	}{
		{description: "case 1: git repository is not valid", gitRepository: git_repository_test_helper.DefaultInvalidGitRepository, dockerFilePath: &docker_test_helper.DefaultDockerFilePath, expectedError: errors.Wrap(git_repository_test_helper.DefaultInvalidNewGitRepositoryParamsError, docker.ErrInvalidGitRepositoryParam.Error())},
		{description: "case 2: dockerfile path is nil", gitRepository: git_repository_test_helper.DefaultValidGitRepository, dockerFilePath: nil, expectedError: nil},
		{description: "case 3: all fields are set", gitRepository: git_repository_test_helper.DefaultValidGitRepository, dockerFilePath: &docker_test_helper.DefaultDockerFilePath, expectedError: nil},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i := docker.Docker{
				GitRepository:  git_repository.GitRepository(tc.gitRepository),
				DockerFilePath: tc.dockerFilePath,
			}

			// verify:
			if err := i.Validate(); err != nil {
				assert.Equal(t, tc.expectedError.Error(), i.Validate().Error())
			} else {
				assert.Equal(t, tc.expectedError, i.Validate()) // <- should be nil
			}
		})
	}
}

func TestNewDocker(t *testing.T) {
	// setup:
	testCases := []struct {
		description    string
		params         docker.NewDockerParams
		expectedResult *docker.Docker
		expectedError  error
	}{
		{
			description: "case 1: invalid git repository",
			params: docker.NewDockerParams{
				GitRepository:  git_repository_test_helper.DefaultInvalidNewGitRepositoryParams,
				DockerFilePath: &docker_test_helper.DefaultDockerFilePath,
			},
			expectedError:  errors.Wrap(git_repository_test_helper.DefaultInvalidNewGitRepositoryParamsError, docker.ErrInvalidGitRepositoryParam.Error()),
			expectedResult: nil,
		},
		{
			description: "case 2: docker file path nil",
			params: docker.NewDockerParams{
				GitRepository:  git_repository_test_helper.DefaultValidNewGitRepositoryParams,
				DockerFilePath: nil,
			},
			expectedError: nil,
			expectedResult: &docker.Docker{
				GitRepository:  git_repository_test_helper.DefaultValidGitRepository,
				DockerFilePath: nil,
			},
		},
		{
			description: "case 3: all params properly set",
			params: docker.NewDockerParams{
				GitRepository:  git_repository_test_helper.DefaultValidNewGitRepositoryParams,
				DockerFilePath: &docker_test_helper.DefaultDockerFilePath,
			},
			expectedError: nil,
			expectedResult: &docker.Docker{
				GitRepository:  git_repository_test_helper.DefaultValidGitRepository,
				DockerFilePath: &docker_test_helper.DefaultDockerFilePath,
			},
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// execute:
			i, err := docker.NewDocker(tc.params)

			// verify:
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, nil, err)
			}
			assert.Equal(t, tc.expectedResult, i)
		})
	}
}
