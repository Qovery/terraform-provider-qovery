//go:build unit && !integration

package member_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

func TestInviteRequestValidate(t *testing.T) {
	t.Parallel()

	validRequest := func() member.InviteRequest {
		return member.InviteRequest{
			Email:  "dev@company.com",
			RoleID: "01234567-8901-2345-6789-012345678901",
		}
	}

	testCases := []struct {
		TestName    string
		Mutate      func(r *member.InviteRequest)
		ExpectedErr string
	}{
		{
			TestName: "success",
			Mutate:   func(r *member.InviteRequest) {},
		},
		{
			TestName:    "error_empty_email",
			Mutate:      func(r *member.InviteRequest) { r.Email = "" },
			ExpectedErr: member.ErrInvalidInviteRequest.Error(),
		},
		{
			TestName:    "error_malformed_email",
			Mutate:      func(r *member.InviteRequest) { r.Email = "not-an-email" },
			ExpectedErr: member.ErrInvalidInviteRequest.Error(),
		},
		{
			TestName:    "error_empty_role_id",
			Mutate:      func(r *member.InviteRequest) { r.RoleID = "" },
			ExpectedErr: member.ErrInvalidInviteRequest.Error(),
		},
		{
			TestName:    "error_malformed_role_id",
			Mutate:      func(r *member.InviteRequest) { r.RoleID = "not-a-uuid" },
			ExpectedErr: member.ErrInvalidInviteRequest.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			request := validRequest()
			tc.Mutate(&request)
			err := request.Validate()
			if tc.ExpectedErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.ExpectedErr)
		})
	}
}

func TestUpdateRoleRequestValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		RoleID      string
		ExpectedErr string
	}{
		{TestName: "success", RoleID: "01234567-8901-2345-6789-012345678901"},
		{TestName: "error_empty_role_id", RoleID: "", ExpectedErr: member.ErrInvalidUpdateRoleRequest.Error()},
		{TestName: "error_malformed_role_id", RoleID: "not-a-uuid", ExpectedErr: member.ErrInvalidUpdateRoleRequest.Error()},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			err := member.UpdateRoleRequest{RoleID: tc.RoleID}.Validate()
			if tc.ExpectedErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.ExpectedErr)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	assert.NoError(t, member.ValidateEmail("dev@company.com"))
	assert.ErrorContains(t, member.ValidateEmail(""), member.ErrInvalidEmailParam.Error())
	assert.ErrorContains(t, member.ValidateEmail("nope"), member.ErrInvalidEmailParam.Error())
}
