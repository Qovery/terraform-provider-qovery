package services

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
)

// Ensure containerRegistryService defined types fully satisfy the registry.Service interface.
var _ gittoken.Service = gitTokenService{}

// containerRegistryService implements the interface registry.Service.
type gitTokenService struct {
	client *qovery.APIClient
}

func NewGitTokenService(client *qovery.APIClient) (gittoken.Service, error) {
	return &gitTokenService{
		client: client,
	}, nil
}

func (g gitTokenService) Create(ctx context.Context, organizationID string, params gittoken.GitTokenParams) (*qovery.GitTokenResponse, error) {
	gitTokenType, err := qovery.NewGitProviderEnumFromValue(params.Type)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get git token type")
	}
	gitTokenResponse, resp, err := g.client.OrganizationMainCallsAPI.
		CreateGitToken(ctx, organizationID).
		GitTokenRequest(qovery.GitTokenRequest{
			Name:        params.Name,
			Description: params.Description,
			Type:        *gitTokenType,
			Token:       params.Token,
			Workspace:   params.BitbucketWorkspace,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewCreateAPIError(apierrors.APIGitToken, "unknown", resp, err)
		return nil, apiErr
	}

	return gitTokenResponse, nil
}

func (g gitTokenService) Get(ctx context.Context, organizationID string, gitTokenID string) (*qovery.GitTokenResponse, error) {
	gitTokenResponse, resp, err := g.client.OrganizationMainCallsAPI.
		GetOrganizationGitToken(ctx, organizationID, gitTokenID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewReadAPIError(apierrors.APIGitToken, gitTokenID, resp, err)
		return nil, apiErr
	}

	return gitTokenResponse, nil
}

func (g gitTokenService) Update(ctx context.Context, organizationID string, gitTokenID string, params gittoken.GitTokenParams) (*qovery.GitTokenResponse, error) {
	gitTokenType, err := qovery.NewGitProviderEnumFromValue(params.Type)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get git token type")
	}
	gitTokenResponse, resp, err := g.client.OrganizationMainCallsAPI.
		EditGitToken(ctx, organizationID, gitTokenID).
		GitTokenRequest(qovery.GitTokenRequest{
			Name:        params.Name,
			Description: params.Description,
			Type:        *gitTokenType,
			Token:       params.Token,
			Workspace:   params.BitbucketWorkspace,
		}).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewUpdateAPIError(apierrors.APIGitToken, gitTokenID, resp, err)
		return nil, apiErr
	}

	return gitTokenResponse, nil
}

func (g gitTokenService) Delete(ctx context.Context, organizationID string, gitTokenID string) error {
	resp, err := g.client.OrganizationMainCallsAPI.
		DeleteGitToken(ctx, organizationID, gitTokenID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		apiErr := apierrors.NewDeleteAPIError(apierrors.APIGitToken, "unknown", resp, err)
		return apiErr
	}

	return nil
}
