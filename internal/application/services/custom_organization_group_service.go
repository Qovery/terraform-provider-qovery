package services

import (
	"context"
	"github.com/qovery/terraform-provider-qovery/internal/domain/custom_organization_role"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var _ custom_organization_role.Service = customOrganizationRoleService{}

type customOrganizationRoleService struct {
	customOrganizationRoleRepository custom_organization_role.Repository
}

func NewCustomOrganizationRoleService(customOrganizationRoleRepository custom_organization_role.Repository) (custom_organization_role.Service, error) {
	return &customOrganizationRoleService{customOrganizationRoleRepository}, nil
}

func (s customOrganizationRoleService) Create(ctx context.Context, organizationID string, request custom_organization_role.UpsertServiceRequest) (*custom_organization_role.CustomOrganizationRole, error) {
	if err := s.checkID(organizationID); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleOrganizationIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrFailedToCreateCustomOrganizationRole.Error())
	}

	newCustomOrganizationRole, err := s.customOrganizationRoleRepository.Create(ctx, organizationID, request.CustomOrganizationRoleUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrFailedToCreateCustomOrganizationRole.Error())
	}

	return newCustomOrganizationRole, nil
}

func (s customOrganizationRoleService) Get(ctx context.Context, organizationId string, customOrganizationRoleId string) (*custom_organization_role.CustomOrganizationRole, error) {
	if err := s.checkID(customOrganizationRoleId); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleIdParam.Error())
	}

	fetchedCustomOrganizationRole, err := s.customOrganizationRoleRepository.Get(ctx, organizationId, customOrganizationRoleId)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrFailedToGetCustomOrganizationRole.Error())
	}

	return fetchedCustomOrganizationRole, nil
}

func (s customOrganizationRoleService) Update(ctx context.Context, organizationId string, customOrganizationRoleId string, request custom_organization_role.UpsertServiceRequest) (*custom_organization_role.CustomOrganizationRole, error) {
	if err := s.checkID(organizationId); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleOrganizationIdParam.Error())
	}

	if err := s.checkID(customOrganizationRoleId); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrFailedToUpdateCustomOrganizationRole.Error())
	}
	fetchedCustomOrganizationRole, err := s.customOrganizationRoleRepository.Update(ctx, organizationId, customOrganizationRoleId, request.CustomOrganizationRoleUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrFailedToUpdateCustomOrganizationRole.Error())
	}

	return fetchedCustomOrganizationRole, nil
}

func (s customOrganizationRoleService) Delete(ctx context.Context, organizationId string, customOrganizationRoleId string) error {
	if err := s.checkID(customOrganizationRoleId); err != nil {
		return errors.Wrap(err, custom_organization_role.ErrFailedToDeleteCustomOrganizationRole.Error())
	}

	err := s.customOrganizationRoleRepository.Delete(ctx, organizationId, customOrganizationRoleId)
	return err
}

func (s customOrganizationRoleService) checkID(customOrganizationRoleId string) error {
	if customOrganizationRoleId == "" {
		return custom_organization_role.ErrInvalidCustomOrganizationRoleIdParam
	}

	if _, err := uuid.Parse(customOrganizationRoleId); err != nil {
		return errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleIdParam.Error())
	}

	return nil
}
