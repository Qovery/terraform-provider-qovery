package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Create calls Qovery's API to create a scaleway cluster credentials on an organization using the given organizationID and request.
func (c credentialsScalewayQoveryRepository) Create(ctx context.Context, organizationID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsApi.
		CreateScalewayCredentials(ctx, organizationID).
		ScalewayCredentialsRequest(convertDomainUpsertScalewayRequestToQovery(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceScalewayCredentials, request.Name, resp, err)
	}

	return convertQoveryCredentialsToDomain(organizationID, creds)
}
