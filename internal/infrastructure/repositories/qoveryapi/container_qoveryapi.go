package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
)

// Ensure containerQoveryAPI defined types fully satisfy the container.Repository interface.
var _ container.Repository = containerQoveryAPI{}

// containerQoveryAPI implements the interface container.Repository.
type containerQoveryAPI struct {
	client *qovery.APIClient
}

// newContainerQoveryAPI return a new instance of a container.Repository that uses Qovery's API.
func newContainerQoveryAPI(client *qovery.APIClient) (container.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &containerQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a container for an organization using the given organizationID and request.
func (c containerQoveryAPI) Create(ctx context.Context, environmentID string, request container.UpsertRepositoryRequest) (*container.Container, error) {
	req, err := newQoveryContainerRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	newContainer, resp, err := c.client.ContainersAPI.
		CreateContainer(ctx, environmentID).
		ContainerRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainer, request.Name, resp, err)
	}

	// Create custom domains
	if !request.CustomDomains.IsEmpty() {
		for _, customDomain := range request.CustomDomains.Create {
			_, resp, err := c.client.ContainerCustomDomainAPI.
				CreateContainerCustomDomain(ctx, newContainer.Id).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, request.Name, resp, err)
			}
		}
	}

	// Attach container to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, newContainer.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainer, request.Name, resp, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.CONTAINER, newContainer.Id, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainer, newContainer.Id, nil, err)
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, newContainer.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainer, newContainer.Id, resp, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.ContainerCustomDomainAPI.ListContainerCustomDomain(ctx, newContainer.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, newContainer.Id, resp, err)
	}

	return newDomainContainerFromQovery(newContainer, deploymentStage.Id, request.AdvancedSettingsJson, customDomains)
}

// Get calls Qovery's API to retrieve a container using the given containerID.
func (c containerQoveryAPI) Get(ctx context.Context, containerID string, advancedSettingsJsonFromState string, isTriggeredFromImport bool) (*container.Container, error) {
	container, resp, err := c.client.ContainerMainCallsAPI.
		GetContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceContainer, containerID, resp, err)
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainer, container.Id, resp, err)
	}

	// Get advanced settings
	advancedSettingsAsJson, err := advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).ReadServiceAdvancedSettings(domain.CONTAINER, container.Id, advancedSettingsJsonFromState, isTriggeredFromImport)
	if err != nil {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceContainer, containerID, nil, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.ContainerCustomDomainAPI.ListContainerCustomDomain(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, container.Id, resp, err)
	}

	return newDomainContainerFromQovery(container, deploymentStage.Id, *advancedSettingsAsJson, customDomains)
}

// Update calls Qovery's API to update a container using the given containerID and request.
func (c containerQoveryAPI) Update(ctx context.Context, containerID string, request container.UpsertRepositoryRequest) (*container.Container, error) {
	req, err := newQoveryContainerRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	container, resp, err := c.client.ContainerMainCallsAPI.
		EditContainer(ctx, containerID).
		ContainerRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceContainer, containerID, resp, err)
	}

	// Create custom domains
	if !request.CustomDomains.IsEmpty() {
		for _, customDomain := range request.CustomDomains.Delete {
			_, err := c.client.ContainerCustomDomainAPI.
				DeleteContainerCustomDomain(ctx, containerID, customDomain.Id).Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, request.Name, resp, err)
			}
		}
		for _, customDomain := range request.CustomDomains.Update {
			_, resp, err := c.client.ContainerCustomDomainAPI.
				EditContainerCustomDomain(ctx, containerID, customDomain.Id).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, request.Name, resp, err)
			}
		}
		for _, customDomain := range request.CustomDomains.Create {
			_, resp, err := c.client.ContainerCustomDomainAPI.
				CreateContainerCustomDomain(ctx, containerID).
				CustomDomainRequest(
					qovery.CustomDomainRequest{
						Domain:              customDomain.Domain,
						GenerateCertificate: customDomain.GenerateCertificate,
						UseCdn:              customDomain.UseCdn,
					}).
				Execute()
			if err != nil || resp.StatusCode >= 400 {
				return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, request.Name, resp, err)
			}
		}
	}

	// Attach container to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsAPI.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, container.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceContainer, request.Name, resp, err)
		}
	}

	// Update advanced settings
	err = advanced_settings.NewServiceAdvancedSettingsService(c.client.GetConfig()).UpdateServiceAdvancedSettings(domain.CONTAINER, container.Id, request.AdvancedSettingsJson)
	if err != nil {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceContainer, container.Id, nil, err)
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsAPI.GetServiceDeploymentStage(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceContainer, container.Id, resp, err)
	}

	// Get custom domains
	customDomains, _, err := c.client.ContainerCustomDomainAPI.ListContainerCustomDomain(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceContainerCustomDomain, container.Id, resp, err)
	}

	return newDomainContainerFromQovery(container, deploymentStage.Id, request.AdvancedSettingsJson, customDomains)
}

// Delete calls Qovery's API to deletes a container using the given containerID.
func (c containerQoveryAPI) Delete(ctx context.Context, containerID string) error {
	_, resp, err := c.client.ContainerMainCallsAPI.
		GetContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// if the container is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceContainer, containerID, resp, err)
	}

	resp, err = c.client.ContainerMainCallsAPI.
		DeleteContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceContainer, containerID, resp, err)
	}

	return nil
}
