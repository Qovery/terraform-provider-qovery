package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// Ensure containerDeploymentQoveryAPI defined types fully satisfy the deployment.Repository interface.
var _ deployment.Repository = containerDeploymentQoveryAPI{}

// containerDeploymentQoveryAPI implements the interface deployment.Repository.
type containerDeploymentQoveryAPI struct {
	client *qovery.APIClient
}

// newContainerDeploymentQoveryAPI return a new instance of a deployment.Repository that uses Qovery's API.
func newContainerDeploymentQoveryAPI(client *qovery.APIClient) (deployment.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &containerDeploymentQoveryAPI{
		client: client,
	}, nil
}

// GetStatus calls Qovery's API to get the status of a container using the given containerID.
func (c containerDeploymentQoveryAPI) GetStatus(ctx context.Context, containerID string) (*status.Status, error) {
	containerStatus, resp, err := c.client.ContainerMainCallsApi.
		GetContainerStatus(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceContainerStatus, containerID, resp, err)
	}

	return newDomainStatusFromQovery(containerStatus)
}

// Deploy calls Qovery's API to deploy a container using the given containerID.
func (c containerDeploymentQoveryAPI) Deploy(ctx context.Context, containerID string, imageTag string) (*status.Status, error) {
	containerStatus, resp, err := c.client.ContainerActionsApi.
		DeployContainer(ctx, containerID).
		ContainerDeployRequest(qovery.ContainerDeployRequest{
			ImageTag: imageTag,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewDeployApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	return newDomainStatusFromQovery(containerStatus)
}

// Redeploy calls Qovery's API to redeploy a container using the given containerID.
func (c containerDeploymentQoveryAPI) Redeploy(ctx context.Context, containerID string) (*status.Status, error) {
	containerStatus, resp, err := c.client.ContainerActionsApi.
		RedeployContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewRedeployApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	return newDomainStatusFromQovery(containerStatus)
}

// Stop calls Qovery's API to stop a container using the given containerID.
func (c containerDeploymentQoveryAPI) Stop(ctx context.Context, containerID string) (*status.Status, error) {
	containerStatus, resp, err := c.client.ContainerActionsApi.
		StopContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewStopApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	return newDomainStatusFromQovery(containerStatus)
}
