package argoCdCredentials

import "context"

// Repository represents the interface to implement to handle the persistence of ArgoCD credentials.
type Repository interface {
	Create(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Get(ctx context.Context, clusterID string) (*ArgoCdCredentials, error)
	Update(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Delete(ctx context.Context, clusterID string) error
}
