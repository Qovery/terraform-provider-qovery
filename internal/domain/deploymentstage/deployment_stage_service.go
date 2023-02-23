package deploymentstage

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateDeploymentStage = errors.New("failed to create deployment stage")
	ErrFailedToGetDeploymentStage    = errors.New("failed to get deployment stage")
	ErrFailedToUpdateDeploymentStage = errors.New("failed to update deployment stage")
	ErrFailedToDeleteDeploymentStage = errors.New("failed to delete deployment stage")
)

// Service represents the interface to implement to handle the domain logic of a DeploymentStage
type Service interface {
	Create(ctx context.Context, environmentID string, request UpsertServiceRequest) (*DeploymentStage, error)
	Get(ctx context.Context, environmentId string, deploymentStageId string) (*DeploymentStage, error)
	Update(ctx context.Context, deploymentStageId string, request UpsertServiceRequest) (*DeploymentStage, error)
	Delete(ctx context.Context, deploymentStageId string) error
}

// UpsertServiceRequest represents the parameters needed to create & update a Deployment Stage
type UpsertServiceRequest struct {
	DeploymentStageUpsertRequest UpsertRepositoryRequest
	// TODO (mzo) put services here ?
}

// Validate returns an error to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) Validate() error {
	if err := r.DeploymentStageUpsertRequest.Validate(); err != nil {
		return errors.Wrap(err, ErrInvalidUpsertRequest.Error())
	}

	return nil
}

// IsValid returns a bool to tell whether the UpsertServiceRequest is valid or not.
func (r UpsertServiceRequest) IsValid() bool {
	return r.Validate() == nil
}
