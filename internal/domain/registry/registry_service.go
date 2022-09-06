package registry

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateRegistry = errors.New("failed to create registry")
	ErrFailedToGetRegistry    = errors.New("failed to get registry")
	ErrFailedToUpdateRegistry = errors.New("failed to update registry")
	ErrFailedToDeleteRegistry = errors.New("failed to delete registry")
)

// Service represents the interface to implement to handle the domain logic of an Registry.
type Service interface {
	Create(ctx context.Context, organizationID string, request UpsertRequest) (*Registry, error)
	Get(ctx context.Context, organizationID string, registryID string) (*Registry, error)
	Update(ctx context.Context, organizationID string, registryID string, request UpsertRequest) (*Registry, error)
	Delete(ctx context.Context, organizationID string, registryID string) error
}
