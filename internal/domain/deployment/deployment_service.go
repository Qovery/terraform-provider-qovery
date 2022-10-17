package deployment

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

//go:generate mockery --testonly --with-expecter --name=Service --structname=DeploymentService --filename=deployment_service_mock.go --output=../../application/services/mocks_test/ --outpkg=mocks_test
var (
	ErrInvalidResourceIDParam = errors.New("invalid resource id param")
	ErrUnexpectedState        = errors.New("unexpected state")
	ErrFailedToGetStatus      = errors.New("failed to get status")
	ErrFailedToUpdateState    = errors.New("failed to update state")
	ErrFailedToDeploy         = errors.New("failed to deploy")
	ErrFailedToRestart        = errors.New("failed to stop")
	ErrFailedToStop           = errors.New("failed to restart")
)

// Service represents the interface to implement to handle the domain logic of a deployment.
type Service interface {
	GetStatus(ctx context.Context, resourceID string) (*status.Status, error)
	UpdateState(ctx context.Context, resourceID string, desiredState status.State, version string) (*status.Status, error)
	Deploy(ctx context.Context, resourceID string, version string) (*status.Status, error)
	Restart(ctx context.Context, resourceID string) (*status.Status, error)
	Stop(ctx context.Context, resourceID string) (*status.Status, error)
}
