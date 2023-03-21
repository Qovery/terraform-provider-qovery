package qoveryapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

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

	newContainer, resp, err := c.client.ContainersApi.
		CreateContainer(ctx, environmentID).
		ContainerRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, request.Name, resp, err)
	}

	// Attach container to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, newContainer.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, request.Name, resp, err)
		}
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, newContainer.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, newContainer.Id, resp, err)
	}

	// Handle container adv settings
	advSettings, settingsErr := handleContainerAdvSettings(nil, newContainer.Id, ctx, c.client)
	if settingsErr != nil {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, newContainer.Id, resp, err)
	}

	return newDomainContainerFromQovery(newContainer, deploymentStage.Id, advSettings)
}

// Get calls Qovery's API to retrieve a container using the given containerID.
func (c containerQoveryAPI) Get(ctx context.Context, containerID string) (*container.Container, error) {
	container, resp, err := c.client.ContainerMainCallsApi.
		GetContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, container.Id, resp, err)
	}

	// Get container adv settings
	advSettings, settingsErr := handleContainerAdvSettings(nil, container.Id, ctx, c.client)
	if settingsErr != nil {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, container.Id, resp, err)
	}

	return newDomainContainerFromQovery(container, deploymentStage.Id, advSettings)
}

// Update calls Qovery's API to update a container using the given containerID and request.
func (c containerQoveryAPI) Update(ctx context.Context, containerID string, request container.UpsertRepositoryRequest) (*container.Container, error) {
	req, err := newQoveryContainerRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, container.ErrInvalidUpsertRequest.Error())
	}

	container, resp, err := c.client.ContainerMainCallsApi.
		EditContainer(ctx, containerID).
		ContainerRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	// Attach container to deployment stage
	if len(request.DeploymentStageID) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, request.DeploymentStageID, container.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, request.Name, resp, err)
		}
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, container.Id, resp, err)
	}

	// Handle container adv settings
	advSettings, settingsErr := handleContainerAdvSettings(nil, container.Id, ctx, c.client)
	if settingsErr != nil {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, container.Id, resp, err)
	}

	return newDomainContainerFromQovery(container, deploymentStage.Id, advSettings)
}

// Delete calls Qovery's API to deletes a container using the given containerID.
func (c containerQoveryAPI) Delete(ctx context.Context, containerID string) error {
	_, resp, err := c.client.ContainerMainCallsApi.
		GetContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			// if the container is not found, then it has already been deleted
			return nil
		}
		return apierrors.NewDeleteApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	resp, err = c.client.ContainerMainCallsApi.
		DeleteContainer(ctx, containerID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceContainer, containerID, resp, err)
	}

	return nil
}

func fromContainerAdvancedSettings(s *qovery.ContainerAdvancedSettings) (map[string]interface{}, error) {
	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return nil, marshalErr
	}

	var unmarshal map[string]interface{}
	if unmarshalErr := json.Unmarshal(resp, &unmarshal); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return unmarshal, nil
}

func toContainerAdvancedSettings(s map[string]interface{}) (qovery.ContainerAdvancedSettings, error) {
	resp, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		return qovery.ContainerAdvancedSettings{}, marshalErr
	}

	var result qovery.ContainerAdvancedSettings
	if unmarshalErr := json.Unmarshal(resp, &result); unmarshalErr != nil {
		return qovery.ContainerAdvancedSettings{}, unmarshalErr
	}

	return result, nil
}

func handleContainerAdvSettings(containerSettings map[string]interface{}, containerId string, ctx context.Context, client *qovery.APIClient) (map[string]interface{}, error) {
	// Get container adv settings
	var containerAdvSettings *qovery.ContainerAdvancedSettings
	var resp *http.Response
	advSettings, err := toContainerAdvancedSettings(containerSettings)
	if err != nil {
		return nil, err
	}
	if containerSettings != nil && len(containerSettings) > 0 {
		containerAdvSettings, resp, err = client.ContainerConfigurationApi.EditContainerAdvancedSettings(ctx, containerId).ContainerAdvancedSettings(advSettings).Execute()
	} else {
		containerAdvSettings, resp, err = client.ContainerConfigurationApi.GetContainerAdvancedSettings(ctx, containerId).Execute()
	}
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewApiErrorFromError(err)
	}

	mapSettings, mapErr := fromContainerAdvancedSettings(containerAdvSettings)
	if mapErr != nil {
		return nil, err
	}

	return mapSettings, nil
}
