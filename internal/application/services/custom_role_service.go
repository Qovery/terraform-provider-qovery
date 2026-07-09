package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

// Ensure customRoleService defined type fully satisfy the customrole.Service interface.
var _ customrole.Service = customRoleService{}

// customRoleService implements the interface customrole.Service.
type customRoleService struct {
	customRoleRepository customrole.Repository
}

func NewCustomRoleService(customRoleRepository customrole.Repository) (customrole.Service, error) {
	if customRoleRepository == nil {
		return nil, ErrInvalidRepository
	}
	return &customRoleService{customRoleRepository: customRoleRepository}, nil
}

func (s customRoleService) Create(ctx context.Context, organizationID string, request customrole.UpsertRequest) (*customrole.CustomRole, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToCreateCustomRole.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToCreateCustomRole.Error())
	}
	role, err := s.customRoleRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToCreateCustomRole.Error())
	}
	return role, nil
}

func (s customRoleService) Get(ctx context.Context, organizationID string, customRoleID string) (*customrole.CustomRole, error) {
	if err := s.checkIDs(organizationID, customRoleID); err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToGetCustomRole.Error())
	}
	role, err := s.customRoleRepository.Get(ctx, organizationID, customRoleID)
	if err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToGetCustomRole.Error())
	}
	return role, nil
}

func (s customRoleService) Update(ctx context.Context, organizationID string, customRoleID string, request customrole.UpsertRequest) (*customrole.CustomRole, error) {
	if err := s.checkIDs(organizationID, customRoleID); err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToUpdateCustomRole.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToUpdateCustomRole.Error())
	}
	role, err := s.customRoleRepository.Update(ctx, organizationID, customRoleID, request)
	if err != nil {
		return nil, errors.Wrap(err, customrole.ErrFailedToUpdateCustomRole.Error())
	}
	return role, nil
}

func (s customRoleService) Delete(ctx context.Context, organizationID string, customRoleID string) error {
	if err := s.checkIDs(organizationID, customRoleID); err != nil {
		return errors.Wrap(err, customrole.ErrFailedToDeleteCustomRole.Error())
	}
	if err := s.customRoleRepository.Delete(ctx, organizationID, customRoleID); err != nil {
		return errors.Wrap(err, customrole.ErrFailedToDeleteCustomRole.Error())
	}
	return nil
}

func (s customRoleService) checkOrganizationID(organizationID string) error {
	return validateUUIDParam(organizationID, customrole.ErrInvalidOrganizationIdParam)
}

func (s customRoleService) checkIDs(organizationID string, customRoleID string) error {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return err
	}
	return validateUUIDParam(customRoleID, customrole.ErrInvalidCustomRoleIdParam)
}
