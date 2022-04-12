package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) CreateOrganization(ctx context.Context, request qovery.OrganizationRequest) (*qovery.OrganizationResponse, *apierrors.APIError) {
	organization, res, err := c.api.OrganizationMainCallsApi.
		CreateOrganization(ctx).
		OrganizationRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceOrganization, request.Name, res, err)
	}
	return organization, nil
}

func (c *Client) GetOrganization(ctx context.Context, organizationID string) (*qovery.OrganizationResponse, *apierrors.APIError) {
	organization, res, err := c.api.OrganizationMainCallsApi.
		GetOrganization(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceOrganization, organizationID, res, err)
	}
	return organization, nil
}

func (c *Client) UpdateOrganization(ctx context.Context, organizationID string, request qovery.OrganizationEditRequest) (*qovery.OrganizationResponse, *apierrors.APIError) {
	organization, res, err := c.api.OrganizationMainCallsApi.
		EditOrganization(ctx, organizationID).
		OrganizationEditRequest(request).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceOrganization, organizationID, res, err)
	}
	return organization, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, organizationID string) *apierrors.APIError {
	res, err := c.api.OrganizationMainCallsApi.
		DeleteOrganization(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceOrganization, organizationID, res, err)
	}
	return nil
}
