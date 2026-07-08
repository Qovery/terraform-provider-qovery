package qoveryapi

import (
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

// newDomainApiTokenFromCreateResponse converts the create endpoint response into a domain
// ApiToken. This is the only place the secret token value is available.
func newDomainApiTokenFromCreateResponse(organizationID string, res *qovery.OrganizationApiTokenCreate) (*apitoken.ApiToken, error) {
	id, err := parseUUID(res.Id, apitoken.ErrInvalidApiTokenIdParam)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(organizationID, apitoken.ErrInvalidOrganizationIdParam)
	if err != nil {
		return nil, err
	}

	return &apitoken.ApiToken{
		ID:             id,
		OrganizationID: orgID,
		Name:           res.GetName(),
		Description:    res.Description,
		RoleID:         res.GetRoleId(),
		Token:          res.Token,
	}, nil
}

// newDomainApiTokenFromListItem converts a list endpoint item into a domain ApiToken.
// The secret token value is never returned by the list endpoint, so Token is always nil.
func newDomainApiTokenFromListItem(organizationID string, res qovery.OrganizationApiToken) (*apitoken.ApiToken, error) {
	id, err := parseUUID(res.Id, apitoken.ErrInvalidApiTokenIdParam)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(organizationID, apitoken.ErrInvalidOrganizationIdParam)
	if err != nil {
		return nil, err
	}

	return &apitoken.ApiToken{
		ID:             id,
		OrganizationID: orgID,
		Name:           res.GetName(),
		Description:    res.Description,
		RoleID:         res.GetRoleId(),
		Token:          nil,
	}, nil
}
