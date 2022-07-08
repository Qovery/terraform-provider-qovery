package api

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func (o organizationAPI) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	orga, res, err := o.client.OrganizationMainCallsApi.
		EditOrganization(ctx, organizationID).
		OrganizationEditRequest(convertDomainUpdateRequestToQovery(request)).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, newUpdateOrganizationError(organizationID, res, err)
	}

	return convertQoveryOrganizationToDomain(orga)
}
