//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
)

func TestValidateGitTokenWorkspace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Type        types.String
		Workspace   types.String
		ExpectError bool
	}{
		{TestName: "bitbucket_with_a_workspace", Type: types.StringValue("BITBUCKET"), Workspace: types.StringValue("team"), ExpectError: false},
		{TestName: "bitbucket_without_a_workspace", Type: types.StringValue("BITBUCKET"), Workspace: types.StringNull(), ExpectError: true},
		{TestName: "bitbucket_with_a_blank_workspace", Type: types.StringValue("BITBUCKET"), Workspace: types.StringValue(" "), ExpectError: true},
		{TestName: "bitbucket_with_an_unknown_workspace", Type: types.StringValue("BITBUCKET"), Workspace: types.StringUnknown(), ExpectError: false},
		{TestName: "unknown_type", Type: types.StringUnknown(), Workspace: types.StringNull(), ExpectError: false},
		{TestName: "github_without_a_workspace", Type: types.StringValue("GITHUB"), Workspace: types.StringNull(), ExpectError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			err := validateGitTokenWorkspace(tc.Type, tc.Workspace)
			if tc.ExpectError {
				assert.ErrorIs(t, err, errGitTokenBitbucketWorkspaceRequired)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGitTokenStateFromAPI(t *testing.T) {
	t.Parallel()

	response := func(description, workspace *string) qovery.GitTokenResponse {
		return qovery.GitTokenResponse{Id: "token-id", Name: "token", Type: qovery.GITPROVIDERENUM_BITBUCKET, Description: description, Workspace: workspace}
	}

	testCases := []struct {
		TestName          string
		Response          qovery.GitTokenResponse
		Prior             GitToken
		ExpectDescription types.String
		ExpectWorkspace   types.String
	}{
		{
			TestName:          "api_values_win",
			Response:          response(new("set from the Console"), new("console-team")),
			Prior:             GitToken{Description: types.StringValue("old"), BitbucketWorkspace: types.StringValue("team")},
			ExpectDescription: types.StringValue("set from the Console"),
			ExpectWorkspace:   types.StringValue("console-team"),
		},
		{
			TestName:          "a_description_cleared_from_the_console_shows_up",
			Response:          response(nil, new("team")),
			Prior:             GitToken{Description: types.StringValue("old"), BitbucketWorkspace: types.StringValue("team")},
			ExpectDescription: types.StringNull(),
			ExpectWorkspace:   types.StringValue("team"),
		},
		{
			TestName:          "an_empty_api_description_keeps_a_null_prior",
			Response:          response(new(""), nil),
			Prior:             GitToken{Description: types.StringNull(), BitbucketWorkspace: types.StringNull()},
			ExpectDescription: types.StringNull(),
			ExpectWorkspace:   types.StringNull(),
		},
		{
			TestName:          "import_records_the_remote_values",
			Response:          response(new("set from the Console"), new("team")),
			Prior:             GitToken{OrganizationId: types.StringValue("organization-id")},
			ExpectDescription: types.StringValue("set from the Console"),
			ExpectWorkspace:   types.StringValue("team"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			tc.Prior.OrganizationId = types.StringValue("organization-id")
			got := gitTokenStateFromAPI(tc.Response, tc.Prior)

			assert.Equal(t, types.StringValue("token-id"), got.ID)
			assert.Equal(t, types.StringValue("organization-id"), got.OrganizationId)
			assert.Equal(t, types.StringValue("BITBUCKET"), got.Type)
			assert.Equal(t, tc.ExpectDescription, got.Description)
			assert.Equal(t, tc.ExpectWorkspace, got.BitbucketWorkspace)
			assert.Equal(t, tc.Prior.Token, got.Token, "the API never returns the token: it comes from prior")
		})
	}
}
