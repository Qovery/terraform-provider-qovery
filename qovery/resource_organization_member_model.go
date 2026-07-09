package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

type OrganizationMember struct {
	ID               types.String `tfsdk:"id"`
	OrganizationId   types.String `tfsdk:"organization_id"`
	Email            types.String `tfsdk:"email"`
	RoleId           types.String `tfsdk:"role_id"`
	UserId           types.String `tfsdk:"user_id"`
	InvitationStatus types.String `tfsdk:"invitation_status"`
}

func (m OrganizationMember) toInviteRequest() member.InviteRequest {
	return member.InviteRequest{
		Email:  ToString(m.Email),
		RoleID: ToString(m.RoleId),
	}
}

func (m OrganizationMember) toUpdateRoleRequest() member.UpdateRoleRequest {
	return member.UpdateRoleRequest{
		RoleID: ToString(m.RoleId),
	}
}

func convertDomainMemberToOrganizationMember(m member.Member) OrganizationMember {
	return OrganizationMember{
		ID:               FromString(m.ID),
		OrganizationId:   FromString(m.OrganizationID.String()),
		Email:            FromString(m.Email),
		RoleId:           FromStringPointer(m.RoleID),
		UserId:           FromStringPointer(m.UserID),
		InvitationStatus: FromString(m.InvitationStatus),
	}
}
