package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

var _ argoCdCredentials.Service = argoCdCredentialsService{}

type argoCdCredentialsService struct {
	repo argoCdCredentials.Repository
}

func NewArgoCdCredentialsService(repo argoCdCredentials.Repository) (argoCdCredentials.Service, error) {
	if repo == nil {
		return nil, ErrInvalidRepository
	}
	return &argoCdCredentialsService{repo: repo}, nil
}

func (s argoCdCredentialsService) Create(ctx context.Context, clusterID string, request argoCdCredentials.UpsertRequest) (*argoCdCredentials.ArgoCdCredentials, error) {
	if err := s.checkClusterID(clusterID); err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToCreateArgoCdCredentials.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToCreateArgoCdCredentials.Error())
	}
	res, err := s.repo.Create(ctx, clusterID, request)
	if err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToCreateArgoCdCredentials.Error())
	}
	return res, nil
}

func (s argoCdCredentialsService) Get(ctx context.Context, clusterID string) (*argoCdCredentials.ArgoCdCredentials, error) {
	if err := s.checkClusterID(clusterID); err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToGetArgoCdCredentials.Error())
	}
	res, err := s.repo.Get(ctx, clusterID)
	if err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToGetArgoCdCredentials.Error())
	}
	return res, nil
}

func (s argoCdCredentialsService) Update(ctx context.Context, clusterID string, request argoCdCredentials.UpsertRequest) (*argoCdCredentials.ArgoCdCredentials, error) {
	if err := s.checkClusterID(clusterID); err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToUpdateArgoCdCredentials.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToUpdateArgoCdCredentials.Error())
	}
	res, err := s.repo.Update(ctx, clusterID, request)
	if err != nil {
		return nil, errors.Wrap(err, argoCdCredentials.ErrFailedToUpdateArgoCdCredentials.Error())
	}
	return res, nil
}

func (s argoCdCredentialsService) Delete(ctx context.Context, clusterID string) error {
	if err := s.checkClusterID(clusterID); err != nil {
		return errors.Wrap(err, argoCdCredentials.ErrFailedToDeleteArgoCdCredentials.Error())
	}
	if err := s.repo.Delete(ctx, clusterID); err != nil {
		return errors.Wrap(err, argoCdCredentials.ErrFailedToDeleteArgoCdCredentials.Error())
	}
	return nil
}

func (s argoCdCredentialsService) checkClusterID(clusterID string) error {
	return validateUUIDParam(clusterID, argoCdCredentials.ErrInvalidClusterIdParam)
}
