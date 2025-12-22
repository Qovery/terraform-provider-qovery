package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
)

// Ensure terraformServiceQoveryAPI defined types fully satisfy the terraformservice.Repository interface.
var _ terraformservice.Repository = terraformServiceQoveryAPI{}

// terraformServiceQoveryAPI implements the interface terraformservice.Repository.
type terraformServiceQoveryAPI struct {
	client *qovery.APIClient
}

// newTerraformServiceQoveryAPI return a new instance of a terraformservice.Repository that uses Qovery's API.
func newTerraformServiceQoveryAPI(client *qovery.APIClient) (terraformservice.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &terraformServiceQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a terraform service for an environment using the given environmentID and request.
func (c terraformServiceQoveryAPI) Create(ctx context.Context, environmentID string, request terraformservice.UpsertRepositoryRequest) (*terraformservice.TerraformService, error) {
	req, err := newQoveryTerraformRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	newTerraform, resp, err := c.client.TerraformsAPI.
		CreateTerraform(ctx, environmentID).
		TerraformRequest(*req).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceTerraformService, request.Name, resp, err)
	}

	// Attach terraform service to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, newTerraform.Id).Execute()
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceTerraformService, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.TERRAFORM, newTerraform.Id, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceTerraformService, request.Name, nil, err)
	}

	// Get terraform service deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, newTerraform.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceTerraformService, newTerraform.Id, resp, err)
	}

	return newDomainTerraformServiceFromQovery(newTerraform, deploymentStage.Id, request.AdvancedSettingsJson)
}

// Get calls Qovery's API to retrieve a terraform service using the given terraformServiceID.
func (c terraformServiceQoveryAPI) Get(ctx context.Context, terraformServiceID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*terraformservice.TerraformService, error) {
	terraform, resp, err := c.client.TerraformMainCallsAPI.
		GetTerraform(ctx, terraformServiceID).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceTerraformService, terraformServiceID, resp, err)
	}

	// Get terraform service deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, terraform.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceTerraformService, terraform.Id, resp, err)
	}

	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).ReadServiceAdvancedSettings(domain.TERRAFORM, terraformServiceID, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceTerraformService, terraformServiceID, nil, err)
	}

	return newDomainTerraformServiceFromQovery(terraform, deploymentStage.Id, *advancedSettingsAsJson)
}

// Update calls Qovery's API to update a terraform service using the given terraformServiceID and request.
func (c terraformServiceQoveryAPI) Update(ctx context.Context, terraformServiceID string, request terraformservice.UpsertRepositoryRequest) (*terraformservice.TerraformService, error) {
	req, err := newQoveryTerraformRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, terraformservice.ErrInvalidTerraformServiceUpsertRequest.Error())
	}

	terraform, resp, err := c.client.TerraformMainCallsAPI.
		EditTerraform(ctx, terraformServiceID).
		TerraformRequest(*req).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceTerraformService, terraformServiceID, resp, err)
	}

	// Attach terraform service to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, terraform.Id).Execute()
		if err != nil || (response != nil && response.StatusCode >= 400) {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceTerraformService, request.Name, response, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.TERRAFORM, terraformServiceID, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceTerraformService, request.Name, nil, err)
	}

	// Get terraform service deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, terraform.Id).Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceTerraformService, terraform.Id, resp, err)
	}

	return newDomainTerraformServiceFromQovery(terraform, deploymentStage.Id, request.AdvancedSettingsJson)
}

// Delete calls Qovery's API to deletes a terraform service using the given terraformServiceID.
func (c terraformServiceQoveryAPI) Delete(ctx context.Context, terraformServiceID string) error {
	_, resp, err := c.client.TerraformMainCallsAPI.
		GetTerraform(ctx, terraformServiceID).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		if resp != nil && resp.StatusCode == 404 {
			// if the terraform service is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceTerraformService, terraformServiceID, resp, err)
	}

	resp, err = c.client.TerraformMainCallsAPI.
		DeleteTerraform(ctx, terraformServiceID).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceTerraformService, terraformServiceID, resp, err)
	}

	return nil
}

// List calls Qovery's API to list terraform services for an environment using the given environmentID.
func (c terraformServiceQoveryAPI) List(ctx context.Context, environmentID string) ([]terraformservice.TerraformService, error) {
	terraformList, resp, err := c.client.TerraformsAPI.
		ListTerraforms(ctx, environmentID).
		Execute()
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceTerraformService, environmentID, resp, err)
	}

	services := make([]terraformservice.TerraformService, 0, len(terraformList.GetResults()))
	for _, tf := range terraformList.GetResults() {
		service, err := newDomainTerraformServiceFromQovery(&tf, "", "")
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert terraform service")
		}
		services = append(services, *service)
	}

	return services, nil
}
