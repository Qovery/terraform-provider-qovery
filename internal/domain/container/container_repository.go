package container

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ContainerRepository --filename=container_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

// Repository represents the interface to implement to handle the persistence of a Container.
type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*Container, error)
	Get(ctx context.Context, containerID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*Container, error)
	Update(ctx context.Context, containerID string, request UpsertRepositoryRequest) (*Container, error)
	Delete(ctx context.Context, containerID string) error
}

// UpsertRepositoryRequest represents the parameters needed to create & update a Variable.
type UpsertRepositoryRequest struct {
	RegistryID           string
	Name                 string `validate:"required"`
	IconUri              *string
	ImageName            string
	Tag                  string
	AutoPreview          *bool
	Entrypoint           *string
	CPU                  *int32
	Memory               *int32
	EphemeralStorage     *int32
	MinRunningInstances  *int32
	MaxRunningInstances  *int32
	Arguments            []string
	Storages             []storage.UpsertRequest
	Ports                []port.UpsertRequest
	DeploymentStageID    string
	IsSkipped            bool
	Healthchecks         qovery.Healthcheck
	AdvancedSettingsJson string
	CustomDomains        client.CustomDomainsDiff
	AutoDeploy           qovery.NullableBool
	AnnotationsGroupIds  []string
	LabelsGroupIds       []string
	Autoscaling          *autoscaling.AutoscalingPolicy
	// GitSource is the Git repository Qovery builds the image from. When set, RegistryID, ImageName and Tag are empty.
	GitSource *docker.Docker
	// ImageAsSource sends the image as the container source, which switches a container with a Git repository source back to it.
	ImageAsSource bool
	// BuildSettings applies to a container with a Git repository source only.
	BuildSettings *qovery.BuildSettings
}

// Validate returns an error to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.validateSource(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// validateSource checks that the request names either an image or a Git repository source.
func (r UpsertRepositoryRequest) validateSource() error {
	if r.GitSource != nil {
		if r.RegistryID != "" || r.ImageName != "" || r.Tag != "" || r.ImageAsSource {
			return ErrImageWithGitSource
		}
		return r.GitSource.Validate()
	}

	if r.RegistryID == "" {
		return ErrInvalidRegistryIDParam
	}
	if r.ImageName == "" {
		return ErrInvalidImageNameParam
	}
	if r.Tag == "" {
		return ErrInvalidTagParam
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
