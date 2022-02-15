package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Organization struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Plan        types.String `tfsdk:"plan"`
	Description types.String `tfsdk:"description"`
}

func (org Organization) toCreateOrganizationRequest() qovery.OrganizationRequest {
	return qovery.OrganizationRequest{
		Name:        toString(org.Name),
		Plan:        toString(org.Plan),
		Description: toStringPointer(org.Description),
	}
}

func (org Organization) toUpdateOrganizationRequest() qovery.OrganizationEditRequest {
	return qovery.OrganizationEditRequest{
		Name:        toString(org.Name),
		Description: toStringPointer(org.Description),
	}
}

func convertResponseToOrganization(organization *qovery.OrganizationResponse) Organization {
	return Organization{
		Id:          fromString(organization.Id),
		Name:        fromString(organization.Name),
		Plan:        fromString(organization.Plan),
		Description: fromStringPointer(organization.Description),
	}
}
