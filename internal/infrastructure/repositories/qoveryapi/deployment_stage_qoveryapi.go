package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

type deploymentStageQoveryAPI struct {
	client *qovery.APIClient
}

func newDeploymentStageQoveryAPI(client *qovery.APIClient) (deploymentstage.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &deploymentStageQoveryAPI{
		client: client,
	}, nil
}

func (c deploymentStageQoveryAPI) Create(ctx context.Context, environmentId string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStageCreated, resp, err := c.client.DeploymentStageMainCallsApi.
		CreateEnvironmentDeploymentStage(ctx, environmentId).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeploymentStage, request.Name, resp, err)
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStageCreated.Id,
		EnvironmentID:     deploymentStageCreated.Environment.Id,
		Name:              *deploymentStageCreated.Name,
		Description:       *deploymentStageCreated.Description,
	})
}

func (c deploymentStageQoveryAPI) Get(ctx context.Context, environmentId string, deploymentStageID string) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetDeploymentStage(ctx, deploymentStageID).Execute()
	if deploymentStage == nil {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
	})
}

func (c deploymentStageQoveryAPI) Update(ctx context.Context, deploymentStageID string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.
		EditDeploymentStage(ctx, deploymentStageID).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
	})
}

func (c deploymentStageQoveryAPI) Delete(ctx context.Context, deploymentStageID string) error {
	resp, err := c.client.DeploymentStageMainCallsApi.
		DeleteDeploymentStage(ctx, deploymentStageID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return nil
}
