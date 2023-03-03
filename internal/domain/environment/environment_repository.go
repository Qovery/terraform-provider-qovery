package environment

//go:generate mockery --testonly --with-expecter --name=Repository --structname=EnvironmentRepository --filename=environment_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Repository represents the interface to implement to handle the persistence of an environment.
type Repository interface {
	Create(ctx context.Context, projectID string, request CreateRepositoryRequest) (*Environment, error)
	Get(ctx context.Context, environmentID string) (*Environment, error)
	Update(ctx context.Context, environmentID string, request UpdateRepositoryRequest) (*Environment, error)
	Delete(ctx context.Context, environmentID string) error
	Exists(ctx context.Context, environmentId string) bool
}

// CreateRepositoryRequest represents the parameters needed to create an Environment.
type CreateRepositoryRequest struct {
	Name      string `validate:"required"`
	ClusterID *string
	Mode      *Mode
}

// Validate returns an error to tell whether the CreateRepositoryRequest is valid or not.
func (r CreateRepositoryRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidCreateRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpdateRepositoryRequest is valid or not.
func (r CreateRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}

// UpdateRepositoryRequest represents the parameters needed to update an Environment.
type UpdateRepositoryRequest struct {
	Name *string
	Mode *Mode
}

// Validate returns an error to tell whether the UpdateRepositoryRequest is valid or not.
func (r UpdateRepositoryRequest) Validate() error {
	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpdateRepositoryRequest is valid or not.
func (r UpdateRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
