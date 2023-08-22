package project

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	ErrFailedToCreateProject = errors.New("failed to create project")
	ErrFailedToGetProject    = errors.New("failed to get project")
	ErrFailedToUpdateProject = errors.New("failed to update project")
	ErrFailedToDeleteProject = errors.New("failed to delete project")
)

// Service represents the interface to implement to handle the domain logic of an Project.
type Service interface {
	Create(ctx context.Context, organizationID string, request UpsertServiceRequest) (*Project, error)
	Get(ctx context.Context, projectID string) (*Project, error)
	Update(ctx context.Context, projectID string, request UpsertServiceRequest) (*Project, error)
	Delete(ctx context.Context, projectID string) error
}

// UpsertServiceRequest represents the parameters needed to create & update a Variable.
type UpsertServiceRequest struct {
	ProjectUpsertRequest       UpsertRepositoryRequest
	EnvironmentVariables       variable.DiffRequest
	EnvironmentVariableAliases variable.DiffRequest
	Secrets                    secret.DiffRequest
	SecretAliases              secret.DiffRequest
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) Validate() error {
	if err := r.ProjectUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
