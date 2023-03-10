package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// newDomainOrganizationFromQovery takes a qovery.Organization returned by the API client and turns it into the domain model organization.Organization.
func newDomainOrganizationFromQovery(orga *qovery.Organization) (*organization.Organization, error) {
	if orga == nil {
		return nil, organization.ErrNilOrganization
	}

	return organization.NewOrganization(organization.NewOrganizationParams{
		OrganizationID: orga.GetId(),
		Name:           orga.GetName(),
		Plan:           string(orga.Plan),
		Description:    orga.Description.Get(),
	})
}

// newQoveryOrganizationEditRequestFromDomain takes the domain request organization.UpdateRequest and turns it into a qovery.OrganizationEditRequest to make the api call.
func newQoveryOrganizationEditRequestFromDomain(request organization.UpdateRequest) qovery.OrganizationEditRequest {
	return qovery.OrganizationEditRequest{
		Name:        request.Name,
		Description: request.Description,
	}
}
