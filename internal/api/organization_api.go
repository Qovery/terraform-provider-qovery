package api

import (
	"net/http"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

type organizationAPI struct {
	client *qovery.APIClient
}

func NewOrganizationAPI(client *qovery.APIClient) organization.API {
	return &organizationAPI{
		client: client,
	}
}

//
// Errors
//

func newReadOrganizationError(resourceID string, resp *http.Response, err error) *apiError {
	return newReadError(apiResourceOrganization, resourceID, resp, err)
}

func newUpdateOrganizationError(resourceID string, resp *http.Response, err error) *apiError {
	return newUpdateError(apiResourceOrganization, resourceID, resp, err)
}

//
// Convertors
//

func convertQoveryOrganizationToDomain(orga *qovery.Organization) (*organization.Organization, error) {
	plan, err := organization.NewPlanFromString(string(orga.Plan))
	if err != nil {
		return nil, err
	}

	orgaDomain := organization.
		NewOrganization(orga.Id, orga.Name, *plan).
		WithDescription(orga.Description)

	return &orgaDomain, nil
}

func convertDomainUpdateRequestToQovery(request organization.UpdateRequest) qovery.OrganizationEditRequest {
	return qovery.OrganizationEditRequest{
		Name:        request.Name,
		Description: request.Description,
	}
}
