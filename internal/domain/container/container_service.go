package container

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	ErrFailedToCreateContainer = errors.New("failed to create container")
	ErrFailedToGetContainer    = errors.New("failed to get container")
	ErrFailedToUpdateContainer = errors.New("failed to update container")
	ErrFailedToDeleteContainer = errors.New("failed to delete container")
)

// Service represents the interface to implement to handle the domain logic of an Container.
type Service interface {
	Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*Container, error)
	Get(ctx context.Context, containerID string) (*Container, error)
	Update(ctx context.Context, containerID string, request UpsertServiceRequest) (*Container, error)
	Delete(ctx context.Context, containerID string) error
}

// UpsertServiceRequest represents the parameters needed to create & update a Variable.
type UpsertServiceRequest struct {
	ContainerUpsertRequest UpsertRepositoryRequest
	EnvironmentVariables   variable.DiffRequest
	Secrets                secret.DiffRequest
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) Validate() error {
	if err := r.ContainerUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
