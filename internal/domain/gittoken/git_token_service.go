package gittoken

import (
	"context"

	"github.com/qovery/qovery-client-go"
)

type GitTokenParams struct {
	Name               string
	Description        *string
	Type               string
	Token              string
	BitbucketWorkspace *string
}

type Service interface {
	Create(ctx context.Context, organizationID string, params GitTokenParams) (*qovery.GitTokenResponse, error)
	Get(ctx context.Context, organizationID string, gitTokenID string) (*qovery.GitTokenResponse, error)
	Update(ctx context.Context, organizationID string, gitTokenID string, params GitTokenParams) (*qovery.GitTokenResponse, error)
	Delete(ctx context.Context, organizationID string, gitTokenID string) error
}
