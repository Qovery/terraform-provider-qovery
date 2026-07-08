package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

type ApiToken struct {
	ID             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	RoleId         types.String `tfsdk:"role_id"`
	Token          types.String `tfsdk:"token"`
}

func (t ApiToken) toCreateRequest() apitoken.CreateRequest {
	return apitoken.CreateRequest{
		Name:        ToString(t.Name),
		Description: ToStringPointer(t.Description),
		RoleID:      ToString(t.RoleId),
	}
}

// convertDomainApiTokenToApiToken converts a domain api token into its terraform model.
// The API returns the secret token value only at creation time, so on Read the value kept
// in tokenFromState is preserved instead of the (always nil) value coming from the API.
func convertDomainApiTokenToApiToken(apiToken apitoken.ApiToken, tokenFromState types.String) ApiToken {
	token := tokenFromState
	if apiToken.Token != nil {
		token = FromStringPointer(apiToken.Token)
	}

	return ApiToken{
		ID:             FromString(apiToken.ID.String()),
		OrganizationId: FromString(apiToken.OrganizationID.String()),
		Name:           FromString(apiToken.Name),
		Description:    FromStringPointer(apiToken.Description),
		RoleId:         FromString(apiToken.RoleID),
		Token:          token,
	}
}
