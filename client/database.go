package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

type DatabaseResponse struct {
	DatabaseResponse *qovery.DatabaseResponse
	DatabaseStatus   *qovery.Status
}

type DatabaseCreateParams struct {
	DatabaseRequest qovery.DatabaseRequest
	DesiredState    string
}

type DatabaseUpdateParams struct {
	DatabaseEditRequest qovery.DatabaseEditRequest
	DesiredState        string
}

func (c *Client) CreateDatabase(ctx context.Context, environmentID string, params DatabaseCreateParams) (*DatabaseResponse, *apierrors.APIError) {
	database, res, err := c.api.DatabasesApi.
		CreateDatabase(ctx, environmentID).
		DatabaseRequest(params.DatabaseRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, params.DatabaseRequest.Name, res, err)
	}
	return c.updateDatabase(ctx, database, params.DesiredState)
}

func (c *Client) GetDatabase(ctx context.Context, databaseID string) (*DatabaseResponse, *apierrors.APIError) {
	database, res, err := c.api.DatabaseMainCallsApi.
		GetDatabase(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	status, apiErr := c.getDatabaseStatus(ctx, databaseID)
	if apiErr != nil {
		return nil, apiErr
	}

	return &DatabaseResponse{
		DatabaseResponse: database,
		DatabaseStatus:   status,
	}, nil
}

func (c *Client) UpdateDatabase(ctx context.Context, databaseID string, params DatabaseUpdateParams) (*DatabaseResponse, *apierrors.APIError) {
	database, res, err := c.api.DatabaseMainCallsApi.
		EditDatabase(ctx, databaseID).
		DatabaseEditRequest(params.DatabaseEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceDatabase, databaseID, res, err)
	}
	// FIXME restart the database if the configuration has changed
	return c.updateDatabase(ctx, database, params.DesiredState)
}

func (c *Client) DeleteDatabase(ctx context.Context, databaseID string) *apierrors.APIError {
	res, err := c.api.DatabaseMainCallsApi.
		DeleteDatabase(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		return apierrors.NewDeleteError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	checker := newDatabaseStatusCheckerWaitFunc(c, databaseID, "DELETED")
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return apiErr
	}
	return nil
}

func (c *Client) updateDatabase(ctx context.Context, database *qovery.DatabaseResponse, desiredState string) (*DatabaseResponse, *apierrors.APIError) {
	status, apiErr := c.updateDatabaseStatus(ctx, database, desiredState)
	if apiErr != nil {
		return nil, apiErr
	}

	return &DatabaseResponse{
		DatabaseResponse: database,
		DatabaseStatus:   status,
	}, nil
}

func (c *Client) deployDatabase(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getDatabaseStatus(ctx, databaseID)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case databaseStateRunning:
		return status, nil
	case "DEPLOYMENT_ERROR":
		return c.restartDatabase(ctx, databaseID)
	default:
		_, res, err := c.api.DatabaseActionsApi.
			DeployDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceDatabase, databaseID, res, err)
		}
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, databaseStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getDatabaseStatus(ctx, databaseID)
}

func (c *Client) stopDatabase(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getDatabaseStatus(ctx, databaseID)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case databaseStateStopped:
		return status, nil
	default:
		_, res, err := c.api.DatabaseActionsApi.
			StopDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceDatabase, databaseID, res, err)
		}
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, databaseStateStopped)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getDatabaseStatus(ctx, databaseID)
}

func (c *Client) restartDatabase(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	finalStateChecker := newDatabaseFinalStateCheckerWaitFunc(c, databaseID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.DatabaseActionsApi.
		RestartDatabase(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRestartError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, databaseStateRunning)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getDatabaseStatus(ctx, databaseID)
}
