package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Create handles the domain logic to create a scaleway cluster credentials.
func (c credentialsScalewayService) Create(ctx context.Context, organizationID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return c.credentialsScalewayRepository.Create(ctx, organizationID, request)
}
