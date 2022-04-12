package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) CreateAWSCredentials(ctx context.Context, organizationID string, request qovery.AwsCredentialsRequest) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		CreateAWSCredentials(ctx, organizationID).
		AwsCredentialsRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceAWSCredentials, request.Name, res, err)
	}
	return credentials, nil
}

func (c *Client) GetAWSCredentials(ctx context.Context, organizationID string, credentialsID string) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		ListAWSCredentials(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceAWSCredentials, credentialsID, res, err)
	}

	for _, creds := range credentials.GetResults() {
		if credentialsID == *creds.Id {
			return &creds, nil
		}
	}
	return nil, apierrors.NewReadError(apierrors.APIResourceAWSCredentials, credentialsID, res, err)
}

func (c *Client) UpdateAWSCredentials(ctx context.Context, organizationID string, credentialsID string, request qovery.AwsCredentialsRequest) (*qovery.ClusterCredentialsResponse, *apierrors.APIError) {
	credentials, res, err := c.api.CloudProviderCredentialsApi.
		EditAWSCredentials(ctx, organizationID, credentialsID).
		AwsCredentialsRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceAWSCredentials, credentialsID, res, err)
	}
	return credentials, nil
}

func (c *Client) DeleteAWSCredentials(ctx context.Context, organizationID string, credentialsID string) *apierrors.APIError {
	res, err := c.api.CloudProviderCredentialsApi.
		DeleteAWSCredentials(ctx, organizationID, credentialsID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceAWSCredentials, credentialsID, res, err)
	}
	return nil
}
