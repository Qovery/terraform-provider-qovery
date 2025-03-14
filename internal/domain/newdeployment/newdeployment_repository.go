package newdeployment

import (
	"context"

	"github.com/google/uuid"
)

type EnvironmentRepository interface {
	Deploy(ctx context.Context, newDeployment Deployment) (*Deployment, error)
	ReDeploy(ctx context.Context, newDeployment Deployment) (*Deployment, error)
	Stop(ctx context.Context, newDeployment Deployment) (*Deployment, error)
	Restart(ctx context.Context, newDeployment Deployment) (*Deployment, error)
	Delete(ctx context.Context, newDeployment Deployment) (*Deployment, error)
}

type DeploymentStatusRepository interface {
	WaitForTerminatedState(ctx context.Context, environmentId uuid.UUID) error
	WaitForExpectedDesiredState(ctx context.Context, newDeployment Deployment) error
	CheckEnvironmentExists(ctx context.Context, environmentId uuid.UUID) (error, int)
}
