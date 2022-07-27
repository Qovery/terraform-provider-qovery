package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// Update handles the domain logic to update an organization.
func (o organizationService) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return o.organizationRepository.Update(ctx, organizationID, request)
}
