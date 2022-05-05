package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationSecrets(ctx context.Context, environmentID string) ([]*qovery.Secret, *apierrors.APIError) {
	vars, res, err := c.api.ApplicationSecretApi.
		ListApplicationSecrets(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationSecret, environmentID, res, err)
	}
	return secretResponseListToArray(vars, qovery.ENVIRONMENTVARIABLESCOPEENUM_APPLICATION), nil
}

func (c *Client) updateApplicationSecrets(ctx context.Context, environmentID string, request SecretsDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationSecretApi.
			DeleteApplicationSecret(ctx, environmentID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationSecret, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.ApplicationSecretApi.
			EditApplicationSecret(ctx, environmentID, variable.Id).
			SecretEditRequest(variable.SecretEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationSecret, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.ApplicationSecretApi.
			CreateApplicationSecret(ctx, environmentID).
			SecretRequest(variable.SecretRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationSecret, variable.Key, res, err)
		}
	}
	return nil
}
