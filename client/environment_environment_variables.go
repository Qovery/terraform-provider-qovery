package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getEnvironmentEnvironmentVariables(ctx context.Context, environmentID string) ([]*qovery.EnvironmentVariable, *apierrors.APIError) {
	vars, res, err := c.api.EnvironmentVariableApi.
		ListEnvironmentEnvironmentVariable(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceEnvironmentEnvironmentVariable, environmentID, res, err)
	}
	return environmentVariableResponseListToArray(vars), nil
}

func (c *Client) updateEnvironmentEnvironmentVariables(ctx context.Context, environment *qovery.Environment, request EnvironmentVariablesDiff) *apierrors.APIError {
	for _, variable := range request.Delete {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.deleteEnvironmentEnvironmentVariable(ctx, environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.deleteProjectEnvironmentVariable(ctx, environment.Project.Id, variable); err != nil {
				return err
			}
		}
	}

	for _, variable := range request.Update {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.editEnvironmentEnvironmentVariable(ctx, environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.editProjectEnvironmentVariable(ctx, environment.Project.Id, variable); err != nil {
				return err
			}
		}
	}

	for _, variable := range request.Create {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.createEnvironmentEnvironmentVariable(ctx, environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.createProjectEnvironmentVariable(ctx, environment.Project.Id, variable); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) createEnvironmentEnvironmentVariable(ctx context.Context, environmentID string, variable EnvironmentVariableCreateRequest) *apierrors.APIError {
	_, res, err := c.api.EnvironmentVariableApi.
		CreateEnvironmentEnvironmentVariable(ctx, environmentID).
		EnvironmentVariableRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewCreateError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Key, res, err)
	}
	return nil
}

func (c *Client) editEnvironmentEnvironmentVariable(ctx context.Context, environmentID string, variable EnvironmentVariableUpdateRequest) *apierrors.APIError {
	_, res, err := c.api.EnvironmentVariableApi.
		EditEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id).
		EnvironmentVariableEditRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}

func (c *Client) deleteEnvironmentEnvironmentVariable(ctx context.Context, environmentID string, variable EnvironmentVariableDeleteRequest) *apierrors.APIError {
	res, err := c.api.EnvironmentVariableApi.
		DeleteEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewDeleteError(apierrors.APIResourceEnvironmentEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}
