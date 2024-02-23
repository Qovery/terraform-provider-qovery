package deploymentstage

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Repository represents the interface to implement to handle the persistence of a DeploymentStage.
type Repository interface {
	Create(ctx context.Context, environmentID string, request UpsertRepositoryRequest) (*DeploymentStage, error)
	Get(ctx context.Context, environmentID string, deploymentStageID string) (*DeploymentStage, error)
	GetAllByEnvironmentID(ctx context.Context, environmentID string) (*[]DeploymentStage, error)
	Update(ctx context.Context, deploymentStageID string, request UpsertRepositoryRequest) (*DeploymentStage, error)
	Delete(ctx context.Context, deploymentStageID string) error
}

// UpsertRepositoryRequest represents the parameters needed to create & update a DeploymentStage
type UpsertRepositoryRequest struct {
	Name        string `validate:"required"`
	Description string
	IsAfter     *string
	IsBefore    *string
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
