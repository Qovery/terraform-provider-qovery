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
	// ErrNotFoundInList is returned when the agent cluster is not yet visible in the ArgoCD
	// live cluster list. ListArgoCdDestinationClusterMappings only returns clusters ArgoCD has
	// actively discovered; a freshly-saved mapping may not appear until ArgoCD has polled that
	// agent. This is indistinguishable from a deletion, so callers should treat it as a soft
	// "not-found" and preserve existing state rather than surfacing an error to the user.
	ErrNotFoundInList = errors.New("agent cluster not yet visible in argocd cluster list")
	// ErrNotFound is returned when the agent cluster IS present in ArgoCD's live list but the
	// specific destination mapping is absent. Because ArgoCD has polled the agent and no longer
	// reports the mapping, it was genuinely removed out-of-band. Callers should treat this as a
	// real deletion and drop the resource from state so Terraform plans a re-create.
	ErrNotFound = errors.New("argocd destination cluster mapping not found")
)

type Service interface {
	Create(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Get(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) (*ArgoCdDestinationClusterMapping, error)
	Update(ctx context.Context, orgID string, request UpsertRequest) (*ArgoCdDestinationClusterMapping, error)
	Delete(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) error
}
