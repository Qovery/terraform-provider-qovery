package qoveryapi

import (
	"context"
	"github.com/qovery/terraform-provider-qovery/internal/domain/custom_organization_role"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

var _ custom_organization_role.Repository = customOrganizationRoleQoveryAPI{}

type customOrganizationRoleQoveryAPI struct {
	client *qovery.APIClient
}

func newCustomOrganizationRoleQoveryAPI(client *qovery.APIClient) (custom_organization_role.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &customOrganizationRoleQoveryAPI{
		client: client,
	}, nil
}

func (c customOrganizationRoleQoveryAPI) Create(ctx context.Context, organizationId string, request custom_organization_role.UpsertRequest) (*custom_organization_role.CustomOrganizationRole, error) {
	req, err := newQoveryCustomOrganizationRoleRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleRequest.Error())
	}

	newCustomOrganizationRole, resp, err := c.client.OrganizationCustomRoleAPI.
		CreateOrganizationCustomRole(ctx, organizationId).
		OrganizationCustomRoleCreateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APICustomOrganizationRole, request.Name, resp, err)
	}

	updateReq, err := newQoveryCustomOrganizationRoleUpdateRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleRequest.Error())
	}

	customOrganizationRole, resp, err := c.client.OrganizationCustomRoleAPI.
		EditOrganizationCustomRole(ctx, organizationId, *newCustomOrganizationRole.Id).
		OrganizationCustomRoleUpdateRequest(*updateReq).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APICustomOrganizationRole, *newCustomOrganizationRole.Id, resp, err)
	}

	return newDomainCustomOrganizationRoleFromQovery(customOrganizationRole)
}

func (c customOrganizationRoleQoveryAPI) Get(ctx context.Context, organizationGroupId string, customOrganizationRoleId string) (*custom_organization_role.CustomOrganizationRole, error) {
	customOrganizationRole, resp, err := c.client.OrganizationCustomRoleAPI.
		GetOrganizationCustomRole(ctx, organizationGroupId, customOrganizationRoleId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APICustomOrganizationRole, customOrganizationRoleId, resp, err)
	}

	return newDomainCustomOrganizationRoleFromQovery(customOrganizationRole)
}

func (c customOrganizationRoleQoveryAPI) Update(ctx context.Context, organizationId string, customOrganizationRoleId string, request custom_organization_role.UpsertRequest) (*custom_organization_role.CustomOrganizationRole, error) {
	req, err := newQoveryCustomOrganizationRoleUpdateRequestFromDomain(request)
	if err != nil {
		return nil, errors.Wrap(err, custom_organization_role.ErrInvalidCustomOrganizationRoleRequest.Error())
	}

	customOrganizationRole, resp, err := c.client.OrganizationCustomRoleAPI.
		EditOrganizationCustomRole(ctx, organizationId, customOrganizationRoleId).
		OrganizationCustomRoleUpdateRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APICustomOrganizationRole, customOrganizationRoleId, resp, err)
	}

	return newDomainCustomOrganizationRoleFromQovery(customOrganizationRole)
}

func (c customOrganizationRoleQoveryAPI) Delete(ctx context.Context, organizationId string, customOrganizationRoleId string) error {
	_, resp, err := c.client.OrganizationCustomRoleAPI.
		GetOrganizationCustomRole(ctx, organizationId, customOrganizationRoleId).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil
		}
		return apierrors.NewDeleteAPIError(apierrors.APICustomOrganizationRole, customOrganizationRoleId, resp, err)
	}

	resp, err = c.client.OrganizationCustomRoleAPI.
		DeleteOrganizationCustomRole(ctx, organizationId, customOrganizationRoleId).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APICustomOrganizationRole, customOrganizationRoleId, resp, err)
	}

	return nil
}
