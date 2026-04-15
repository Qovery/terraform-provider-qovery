package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/credentials"
)

// credentialsEksAnywhereVsphereQoveryAPI implements the interface credentials.EksAnywhereVsphereRepository.
type credentialsEksAnywhereVsphereQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface credentials.EksAnywhereVsphereRepository by credentialsEksAnywhereVsphereQoveryAPI at compile time.
var _ credentials.EksAnywhereVsphereRepository = credentialsEksAnywhereVsphereQoveryAPI{}

// newCredentialsEksAnywhereVsphereQoveryAPI return a new instance of a credentials.EksAnywhereVsphereRepository that uses Qovery's API.
func newCredentialsEksAnywhereVsphereQoveryAPI(client *qovery.APIClient) (credentials.EksAnywhereVsphereRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &credentialsEksAnywhereVsphereQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create an eks anywhere vsphere cluster credentials on an organization using the given organizationID and request.
func (c credentialsEksAnywhereVsphereQoveryAPI) Create(ctx context.Context, organizationID string, request credentials.UpsertEksAnywhereVsphereRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		CreateAWSCredentials(ctx, organizationID).
		AwsCredentialsRequest(newQoveryEksAnywhereVsphereCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceEksAnywhereVsphereCredentials, request.Name, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Get calls Qovery's API to retrieve an eks anywhere vsphere cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsEksAnywhereVsphereQoveryAPI) Get(ctx context.Context, organizationID string, credentialsID string) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		GetAWSCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceEksAnywhereVsphereCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Update calls Qovery's API to update an eks anywhere vsphere cluster credentials from an organization using the given organizationID, credentialsID and request.
func (c credentialsEksAnywhereVsphereQoveryAPI) Update(ctx context.Context, organizationID string, credentialsID string, request credentials.UpsertEksAnywhereVsphereRequest) (*credentials.Credentials, error) {
	creds, resp, err := c.client.CloudProviderCredentialsAPI.
		EditAWSCredentials(ctx, organizationID, credentialsID).
		AwsCredentialsRequest(newQoveryEksAnywhereVsphereCredentialsRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceEksAnywhereVsphereCredentials, credentialsID, resp, err)
	}

	return newDomainCredentialsFromQovery(organizationID, creds)
}

// Delete calls Qovery's API to delete an eks anywhere vsphere cluster credentials from an organization using the given organizationID and credentialsID.
func (c credentialsEksAnywhereVsphereQoveryAPI) Delete(ctx context.Context, organizationID string, credentialsID string) error {
	resp, err := c.client.CloudProviderCredentialsAPI.
		DeleteAWSCredentials(ctx, credentialsID, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceEksAnywhereVsphereCredentials, credentialsID, resp, err)
	}

	return nil
}
