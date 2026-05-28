package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

var _ argoCdDestinationClusterMapping.Service = argoCdDestinationClusterMappingService{}

type argoCdDestinationClusterMappingService struct {
	repo argoCdDestinationClusterMapping.Repository
}

func NewArgoCdDestinationClusterMappingService(repo argoCdDestinationClusterMapping.Repository) (argoCdDestinationClusterMapping.Service, error) {
	if repo == nil {
		return nil, ErrInvalidRepository
	}
	return &argoCdDestinationClusterMappingService{repo: repo}, nil
}

func (s argoCdDestinationClusterMappingService) Create(ctx context.Context, orgID string, request argoCdDestinationClusterMapping.UpsertRequest) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	if err := s.checkOrgID(orgID); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToCreateArgoCdDestinationClusterMapping.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToCreateArgoCdDestinationClusterMapping.Error())
	}
	res, err := s.repo.Create(ctx, orgID, request)
	if err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToCreateArgoCdDestinationClusterMapping.Error())
	}
	return res, nil
}

func (s argoCdDestinationClusterMappingService) Get(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	if err := s.checkOrgID(orgID); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToGetArgoCdDestinationClusterMapping.Error())
	}
	if err := s.checkAgentClusterID(agentClusterID); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToGetArgoCdDestinationClusterMapping.Error())
	}
	res, err := s.repo.Get(ctx, orgID, agentClusterID, argocdClusterUrl)
	if err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToGetArgoCdDestinationClusterMapping.Error())
	}
	return res, nil
}

func (s argoCdDestinationClusterMappingService) Update(ctx context.Context, orgID string, request argoCdDestinationClusterMapping.UpsertRequest) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	if err := s.checkOrgID(orgID); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToUpdateArgoCdDestinationClusterMapping.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToUpdateArgoCdDestinationClusterMapping.Error())
	}
	res, err := s.repo.Update(ctx, orgID, request)
	if err != nil {
		return nil, errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToUpdateArgoCdDestinationClusterMapping.Error())
	}
	return res, nil
}

func (s argoCdDestinationClusterMappingService) Delete(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) error {
	if err := s.checkOrgID(orgID); err != nil {
		return errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToDeleteArgoCdDestinationClusterMapping.Error())
	}
	if err := s.checkAgentClusterID(agentClusterID); err != nil {
		return errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToDeleteArgoCdDestinationClusterMapping.Error())
	}
	if err := s.repo.Delete(ctx, orgID, agentClusterID, argocdClusterUrl); err != nil {
		return errors.Wrap(err, argoCdDestinationClusterMapping.ErrFailedToDeleteArgoCdDestinationClusterMapping.Error())
	}
	return nil
}

func (s argoCdDestinationClusterMappingService) checkOrgID(orgID string) error {
	if orgID == "" {
		return argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam
	}
	if _, err := uuid.Parse(orgID); err != nil {
		return errors.Wrap(err, argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam.Error())
	}
	return nil
}

func (s argoCdDestinationClusterMappingService) checkAgentClusterID(agentClusterID string) error {
	if agentClusterID == "" {
		return argoCdDestinationClusterMapping.ErrInvalidAgentClusterIdParam
	}
	if _, err := uuid.Parse(agentClusterID); err != nil {
		return errors.Wrap(err, argoCdDestinationClusterMapping.ErrInvalidAgentClusterIdParam.Error())
	}
	return nil
}
