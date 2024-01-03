package qoveryapi

import (
	"context"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

var _ helmRepository.Repository = helmRepositoryQoveryAPI{}

type helmRepositoryQoveryAPI struct {
	client *qovery.APIClient
}

func newHelmRepositoryQoveryAPI(client *qovery.APIClient) (helmRepository.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &helmRepositoryQoveryAPI{
		client: client,
	}, nil
}

func (h helmRepositoryQoveryAPI) Create(ctx context.Context, organizationID string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	req, err := newQoveryHelmRepositoryRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrInvalidUpsertRequest.Error())
	}

	reg, resp, err := h.client.HelmRepositoriesAPI.
		CreateHelmRepository(ctx, organizationID).
		HelmRepositoryRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewCreateAPIError(apierrors.APIResourceHelmRepository, request.Name, resp, err)
		return nil, apiErr
	}

	return newDomainHelmRepositoryFromQovery(reg, organizationID)
}

func (h helmRepositoryQoveryAPI) Get(ctx context.Context, organizationID string, repositoryId string) (*helmRepository.HelmRepository, error) {
	reg, resp, err := h.client.HelmRepositoriesAPI.
		GetHelmRepository(ctx, organizationID, repositoryId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceHelmRepository, repositoryId, resp, err)
	}

	return newDomainHelmRepositoryFromQovery(reg, organizationID)
}

func (h helmRepositoryQoveryAPI) Update(ctx context.Context, organizationID string, repositoryId string, request helmRepository.UpsertRequest) (*helmRepository.HelmRepository, error) {
	req, err := newQoveryHelmRepositoryRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, helmRepository.ErrInvalidUpsertRequest.Error())
	}

	reg, resp, err := h.client.HelmRepositoriesAPI.
		EditHelmRepository(ctx, organizationID, repositoryId).
		HelmRepositoryRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceHelmRepository, repositoryId, resp, err)
	}

	return newDomainHelmRepositoryFromQovery(reg, organizationID)
}

func (h helmRepositoryQoveryAPI) Delete(ctx context.Context, organizationID string, repositoryId string) error {
	resp, err := h.client.HelmRepositoriesAPI.
		DeleteHelmRepository(ctx, organizationID, repositoryId).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceHelmRepository, repositoryId, resp, err)
	}

	return nil
}
