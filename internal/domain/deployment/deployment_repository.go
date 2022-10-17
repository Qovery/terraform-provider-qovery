package deployment

//go:generate mockery --testonly --with-expecter --name=Repository --structname=DeploymentRepository --filename=deployment_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// Repository represents the interface to implement to handle the deployments of qovery services.
type Repository interface {
	GetStatus(ctx context.Context, resourceID string) (*status.Status, error)
	Deploy(ctx context.Context, resourceID string, version string) (*status.Status, error)
	Restart(ctx context.Context, resourceID string) (*status.Status, error)
	Stop(ctx context.Context, resourceID string) (*status.Status, error)
}
