package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsGcpQoveryAPI implements the interface credentials.GcpRepository.
type credentialsGcpQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface credentials.GcpRepository by credentialsGcpQoveryAPI at compile time.
var _ credentials.GcpRepository = credentialsGcpQoveryAPI{}

// newCredentialsGcpQoveryAPI return a new instance of a credentials.GcpRepository that uses Qovery's API.
func newCredentialsGcpQoveryAPI(client *qovery.APIClient) (credentials.GcpRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &credentialsGcpQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a gcp cluster credentials on an organization using the given organizationID and request.
func (c credentialsGcpQoveryAPI) Create(ctx context.Context, organizationID string, request credentials.UpsertGcpRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		CreateGcpCredentials(ctx, organizationID).
		GcpCredentialsRequest(newQoveryGcpCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceGCPCredentials, request.Name, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Get calls Qovery's API to retrieve a gcp cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsGcpQoveryAPI) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		GetGcpCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceGCPCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Update calls Qovery's API to update a gcp cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsGcpQoveryAPI) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertGcpRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		EditGcpCredentials(ctx, organizationID, credentialsID).
		GcpCredentialsRequest(newQoveryGcpCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceGCPCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Delete calls Qovery's API to delete a gcp cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsGcpQoveryAPI) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsAPI.
		DeleteGcpCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceGCPCredentials, credentialsID, resp, err)
	}

	return nil
}
