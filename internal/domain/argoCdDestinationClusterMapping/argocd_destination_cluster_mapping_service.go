package argoCdDestinationClusterMapping

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrFailedToCreateArgoCdDestinationClusterMapping = errors.New("failed to create argocd destination cluster mapping")
	ErrFailedToGetArgoCdDestinationClusterMapping    = errors.New("failed to get argocd destination cluster mapping")
	ErrFailedToUpdateArgoCdDestinationClusterMapping = errors.New("failed to update argocd destination cluster mapping")
	ErrFailedToDeleteArgoCdDestinationClusterMapping = errors.New("failed to delete argocd destination cluster mapping")
	ErrInvalidOrganizationIdParam                    = errors.New("invalid organization id")
	ErrInvalidAgentClusterIdParam                    = errors.New("invalid agent cluster id")
	// ErrNotFoundInList is returned when the mapping is not yet visible in the ArgoCD live
	// cluster list. This happens because ListArgoCdDestinationClusterMappings only returns
	// clusters ArgoCD has actively discovered; a freshly-saved mapping may not appear until
	// ArgoCD has polled that destination. Callers should treat this as a soft "not-found"
	// and preserve existing state rather than surfacing an error to the user.
	ErrNotFoundInList = errors.New("mapping not found in argocd cluster list")
)

type Service interface {
	Create(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Get(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) (*ArgoCdDestinationClusterMapping, error)
	Update(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Delete(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) error
}
