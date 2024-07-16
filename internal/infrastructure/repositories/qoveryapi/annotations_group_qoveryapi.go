package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

var _ annotations_group.Repository = annotationsGroupQoveryAPI{}

type annotationsGroupQoveryAPI struct {
	client *qovery.APIClient
}

func newAnnotationsGroupQoveryAPI(client *qovery.APIClient) (annotations_group.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &annotationsGroupQoveryAPI{
		client: client,
	}, nil
}

func (c annotationsGroupQoveryAPI) Create(ctx context.Context, organizationId string, request annotations_group.UpsertRequest) (*annotations_group.AnnotationsGroup, error) {
	req, err := newQoveryAnnotationsGroupRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupRequest.Error())
	}

	newAnnotationsGroup, resp, err := c.client.OrganizationAnnotationsGroupAPI.
		CreateOrganizationAnnotationsGroup(ctx, organizationId).
		OrganizationAnnotationsGroupCreateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceAnnotationsGroup, request.Name, resp, err)
	}

	return newDomainAnnotationsGroupFromQovery(newAnnotationsGroup)
}

func (c annotationsGroupQoveryAPI) Get(ctx context.Context, organizationGroupId string, annotationsGroupId string) (*annotations_group.AnnotationsGroup, error) {
	annotationsGroup, resp, err := c.client.OrganizationAnnotationsGroupAPI.
		GetOrganizationAnnotationsGroup(ctx, organizationGroupId, annotationsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceAnnotationsGroup, annotationsGroupId, resp, err)
	}

	return newDomainAnnotationsGroupFromQovery(annotationsGroup)
}

func (c annotationsGroupQoveryAPI) Update(ctx context.Context, organizationGroupId string, annotationsGroupId string, request annotations_group.UpsertRequest) (*annotations_group.AnnotationsGroup, error) {
	req, err := newQoveryAnnotationsGroupRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, annotations_group.ErrInvalidAnnotationsGroupRequest.Error())
	}

	annotationsGroup, resp, err := c.client.OrganizationAnnotationsGroupAPI.
		EditOrganizationAnnotationsGroup(ctx, organizationGroupId, annotationsGroupId).
		OrganizationAnnotationsGroupCreateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceAnnotationsGroup, annotationsGroupId, resp, err)
	}

	return newDomainAnnotationsGroupFromQovery(annotationsGroup)
}

func (c annotationsGroupQoveryAPI) Delete(ctx context.Context, organizationId string, annotationsGroupId string) error {
	_, resp, err := c.client.OrganizationAnnotationsGroupAPI.
		GetOrganizationAnnotationsGroup(ctx, organizationId, annotationsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceAnnotationsGroup, annotationsGroupId, resp, err)
	}

	resp, err = c.client.OrganizationAnnotationsGroupAPI.
		DeleteOrganizationAnnotationsGroup(ctx, organizationId, annotationsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceAnnotationsGroup, annotationsGroupId, resp, err)
	}

	return nil
}
