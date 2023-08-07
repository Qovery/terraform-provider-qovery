package environment

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	ErrFailedToCreateEnvironment = errors.New("failed to create environment")
	ErrFailedToGetEnvironment    = errors.New("failed to get environment")
	ErrFailedToUpdateEnvironment = errors.New("failed to update environment")
	ErrFailedToDeleteEnvironment = errors.New("failed to delete environment")
)

// Service represents the interface to implement to handle the domain logic of an Environment.
type Service interface {
	Create(ctx context.Context, projectID string, request CreateServiceRequest) (*Environment, error)
	Get(ctx context.Context, environmentID string) (*Environment, error)
	Update(ctx context.Context, environmentID string, request UpdateServiceRequest) (*Environment, error)
	Delete(ctx context.Context, environmentID string) error
}

// CreateServiceRequest represents the parameters needed to create an Environment.
type CreateServiceRequest struct {
	EnvironmentCreateRequest     CreateRepositoryRequest
	EnvironmentVariables         variable.DiffRequest
	EnvironmentVariableAliases   variable.DiffRequest
	EnvironmentVariableOverrides variable.DiffRequest
	Secrets                      secret.DiffRequest
	SecretAliases                secret.DiffRequest
	SecretOverrides              secret.DiffRequest
}

// Validate returns an error to tell whether the CreateServiceRequest is valid or not.
func (r CreateServiceRequest) Validate() error {
	if err := r.EnvironmentCreateRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidCreateRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidCreateRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidCreateRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the CreateServiceRequest is valid or not.
func (r CreateServiceRequest) IsValid() bool {
	return r.Validate() == nil
}

// UpdateServiceRequest represents the parameters needed to update an Environment.
type UpdateServiceRequest struct {
	EnvironmentUpdateRequest     UpdateRepositoryRequest
	EnvironmentVariables         variable.DiffRequest
	EnvironmentVariableAliases   variable.DiffRequest
	EnvironmentVariableOverrides variable.DiffRequest
	Secrets                      secret.DiffRequest
	SecretAliases                secret.DiffRequest
	SecretOverrides              secret.DiffRequest
}

// Validate returns an error to tell whether the UpdateServiceRequest is valid or not.
func (r UpdateServiceRequest) Validate() error {
	if err := r.EnvironmentUpdateRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpdateRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpdateServiceRequest is valid or not.
func (r UpdateServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
