package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// Update calls Qovery's API to update an organization using the given ID and request.
func (o organizationQoveryRepository) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	orga, resp, err := o.client.OrganizationMainCallsApi.
		EditOrganization(ctx, organizationID).
		OrganizationEditRequest(convertDomainUpdateRequestToQovery(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceOrganization, organizationID, resp, err)
	}

	return convertQoveryOrganizationToDomain(orga)
}
