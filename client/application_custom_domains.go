package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationCustomDomains(ctx context.Context, applicationID string) ([]*qovery.CustomDomain, *apierrors.APIError) {
	applicationDomains, res, err := c.api.ApplicationCustomDomainAPI.
		ListApplicationCustomDomain(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationCustomDomain, applicationID, res, err)
	}
	return customDomainResponseListToArray(applicationDomains), nil
}

func (c *Client) updateApplicationCustomDomains(ctx context.Context, applicationID string, request CustomDomainsDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationCustomDomainAPI.
			DeleteCustomDomain(ctx, applicationID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationCustomDomain, variable.Id, res, err)
		}
	}

	for _, customDomainToUpdate := range request.Update {
		_, res, err := c.api.ApplicationCustomDomainAPI.
			EditCustomDomain(ctx, applicationID, customDomainToUpdate.Id).
			CustomDomainRequest(customDomainToUpdate.CustomDomainRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationCustomDomain, customDomainToUpdate.Id, res, err)
		}
	}

	for _, customDomainToCreate := range request.Create {
		_, res, err := c.api.ApplicationCustomDomainAPI.
			CreateApplicationCustomDomain(ctx, applicationID).
			CustomDomainRequest(customDomainToCreate.CustomDomainRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationCustomDomain, customDomainToCreate.Domain, res, err)
		}
	}
	return nil
}
