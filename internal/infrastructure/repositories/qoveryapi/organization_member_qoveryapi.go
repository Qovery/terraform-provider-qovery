package qoveryapi

import (
	"context"
	"strings"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

// Ensure organizationMemberQoveryAPI defined type fully satisfy the member.Repository interface.
var _ member.Repository = organizationMemberQoveryAPI{}

// organizationMemberQoveryAPI implements the interface member.Repository.
type organizationMemberQoveryAPI struct {
	client *qovery.APIClient
}

func newOrganizationMemberQoveryAPI(client *qovery.APIClient) (member.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &organizationMemberQoveryAPI{client: client}, nil
}

// Create calls Qovery's API to invite a member to the organization.
func (a organizationMemberQoveryAPI) Create(ctx context.Context, organizationID string, request member.InviteRequest) (*member.Member, error) {
	roleID := request.RoleID
	res, resp, err := a.client.MembersAPI.
		PostInviteMember(ctx, organizationID).
		InviteMemberRequest(qovery.InviteMemberRequest{
			Email:  request.Email,
			RoleId: &roleID,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceOrganizationMember, request.Email, resp, err)
	}

	return newDomainMemberFromInvite(organizationID, res, &roleID)
}

// Get retrieves a member by email. The API exposes no get-single endpoint, so the pending
// invitations are listed first (invite lifecycle state), then the active members.
func (a organizationMemberQoveryAPI) Get(ctx context.Context, organizationID string, email string) (*member.Member, error) {
	invites, resp, err := a.client.MembersAPI.
		GetOrganizationInvitedMembers(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
	}
	for _, invite := range invites.GetResults() {
		if strings.EqualFold(invite.Email, email) {
			return newDomainMemberFromInvite(organizationID, &invite, nil)
		}
	}

	members, resp, err := a.client.MembersAPI.
		GetOrganizationMembers(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
	}
	for _, m := range members.GetResults() {
		if strings.EqualFold(m.Email, email) {
			return newDomainMemberFromQoveryMember(organizationID, m)
		}
	}

	return nil, apierrors.NewNotFoundAPIError(apierrors.APIResourceOrganizationMember, email)
}

// Update changes the role of a member. Active members have a dedicated endpoint; a pending
// invitation has none, so it is deleted and re-created with the new role (the invite id —
// and therefore the resource id — changes).
func (a organizationMemberQoveryAPI) Update(ctx context.Context, organizationID string, email string, request member.UpdateRoleRequest) (*member.Member, error) {
	current, err := a.Get(ctx, organizationID, email)
	if err != nil {
		return nil, err
	}

	if current.UserID == nil {
		resp, err := a.client.MembersAPI.
			DeleteInviteMember(ctx, organizationID, current.ID).
			Execute()
		if err != nil || resp.StatusCode >= 300 {
			return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
		}
		return a.Create(ctx, organizationID, member.InviteRequest{Email: current.Email, RoleID: request.RoleID})
	}

	resp, err := a.client.MembersAPI.
		EditOrganizationMemberRole(ctx, organizationID).
		MemberRoleUpdateRequest(qovery.MemberRoleUpdateRequest{
			UserId: *current.UserID,
			RoleId: request.RoleID,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
	}

	// The member list is served from Auth0, so this read-after-write may lag and return the old
	// role. The edit above succeeded, so the requested role is authoritative: force it to avoid a
	// "provider produced inconsistent result after apply" error on the (Required, known) role_id.
	updated, err := a.Get(ctx, organizationID, email)
	if err != nil {
		return nil, err
	}
	roleID := request.RoleID
	updated.RoleID = &roleID
	return updated, nil
}

// Delete removes a member from the organization: a pending invitation is cancelled, an
// active member is removed by user id.
func (a organizationMemberQoveryAPI) Delete(ctx context.Context, organizationID string, email string) error {
	current, err := a.Get(ctx, organizationID, email)
	if err != nil {
		return err
	}

	if current.UserID == nil {
		resp, err := a.client.MembersAPI.
			DeleteInviteMember(ctx, organizationID, current.ID).
			Execute()
		if err != nil || resp.StatusCode >= 300 {
			return apierrors.NewDeleteAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
		}
		return nil
	}

	resp, err := a.client.MembersAPI.
		DeleteMember(ctx, organizationID).
		DeleteMemberRequest(qovery.DeleteMemberRequest{UserId: *current.UserID}).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceOrganizationMember, email, resp, err)
	}
	return nil
}
