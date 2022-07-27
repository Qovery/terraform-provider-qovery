package service

import (
	"github.com/qovery/terraform-provider-qovery/internal/domain/common"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// organizationService implements the interface organization.Service.
type organizationService struct {
	organizationRepository organization.Repository
}

// NewOrganizationService return a new instance of an organization.Service that uses the given organization.Repository.
func NewOrganizationService(organizationRepository organization.Repository) (organization.Service, error) {
	if organizationRepository == nil {
		return nil, common.ErrInvalidRepository
	}

	return &organizationService{
		organizationRepository: organizationRepository,
	}, nil
}
