package services

import (
	"context"

	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

// Ensure apiTokenService defined type fully satisfy the apitoken.Service interface.
var _ apitoken.Service = apiTokenService{}

// apiTokenService implements the interface apitoken.Service.
type apiTokenService struct {
	repo apitoken.Repository
}

func NewApiTokenService(repo apitoken.Repository) (apitoken.Service, error) {
	if repo == nil {
		return nil, ErrInvalidRepository
	}
	return &apiTokenService{repo: repo}, nil
}

func (s apiTokenService) Create(ctx context.Context, organizationID string, request apitoken.CreateRequest) (*apitoken.ApiToken, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToCreateApiToken.Error())
	}
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToCreateApiToken.Error())
	}
	res, err := s.repo.Create(ctx, organizationID, request)
	if err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToCreateApiToken.Error())
	}
	return res, nil
}

func (s apiTokenService) Get(ctx context.Context, organizationID string, apiTokenID string) (*apitoken.ApiToken, error) {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToGetApiToken.Error())
	}
	if err := s.checkApiTokenID(apiTokenID); err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToGetApiToken.Error())
	}
	res, err := s.repo.Get(ctx, organizationID, apiTokenID)
	if err != nil {
		return nil, errors.Wrap(err, apitoken.ErrFailedToGetApiToken.Error())
	}
	return res, nil
}

func (s apiTokenService) Delete(ctx context.Context, organizationID string, apiTokenID string) error {
	if err := s.checkOrganizationID(organizationID); err != nil {
		return errors.Wrap(err, apitoken.ErrFailedToDeleteApiToken.Error())
	}
	if err := s.checkApiTokenID(apiTokenID); err != nil {
		return errors.Wrap(err, apitoken.ErrFailedToDeleteApiToken.Error())
	}
	if err := s.repo.Delete(ctx, organizationID, apiTokenID); err != nil {
		return errors.Wrap(err, apitoken.ErrFailedToDeleteApiToken.Error())
	}
	return nil
}

func (s apiTokenService) checkOrganizationID(organizationID string) error {
	return validateUUIDParam(organizationID, apitoken.ErrInvalidOrganizationIdParam)
}

func (s apiTokenService) checkApiTokenID(apiTokenID string) error {
	return validateUUIDParam(apiTokenID, apitoken.ErrInvalidApiTokenIdParam)
}
