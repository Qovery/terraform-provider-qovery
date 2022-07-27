package organization

import (
	"context"
)

// Repository represents the interface to implement to handle the persistence of an Organization.
type Repository interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}
