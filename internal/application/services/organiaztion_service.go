package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// Ensure organizationService defined type fully satisfy the organization.Service interface.
var _ organization.Service = organizationService{}

// organizationService implements the interface organization.Service.
type organizationService struct {
	organizationRepository organization.Repository
}

// NewOrganizationService return a new instance of an organization.Service that uses the given organization.Repository.
func NewOrganizationService(organizationRepository organization.Repository) (organization.Service, error) {
	if organizationRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &organizationService{
		organizationRepository: organizationRepository,
	}, nil
}

// Get handles the domain logic to retrieve an organization.
func (c organizationService) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, organization.ErrFailedToGetOrganization.Error())
	}

	orga, err := c.organizationRepository.Get(ctx, organizationID)
	if err != nil {
		return nil, errors.Wrap(err, organization.ErrFailedToGetOrganization.Error())
	}

	return orga, nil
}

// Update handles the domain logic to update an organization.
func (c organizationService) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, organization.ErrFailedToUpdateOrganization.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, organization.ErrFailedToUpdateOrganization.Error())
	}

	orga, err := c.organizationRepository.Update(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, organization.ErrFailedToUpdateOrganization.Error())
	}

	return orga, nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c organizationService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return organization.ErrInvalidOrganizationIDParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, organization.ErrInvalidOrganizationIDParam.Error())
	}

	return nil
}
