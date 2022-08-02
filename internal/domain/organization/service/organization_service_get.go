package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// Get handles the domain logic to retrieve an organization.
func (o organizationService) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	return o.organizationRepository.Get(ctx, organizationID)
}
