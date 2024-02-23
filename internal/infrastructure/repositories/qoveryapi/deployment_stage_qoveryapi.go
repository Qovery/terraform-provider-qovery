package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
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

func (c deploymentStageQoveryAPI) Create(ctx context.Context, environmentID string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStageCreated, resp, err := c.client.DeploymentStageMainCallsAPI.
		CreateEnvironmentDeploymentStage(ctx, environmentID).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
	}

	if request.IsAfter != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveAfterDeploymentStage(ctx, deploymentStageCreated.Id, *request.IsAfter).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	if request.IsBefore != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveBeforeDeploymentStage(ctx, deploymentStageCreated.Id, *request.IsBefore).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStageCreated.Id,
		EnvironmentID:     deploymentStageCreated.Environment.Id,
		Name:              *deploymentStageCreated.Name,
		Description:       *deploymentStageCreated.Description,
		IsAfter:           request.IsAfter,
		IsBefore:          request.IsBefore,
	})
}

func (c deploymentStageQoveryAPI) Get(ctx context.Context, environmentID string, deploymentStageID string) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetDeploymentStage(ctx, deploymentStageID).Execute()
	if deploymentStage == nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
	})
}

func (c deploymentStageQoveryAPI) GetAllByEnvironmentID(ctx context.Context, environmentID string) (*[]deploymentstage.DeploymentStage, error) {
	result, resp, err := c.client.DeploymentStageMainCallsAPI.ListEnvironmentDeploymentStage(ctx, environmentID).Execute()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 200 {
		return nil, errors.New("Wrong environment id")
	}

	var array []deploymentstage.DeploymentStage
	for _, deploymentStage := range result.Results {
		stage, err := deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
			DeploymentStageID: deploymentStage.Id,
			EnvironmentID:     deploymentStage.Environment.Id,
			Name:              *deploymentStage.Name,
			Description:       *deploymentStage.Description,
		})
		if err == nil {
			array = append(array, *stage)
		}
	}

	return &array, nil
}

func (c deploymentStageQoveryAPI) Update(ctx context.Context, deploymentStageID string, request deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.
		EditDeploymentStage(ctx, deploymentStageID).
		DeploymentStageRequest(qovery.DeploymentStageRequest{
			Name:        request.Name,
			Description: *qovery.NewNullableString(&request.Description),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	if request.IsAfter != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveAfterDeploymentStage(ctx, deploymentStageID, *request.IsAfter).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	if request.IsBefore != nil {
		_, resp, err = c.client.DeploymentStageMainCallsAPI.
			MoveBeforeDeploymentStage(ctx, deploymentStageID, *request.IsBefore).
			Execute()
		if err != nil || resp.StatusCode >= 400 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceDeploymentStage, request.Name, resp, err)
		}
	}

	return deploymentstage.NewDeploymentStage(deploymentstage.NewDeploymentStageParams{
		DeploymentStageID: deploymentStage.Id,
		EnvironmentID:     deploymentStage.Environment.Id,
		Name:              *deploymentStage.Name,
		Description:       *deploymentStage.Description,
		IsAfter:           request.IsAfter,
		IsBefore:          request.IsBefore,
	})
}

func (c deploymentStageQoveryAPI) Delete(ctx context.Context, deploymentStageID string) error {
	_, resp, err := c.client.DeploymentStageMainCallsAPI.GetDeploymentStage(ctx, deploymentStageID).Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// if the deployment stage is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewReadAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	resp, err = c.client.DeploymentStageMainCallsAPI.
		DeleteDeploymentStage(ctx, deploymentStageID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceDeploymentStage, deploymentStageID, resp, err)
	}

	return nil
}
