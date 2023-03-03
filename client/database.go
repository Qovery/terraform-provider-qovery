package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type DatabaseResponse struct {
	DatabaseResponse     *qovery.Database
	DatabaseStatus       *qovery.Status
	DatabaseCredentials  *qovery.Credentials
	DatabaseInternalHost string
	DeploymentStageId    string
}

type DatabaseCreateParams struct {
	DatabaseRequest   qovery.DatabaseRequest
	DeploymentStageId string
}

type DatabaseUpdateParams struct {
	DatabaseEditRequest qovery.DatabaseEditRequest
	DeploymentStageId   string
}

func (c *Client) CreateDatabase(ctx context.Context, environmentID string, params *DatabaseCreateParams) (*DatabaseResponse, *apierrors.APIError) {
	database, res, err := c.api.DatabasesApi.
		CreateDatabase(ctx, environmentID).
		DatabaseRequest(params.DatabaseRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, params.DatabaseRequest.Name, res, err)
	}

	// Attach database to deployment stage
	if len(params.DeploymentStageId) > 0 {
		_, response, err := c.api.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, params.DeploymentStageId, database.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, params.DeploymentStageId, response, err)
		}
	}

	// Get database deployment stage
	deploymentStage, resp, err := c.api.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, database.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, database.Id, resp, err)
	}

	return c.updateDatabase(ctx, database, deploymentStage.Id)
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

	credentials, apiErr := c.GetDatabaseCredentials(ctx, databaseID)
	if apiErr != nil {
		return nil, apiErr
	}

	hostInternal, apiErr := c.getDatabaseHostInternal(ctx, database)
	if apiErr != nil {
		return nil, apiErr
	}

	// Get database deployment stage
	deploymentStage, resp, err := c.api.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, database.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, database.Id, resp, err)
	}

	return &DatabaseResponse{
		DatabaseResponse:     database,
		DatabaseStatus:       status,
		DatabaseCredentials:  credentials,
		DatabaseInternalHost: hostInternal,
		DeploymentStageId:    deploymentStage.Id,
	}, nil
}

func (c *Client) getDatabaseHostInternal(ctx context.Context, database *qovery.Database) (string, *apierrors.APIError) {
	environmentVariables, apiErr := c.getEnvironmentBuiltInEnvironmentVariables(ctx, database.Environment.Id)
	if apiErr != nil {
		return "", apiErr
	}

	// Get all environment variables associated to this database,
	// and pick only the elements that I need to construct my struct below
	// Context: since I need to get the internal host of my database and this information is only available via the environment env vars,
	// then we list all env vars from the environment where the database is to take it.
	// FIXME - it's a really bad idea of doing that but I have no choice... If we change the way we structure environment variable backend side, then we will be f***ed up :/
	hostInternalKey := fmt.Sprintf("QOVERY_%s_Z%s_HOST_INTERNAL", database.Type, strings.ToUpper(strings.Split(database.Id, "-")[0]))
	// Expected host internal key syntax is `QOVERY_{DB-TYPE}_Z{DB-ID}_HOST_INTERNAL`
	hostInternal := ""
	for _, env := range environmentVariables {
		if env.Key == hostInternalKey {
			hostInternal = env.Value
			break
		}
	}

	return hostInternal, nil
}

func (c *Client) GetDatabaseCredentials(ctx context.Context, databaseID string) (*qovery.Credentials, *apierrors.APIError) {
	credentials, res, err := c.api.DatabaseMainCallsApi.
		GetDatabaseMasterCredentials(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	return credentials, nil
}

func (c *Client) UpdateDatabase(ctx context.Context, databaseID string, params *DatabaseUpdateParams) (*DatabaseResponse, *apierrors.APIError) {
	database, res, err := c.api.DatabaseMainCallsApi.
		EditDatabase(ctx, databaseID).
		DatabaseEditRequest(params.DatabaseEditRequest).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceDatabase, databaseID, res, err)
	}
	// Attach database to deployment stage
	if len(params.DeploymentStageId) > 0 {
		_, response, err := c.api.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, params.DeploymentStageId, database.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, params.DeploymentStageId, response, err)
		}
	}

	// Get database deployment stage
	deploymentStage, resp, err := c.api.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, database.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateError(apierrors.APIResourceDatabase, database.Id, resp, err)
	}

	return c.updateDatabase(ctx, database, deploymentStage.Id)
}

func (c *Client) DeleteDatabase(ctx context.Context, databaseID string) *apierrors.APIError {
	database, res, err := c.api.DatabaseMainCallsApi.
		GetDatabase(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		if res.StatusCode == 404 {
			// if the database is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	envChecker := newEnvironmentFinalStateCheckerWaitFunc(c, database.Environment.Id)
	if apiErr := wait(ctx, envChecker, nil); apiErr != nil {
		return apiErr
	}

	res, err = c.api.DatabaseMainCallsApi.
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

func (c *Client) updateDatabase(ctx context.Context, database *qovery.Database, deploymentStageId string) (*DatabaseResponse, *apierrors.APIError) {
	credentials, apiErr := c.GetDatabaseCredentials(ctx, database.Id)
	if apiErr != nil {
		return nil, apiErr
	}

	hostInternal, apiErr := c.getDatabaseHostInternal(ctx, database)
	if apiErr != nil {
		return nil, apiErr
	}

	return &DatabaseResponse{
		DatabaseResponse:     database,
		DatabaseCredentials:  credentials,
		DatabaseInternalHost: hostInternal,
		DeploymentStageId:    deploymentStageId,
	}, nil
}

func (c *Client) deployDatabase(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	status, apiErr := c.getDatabaseStatus(ctx, databaseID)
	if apiErr != nil {
		return nil, apiErr
	}

	switch status.State {
	case qovery.STATEENUM_RUNNING:
		return status, nil
	case qovery.STATEENUM_DEPLOYMENT_ERROR:
		return c.redeployDatabase(ctx, databaseID)
	default:
		_, res, err := c.api.DatabaseActionsApi.
			DeployDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewDeployError(apierrors.APIResourceDatabase, databaseID, res, err)
		}
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, qovery.STATEENUM_RUNNING)
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
	case qovery.STATEENUM_STOPPED:
		return status, nil
	default:
		_, res, err := c.api.DatabaseActionsApi.
			StopDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewStopError(apierrors.APIResourceDatabase, databaseID, res, err)
		}
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, qovery.STATEENUM_STOPPED)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getDatabaseStatus(ctx, databaseID)
}

func (c *Client) redeployDatabase(ctx context.Context, databaseID string) (*qovery.Status, *apierrors.APIError) {
	finalStateChecker := newDatabaseFinalStateCheckerWaitFunc(c, databaseID)
	if apiErr := wait(ctx, finalStateChecker, nil); apiErr != nil {
		return nil, apiErr
	}

	_, res, err := c.api.DatabaseActionsApi.
		RedeployDatabase(ctx, databaseID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewRedeployError(apierrors.APIResourceDatabase, databaseID, res, err)
	}

	statusChecker := newDatabaseStatusCheckerWaitFunc(c, databaseID, qovery.STATEENUM_RUNNING)
	if apiErr := wait(ctx, statusChecker, nil); apiErr != nil {
		return nil, apiErr
	}
	return c.getDatabaseStatus(ctx, databaseID)
}
