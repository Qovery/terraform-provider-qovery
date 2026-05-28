package argoCdCredentials

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateArgoCdCredentials = errors.New("failed to create argocd credentials")
	ErrFailedToGetArgoCdCredentials    = errors.New("failed to get argocd credentials")
	ErrFailedToUpdateArgoCdCredentials = errors.New("failed to update argocd credentials")
	ErrFailedToDeleteArgoCdCredentials = errors.New("failed to delete argocd credentials")
	ErrInvalidClusterIdParam           = errors.New("invalid cluster id")
)

type Service interface {
	Create(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Get(ctx context.Context, clusterID string) (*ArgoCdCredentials, error)
	Update(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Delete(ctx context.Context, clusterID string) error
}
