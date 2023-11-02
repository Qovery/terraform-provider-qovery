package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsScalewayQoveryAPI implements the interface credentials.ScalewayRepository.
type credentialsScalewayQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface credentials.ScalewayRepository  by credentialsScalewayQoveryAPI at compile time.
var _ credentials.ScalewayRepository = credentialsScalewayQoveryAPI{}

// newCredentialsScalewayQoveryAPI return a new instance of a credentials.ScalewayRepository that uses Qovery's API.
func newCredentialsScalewayQoveryAPI(client *qovery.APIClient) (credentials.ScalewayRepository, error) {
	if client == nil {
		return nil, common.ErrInvalidQoveryClient
	}

	return &credentialsScalewayQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a scaleway cluster credentials on an organization using the given organizationID and request.
func (c credentialsScalewayQoveryAPI) Create(ctx context.Context, organizationID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		CreateScalewayCredentials(ctx, organizationID).
		ScalewayCredentialsRequest(newQoveryScalewayCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceScalewayCredentials, request.Name, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Get calls Qovery's API to retrieve an scaleway cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsScalewayQoveryAPI) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		GetScalewayCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceScalewayCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Update calls Qovery's API to update a scaleway cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsScalewayQoveryAPI) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertScalewayRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		EditScalewayCredentials(ctx, organizationID, credentialsID).
		ScalewayCredentialsRequest(newQoveryScalewayCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceScalewayCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Delete calls Qovery's API to delete a scaleway cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsScalewayQoveryAPI) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsAPI.
		DeleteScalewayCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceScalewayCredentials, credentialsID, resp, err)
	}

	return nil
}
