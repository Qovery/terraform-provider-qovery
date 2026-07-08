package apitoken

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ApiTokenRepository --filename=api_token_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import "context"

// Repository represents the interface to implement to handle the persistence of organization api tokens.
// The API exposes no update endpoint and no get-single endpoint: Get is implemented by listing the
// organization tokens and filtering by id.
type Repository interface {
	Create(ctx context.Context, organizationID string, request CreateRequest) (*ApiToken, error)
	Get(ctx context.Context, organizationID string, apiTokenID string) (*ApiToken, error)
	Delete(ctx context.Context, organizationID string, apiTokenID string) error
}
