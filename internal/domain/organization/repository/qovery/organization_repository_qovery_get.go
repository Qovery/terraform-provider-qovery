package qovery

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// Get calls Qovery's API to retrieve an organization using the given ID.
func (o organizationQoveryRepository) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	orga, resp, err := o.client.OrganizationMainCallsApi.
		GetOrganization(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceOrganization, organizationID, resp, err)
	}

	return convertQoveryOrganizationToDomain(orga)
}
