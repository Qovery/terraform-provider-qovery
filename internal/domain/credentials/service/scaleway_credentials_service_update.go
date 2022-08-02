package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Update handles the domain logic to update a scaleway cluster credentials.
func (c credentialsScalewayService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return c.credentialsScalewayRepository.Update(ctx, organizationID, credentialsID, request)
}
