//go:build unit && !integration

package qoveryapi

import (
	"testing"
	"time"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

const testConvOrganizationID = "01234567-8901-2345-6789-012345678901"

func TestNewDomainMemberFromInvite(t *testing.T) {
	t.Parallel()

	roleID := "11234567-8901-2345-6789-012345678901"
	invite := &qovery.InviteMember{
		Id:               "21234567-8901-2345-6789-012345678901",
		CreatedAt:        time.Now(),
		Email:            "dev@company.com",
		Role:             qovery.INVITEMEMBERROLEENUM_ADMIN,
		InvitationLink:   "https://console.qovery.com/invite",
		InvitationStatus: qovery.INVITESTATUSENUM_PENDING,
		Inviter:          "someone",
		RoleId:           &roleID,
	}

	t.Run("nil invite", func(t *testing.T) {
		t.Parallel()
		res, err := newDomainMemberFromInvite(testConvOrganizationID, nil, nil)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidMember.Error())
	})

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		res, err := newDomainMemberFromInvite("invalid", invite, nil)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidOrganizationIdParam.Error())
	})

	t.Run("success with role id from api", func(t *testing.T) {
		t.Parallel()
		res, err := newDomainMemberFromInvite(testConvOrganizationID, invite, nil)
		assert.NoError(t, err)
		assert.Equal(t, invite.Id, res.ID)
		assert.Equal(t, testConvOrganizationID, res.OrganizationID.String())
		assert.Equal(t, invite.Email, res.Email)
		assert.Equal(t, &roleID, res.RoleID)
		assert.Nil(t, res.UserID)
		assert.Equal(t, member.StatusPending, res.InvitationStatus)
	})

	t.Run("role id falls back to requested value when api returns none", func(t *testing.T) {
		t.Parallel()
		legacy := *invite
		legacy.RoleId = nil
		requested := "31234567-8901-2345-6789-012345678901"
		res, err := newDomainMemberFromInvite(testConvOrganizationID, &legacy, &requested)
		assert.NoError(t, err)
		assert.Equal(t, &requested, res.RoleID)
	})

	t.Run("expired invite keeps expired status", func(t *testing.T) {
		t.Parallel()
		expired := *invite
		expired.InvitationStatus = qovery.INVITESTATUSENUM_EXPIRED
		res, err := newDomainMemberFromInvite(testConvOrganizationID, &expired, nil)
		assert.NoError(t, err)
		assert.Equal(t, member.StatusExpired, res.InvitationStatus)
	})
}

func TestNewDomainMemberFromQoveryMember(t *testing.T) {
	t.Parallel()

	roleID := "11234567-8901-2345-6789-012345678901"
	apiMember := qovery.Member{
		Id:        "github|12345",
		CreatedAt: time.Now(),
		Email:     "dev@company.com",
		RoleId:    &roleID,
	}

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		res, err := newDomainMemberFromQoveryMember("invalid", apiMember)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidOrganizationIdParam.Error())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		res, err := newDomainMemberFromQoveryMember(testConvOrganizationID, apiMember)
		assert.NoError(t, err)
		assert.Equal(t, "github|12345", res.ID)
		assert.NotNil(t, res.UserID)
		assert.Equal(t, "github|12345", *res.UserID)
		assert.Equal(t, &roleID, res.RoleID)
		assert.Equal(t, member.StatusAccepted, res.InvitationStatus)
	})
}
