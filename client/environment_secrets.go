package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getEnvironmentSecrets(ctx context.Context, environmentID string) ([]*qovery.Secret, *apierrors.APIError) {
	vars, res, err := c.api.EnvironmentSecretApi.
		ListEnvironmentSecrets(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceEnvironmentSecret, environmentID, res, err)
	}
	return secretResponseListToArray(vars, qovery.APIVARIABLESCOPEENUM_ENVIRONMENT), nil
}

func (c *Client) updateEnvironmentSecrets(ctx context.Context, environmentID string, request SecretsDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.EnvironmentSecretApi.
			DeleteEnvironmentSecret(ctx, environmentID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceEnvironmentSecret, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.EnvironmentSecretApi.
			EditEnvironmentSecret(ctx, environmentID, variable.Id).
			SecretEditRequest(variable.SecretEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceEnvironmentSecret, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.EnvironmentSecretApi.
			CreateEnvironmentSecret(ctx, environmentID).
			SecretRequest(variable.SecretRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceEnvironmentSecret, variable.Key, res, err)
		}
	}
	return nil
}
