package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Get calls Qovery's API to retrieve an scaleway cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsScalewayQoveryRepository) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	credsList, resp, err := c.client.CloudProviderCredentialsApi.
		ListScalewayCredentials(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceScalewayCredentials, credentialsID, resp, err)
	}

	for _, creds := range credsList.GetResults() {
		if credentialsID == *creds.Id {
			return convertQoveryCredentialsToDomain(organizationID, &creds)
		}
	}

	// NOTE: Force status 404 since we didn't find the credential.
	// The status is used to generate the proper error return by the provider.
	resp.StatusCode = 404
	return nil, apierrors.NewReadApiError(apierrors.ApiResourceScalewayCredentials, credentialsID, resp, err)
}
