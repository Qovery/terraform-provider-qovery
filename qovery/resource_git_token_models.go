package qovery

import (
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
