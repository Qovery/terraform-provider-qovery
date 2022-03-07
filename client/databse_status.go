package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/client/apierrors"
)

const (
	databaseStateRunning = "RUNNING"
	databaseStateStopped = "STOPPED"
)

func (c *Client) GetDatabaseStatus(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	status, res, err := c.api.DatabaseMainCallsApi.
		GetDatabaseStatus(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceDatabaseStatus, databaseID, res, err)
	}
	return status, nil
}

func (c *Client) updateDatabaseStatus(ctx context.Context, database *qovery.DatabaseResponse, desiredState string) (*qovery.Status, *apierrors.APIError) {
	// wait until we can stop the database - otherwise it will fail
	checker := newDatabaseFinalStateCheckerWaitFunc(c, database.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	status, apiErr := c.GetDatabaseStatus(ctx, database.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	if status.State != desiredState {
		switch desiredState {
		case databaseStateRunning:
			return c.deployDatabase(ctx, database.Id)
		case databaseStateStopped:
			return c.stopDatabase(ctx, database.Id)
		}
	}

	deploymentStatus := status.ServiceDeploymentStatus.Get()
	if deploymentStatus != nil && *deploymentStatus == "OUT_OF_DATE" {
		return c.restartDatabase(ctx, database.Id)
	}

	return status, nil
}
