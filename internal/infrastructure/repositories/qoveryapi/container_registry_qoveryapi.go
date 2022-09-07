package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// Ensure containerRegistryQoveryAPI defined types fully satisfy the registry.Repository interface.
var _ registry.Repository = containerRegistryQoveryAPI{}

// containerRegistryQoveryAPI implements the interface registry.Repository.
type containerRegistryQoveryAPI struct {
	client *qovery.APIClient
}

// newContainerRegistryQoveryAPI return a new instance of a registry.Repository that uses Qovery's API.
func newContainerRegistryQoveryAPI(client *qovery.APIClient) (registry.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &containerRegistryQoveryAPI{
		client: client,
	}, nil
}

// Create calls Qovery's API to create a registry for an organization using the given organizationID and request.
func (c containerRegistryQoveryAPI) Create(ctx context.Context, organizationID string, request registry.UpsertRequest) (*registry.Registry, error) {
	req, err := newQoveryContainerRegistryRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, registry.ErrInvalidUpsertRequest.Error())
	}

	reg, resp, err := c.client.ContainerRegistriesApi.
		CreateContainerRegistry(ctx, organizationID).
		ContainerRegistryRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewCreateApiError(apierrors.ApiResourceContainerRegistry, request.Name, resp, err)
		return nil, apiErr
	}

	return newDomainRegistryFromQovery(reg, organizationID)
}

// Get calls Qovery's API to retrieve a  registry using the given registryID.
func (c containerRegistryQoveryAPI) Get(ctx context.Context, organizationID string, registryID string) (*registry.Registry, error) {
	reg, resp, err := c.client.ContainerRegistriesApi.
		GetContainerRegistry(ctx, organizationID, registryID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceContainerRegistry, registryID, resp, err)
	}

	return newDomainRegistryFromQovery(reg, organizationID)
}

// Update calls Qovery's API to update a registry using the given registryID and request.
func (c containerRegistryQoveryAPI) Update(ctx context.Context, organizationID string, registryID string, request registry.UpsertRequest) (*registry.Registry, error) {
	req, err := newQoveryContainerRegistryRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, registry.ErrInvalidUpsertRequest.Error())
	}

	reg, resp, err := c.client.ContainerRegistriesApi.
		EditContainerRegistry(ctx, organizationID, registryID).
		ContainerRegistryRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceContainerRegistry, registryID, resp, err)
	}

	return newDomainRegistryFromQovery(reg, organizationID)
}

// Delete calls Qovery's API to deletes a registry using the given registryID.
func (c containerRegistryQoveryAPI) Delete(ctx context.Context, organizationID string, registryID string) error {
	resp, err := c.client.ContainerRegistriesApi.
		DeleteContainerRegistry(ctx, organizationID, registryID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteApiError(apierrors.ApiResourceContainerRegistry, registryID, resp, err)
	}

	return nil
}
