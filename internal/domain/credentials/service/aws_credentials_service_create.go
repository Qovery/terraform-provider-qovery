package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Create handles the domain logic to create an aws cluster credentials.
func (c credentialsAwsService) Create(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return c.credentialsAwsRepository.Create(ctx, organizationID, request)
}
