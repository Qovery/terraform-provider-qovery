package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Update calls Qovery's API to update a scaleway cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsScalewayQoveryRepository) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsApi.
		EditScalewayCredentials(ctx, organizationID, credentialsID).
		ScalewayCredentialsRequest(convertDomainUpsertScalewayRequestToQovery(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceScalewayCredentials, credentialsID, resp, err)
	}

	return convertQoveryCredentialsToDomain(organizationID, creds)
}
