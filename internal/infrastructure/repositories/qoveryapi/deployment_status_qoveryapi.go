package qoveryapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

type deploymentStatusQoveryAPI struct {
	client *qovery.APIClient
}

func newDeploymentStatusQoveryAPI(client *qovery.APIClient) (newdeployment.DeploymentStatusRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &deploymentStatusQoveryAPI{
		client: client,
	}, nil
}

func (d deploymentStatusQoveryAPI) WaitForTerminatedState(ctx context.Context, environmentId uuid.UUID) error {
	checkEnvironmentStatus := d.newEnvironmentWaitForTerminalStateBeforeDeploying(environmentId)
	err := waitWithDefaultTimeout(ctx, checkEnvironmentStatus)
	if err != nil {
		return err
	}

	return nil
}

func (d deploymentStatusQoveryAPI) WaitForExpectedDesiredState(ctx context.Context, newDeployment newdeployment.Deployment) error {
	checkEnvironmentStatus := d.newEnvironmentWaitForExpectedDesiredState(*newDeployment.EnvironmentId, newDeployment.DesiredState)
	err := waitWithDefaultTimeout(ctx, checkEnvironmentStatus)
	if err != nil {
		return err
	}

	return nil
}

type waitFunc func(ctx context.Context) (bool, error)

func waitWithDefaultTimeout(ctx context.Context, f waitFunc) error {
	defaultWaitTimeout := 4 * time.Hour
	return wait(ctx, f, &defaultWaitTimeout)
}

func wait(ctx context.Context, f waitFunc, timeout *time.Duration) error {
	// Run the function once before waiting
	ok, err := f(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ticker := time.NewTicker(10 * time.Second)
	timeoutTicker := time.NewTicker(*timeout)

	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case <-ticker.C:
			ok, apiErr := f(ctx)
			if apiErr != nil {
				return apiErr
			}
			if ok {
				return nil
			}
		}
	}
}

func (d deploymentStatusQoveryAPI) newEnvironmentWaitForTerminalStateBeforeDeploying(environmentId uuid.UUID) waitFunc {
	return func(ctx context.Context) (bool, error) {
		status, response, err := d.client.EnvironmentMainCallsApi.GetEnvironmentStatus(ctx, environmentId.String()).Execute()
		if err != nil || response.StatusCode >= 400 {
			return false, err
		}

		switch status.State {
		// In progress
		case "BUILDING", "CANCELING", "DELETE_QUEUED", "DELETING", "DEPLOYING", "STOPPING",
			"STOP_QUEUED", "RESTART_QUEUED", "RESTARTING", "DEPLOYMENT_QUEUED", "QUEUED":
			tflog.Info(ctx, fmt.Sprintf("Environment deployment in progress with current status %s...", status.State))
			return false, nil
		// Finished with error
		case "READY", "DEPLOYMENT_ERROR", "DELETE_ERROR", "STOP_ERROR", "RESTART_ERROR", "RUNNING",
			"STOPPED", "DEPLOYED", "DELETED", "RESTARTED", "CANCELED":
			return true, nil
		}

		// Unexpected status
		return false, errors.New(fmt.Sprintf("Unexpected deployment status having status: %s", status.State))
	}
}

func (d deploymentStatusQoveryAPI) newEnvironmentWaitForExpectedDesiredState(environmentId uuid.UUID, desiredState newdeployment.DeploymentDesiredState) waitFunc {
	return func(ctx context.Context) (bool, error) {
		status, response, err := d.client.EnvironmentMainCallsApi.GetEnvironmentStatus(ctx, environmentId.String()).Execute()
		if err != nil {
			if response.StatusCode == 404 && desiredState == newdeployment.DELETED {
				return true, nil
			}
			return false, err
		}

		switch status.State {
		// In progress
		case "BUILDING", "CANCELING", "DELETE_QUEUED", "DELETING", "DEPLOYING", "STOPPING",
			"STOP_QUEUED", "RESTART_QUEUED", "RESTARTING", "DEPLOYMENT_QUEUED", "QUEUED", "READY":
			tflog.Info(ctx, fmt.Sprintf("Environment deployment in progress with target status %s...", desiredState))
			return false, nil
		// Finished with error
		case "DEPLOYMENT_ERROR", "DELETE_ERROR", "STOP_ERROR", "RESTART_ERROR":
			return false, errors.New(fmt.Sprintf("Environment deployment failed with final status: %s", status.State))
		// Finished with success
		case "RUNNING", "STOPPED", "DEPLOYED", "DELETED", "RESTARTED", "CANCELED":
			return true, nil
		}

		// Unexpected status
		return false, errors.New(fmt.Sprintf("Unexpected deployment status having status: %s", status.State))
	}
}
