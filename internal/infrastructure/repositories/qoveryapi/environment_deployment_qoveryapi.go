package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// Ensure environmentDeploymentQoveryAPI defined types fully satisfy the deployment.Repository interface.
var _ deployment.Repository = environmentDeploymentQoveryAPI{}

// environmentDeploymentQoveryAPI implements the interface deployment.Repository.
type environmentDeploymentQoveryAPI struct {
	client *qovery.APIClient
}

// newEnvironmentDeploymentQoveryAPI return a new instance of a deployment.Repository that uses Qovery's API.
func newEnvironmentDeploymentQoveryAPI(client *qovery.APIClient) (deployment.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &environmentDeploymentQoveryAPI{
		client: client,
	}, nil
}

// GetStatus calls Qovery's API to get the status of an environment using the given environmentID.
func (c environmentDeploymentQoveryAPI) GetStatus(ctx context.Context, environmentID string) (*status.Status, error) {
	environmentStatus, resp, err := c.client.EnvironmentMainCallsApi.
		GetEnvironmentStatus(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceEnvironmentStatus, environmentID, resp, err)
	}

	return newDomainEnvironmentStatusFromQovery(environmentStatus)
}

// Deploy calls Qovery's API to deploy an environment using the given environmentID.
func (c environmentDeploymentQoveryAPI) Deploy(ctx context.Context, environmentID string, imageTag string) (*status.Status, error) {
	environmentStatus, resp, err := c.client.EnvironmentActionsApi.
		DeployEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewDeployApiError(apierrors.ApiResourceEnvironment, environmentID, resp, err)
	}

	return newDomainStatusFromQovery(environmentStatus)
}

// Redeploy calls Qovery's API to redeploy an environment using the given environmentID.
func (c environmentDeploymentQoveryAPI) Redeploy(ctx context.Context, environmentID string) (*status.Status, error) {
	_, resp, err := c.client.EnvironmentActionsApi.
		RedeployEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewRedeployApiError(apierrors.ApiResourceEnvironment, environmentID, resp, err)
	}

	return c.GetStatus(ctx, environmentID)
}

// Stop calls Qovery's API to stop an environment using the given environmentID.
func (c environmentDeploymentQoveryAPI) Stop(ctx context.Context, environmentID string) (*status.Status, error) {
	environmentStatus, resp, err := c.client.EnvironmentActionsApi.
		StopEnvironment(ctx, environmentID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewStopApiError(apierrors.ApiResourceEnvironment, environmentID, resp, err)
	}

	return newDomainEnvironmentStatusFromQovery(environmentStatus)
}
