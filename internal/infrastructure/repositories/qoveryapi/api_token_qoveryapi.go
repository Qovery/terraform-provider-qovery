package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

// Ensure apiTokenQoveryAPI defined type fully satisfy the apitoken.Repository interface.
var _ apitoken.Repository = apiTokenQoveryAPI{}

// apiTokenQoveryAPI implements the interface apitoken.Repository.
type apiTokenQoveryAPI struct {
	client *qovery.APIClient
}

func newApiTokenQoveryAPI(client *qovery.APIClient) (apitoken.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &apiTokenQoveryAPI{client: client}, nil
}

// Create calls Qovery's API to create an organization api token.
// The response is the only moment the secret token value is returned by the API.
func (a apiTokenQoveryAPI) Create(ctx context.Context, organizationID string, request apitoken.CreateRequest) (*apitoken.ApiToken, error) {
	roleID := request.RoleID
	res, resp, err := a.client.OrganizationApiTokenAPI.
		CreateOrganizationApiToken(ctx, organizationID).
		OrganizationApiTokenCreateRequest(qovery.OrganizationApiTokenCreateRequest{
			Name:        request.Name,
			Description: request.Description,
			RoleId:      *qovery.NewNullableString(&roleID),
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceOrganizationApiToken, organizationID, resp, err)
	}

	return newDomainApiTokenFromCreateResponse(organizationID, res)
}

// Get calls Qovery's API to retrieve an organization api token by id.
// The API exposes no get-single endpoint, so the organization tokens are listed and filtered by id.
func (a apiTokenQoveryAPI) Get(ctx context.Context, organizationID string, apiTokenID string) (*apitoken.ApiToken, error) {
	tokens, resp, err := a.client.OrganizationApiTokenAPI.
		ListOrganizationApiTokens(ctx, organizationID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceOrganizationApiToken, apiTokenID, resp, err)
	}

	for _, token := range tokens.GetResults() {
		if token.Id == apiTokenID {
			return newDomainApiTokenFromListItem(organizationID, token)
		}
	}

	return nil, apierrors.NewNotFoundAPIError(apierrors.APIResourceOrganizationApiToken, apiTokenID)
}

// Delete calls Qovery's API to delete an organization api token.
func (a apiTokenQoveryAPI) Delete(ctx context.Context, organizationID string, apiTokenID string) error {
	resp, err := a.client.OrganizationApiTokenAPI.
		DeleteOrganizationApiToken(ctx, organizationID, apiTokenID).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceOrganizationApiToken, apiTokenID, resp, err)
	}

	return nil
}
