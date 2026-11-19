//go:build unit && !integration

package container_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
)

func newValidContainer() container.Container {
	return container.Container{
		ID:                  uuid.New(),
		EnvironmentID:       uuid.New(),
		RegistryID:          uuid.New(),
		Name:                "my-container",
		IconUri:             "app://icon",
		ImageName:           "nginx",
		Tag:                 "latest",
		CPU:                 500,
		Memory:              512,
		MinRunningInstances: 1,
		MaxRunningInstances: 1,
	}
}

func TestContainer_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, newValidContainer().Validate())
	})

	// Scale-to-zero: min_running_instances = 0 is allowed (KEDA). The `required`
	// validator tag rejected the int32 zero value, so this guards the gte=0 fix.
	t.Run("min_running_instances zero is valid", func(t *testing.T) {
		t.Parallel()
		c := newValidContainer()
		c.MinRunningInstances = 0
		assert.NoError(t, c.Validate())
	})

	t.Run("invalid autoscaling fails", func(t *testing.T) {
		t.Parallel()
		c := newValidContainer()
		c.Autoscaling = &autoscaling.AutoscalingPolicy{} // no scalers
		assert.Error(t, c.Validate())
	})
}

func TestUpsertRepositoryRequest_ValidateSource(t *testing.T) {
	t.Parallel()

	gitSource := &docker.Docker{GitRepository: git_repository.GitRepository{Url: "https://github.com/acme/api.git"}}

	testCases := []struct {
		TestName      string
		Request       container.UpsertRepositoryRequest
		ExpectedError error
	}{
		{
			TestName: "image",
			Request:  container.UpsertRepositoryRequest{Name: "api", RegistryID: uuid.NewString(), ImageName: "nginx", Tag: "1.27"},
		},
		{
			TestName: "image_as_source",
			Request:  container.UpsertRepositoryRequest{Name: "api", RegistryID: uuid.NewString(), ImageName: "nginx", Tag: "1.27", ImageAsSource: true},
		},
		{
			TestName: "git_repository_source",
			Request:  container.UpsertRepositoryRequest{Name: "api", GitSource: gitSource},
		},
		{
			TestName:      "image_without_tag",
			Request:       container.UpsertRepositoryRequest{Name: "api", RegistryID: uuid.NewString(), ImageName: "nginx"},
			ExpectedError: container.ErrInvalidTagParam,
		},
		{
			TestName:      "neither_image_nor_git_repository_source",
			Request:       container.UpsertRepositoryRequest{Name: "api"},
			ExpectedError: container.ErrInvalidRegistryIDParam,
		},
		{
			TestName:      "image_and_git_repository_source",
			Request:       container.UpsertRepositoryRequest{Name: "api", GitSource: gitSource, ImageName: "nginx"},
			ExpectedError: container.ErrImageWithGitSource,
		},
		{
			TestName:      "git_repository_source_without_url",
			Request:       container.UpsertRepositoryRequest{Name: "api", GitSource: &docker.Docker{}},
			ExpectedError: git_repository.ErrInvalidURLParam,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			err := tc.Request.Validate()
			if tc.ExpectedError == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.ExpectedError)
		})
	}
}
