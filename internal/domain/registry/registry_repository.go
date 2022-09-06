package registry

//go:generate mockery --testonly --with-expecter --name=Repository --structname=RegistryRepository --filename=registry_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of a Registry.
// registryID can be either a registryID, environmentID, application or containerID
type Repository interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*Registry, error)
	Get(ctx context.Context, organizationID string, registryID string) (*Registry, error)
	Update(ctx context.Context, organizationID string, registryID string, request UpsertRequest) (*Registry, error)
	Delete(ctx context.Context, organizationID string, registryID string) error
}
