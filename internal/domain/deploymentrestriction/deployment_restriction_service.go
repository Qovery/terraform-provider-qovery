package deploymentrestriction

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain"
)

type DeploymentRestrictionService struct {
	client qovery.APIClient
}

func NewDeploymentRestrictionService(apiClient qovery.APIClient) (DeploymentRestrictionService, error) {
	return DeploymentRestrictionService{
		client: apiClient,
	}, nil
}

func (service DeploymentRestrictionService) GetServiceDeploymentRestrictions(ctx context.Context, serviceId string, serviceType int) ([]ServiceDeploymentRestriction, *apierrors.APIError) {
	deploymentRestrictions := make([]ServiceDeploymentRestriction, 0)

	switch serviceType {
	case domain.APPLICATION:
		deploymentRestrictionsResponse, res, err := service.client.ApplicationDeploymentRestrictionAPI.
			GetApplicationDeploymentRestrictions(ctx, serviceId).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceId, res, err)
		}
		for _, deploymentRestriction := range deploymentRestrictionsResponse.Results {
			id := deploymentRestriction.Id
			deploymentRestrictions = append(deploymentRestrictions, ServiceDeploymentRestriction{
				Id:    &id,
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			})
		}
	case domain.JOB:
		deploymentRestrictionsResponse, res, err := service.client.JobDeploymentRestrictionAPI.
			GetJobDeploymentRestrictions(ctx, serviceId).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceId, res, err)
		}
		for _, deploymentRestriction := range deploymentRestrictionsResponse.Results {
			id := deploymentRestriction.Id
			deploymentRestrictions = append(deploymentRestrictions, ServiceDeploymentRestriction{
				Id:    &id,
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			})
		}
	case domain.HELM:
		deploymentRestrictionsResponse, res, err := service.client.HelmDeploymentRestrictionAPI.
			GetHelmDeploymentRestrictions(ctx, serviceId).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceId, res, err)
		}
		for _, deploymentRestriction := range deploymentRestrictionsResponse.Results {
			id := deploymentRestriction.Id
			deploymentRestrictions = append(deploymentRestrictions, ServiceDeploymentRestriction{
				Id:    &id,
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			})
		}
	default:
		return nil, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceId, nil, nil)
	}

	return deploymentRestrictions, nil
}

func (service DeploymentRestrictionService) UpdateServiceDeploymentRestrictions(ctx context.Context, serviceId string, serviceType int, request ServiceDeploymentRestrictionsDiff) *apierrors.APIError {
	switch serviceType {
	case domain.APPLICATION:
		for _, deploymentRestrictionId := range request.Delete {
			res, err := service.client.ApplicationDeploymentRestrictionAPI.DeleteApplicationDeploymentRestriction(ctx, serviceId, deploymentRestrictionId).Execute()
			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewDeleteError(apierrors.APIResourceServiceDeploymentRestriction, deploymentRestrictionId, res, err)
			}
		}

		// Nothing to do for request.Update as we don't have the id
		// If a property is changed, it is automatically deleted / re-created

		for _, deploymentRestriction := range request.Create {
			_, res, err := service.client.ApplicationDeploymentRestrictionAPI.CreateApplicationDeploymentRestriction(ctx, serviceId).ApplicationDeploymentRestrictionRequest(qovery.ApplicationDeploymentRestrictionRequest{
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			}).Execute()

			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewCreateError(apierrors.APIResourceServiceDeploymentRestriction, "no id", res, err)
			}
		}
	case domain.JOB:
		for _, deploymentRestrictionId := range request.Delete {
			res, err := service.client.JobDeploymentRestrictionAPI.DeleteJobDeploymentRestriction(ctx, serviceId, deploymentRestrictionId).Execute()
			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewDeleteError(apierrors.APIResourceServiceDeploymentRestriction, deploymentRestrictionId, res, err)
			}
		}

		// Nothing to do for request.Update as we don't have the id
		// If a property is changed, it is automatically deleted / re-created

		for _, deploymentRestriction := range request.Create {
			_, res, err := service.client.JobDeploymentRestrictionAPI.CreateJobDeploymentRestriction(ctx, serviceId).JobDeploymentRestrictionRequest(qovery.JobDeploymentRestrictionRequest{
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			}).Execute()

			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewCreateError(apierrors.APIResourceServiceDeploymentRestriction, "no id", res, err)
			}
		}
	case domain.HELM:
		for _, deploymentRestrictionId := range request.Delete {
			res, err := service.client.HelmDeploymentRestrictionAPI.DeleteHelmDeploymentRestriction(ctx, serviceId, deploymentRestrictionId).Execute()
			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewDeleteError(apierrors.APIResourceServiceDeploymentRestriction, deploymentRestrictionId, res, err)
			}
		}

		// Nothing to do for request.Update as we don't have the id
		// If a property is changed, it is automatically deleted / re-created

		for _, deploymentRestriction := range request.Create {
			_, res, err := service.client.HelmDeploymentRestrictionAPI.CreateHelmDeploymentRestriction(ctx, serviceId).HelmDeploymentRestrictionRequest(qovery.HelmDeploymentRestrictionRequest{
				Mode:  deploymentRestriction.Mode,
				Type:  deploymentRestriction.Type,
				Value: deploymentRestriction.Value,
			}).Execute()

			if err != nil || res.StatusCode >= 400 {
				return apierrors.NewCreateError(apierrors.APIResourceServiceDeploymentRestriction, "no id", res, err)
			}
		}
	default:
		return apierrors.NewError(apierrors.APIActionNotSupported, apierrors.APIResourceServiceDeploymentRestriction, serviceId, nil, nil)
	}

	return nil
}
