package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
)

var _ labels_group.Service = labelsGroupService{}

type labelsGroupService struct {
	labelsGroupRepository labels_group.Repository
}

func NewLabelsGroupService(labelsGroupRepository labels_group.Repository) (labels_group.Service, error) {
	return &labelsGroupService{labelsGroupRepository}, nil
}

func (s labelsGroupService) Create(ctx context.Context, organizationID string, request labels_group.UpsertServiceRequest) (*labels_group.LabelsGroup, error) {
	if err := s.checkID(organizationID); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrInvalidLabelsGroupOrganizationIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrFailedToCreateLabelsGroup.Error())
	}

	newLabelsGroup, err := s.labelsGroupRepository.Create(ctx, organizationID, request.LabelsGroupUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, labels_group.ErrFailedToCreateLabelsGroup.Error())
	}

	return newLabelsGroup, nil
}

func (s labelsGroupService) Get(ctx context.Context, organizationId string, labelsGroupId string) (*labels_group.LabelsGroup, error) {
	if err := s.checkID(labelsGroupId); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrInvalidLabelsGroupIdParam.Error())
	}

	fetchedLabelsGroup, err := s.labelsGroupRepository.Get(ctx, organizationId, labelsGroupId)
	if err != nil {
		return nil, errors.Wrap(err, labels_group.ErrFailedToGetLabelsGroup.Error())
	}

	return fetchedLabelsGroup, nil
}

func (s labelsGroupService) Update(ctx context.Context, organizationId string, labelsGroupId string, request labels_group.UpsertServiceRequest) (*labels_group.LabelsGroup, error) {
	if err := s.checkID(organizationId); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrInvalidLabelsGroupOrganizationIdParam.Error())
	}

	if err := s.checkID(labelsGroupId); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrInvalidLabelsGroupIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, labels_group.ErrFailedToUpdateLabelsGroup.Error())
	}
	fetchedLabelsGroup, err := s.labelsGroupRepository.Update(ctx, organizationId, labelsGroupId, request.LabelsGroupUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, labels_group.ErrFailedToUpdateLabelsGroup.Error())
	}

	return fetchedLabelsGroup, nil
}

func (s labelsGroupService) Delete(ctx context.Context, organizationId string, labelsGroupId string) error {
	if err := s.checkID(labelsGroupId); err != nil {
		return errors.Wrap(err, labels_group.ErrFailedToDeleteLabelsGroup.Error())
	}

	err := s.labelsGroupRepository.Delete(ctx, organizationId, labelsGroupId)
	return err
}

func (s labelsGroupService) checkID(labelsGroupId string) error {
	if labelsGroupId == "" {
		return labels_group.ErrInvalidLabelsGroupIdParam
	}

	if _, err := uuid.Parse(labelsGroupId); err != nil {
		return errors.Wrap(err, labels_group.ErrInvalidLabelsGroupIdParam.Error())
	}

	return nil
}
