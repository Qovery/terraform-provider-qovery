package deploymentrestriction

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type DeploymentRestrictionService struct {
	client qovery.APIClient
}

func NewDeploymentRestrictionService(apiClient qovery.APIClient) (DeploymentRestrictionService, error) {
	return DeploymentRestrictionService{
		client: apiClient,
	}, nil
}

func (service DeploymentRestrictionService) GetServiceDeploymentRestrictions(ctx context.Context, serviceId string) ([]ServiceDeploymentRestriction, *apierrors.APIError) {
	deploymentRestrictionsResponse, res, err := service.client.ServiceMainCallsAPI.
		ListServiceDeploymentRestrictions(ctx, serviceId).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceId, res, err)
	}

	deploymentRestrictions := make([]ServiceDeploymentRestriction, 0)
	for _, deploymentRestriction := range deploymentRestrictionsResponse.Results {
		id := deploymentRestriction.Id
		deploymentRestrictions = append(deploymentRestrictions, ServiceDeploymentRestriction{
			Id:    &id,
			Mode:  deploymentRestriction.Mode,
			Type:  deploymentRestriction.Type,
			Value: deploymentRestriction.Value,
		})
	}

	return deploymentRestrictions, nil
}

func (service DeploymentRestrictionService) UpdateServiceDeploymentRestrictions(ctx context.Context, serviceId string, request ServiceDeploymentRestrictionsDiff) *apierrors.APIError {
	for _, deploymentRestrictionId := range request.Delete {
		res, err := service.client.ServiceMainCallsAPI.DeleteServiceDeploymentRestriction(ctx, serviceId, deploymentRestrictionId).Execute()
		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewDeleteError(apierrors.APIResourceServiceDeploymentRestriction, deploymentRestrictionId, res, err)
		}
	}

	// Nothing to do for request.Update as we don't have the id
	// If a property is changed, it is automatically deleted / re-created

	for _, deploymentRestriction := range request.Create {
		_, res, err := service.client.ServiceMainCallsAPI.CreateServiceDeploymentRestriction(ctx, serviceId).ServiceDeploymentRestrictionRequest(qovery.ServiceDeploymentRestrictionRequest{
			Mode:  deploymentRestriction.Mode,
			Type:  deploymentRestriction.Type,
			Value: deploymentRestriction.Value,
		}).Execute()

		if err != nil || res.StatusCode >= 400 {
			return apierrors.NewCreateError(apierrors.APIResourceServiceDeploymentRestriction, "no id", res, err)
		}
	}

	return nil
}
