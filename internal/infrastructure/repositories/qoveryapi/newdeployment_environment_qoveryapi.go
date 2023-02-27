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
	_, resp, err := c.client.EnvironmentActionsApi.DeployEnvironment(ctx, newDeployment.EnvironmentId.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeployment, newDeployment.EnvironmentId.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) ReDeploy(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	return c.Deploy(ctx, newDeployment)
}

func (c newNewDeploymentQoveryAPI) Stop(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	_, resp, err := c.client.EnvironmentActionsApi.StopEnvironment(ctx, newDeployment.EnvironmentId.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeployment, newDeployment.EnvironmentId.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) Restart(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	_, resp, err := c.client.EnvironmentActionsApi.RestartEnvironment(ctx, newDeployment.EnvironmentId.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeployment, newDeployment.EnvironmentId.String(), resp, err)
	}

	return &newDeployment, nil
}

func (c newNewDeploymentQoveryAPI) Delete(ctx context.Context, newDeployment newdeployment.Deployment) (*newdeployment.Deployment, error) {
	resp, err := c.client.EnvironmentMainCallsApi.DeleteEnvironment(ctx, newDeployment.EnvironmentId.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeployment, newDeployment.EnvironmentId.String(), resp, err)
	}

	return &newDeployment, nil
}
