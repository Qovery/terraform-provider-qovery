package qoveryapi

import (
	"context"

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
	if len(request.DeploymentStageId) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, request.DeploymentStageId, newContainer.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, request.Name, resp, err)
		}
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, newContainer.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, newContainer.Id, resp, err)
	}

	return newDomainContainerFromQovery(newContainer, deploymentStage.Id)
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

	return newDomainContainerFromQovery(container, deploymentStage.Id)
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
	if len(request.DeploymentStageId) > 0 {
		_, response, err := c.client.DeploymentStageMainCallsApi.AttachServiceToDeploymentStage(ctx, request.DeploymentStageId, container.Id).Execute()
		if err != nil || response.StatusCode >= 400 {
			return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, request.Name, resp, err)
		}
	}

	// Get container deployment stage
	deploymentStage, resp, err := c.client.DeploymentStageMainCallsApi.GetServiceDeploymentStage(ctx, container.Id).Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateApiError(apierrors.ApiResourceContainer, container.Id, resp, err)
	}

	return newDomainContainerFromQovery(container, deploymentStage.Id)
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
