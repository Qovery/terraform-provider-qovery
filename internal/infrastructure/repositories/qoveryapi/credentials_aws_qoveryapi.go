package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAwsQoveryAPI implements the interface credentials.AwsRepository.
type credentialsAwsQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface credentials.AwsRepository  by credentialsAwsQoveryAPI at compile time.
var _ credentials.AwsRepository = credentialsAwsQoveryAPI{}

// newCredentialsAwsQoveryAPI return a new instance of a credentials.AwsRepository that uses Qovery's API.
func newCredentialsAwsQoveryAPI(client *qovery.APIClient) (credentials.AwsRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &credentialsAwsQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an aws cluster credentials on an organization using the given organizationID and request.
func (c credentialsAwsQoveryAPI) Create(ctx context.Context, organizationID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		CreateAWSCredentials(ctx, organizationID).
		AwsCredentialsRequest(newQoveryAwsCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceAWSCredentials, request.Name, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Get calls Qovery's API to retrieve an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsAwsQoveryAPI) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		GetAWSCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceAWSCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Update calls Qovery's API to update an aws cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsAwsQoveryAPI) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAwsRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		EditAWSCredentials(ctx, organizationID, credentialsID).
		AwsCredentialsRequest(newQoveryAwsCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceAWSCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Delete calls Qovery's API to delete an aws cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsAwsQoveryAPI) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsAPI.
		DeleteAWSCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceAWSCredentials, credentialsID, resp, err)
	}

	return nil
}
