package newdeployment

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateDeployment        = errors.New("failed to create deployment")
	ErrFailedToGetDeployment           = errors.New("failed to get deployment")
	ErrFailedToDeleteDeployment        = errors.New("failed to delete deployment")
	ErrDesiredStateForbiddenAtCreation = errors.New("Cannot create a deployment having state 'DELETED' or 'RESTARTED'")
	ErrFailedToCheckDeploymentStatus   = errors.New("failed to retrieve deployment status")
	ErrFailedToGetNextDeploymentId     = errors.New("failed to get next deployment id")
)

type Service interface {
	Create(ctx context.Context, params NewDeploymentParams) (*Deployment, error)
	Get(ctx context.Context, params NewDeploymentParams) (*Deployment, error)
	Update(ctx context.Context, params NewDeploymentParams) (*Deployment, error)
	Delete(ctx context.Context, params NewDeploymentParams) error
}
