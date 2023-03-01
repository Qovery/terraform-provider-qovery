package qoveryapi

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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

func (c newNewDeploymentQoveryAPI) GetLastDeploymentId(ctx context.Context, environmentId uuid.UUID) (*string, error) {
	history, resp, err := c.client.EnvironmentDeploymentHistoryApi.ListEnvironmentDeploymentHistory(ctx, environmentId.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceDeployment, environmentId.String(), resp, err)
	}

	deploymentHistory := history.GetResults()
	if len(deploymentHistory) == 0 {
		deploymentId := fmt.Sprintf("%s-0", environmentId)
		return &deploymentId, nil
	}

	lastDeployment := deploymentHistory[len(deploymentHistory)-1]
	id := lastDeployment.Id
	return &id, nil
}

func (c newNewDeploymentQoveryAPI) GetNextDeploymentId(ctx context.Context, environmentId uuid.UUID) (*string, error) {
	lastDeploymentId, err := c.GetLastDeploymentId(ctx, environmentId)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`-(\d+)$`)
	result := re.FindStringSubmatch(*lastDeploymentId)
	if len(result) == 2 {
		version, err := strconv.Atoi(result[1])
		if err != nil {
			return nil, err
		}
		// Simulate next deployment id: ${environment_id}-${version}
		newVersion := fmt.Sprintf("%s-%d", environmentId.String(), version+1)
		return &newVersion, nil
	}

	return nil, errors.New(fmt.Sprintf("Cannot compute next deployment id for environment id %s", environmentId))
}
