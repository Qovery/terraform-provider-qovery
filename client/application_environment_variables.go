package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getApplicationEnvironmentVariables(ctx context.Context, applicationID string) ([]*qovery.EnvironmentVariable, *apierrors.APIError) {
	applicationVariables, res, err := c.api.ApplicationEnvironmentVariableApi.
		ListApplicationEnvironmentVariable(ctx, applicationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceApplicationEnvironmentVariable, applicationID, res, err)
	}
	return environmentVariableResponseListToArray(applicationVariables, qovery.APIVARIABLESCOPEENUM_APPLICATION), nil
}

func (c *Client) updateApplicationEnvironmentVariables(ctx context.Context, applicationID string, request EnvironmentVariablesDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Update {
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			EditApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			EnvironmentVariableEditRequest(variable.EnvironmentVariableEditRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewUpdateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
		}
	}

	for _, variable := range request.Create {
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariable(ctx, applicationID).
			EnvironmentVariableRequest(variable.EnvironmentVariableRequest).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Key, res, err)
		}
	}
	return nil
}

func (c *Client) updateApplicationEnvironmentVariableAliases(
	ctx context.Context,
	applicationID string,
	request EnvironmentVariablesDiff,
	environmentVariablesByName map[string]qovery.EnvironmentVariable,
) *apierrors.APIError {
	// Delete
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && res == nil || (err != nil && res != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentAliasVariable, variable.Id, res, err)
		}
	}

	// Update
	for _, variable := range request.Update {
		// If the variable alias value has been updated, it means it targets a new aliased variable.
		// So delete it firstly and re-create it
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && res == nil || (err != nil && res != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentAliasVariable, variable.Id, res, err)
		}
		// The alias variable value contains the name of the aliased variable
		aliasedVariableId := environmentVariablesByName[*(variable.Value)].Id
		_, res, err = c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariableAlias(ctx, applicationID, aliasedVariableId).
			Key(qovery.Key{Key: variable.Key}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentAliasVariable, variable.Key, res, err)
		}
	}

	// Create
	for _, variable := range request.Create {
		// The alias variable value contains the name of the aliased variable
		aliasedVariableId := environmentVariablesByName[*(variable.Value)].Id
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariableAlias(ctx, applicationID, aliasedVariableId).
			Key(qovery.Key{Key: variable.Key}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentAliasVariable, variable.Key, res, err)
		}
	}
	return nil
}

func (c *Client) updateApplicationEnvironmentVariableOverrides(
	ctx context.Context,
	applicationID string,
	request EnvironmentVariablesDiff,
	environmentVariablesByName map[string]qovery.EnvironmentVariable,
) *apierrors.APIError {
	// Delete
	for _, variable := range request.Delete {
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && res == nil || (err != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentOverrideVariable, variable.Id, res, err)
		}
	}

	// Update
	for _, variable := range request.Update {
		// If the variable override name has been updated, it means it targets a new overrided variable.
		// So delete it firstly and re-create it
		res, err := c.api.ApplicationEnvironmentVariableApi.
			DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil && res == nil || (err != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentOverrideVariable, variable.Id, res, err)
		}
		// The override variable key contains the name of the overridden variable
		overriddenVariableId := environmentVariablesByName[variable.Key].Id
		_, res, err = c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariableOverride(ctx, applicationID, overriddenVariableId).
			Value(qovery.Value{Value: variable.Value}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentOverrideVariable, variable.Key, res, err)
		}
	}

	// Create
	for _, variable := range request.Create {
		// The override variable key contains the name of the overridden variable
		overriddenVariableId := environmentVariablesByName[variable.Key].Id
		_, res, err := c.api.ApplicationEnvironmentVariableApi.
			CreateApplicationEnvironmentVariableOverride(ctx, applicationID, overriddenVariableId).
			Value(qovery.Value{Value: variable.Value}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationEnvironmentOverrideVariable, variable.Key, res, err)
		}
	}
	return nil
}

func (c *Client) updateApplicationSecretAliases(
	ctx context.Context,
	applicationID string,
	request SecretsDiff,
	secretsByName map[string]qovery.Secret,
) *apierrors.APIError {
	// Delete all aliases to remove
	for _, secret := range request.Delete {
		res, err := c.api.ApplicationSecretApi.
			DeleteApplicationSecret(ctx, applicationID, secret.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if (err != nil && res == nil) || (err != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationSecretAlias, secret.Id, res, err)
		}
	}

	// If the secret alias value has been updated, it means it targets a new aliased secret.
	// So delete it firstly and re-create it
	for _, secret := range request.Update {
		res, err := c.api.ApplicationSecretApi.
			DeleteApplicationSecret(ctx, applicationID, secret.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if (err != nil && res == nil) || (err != nil && res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationSecretAlias, secret.Id, res, err)
		}
	}
	for _, secret := range request.Update {
		// The alias secret value contains the name of the aliased secret
		aliasedSecretId := secretsByName[*(secret.Value)].Id
		_, res, err := c.api.ApplicationSecretApi.
			CreateApplicationSecretAlias(ctx, applicationID, aliasedSecretId).
			Key(qovery.Key{Key: secret.Key}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationSecretAlias, secret.Key, res, err)
		}
	}

	// Create new aliases
	for _, secret := range request.Create {
		// The alias secret value contains the name of the aliased secret
		aliasedSecretId := secretsByName[*(secret.Value)].Id
		_, res, err := c.api.ApplicationSecretApi.
			CreateApplicationSecretAlias(ctx, applicationID, aliasedSecretId).
			Key(qovery.Key{Key: secret.Key}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationSecretAlias, secret.Key, res, err)
		}
	}
	return nil
}

func (c *Client) updateApplicationSecretOverrides(
	ctx context.Context,
	applicationID string,
	request SecretsDiff,
	secretsByName map[string]qovery.Secret,
) *apierrors.APIError {
	// Delete all overrides to remove
	for _, secret := range request.Delete {
		res, err := c.api.ApplicationSecretApi.
			DeleteApplicationSecret(ctx, applicationID, secret.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil || (res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationSecretOverride, secret.Id, res, err)
		}
	}

	// If the secret override name has been updated, it means it targets a new overrided secret.
	// So delete it firstly and re-create it
	for _, secret := range request.Update {
		res, err := c.api.ApplicationSecretApi.
			DeleteApplicationSecret(ctx, applicationID, secret.Id).
			Execute()
		// if 404 then ignore (the higher scoped variable could have been deleted, deleting the current scope variable previously so 404 is normal)
		if err != nil || (res.StatusCode >= 400 && res.StatusCode != 404) {
			return apierrors.NewDeleteError(apierrors.APIResourceApplicationSecretOverride, secret.Id, res, err)
		}
	}
	for _, secret := range request.Update {
		// The override secret key contains the name of the overridden secret
		overriddenSecretId := secretsByName[secret.Key].Id
		_, res, err := c.api.ApplicationSecretApi.
			CreateApplicationSecretOverride(ctx, applicationID, overriddenSecretId).
			Value(qovery.Value{Value: secret.Value}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationSecretOverride, secret.Key, res, err)
		}
	}

	// Create new aliases
	for _, secret := range request.Create {
		// The override secret key contains the name of the overridden secret
		overriddenSecretId := secretsByName[secret.Key].Id
		_, res, err := c.api.ApplicationSecretApi.
			CreateApplicationSecretOverride(ctx, applicationID, overriddenSecretId).
			Value(qovery.Value{Value: secret.Value}).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceApplicationSecretOverride, secret.Key, res, err)
		}
	}
	return nil
}
