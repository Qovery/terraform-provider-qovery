package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Update handles the domain logic to update an aws cluster credentials.
func (c credentialsAwsService) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return c.credentialsAwsRepository.Update(ctx, organizationID, credentialsID, request)
}
