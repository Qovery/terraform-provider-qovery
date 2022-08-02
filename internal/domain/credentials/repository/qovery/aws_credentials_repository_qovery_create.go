package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// Create calls Qovery's API to create an aws cluster credentials on an organization using the given organizationID and request.
func (c credentialsAwsQoveryRepository) Create(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsApi.
		CreateAWSCredentials(ctx, organizationID).
		AwsCredentialsRequest(convertDomainUpsertAwsRequestToQovery(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceAWSCredentials, request.Name, resp, err)
	}

	return convertQoveryCredentialsToDomain(organizationID, creds)
}
