package qovery

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// convertQoveryOrganizationToDomain takes a qovery.Organization returned by the API client and turns it into the domain model organization.Organization.
func convertQoveryOrganizationToDomain(orga *qovery.Organization) (*organization.Organization, error) {
	if orga == nil {
		return nil, organization.ErrNilOrganization
	}

	plan, err := organization.NewPlanFromString(string(orga.Plan))
	if err != nil {
		return nil, err
	}

	orgaDomain, err := organization.NewOrganization(orga.Id, orga.Name, *plan)
	if err != nil {
		return nil, err
	}

	orgaDomain.Description = orga.Description

	return orgaDomain, nil
}

// convertDomainUpdateRequestToQovery takes the domain request organization.UpdateRequest and turns it into a qovery.OrganizationEditRequest to make the api call.
func convertDomainUpdateRequestToQovery(request organization.UpdateRequest) qovery.OrganizationEditRequest {
	return qovery.OrganizationEditRequest{
		Name:        request.Name,
		Description: request.Description,
	}
}
