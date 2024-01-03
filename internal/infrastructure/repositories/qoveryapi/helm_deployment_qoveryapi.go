package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

var _ deployment.Repository = helmDeploymentQoveryAPI{}

// helmDeploymentQoveryAPI implements the interface deployment.Repository.
type helmDeploymentQoveryAPI struct {
	client *qovery.APIClient
}

func newHelmDeploymentQoveryAPI(client *qovery.APIClient) (deployment.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &helmDeploymentQoveryAPI{
		client: client,
	}, nil
}

func (h helmDeploymentQoveryAPI) GetStatus(ctx context.Context, helmID string) (*status.Status, error) {
	helmStatus, resp, err := h.client.HelmMainCallsAPI.
		GetHelmStatus(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelmStatus, helmID, resp, err)
	}

	return newDomainStatusFromQovery(helmStatus)
}

func (h helmDeploymentQoveryAPI) Deploy(ctx context.Context, helmID string, version string) (*status.Status, error) {
	helmStatus, resp, err := h.client.HelmActionsAPI.
		DeployHelm(ctx, helmID).
		HelmDeployRequest(qovery.HelmDeployRequest{
			ChartVersion:              &version,
			GitCommitId:               nil,
			ValuesOverrideGitCommitId: nil,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewDeployAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	return newDomainStatusFromQovery(helmStatus)
}

func (h helmDeploymentQoveryAPI) Redeploy(ctx context.Context, helmID string) (*status.Status, error) {
	helmStatus, resp, err := h.client.HelmActionsAPI.
		DeployHelm(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewRedeployAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	return newDomainStatusFromQovery(helmStatus)
}

func (h helmDeploymentQoveryAPI) Stop(ctx context.Context, helmID string) (*status.Status, error) {
	helmStatus, resp, err := h.client.HelmActionsAPI.
		StopHelm(ctx, helmID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewStopAPIError(apierrors.APIResourceHelm, helmID, resp, err)
	}

	return newDomainStatusFromQovery(helmStatus)
}
