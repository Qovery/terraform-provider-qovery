//go:build unit || !integration

package qovery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

func TestConvertDomainMemberToOrganizationMember_RoleIdFallback(t *testing.T) {
	orgID := uuid.New()
	apiRoleID := "11111111-1111-1111-1111-111111111111"
	priorRoleID := "22222222-2222-2222-2222-222222222222"

	testCases := []struct {
		name           string
		domainRoleID   *string
		priorRoleId    types.String
		expectedRoleId types.String
	}{
		{
			name:           "api role wins when present",
			domainRoleID:   &apiRoleID,
			priorRoleId:    types.StringValue(priorRoleID),
			expectedRoleId: types.StringValue(apiRoleID),
		},
		{
			name:           "prior role preserved when api role is nil",
			domainRoleID:   nil,
			priorRoleId:    types.StringValue(priorRoleID),
			expectedRoleId: types.StringValue(priorRoleID),
		},
		{
			name:           "null when api role is nil and no prior role (import)",
			domainRoleID:   nil,
			priorRoleId:    types.StringNull(),
			expectedRoleId: types.StringNull(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			domainMember := member.Member{
				ID:               "invite-id",
				OrganizationID:   orgID,
				Email:            "USER@example.com",
				RoleID:           tc.domainRoleID,
				UserID:           nil,
				InvitationStatus: "PENDING",
			}
			configEmail := types.StringValue("user@example.com")

			got := convertDomainMemberToOrganizationMember(domainMember, configEmail, tc.priorRoleId)

			assert.Equal(t, tc.expectedRoleId, got.RoleId)
			assert.Equal(t, configEmail, got.Email)
			assert.Equal(t, types.StringValue("invite-id"), got.ID)
			assert.Equal(t, types.StringValue(orgID.String()), got.OrganizationId)
			assert.Equal(t, types.StringNull(), got.UserId)
			assert.Equal(t, types.StringValue("PENDING"), got.InvitationStatus)
		})
	}
}
