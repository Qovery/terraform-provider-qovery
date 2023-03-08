package job

//go:generate mockery --testonly --with-expecter --name=Repository --structname=JobRepository --filename=job_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// Repository represents the interface to implement to handle the persistence of a Job.
type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*Job, error)
	Get(ctx context.Context, jobID string) (*Job, error)
	Update(ctx context.Context, jobID string, request UpsertRepositoryRequest) (*Job, error)
	Delete(ctx context.Context, jobID string) error
}

// UpsertRepositoryRequest represents the parameters needed to create & update a Job.
type UpsertRepositoryRequest struct {
	JobID         string `validate:"required"`
	EnvironmentID string `validate:"required"`
	Name          string `validate:"required"`

	AutoPreview        *bool
	Entrypoint         *string
	CPU                *int32
	Memory             *int32
	MaxNbRestart       *uint32
	MaxDurationSeconds *uint32

	Source               JobSource
	Schedule             JobSchedule
	Ports                []port.UpsertRequest
	EnvironmentVariables []variable.UpsertRequest
	Secrets              []secret.UpsertRequest

	State *string
}

// Validate returns an error to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) Validate() error {

	if err := r.Schedule.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := r.Source.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	if err := validator.New().Struct(r); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertRepositoryRequest is valid or not.
func (r UpsertRepositoryRequest) IsValid() bool {
	return r.Validate() == nil
}
