package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationCustomDomains(ctx context.Context, applicationID string) ([]*qovery.CustomDomain, *apierrors.APIError) {
	applicationDomains, res, err := c.api.CustomDomainAPI.
		ListApplicationCustomDomain(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationCustomDomain, applicationID, res, err)
	}
	return customDomainResponseListToArray(applicationDomains), nil
}

func (c *Client) updateApplicationCustomDomains(ctx context.Context, applicationID string, request CustomDomainsDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.CustomDomainAPI.
			DeleteCustomDomain(ctx, applicationID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationCustomDomain, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.CustomDomainAPI.
			EditCustomDomain(ctx, applicationID, variable.Id).
			CustomDomainRequest(variable.CustomDomainRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationCustomDomain, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.CustomDomainAPI.
			CreateApplicationCustomDomain(ctx, applicationID).
			CustomDomainRequest(variable.CustomDomainRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationCustomDomain, variable.Domain, res, err)
		}
	}
	return nil
}
