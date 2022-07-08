package api

import (
	"context"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

func (o organizationAPI) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	orga, res, err := o.client.OrganizationMainCallsApi.
		GetOrganization(ctx, organizationID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, newReadOrganizationError(organizationID, res, err)
	}

	return convertQoveryOrganizationToDomain(orga)
}
