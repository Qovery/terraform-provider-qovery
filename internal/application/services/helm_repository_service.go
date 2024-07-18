package services

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var _ helmRepository.Service = helmRepositoryService{}

// helmRepositoryService implements the interface helmRepository.Service.
type helmRepositoryService struct {
	helmRepositoryRepository helmRepository.Repository
}

func NewHelmRepositoryService(helmRepositoryRepository helmRepository.Repository) (helmRepository.Service, error) {
	if helmRepositoryRepository == nil {
		return nil, ErrInvalidRepository
	}

	return &helmRepositoryService{
		helmRepositoryRepository: helmRepositoryRepository,
	}, nil
}

func (c helmRepositoryService) Create(ctx context.Context, organizationID string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToCreateHelmRepository.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToCreateHelmRepository.Error())
	}

	reg, err := c.helmRepositoryRepository.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToCreateHelmRepository.Error())
	}

	return reg, nil
}

func (c helmRepositoryService) Get(ctx context.Context, organizationID string, repositoryId string) (*helmRepository.HelmRepository, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToGetHelmRepository.Error())
	}

	if err := c.checkRegistryID(repositoryId); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToGetHelmRepository.Error())
	}

	reg, err := c.helmRepositoryRepository.Get(ctx, organizationID, repositoryId)
	if err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToGetHelmRepository.Error())
	}

	return reg, nil
}

func (c helmRepositoryService) Update(ctx context.Context, organizationID string, repositoryId string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToUpdateHelmRepository.Error())
	}

	if err := c.checkRegistryID(repositoryId); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToUpdateHelmRepository.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToUpdateHelmRepository.Error())
	}

	reg, err := c.helmRepositoryRepository.Update(ctx, organizationID, repositoryId, request)
	if err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrFailedToUpdateHelmRepository.Error())
	}

	return reg, nil
}

func (c helmRepositoryService) Delete(ctx context.Context, organizationID string, repositoryId string) error {
	if err := c.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, helmRepository.ErrFailedToDeleteHelmRepository.Error())
	}

	if err := c.checkRegistryID(repositoryId); err != nil {
		return errors.Wrap(err, helmRepository.ErrFailedToDeleteHelmRepository.Error())
	}

	if err := c.helmRepositoryRepository.Delete(ctx, organizationID, repositoryId); err != nil {
		return errors.Wrap(err, helmRepository.ErrFailedToDeleteHelmRepository.Error())
	}

	return nil
}

// checkOrganizationID validates that the given organizationID is valid.
func (c helmRepositoryService) checkOrganizationID(organizationID string) error {
	if organizationID == "" {
		return helmRepository.ErrInvalidOrganizationIdParam
	}

	if _, err := uuid.Parse(organizationID); err != nil {
		return errors.Wrap(err, helmRepository.ErrInvalidOrganizationIdParam.Error())
	}

	return nil
}

// checkRegistryID validates that the given registryID is valid.
func (c helmRepositoryService) checkRegistryID(repositoryId string) error {
	if repositoryId == "" {
		return helmRepository.ErrInvalidRepositoryIdParam
	}

	if _, err := uuid.Parse(repositoryId); err != nil {
		return errors.Wrap(err, helmRepository.ErrInvalidRepositoryIdParam.Error())
	}

	return nil
}
