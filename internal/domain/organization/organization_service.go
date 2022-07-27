package organization

import (
	"context"
)

// Service represents the interface to implement to handle the domain logic of an Organization.
type Service interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}
