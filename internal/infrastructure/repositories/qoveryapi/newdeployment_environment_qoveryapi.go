package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

type newNewDeploymentQoveryAPI struct {
	client *qovery.APIClient
}

func newDeploymentEnvironmentQoveryAPI(client *qovery.APIClient) (newdeployment.EnvironmentRepository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &newNewDeploymentQoveryAPI{
		client: client,
	}, nil
}

func (c newNewDeploymentQoveryAPI) Deploy(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	_, resp, err := c.client.EnvironmentActionsAPI.DeployEnvironment(ctx, newDeployment.EnvironmentID.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeployment, newDeployment.EnvironmentID.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) ReDeploy(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return c.Deploy(ctx, newDeployment)
}

func (c newNewDeploymentQoveryAPI) Stop(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	_, resp, err := c.client.EnvironmentActionsAPI.StopEnvironment(ctx, newDeployment.EnvironmentID.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeployment, newDeployment.EnvironmentID.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) Restart(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	_, resp, err := c.client.EnvironmentActionsAPI.RedeployEnvironment(ctx, newDeployment.EnvironmentID.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeployment, newDeployment.EnvironmentID.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) Delete(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	resp, err := c.client.EnvironmentMainCallsAPI.DeleteEnvironment(ctx, newDeployment.EnvironmentID.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeployment, newDeployment.EnvironmentID.String(), resp, err)
	}

	return &newDeployment, nil
}
