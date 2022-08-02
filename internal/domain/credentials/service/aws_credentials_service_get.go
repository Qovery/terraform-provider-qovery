package service

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Get handles the domain logic to retrieve an aws cluster credentials.
func (c credentialsAwsService) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	return c.credentialsAwsRepository.Get(ctx, organizationID, credentialsID)
}
