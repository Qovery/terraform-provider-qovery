package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

// newDomainMemberFromInvite converts a pending invitation into a domain Member.
// Legacy invitations can carry a nil role id; requestedRoleID (the role sent in the
// invite request, when known) is used as fallback so a fresh Create never yields a
// state role_id diverging from the plan.
func newDomainMemberFromInvite(organizationID string, invite *qovery.InviteMember, requestedRoleID *string) (*member.Member, error) {
	if invite == nil {
		return nil, errors.Wrap(errors.New("nil invite"), member.ErrInvalidMember.Error())
	}
	orgID, err := parseUUID(organizationID, member.ErrInvalidOrganizationIdParam)
	if err != nil {
		return nil, err
	}

	roleID := invite.RoleId
	if roleID == nil {
		roleID = requestedRoleID
	}

	domainMember := &member.Member{
		ID:               invite.Id,
		OrganizationID:   orgID,
		Email:            invite.Email,
		RoleID:           roleID,
		UserID:           nil,
		InvitationStatus: string(invite.InvitationStatus),
	}
	if err := domainMember.Validate(); err != nil {
		return nil, err
	}
	return domainMember, nil
}

// newDomainMemberFromQoveryMember converts an active organization member into a domain
// Member. The API member id is the identity provider subject and doubles as the user_id
// expected by the role-edit and delete endpoints.
func newDomainMemberFromQoveryMember(organizationID string, m qovery.Member) (*member.Member, error) {
	orgID, err := parseUUID(organizationID, member.ErrInvalidOrganizationIdParam)
	if err != nil {
		return nil, err
	}

	userID := m.Id
	domainMember := &member.Member{
		ID:               m.Id,
		OrganizationID:   orgID,
		Email:            m.Email,
		RoleID:           m.RoleId,
		UserID:           &userID,
		InvitationStatus: member.StatusAccepted,
	}
	if err := domainMember.Validate(); err != nil {
		return nil, err
	}
	return domainMember, nil
}
