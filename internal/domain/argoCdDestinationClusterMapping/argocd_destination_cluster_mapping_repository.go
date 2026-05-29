package argoCdDestinationClusterMapping

//go:generate mockery --testonly --with-expecter --name=Repository --structname=ArgoCdDestinationClusterMappingRepository --filename=argocd_destination_cluster_mapping_repository_mock.go --output=../../infrastructure/repositories/mocks_test/ --outpkg=mocks_test

import "context"

// Repository represents the interface to implement to handle the persistence of ArgoCD destination cluster mappings.
type Repository interface {
	Create(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Get(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) (*ArgoCdDestinationClusterMapping, error)
	Update(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Delete(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) error
}
