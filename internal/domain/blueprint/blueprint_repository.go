package blueprint

//go:generate mockery --testonly --with-expecter --name=Repository --structname=BlueprintRepository --filename=blueprint_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of a blueprint.
type Repository interface {
	Create(ctx context.Context, environmentID string, request CreateRequest) (*CreationResult, error)
	Get(ctx context.Context, blueprintID string) (*Blueprint, error)
	Update(ctx context.Context, blueprintID string, request UpdateRequest) error
	// Deploy re-dispatches the persisted blueprint and returns the dispatch ID.
	Deploy(ctx context.Context, blueprintID string) (string, error)
	// GetServiceStatus returns nil when the environment does not report the service yet.
	GetServiceStatus(ctx context.Context, environmentID string, serviceType ServiceType, serviceID string) (*ServiceStatus, error)
	// DeleteService deletes the blueprint's service, which deletes the blueprint too.
	DeleteService(ctx context.Context, serviceType ServiceType, serviceID string) error
	GetOrganizationID(ctx context.Context, environmentID string) (string, error)
	ListCatalog(ctx context.Context, organizationID string) ([]CatalogEntry, error)
}
