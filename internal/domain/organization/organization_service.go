package organization

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToGetOrganization    = errors.New("failed to get organization")
	ErrFailedToUpdateOrganization = errors.New("failed to update organization")
)

// Service represents the interface to implement to handle the domain logic of an Organization.
type Service interface {
	Get(ctx context.Context, organizationID string) (*Organization, error)
	Update(ctx context.Context, organizationID string, request UpdateRequest) (*Organization, error)
}
