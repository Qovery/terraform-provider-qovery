package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
)

// Ensure containerRegistryService defined types fully satisfy the registry.Service interface.
var _ registry.Service = containerRegistryService{}

// containerRegistryService implements the interface registry.Service.
type containerRegistryService struct {
	registryRepository registry.Repository
}

// NewContainerRegistryService return a new instance of a registry.Service that uses the given registry.Repository.
func NewContainerRegistryService(registryRepository registry.Repository) (registry.Service, error) {
	if registryRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &containerRegistryService{
		registryRepository: registryRepository,
	}, nil
}

// Create handles the domain logic to create an aws cluster registry.
func (c containerRegistryService) Create(ctx context.Context, organizationID string, request registry.UpsertRequest) (*registry.Registry, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToCreateRegistry.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToCreateRegistry.Error())
	}

	reg, err := c.registryRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToCreateRegistry.Error())
	}

	return reg, nil
}

// Get handles the domain logic to retrieve an aws cluster registry.
func (c containerRegistryService) Get(ctx context.Context, organizationID string, registryID string) (*registry.Registry, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToGetRegistry.Error())
	}

	if err := c.checkRegistryID(registryID); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToGetRegistry.Error())
	}

	reg, err := c.registryRepository.Get(ctx, organizationID, registryID)
	if err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToGetRegistry.Error())
	}

	return reg, nil
}

// Update handles the domain logic to update an aws cluster registry.
func (c containerRegistryService) Update(ctx context.Context, organizationID string, registryID string, request registry.UpsertRequest) (*registry.Registry, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToUpdateRegistry.Error())
	}

	if err := c.checkRegistryID(registryID); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToUpdateRegistry.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToUpdateRegistry.Error())
	}

	reg, err := c.registryRepository.Update(ctx, organizationID, registryID, request)
	if err != nil {
		return nil, errors.Wrap(err, registry.ErrFailedToUpdateRegistry.Error())
	}

	return reg, nil
}

// Delete handles the domain logic to delete an aws cluster registry.
func (c containerRegistryService) Delete(ctx context.Context, organizationID string, registryID string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, registry.ErrFailedToDeleteRegistry.Error())
	}

	if err := c.checkRegistryID(registryID); err != nil {
		return errors.Wrap(err, registry.ErrFailedToDeleteRegistry.Error())
	}

	if err := c.registryRepository.Delete(ctx, organizationID, registryID); err != nil {
		return errors.Wrap(err, registry.ErrFailedToDeleteRegistry.Error())
	}

	return nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c containerRegistryService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return registry.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, registry.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}

// checkRegistryID validates that the given registryID is valid.
func (c containerRegistryService) checkRegistryID(registryID string) error {
	if registryID == "" {
		return registry.ErrInvalidRegistryIDParam
	}

	if _, err := uuid.Parse(registryID); err != nil {
		return errors.Wrap(err, registry.ErrInvalidRegistryIDParam.Error())
	}

	return nil
}
