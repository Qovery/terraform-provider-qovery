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

// organizationStateFromAPI converts the organization for the resource. description is Optional
// only, so an empty description from the API keeps the null of prior, the plan on apply or the
// state on refresh.
func organizationStateFromAPI(orga *organization.Organization, prior Organization) Organization {
	state := convertDomainOrganizationToTerraform(orga)
	state.Description = optionalStringFromAPI(prior.Description, orga.Description)
	return state
}
