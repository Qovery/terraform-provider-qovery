package qoveryapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

// Ensure customRoleQoveryAPI defined type fully satisfy the customrole.Repository interface.
var _ customrole.Repository = customRoleQoveryAPI{}

// customRoleQoveryAPI implements the interface customrole.Repository.
type customRoleQoveryAPI struct {
	client *qovery.APIClient
}

func newCustomRoleQoveryAPI(client *qovery.APIClient) (customrole.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &customRoleQoveryAPI{client: client}, nil
}

// Create POSTs name+description (the only fields the create endpoint accepts — the server
// seeds a full default matrix), then PUTs the declared permissions overlaid on that matrix.
func (c customRoleQoveryAPI) Create(ctx context.Context, organizationID string, request customrole.UpsertRequest) (*customrole.CustomRole, error) {
	created, resp, err := c.client.OrganizationCustomRoleAPI.
		CreateOrganizationCustomRole(ctx, organizationID).
		OrganizationCustomRoleCreateRequest(qovery.OrganizationCustomRoleCreateRequest{
			Name:        request.Name,
			Description: request.Description,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceOrganizationCustomRole, request.Name, resp, err)
	}

	// Nothing declared: server defaults are already the desired state, skip the PUT.
	if len(request.ClusterPermissions) == 0 && len(request.ProjectPermissions) == 0 {
		return newDomainCustomRoleFromQovery(organizationID, created)
	}

	role, err := c.edit(ctx, organizationID, created.GetId(), created, request)
	if err != nil {
		// Best-effort cleanup: without it a retry would collide on the unique role name.
		_, _ = c.client.OrganizationCustomRoleAPI.DeleteOrganizationCustomRole(ctx, organizationID, created.GetId()).Execute()
		return nil, errors.Wrap(err, customrole.ErrFailedToCreateCustomRole.Error())
	}
	return role, nil
}

func (c customRoleQoveryAPI) Get(ctx context.Context, organizationID string, customRoleID string) (*customrole.CustomRole, error) {
	role, resp, err := c.client.OrganizationCustomRoleAPI.
		GetOrganizationCustomRole(ctx, organizationID, customRoleID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceOrganizationCustomRole, customRoleID, resp, err)
	}
	return newDomainCustomRoleFromQovery(organizationID, role)
}

func (c customRoleQoveryAPI) Update(ctx context.Context, organizationID string, customRoleID string, request customrole.UpsertRequest) (*customrole.CustomRole, error) {
	current, resp, err := c.client.OrganizationCustomRoleAPI.
		GetOrganizationCustomRole(ctx, organizationID, customRoleID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceOrganizationCustomRole, customRoleID, resp, err)
	}
	return c.edit(ctx, organizationID, customRoleID, current, request)
}

func (c customRoleQoveryAPI) Delete(ctx context.Context, organizationID string, customRoleID string) error {
	resp, err := c.client.OrganizationCustomRoleAPI.
		DeleteOrganizationCustomRole(ctx, organizationID, customRoleID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceOrganizationCustomRole, customRoleID, resp, err)
	}
	return nil
}

// edit builds the full-replace PUT payload from the authoritative `current` role and applies it.
// It guards against a nil `current` (nil body on a 2xx) before handing it to
// newQoveryCustomRoleEditRequestFrom, which dereferences it without its own nil check.
func (c customRoleQoveryAPI) edit(ctx context.Context, organizationID string, customRoleID string, current *qovery.OrganizationCustomRole, request customrole.UpsertRequest) (*customrole.CustomRole, error) {
	if current == nil {
		return nil, customrole.ErrInvalidCustomRole
	}
	editRequest, err := newQoveryCustomRoleEditRequestFrom(current, request)
	if err != nil {
		return nil, err
	}
	updated, resp, err := c.client.OrganizationCustomRoleAPI.
		EditOrganizationCustomRole(ctx, organizationID, customRoleID).
		OrganizationCustomRoleUpdateRequest(*editRequest).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceOrganizationCustomRole, customRoleID, resp, err)
	}
	return newDomainCustomRoleFromQovery(organizationID, updated)
}
