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
	return environmentVariableResponseListToArray(applicationVariables), nil
}

func (c *Client) updateApplicationEnvironmentVariables(ctx context.Context, application *qovery.Application, request EnvironmentVariablesDiff) *apierrors.APIError {
	// We need to get the project id to be able to update its environment variables
	environment, err := c.getEnvironment(ctx, application.Environment.Id)
	if err != nil {
		return err
	}

	project, err := c.getProject(ctx, environment.Project.Id)
	if err != nil {
		return err
	}

	for _, variable := range request.Delete {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_APPLICATION:
			if err := c.deleteApplicationEnvironmentVariable(ctx, application.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.deleteEnvironmentEnvironmentVariable(ctx, application.Environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.deleteProjectEnvironmentVariable(ctx, project.Id, variable); err != nil {
				return err
			}
		}
	}

	for _, variable := range request.Update {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_APPLICATION:
			if err := c.editApplicationEnvironmentVariable(ctx, application.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.editEnvironmentEnvironmentVariable(ctx, application.Environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.editProjectEnvironmentVariable(ctx, project.Id, variable); err != nil {
				return err
			}
		}
	}

	for _, variable := range request.Create {
		switch variable.Scope {
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_APPLICATION:
			if err := c.createApplicationEnvironmentVariable(ctx, application.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_ENVIRONMENT:
			if err := c.createEnvironmentEnvironmentVariable(ctx, application.Environment.Id, variable); err != nil {
				return err
			}
		case qovery.ENVIRONMENTVARIABLESCOPEENUM_PROJECT:
			if err := c.createProjectEnvironmentVariable(ctx, project.Id, variable); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) createApplicationEnvironmentVariable(ctx context.Context, applicationID string, variable EnvironmentVariableCreateRequest) *apierrors.APIError {
	_, res, err := c.api.ApplicationEnvironmentVariableApi.
		CreateApplicationEnvironmentVariable(ctx, applicationID).
		EnvironmentVariableRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Key, res, err)
	}
	return nil
}

func (c *Client) editApplicationEnvironmentVariable(ctx context.Context, applicationID string, variable EnvironmentVariableUpdateRequest) *apierrors.APIError {
	_, res, err := c.api.ApplicationEnvironmentVariableApi.
		EditApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
		EnvironmentVariableEditRequest(variable.toRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewUpdateError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}

func (c *Client) deleteApplicationEnvironmentVariable(ctx context.Context, applicationID string, variable EnvironmentVariableDeleteRequest) *apierrors.APIError {
	res, err := c.api.ApplicationEnvironmentVariableApi.
		DeleteApplicationEnvironmentVariable(ctx, applicationID, variable.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return apierrors.NewDeleteError(apierrors.APIResourceApplicationEnvironmentVariable, variable.Id, res, err)
	}
	return nil
}
