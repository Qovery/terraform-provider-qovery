package helmRepository

//go:generate mockery --testonly --with-expecter --name=Repository --structname=RegistryRepository --filename=registry_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of a Helm Repository.
type Repository interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*HelmRepository, error)
	Get(ctx context.Context, organizationID string, registryID string) (*HelmRepository, error)
	Update(ctx context.Context, organizationID string, registryID string, request UpsertRequest) (*HelmRepository, error)
	Delete(ctx context.Context, organizationID string, registryID string) error
}
