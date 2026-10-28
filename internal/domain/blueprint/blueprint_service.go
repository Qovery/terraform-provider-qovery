package blueprint

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateBlueprint = errors.New("failed to create blueprint")
	ErrFailedToGetBlueprint    = errors.New("failed to get blueprint")
	ErrFailedToUpdateBlueprint = errors.New("failed to update blueprint")
	ErrFailedToDeleteBlueprint = errors.New("failed to delete blueprint")
	ErrFailedToResolveTag      = errors.New("failed to resolve blueprint latest tag")
	ErrFailedToGetDefaults     = errors.New("failed to get blueprint variable defaults")
)

// Service represents the interface to implement to handle the domain logic of a blueprint.
type Service interface {
	// Create can return a blueprint alongside an error: the blueprint exists in Qovery but
	// its dispatch or deployment failed, and the caller must still track it.
	Create(ctx context.Context, environmentID string, request CreateRequest) (*Blueprint, error)
	// Get also reads the icon from the service the blueprint materialized.
	Get(ctx context.Context, blueprintID string) (*Blueprint, error)
	Update(ctx context.Context, blueprintID string, request UpdateRequest) (*Blueprint, error)
	Delete(ctx context.Context, blueprintID string) error
	// ResolveCatalogEntry returns the catalog entry of version. Its LatestTag is the only tag Terraform deploys.
	ResolveCatalogEntry(ctx context.Context, environmentID string, version CatalogVersion) (CatalogEntry, error)
	// GetVariableDefaults returns the catalog default of each variable of version that has one.
	GetVariableDefaults(ctx context.Context, environmentID string, version CatalogVersion) (map[string]string, error)
}
