package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

// Ensure jobDeploymentQoveryAPI defined types fully satisfy the deployment.Repository interface.
var _ deployment.Repository = jobDeploymentQoveryAPI{}

// jobDeploymentQoveryAPI implements the interface deployment.Repository.
type jobDeploymentQoveryAPI struct {
	client *qovery.APIClient
}

// newJobDeploymentQoveryAPI return a new instance of a deployment.Repository that uses Qovery's API.
func newJobDeploymentQoveryAPI(client *qovery.APIClient) (deployment.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &jobDeploymentQoveryAPI{
		client: client,
	}, nil
}

// GetStatus calls Qovery's API to get the status of a job using the given jobID.
func (c jobDeploymentQoveryAPI) GetStatus(ctx context.Context, jobID string) (*status.Status, error) {
	jobStatus, resp, err := c.client.JobMainCallsApi.
		GetJobStatus(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceJobStatus, jobID, resp, err)
	}

	return newDomainStatusFromQovery(jobStatus)
}

// Deploy calls Qovery's API to deploy a job using the given jobID.
func (c jobDeploymentQoveryAPI) Deploy(ctx context.Context, jobID string, version string) (*status.Status, error) {
	// TODO(benjaminch): to be checked because we should be able to pass a commit ID
	jobStatus, resp, err := c.client.JobActionsApi.
		DeployJob(ctx, jobID).
		JobDeployRequest(qovery.JobDeployRequest{
			ImageTag: &version,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewDeployApiError(apierrors.ApiResourceJob, jobID, resp, err)
	}

	return newDomainStatusFromQovery(jobStatus)
}

// Redeploy calls Qovery's API to redeploy a job using the given jobID.
func (c jobDeploymentQoveryAPI) Redeploy(ctx context.Context, jobID string) (*status.Status, error) {
	jobStatus, resp, err := c.client.JobActionsApi.
		RedeployJob(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewRedeployApiError(apierrors.ApiResourceJob, jobID, resp, err)
	}

	return newDomainStatusFromQovery(jobStatus)
}

// Stop calls Qovery's API to stop a job using the given jobID.
func (c jobDeploymentQoveryAPI) Stop(ctx context.Context, jobID string) (*status.Status, error) {
	jobStatus, resp, err := c.client.JobActionsApi.
		StopJob(ctx, jobID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewStopApiError(apierrors.ApiResourceJob, jobID, resp, err)
	}

	return newDomainStatusFromQovery(jobStatus)
}
