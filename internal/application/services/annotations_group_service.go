package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
)

var _ annotations_group.Service = annotationsGroupService{}

type annotationsGroupService struct {
	annotationsGroupRepository annotations_group.Repository
}

func NewAnnotationsGroupService(annotationsGroupRepository annotations_group.Repository) (annotations_group.Service, error) {
	return &annotationsGroupService{annotationsGroupRepository}, nil
}

func (s annotationsGroupService) Create(ctx context.Context, organizationID string, request annotations_group.UpsertServiceRequest) (*annotations_group.AnnotationsGroup, error) {
	if err := s.checkID(organizationID); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupOrganizationIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrFailedToCreateAnnotationsGroup.Error())
	}

	newAnnotationsGroup, err := s.annotationsGroupRepository.Create(ctx, organizationID, request.AnnotationsGroupUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrFailedToCreateAnnotationsGroup.Error())
	}

	return newAnnotationsGroup, nil
}

func (s annotationsGroupService) Get(ctx context.Context, organizationId string, annotationsGroupId string) (*annotations_group.AnnotationsGroup, error) {
	if err := s.checkID(annotationsGroupId); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupIdParam.Error())
	}

	fetchedAnnotationsGroup, err := s.annotationsGroupRepository.Get(ctx, organizationId, annotationsGroupId)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrFailedToGetAnnotationsGroup.Error())
	}

	return fetchedAnnotationsGroup, nil
}

func (s annotationsGroupService) Update(ctx context.Context, organizationId string, annotationsGroupId string, request annotations_group.UpsertServiceRequest) (*annotations_group.AnnotationsGroup, error) {
	if err := s.checkID(organizationId); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupOrganizationIdParam.Error())
	}

	if err := s.checkID(annotationsGroupId); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupIdParam.Error())
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrFailedToUpdateAnnotationsGroup.Error())
	}
	fetchedAnnotationsGroup, err := s.annotationsGroupRepository.Update(ctx, organizationId, annotationsGroupId, request.AnnotationsGroupUpsertRequest)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrFailedToUpdateAnnotationsGroup.Error())
	}

	return fetchedAnnotationsGroup, nil
}

func (s annotationsGroupService) Delete(ctx context.Context, organizationId string, annotationsGroupId string) error {
	if err := s.checkID(annotationsGroupId); err != nil {
		return errors.Wrap(err, annotations_group.ErrFailedToDeleteAnnotationsGroup.Error())
	}

	err := s.annotationsGroupRepository.Delete(ctx, organizationId, annotationsGroupId)
	return err
}

func (s annotationsGroupService) checkID(annotationsGroupId string) error {
	if annotationsGroupId == "" {
		return annotations_group.ErrInvalidAnnotationsGroupIdParam
	}

	if _, err := uuid.Parse(annotationsGroupId); err != nil {
		return errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupIdParam.Error())
	}

	return nil
}
