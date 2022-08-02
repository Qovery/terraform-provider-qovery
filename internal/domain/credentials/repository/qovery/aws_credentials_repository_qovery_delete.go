package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// Delete calls Qovery's API to delete an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsAwsQoveryRepository) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsApi.
		DeleteAWSCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceAWSCredentials, credentialsID, resp, err)
	}

	return nil
}
