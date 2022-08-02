package service

import (
	"context"
)

// Delete handles the domain logic to delete a scaleway cluster credentials.
func (c credentialsScalewayService) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	return c.credentialsScalewayRepository.Delete(ctx, organizationID, credentialsID)
}
