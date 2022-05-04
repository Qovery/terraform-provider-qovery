package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func (c *Client) getDatabaseStatus(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	status, res, err := c.api.DatabaseMainCallsApi.
		GetDatabaseStatus(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceDatabaseStatus, databaseID, res, err)
	}

	// Handle READY as STOPPED state
	if status.State == qovery.STATEENUM_READY {
		status.State = qovery.STATEENUM_STOPPED
	}
	return status, nil
}

func (c *Client) updateDatabaseStatus(ctx context.Context, database *qovery.Database, desiredState qovery.StateEnum) (*qovery.Status, *apierrors.APIError) {
	// wait until we can stop the database - otherwise it will fail
	checker := newDatabaseFinalStateCheckerWaitFunc(c, database.Id)
	if apiErr := wait(ctx, checker, nil); apiErr != nil {
		return nil, apiErr
	}

	status, apiErr := c.getDatabaseStatus(ctx, database.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	if status.State != desiredState {
		switch desiredState {
		case qovery.STATEENUM_RUNNING:
			return c.deployDatabase(ctx, database.Id)
		case qovery.STATEENUM_STOPPED:
			return c.stopDatabase(ctx, database.Id)
		}
	}

	deploymentStatus := status.ServiceDeploymentStatus.Get()
	if deploymentStatus != nil && *deploymentStatus == qovery.SERVICEDEPLOYMENTSTATUSENUM_OUT_OF_DATE {
		return c.restartDatabase(ctx, database.Id)
	}

	return status, nil
}
