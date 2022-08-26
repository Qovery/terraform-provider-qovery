package project

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ProjectRepository --filename=project_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Repository represents the interface to implement to handle the persistence of a Project.
// projectID can be either a projectID, environmentID, application or containerID
type Repository interface {
	Create(ctx context.Context, organizationID string, request UpsertRepositoryRequest) (*Project, error)
	Get(ctx context.Context, projectID string) (*Project, error)
	Update(ctx context.Context, projectID string, request UpsertRepositoryRequest) (*Project, error)
	Delete(ctx context.Context, projectID string) error
}

// UpsertRepositoryRequest represents the parameters needed to create & update a Variable.
type UpsertRepositoryRequest struct {
	Name        string `validate:"required"`
	Description *string
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
