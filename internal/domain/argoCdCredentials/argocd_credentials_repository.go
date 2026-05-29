package argoCdCredentials

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ArgoCdCredentialsRepository --filename=argocd_credentials_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import "context"

// Repository represents the interface to implement to handle the persistence of ArgoCD credentials.
type Repository interface {
	Create(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Get(ctx context.Context, clusterID string) (*ArgoCdCredentials, error)
	Update(ctx context.Context, clusterID string, request UpsertRequest) (*ArgoCdCredentials, error)
	Delete(ctx context.Context, clusterID string) error
}
