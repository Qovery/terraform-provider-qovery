package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
)

var _ labels_group.Repository = labelsGroupQoveryAPI{}

type labelsGroupQoveryAPI struct {
	client *qovery.APIClient
}

func newLabelsGroupQoveryAPI(client *qovery.APIClient) (labels_group.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &labelsGroupQoveryAPI{
		client: client,
	}, nil
}

func (c labelsGroupQoveryAPI) Create(ctx context.Context, organizationId string, request labels_group.UpsertRequest) (*labels_group.LabelsGroup, error) {
	req := newQoveryLabelsGroupRequestFromDomain(request)

	newLabelsGroup, resp, err := c.client.OrganizationLabelsGroupAPI.
		CreateOrganizationLabelsGroup(ctx, organizationId).
		OrganizationLabelsGroupCreateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceLabelsGroup, request.Name, resp, err)
	}

	return newDomainLabelsGroupFromQovery(newLabelsGroup)
}

func (c labelsGroupQoveryAPI) Get(ctx context.Context, organizationGroupId string, labelsGroupId string) (*labels_group.LabelsGroup, error) {
	labelsGroup, resp, err := c.client.OrganizationLabelsGroupAPI.
		GetOrganizationLabelssGroup(ctx, organizationGroupId, labelsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceLabelsGroup, labelsGroupId, resp, err)
	}

	return newDomainLabelsGroupFromQovery(labelsGroup)
}

func (c labelsGroupQoveryAPI) Update(ctx context.Context, organizationGroupId string, labelsGroupId string, request labels_group.UpsertRequest) (*labels_group.LabelsGroup, error) {
	req := newQoveryLabelsGroupRequestFromDomain(request)

	labelsGroup, resp, err := c.client.OrganizationLabelsGroupAPI.
		EditOrganizationLabelsGroup(ctx, organizationGroupId, labelsGroupId).
		OrganizationLabelsGroupCreateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceLabelsGroup, labelsGroupId, resp, err)
	}

	return newDomainLabelsGroupFromQovery(labelsGroup)
}

func (c labelsGroupQoveryAPI) Delete(ctx context.Context, organizationId string, labelsGroupId string) error {
	_, resp, err := c.client.OrganizationLabelsGroupAPI.
		GetOrganizationLabelssGroup(ctx, organizationId, labelsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APIResourceLabelsGroup, labelsGroupId, resp, err)
	}

	resp, err = c.client.OrganizationLabelsGroupAPI.
		DeleteOrganizationLabelsGroup(ctx, organizationId, labelsGroupId).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceLabelsGroup, labelsGroupId, resp, err)
	}

	return nil
}
