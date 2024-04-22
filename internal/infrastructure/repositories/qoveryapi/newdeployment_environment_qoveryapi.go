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

func (c newNewDeploymentQoveryAPI) GetLastDeploymentId(ctx context.Context, environmentID uuid.UUID) (*string, error) {
	history, resp, err := c.client.EnvironmentDeploymentHistoryAPI.ListEnvironmentDeploymentHistory(ctx, environmentID.String()).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceDeployment, environmentID.String(), resp, err)
	}

	deploymentHistory := history.GetResults()
	if len(deploymentHistory) == 0 {
		deploymentID := fmt.Sprintf("%s-0", environmentID)
		return &deploymentID, nil
	}

	lastDeployment := deploymentHistory[len(deploymentHistory)-1]
	ID := lastDeployment.Id
	return &ID, nil
}

func (c newNewDeploymentQoveryAPI) GetNextDeploymentId(ctx context.Context, environmentID uuid.UUID) (*string, error) {
	lastDeploymentID, err := c.GetLastDeploymentId(ctx, environmentID)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`-(\d+)$`)
	result := re.FindStringSubmatch(*lastDeploymentID)
	if len(result) == 2 {
		version, err := strconv.Atoi(result[1])
		if err != nil {
			return nil, err
		}
		// Simulate next deployment id: ${environment_id}-${version}
		newVersion := fmt.Sprintf("%s-%d", environmentID.String(), version+1)
		return &newVersion, nil
	}

	return nil, errors.New(fmt.Sprintf("Cannot compute next deployment id for environment id %s", environmentID))
}
