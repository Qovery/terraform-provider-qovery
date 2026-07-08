package apitoken

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateApiToken = errors.New("failed to create api token")
	ErrFailedToGetApiToken    = errors.New("failed to get api token")
	ErrFailedToDeleteApiToken = errors.New("failed to delete api token")
)

// Service represents the interface to implement to handle the domain logic of an organization api token.
type Service interface {
	Create(ctx context.Context, organizationID string, request CreateRequest) (*ApiToken, error)
	Get(ctx context.Context, organizationID string, apiTokenID string) (*ApiToken, error)
	Delete(ctx context.Context, organizationID string, apiTokenID string) error
}
