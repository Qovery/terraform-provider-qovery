package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) CreateScalewayCredentials(ctx context.Context, organizationID string, request qovery.ScalewayCredentialsRequest) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		CreateScalewayCredentials(ctx, organizationID).
		ScalewayCredentialsRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceScalewayCredentials, request.Name, res, err)
	}
	return credentials, nil
}

func (c *Client) GetScalewayCredentials(ctx context.Context, organizationID string, credentialsID string) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		ListScalewayCredentials(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceScalewayCredentials, credentialsID, res, err)
	}

	for _, creds := range credentials.GetResults() {
		if credentialsID == *creds.Id {
			return &creds, nil
		}
	}
	return nil, apierrors.NewReadError(apierrors.APIResourceScalewayCredentials, credentialsID, res, err)
}

func (c *Client) UpdateScalewayCredentials(ctx context.Context, organizationID string, credentialsID string, request qovery.ScalewayCredentialsRequest) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		EditScalewayCredentials(ctx, organizationID, credentialsID).
		ScalewayCredentialsRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceScalewayCredentials, credentialsID, res, err)
	}
	return credentials, nil
}

func (c *Client) DeleteScalewayCredentials(ctx context.Context, organizationID string, credentialsID string) *apierrors.APIError {
	res, err := c.api.CloudProviderCredentialsApi.
		DeleteScalewayCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceScalewayCredentials, credentialsID, res, err)
	}
	return nil
}
