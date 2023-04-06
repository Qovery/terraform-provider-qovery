package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

type Organization struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Plan        types.String `tfsdk:"plan"`
	Description types.String `tfsdk:"description"`
}

func (org Organization) toOrganizationUpdateRequest() organization.UpdateRequest {
	return organization.UpdateRequest{
		Name:        ToString(org.Name),
		Description: ToStringPointer(org.Description),
	}
}

func convertDomainOrganizationToTerraform(organization *organization.Organization) Organization {
	return Organization{
		Id:          FromString(organization.ID.String()),
		Name:        FromString(organization.Name),
		Plan:        fromClientEnum(organization.Plan),
		Description: FromStringPointer(organization.Description),
	}
}
