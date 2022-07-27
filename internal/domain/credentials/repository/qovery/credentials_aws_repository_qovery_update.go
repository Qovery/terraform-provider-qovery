package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Update calls Qovery's API to update an aws cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsAwsQoveryRepository) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsApi.
		EditAWSCredentials(ctx, organizationID, credentialsID).
		AwsCredentialsRequest(convertDomainUpsertAwsRequestToQovery(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceAWSCredentials, credentialsID, resp, err)
	}

	return convertQoveryCredentialsToDomain(organizationID, creds)
}
