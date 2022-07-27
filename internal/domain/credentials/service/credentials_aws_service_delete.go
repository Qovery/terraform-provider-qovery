package service

import (
	"context"
)

// Delete handles the domain logic to delete an aws cluster credentials.
func (c credentialsAwsService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	return c.credentialsAwsRepository.Delete(ctx, organizationID, credentialsID)
}
