package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsAzureQoveryAPI implements the interface credentials.AzureRepository.
type credentialsAzureQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface credentials.AzureRepository by credentialsAzureQoveryAPI at compile time.
var _ credentials.AzureRepository = credentialsAzureQoveryAPI{}

// newCredentialsAzureQoveryAPI return a new instance of a credentials.AzureRepository that uses Qovery's API.
func newCredentialsAzureQoveryAPI(client *qovery.APIClient) (credentials.AzureRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &credentialsAzureQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an azure cluster credentials on an organization using the given organizationID and request.
func (c credentialsAzureQoveryAPI) Create(ctx context.Context, organizationID string, request credentials.UpsertAzureRequest) (*credentials.AzureCredentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		CreateAzureCredentials(ctx, organizationID).
		AzureCredentialsRequest(newQoveryAzureCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceAzureCredentials, request.Name, resp, err)
	}

	return newDomainAzureCredentialsFromQovery(organizationID, creds)
}

// Get calls Qovery's API to retrieve an azure cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsAzureQoveryAPI) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.AzureCredentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		GetAzureCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceAzureCredentials, credentialsID, resp, err)
	}

	return newDomainAzureCredentialsFromQovery(organizationID, creds)
}

// Update calls Qovery's API to update an azure cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsAzureQoveryAPI) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertAzureRequest) (*credentials.AzureCredentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		EditAzureCredentials(ctx, organizationID, credentialsID).
		AzureCredentialsRequest(newQoveryAzureCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceAzureCredentials, credentialsID, resp, err)
	}

	return newDomainAzureCredentialsFromQovery(organizationID, creds)
}

// Delete calls Qovery's API to delete an azure cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsAzureQoveryAPI) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsAPI.
		DeleteAzureCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceAzureCredentials, credentialsID, resp, err)
	}

	return nil
}
