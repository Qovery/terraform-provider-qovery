package job

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

var (
	ErrFailedToCreateJob = errors.New("failed to create job")
	ErrFailedToGetJob    = errors.New("failed to get job")
	ErrFailedToUpdateJob = errors.New("failed to update job")
	ErrFailedToDeleteJob = errors.New("failed to delete job")
)

// Service represents the interface to implement to handle the domain logic of an Job.
type Service interface {
	Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*Job, error)
	Get(ctx context.Context, jobID string) (*Job, error)
	Update(ctx context.Context, jobID string, request UpsertServiceRequest) (*Job, error)
	Delete(ctx context.Context, jobID string) error
}

// UpsertServiceRequest represents the parameters needed to create & update a Job.
type UpsertServiceRequest struct {
	JobUpsertRequest     UpsertRepositoryRequest
	EnvironmentVariables variable.DiffRequest
	Secrets              secret.DiffRequest
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) Validate() error {
	if err := r.JobUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJobUpsertRequest.Error())
	}

	if err := r.EnvironmentVariables.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJobUpsertRequest.Error())
	}

	if err := r.Secrets.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidJobUpsertRequest.Error())
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidJobUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
