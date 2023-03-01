package container

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ContainerRepository --filename=container_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

// Repository represents the interface to implement to handle the persistence of a Container.
type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*Container, error)
	Get(ctx context.Context, containerID string) (*Container, error)
	Update(ctx context.Context, containerID string, request UpsertRepositoryRequest) (*Container, error)
	Delete(ctx context.Context, containerID string) error
}

// UpsertRepositoryRequest represents the parameters needed to create & update a Variable.
type UpsertRepositoryRequest struct {
	RegistryID string `validate:"required"`
	Name       string `validate:"required"`
	ImageName  string `validate:"required"`
	Tag        string `validate:"required"`

	AutoPreview         *bool
	Entrypoint          *string
	CPU                 *int32
	Memory              *int32
	MinRunningInstances *int32
	MaxRunningInstances *int32
	Arguments           []string
	Storages            []storage.UpsertRequest
	Ports               []port.UpsertRequest
	DeploymentStageId   string
}

// Validate returns an error to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
