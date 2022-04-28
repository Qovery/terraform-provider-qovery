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

func (org Organization) toCreateOrganizationRequest() (*qovery.OrganizationRequest, error) {
	plan, err := qovery.NewPlanEnumFromValue(toString(org.Plan))
	if err != nil {
		return nil, err
	}

	return &qovery.OrganizationRequest{
		Name:        toString(org.Name),
		Plan:        *plan,
		Description: toStringPointer(org.Description),
	}, nil
}

func (org Organization) toUpdateOrganizationRequest() qovery.OrganizationEditRequest {
	return qovery.OrganizationEditRequest{
		Name:        toString(org.Name),
		Description: toStringPointer(org.Description),
	}
}

func convertResponseToOrganization(organization *qovery.Organization) Organization {
	return Organization{
		Id:          fromString(organization.Id),
		Name:        fromString(organization.Name),
		Plan:        fromClientEnum(organization.Plan),
		Description: fromStringPointer(organization.Description),
	}
}
