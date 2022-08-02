package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Get handles the domain logic to retrieve a scaleway cluster credentials.
func (c credentialsScalewayService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	return c.credentialsScalewayRepository.Get(ctx, organizationID, credentialsID)
}
