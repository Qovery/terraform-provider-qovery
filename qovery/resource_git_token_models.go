package qovery

import (
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
)

type GitToken struct {
	ID                 types.String `tfsdk:"id"`
	OrganizationId     types.String `tfsdk:"organization_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	Token              types.String `tfsdk:"token"`
	BitbucketWorkspace types.String `tfsdk:"bitbucket_workspace"`
}

func (it GitToken) toUpsertRequest() gittoken.GitTokenParams {
	return gittoken.GitTokenParams{
		Name:               ToString(it.Name),
		Description:        ToStringPointer(it.Description),
		Type:               ToString(it.Type),
		Token:              ToString(it.Token),
		BitbucketWorkspace: ToStringPointer(it.BitbucketWorkspace),
	}
}

func toTerraformObject(organizationID string, token string, gitTokenResponse qovery.GitTokenResponse) GitToken {
	return GitToken{
		ID:                 FromString(gitTokenResponse.Id),
		OrganizationId:     FromString(organizationID),
		Name:               FromString(gitTokenResponse.Name),
		Description:        FromStringPointer(gitTokenResponse.Description),
		Type:               FromString(string(gitTokenResponse.Type)),
		Token:              FromString(token),
		BitbucketWorkspace: FromStringPointer(gitTokenResponse.Workspace),
	}
}

var errGitTokenBitbucketWorkspaceRequired = errors.New("bitbucket_workspace is required when type is BITBUCKET")

// validateGitTokenWorkspace rejects a BITBUCKET token without a workspace, as q-core does for a
// null or blank one. Unknown values pass: they are checked again once known.
func validateGitTokenWorkspace(tokenType types.String, workspace types.String) error {
	if tokenType.IsUnknown() || workspace.IsUnknown() {
		return nil
	}
	if tokenType.ValueString() == string(gittoken.BITBUCKET) && strings.TrimSpace(workspace.ValueString()) == "" {
		return errGitTokenBitbucketWorkspaceRequired
	}
	return nil
}

// gitTokenStateFromAPI converts the git token for the resource. prior is the plan on apply and
// the state on refresh. The API never returns the token value, so it is kept from prior.
// description and bitbucket_workspace are Optional only: an empty value from the API keeps the
// null of prior.
func gitTokenStateFromAPI(response qovery.GitTokenResponse, prior GitToken) GitToken {
	state := toTerraformObject(prior.OrganizationId.ValueString(), prior.Token.ValueString(), response)
	state.Token = prior.Token
	state.Description = optionalStringFromAPI(prior.Description, response.Description)
	state.BitbucketWorkspace = optionalStringFromAPI(prior.BitbucketWorkspace, response.Workspace)
	return state
}
