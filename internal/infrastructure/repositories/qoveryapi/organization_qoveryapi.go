package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// organizationQoveryAPI implements the interface organization.Repository.
type organizationQoveryAPI struct {
	client *qovery.APIClient
}

// NOTE: This forces the implementation of the interface organization.Repository by organizationQoveryAPI at compile time.
var _ organization.Repository = organizationQoveryAPI{}

// newOrganizationQoveryAPI return a new instance of an organization.Repository that uses Qovery's API.
func newOrganizationQoveryAPI(client *qovery.APIClient) (organization.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}

	return &organizationQoveryAPI{
		client: client,
	}, nil
}

// Get calls Qovery's API to retrieve an organization using the given organizationID.
func (c organizationQoveryAPI) Get(ctx context.Context, organizationID string) (*organization.Organization, error) {
	orga, resp, err := c.client.OrganizationMainCallsApi.
		GetOrganization(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadApiError(apierrors.ApiResourceOrganization, organizationID, resp, err)
	}

	return newDomainOrganizationFromQovery(orga)
}

// Update calls Qovery's API to update an organization using the given organizationID and request.
func (c organizationQoveryAPI) Update(ctx context.Context, organizationID string, request organization.UpdateRequest) (*organization.Organization, error) {
	orga, resp, err := c.client.OrganizationMainCallsApi.
		EditOrganization(ctx, organizationID).
		OrganizationEditRequest(newQoveryOrganizationEditRequestFromDomain(request)).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateApiError(apierrors.ApiResourceOrganization, organizationID, resp, err)
	}

	return newDomainOrganizationFromQovery(orga)
}
