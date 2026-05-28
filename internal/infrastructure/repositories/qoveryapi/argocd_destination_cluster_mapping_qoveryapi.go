package qoveryapi

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

var _ argoCdDestinationClusterMapping.Repository = argoCdDestinationClusterMappingQoveryAPI{}

type argoCdDestinationClusterMappingQoveryAPI struct {
	client *qovery.APIClient
}

func newArgoCdDestinationClusterMappingQoveryAPI(client *qovery.APIClient) (argoCdDestinationClusterMapping.Repository, error) {
	if client == nil {
		return nil, ErrInvalidQoveryAPIClient
	}
	return &argoCdDestinationClusterMappingQoveryAPI{client: client}, nil
}

func (a argoCdDestinationClusterMappingQoveryAPI) Create(ctx context.Context, orgID string, request argoCdDestinationClusterMapping.UpsertRequest) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	req := qovery.NewArgoCdDestinationClusterMappingRequest(request.AgentClusterId, request.ArgocdClusterUrl, request.ClusterId)
	res, resp, err := a.client.ArgoCDAPI.
		SaveArgoCdDestinationClusterMapping(ctx, orgID).
		ArgoCdDestinationClusterMappingRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewCreateAPIError(apierrors.APIResourceArgoCdDestinationClusterMapping, request.AgentClusterId, resp, err)
	}
	return newDomainArgoCdDestinationClusterMappingFromResponse(orgID, res)
}

func (a argoCdDestinationClusterMappingQoveryAPI) Get(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	list, resp, err := a.client.ArgoCDAPI.
		ListArgoCdDestinationClusterMappings(ctx, orgID).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewReadAPIError(apierrors.APIResourceArgoCdDestinationClusterMapping, agentClusterID, resp, err)
	}

	// agentFound lets us distinguish a genuine deletion from ArgoCD eventual consistency:
	// if the agent cluster is present in the live list but the mapping is not among its linked
	// clusters, ArgoCD has polled the agent and no longer reports the mapping → genuine deletion.
	// If the agent cluster is absent entirely, ArgoCD has not polled it yet → soft not-found.
	agentFound := false
	for _, instance := range list.GetResults() {
		if instance.AgentClusterId != agentClusterID {
			continue
		}
		agentFound = true
		for _, linked := range instance.GetLinkedClusters() {
			if linked.ArgocdClusterUrl == argocdClusterUrl {
				orgUUID, err := parseUUID(orgID, argoCdDestinationClusterMapping.ErrInvalidOrganizationIDParam)
				if err != nil {
					return nil, err
				}
				agentUUID, err := parseUUID(agentClusterID, argoCdDestinationClusterMapping.ErrInvalidAgentClusterIDParam)
				if err != nil {
					return nil, err
				}
				clusterUUID, err := parseUUID(linked.QoveryClusterId, argoCdDestinationClusterMapping.ErrInvalidClusterIDParam)
				if err != nil {
					return nil, err
				}
				return &argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping{
					OrganizationID:   orgUUID,
					AgentClusterID:   agentUUID,
					ArgocdClusterUrl: argocdClusterUrl,
					ClusterID:        clusterUUID,
				}, nil
			}
		}
	}

	if agentFound {
		return nil, argoCdDestinationClusterMapping.ErrNotFound
	}
	return nil, argoCdDestinationClusterMapping.ErrNotFoundInList
}

func (a argoCdDestinationClusterMappingQoveryAPI) Update(ctx context.Context, orgID string, request argoCdDestinationClusterMapping.UpsertRequest) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	req := qovery.NewArgoCdDestinationClusterMappingRequest(request.AgentClusterId, request.ArgocdClusterUrl, request.ClusterId)
	res, resp, err := a.client.ArgoCDAPI.
		SaveArgoCdDestinationClusterMapping(ctx, orgID).
		ArgoCdDestinationClusterMappingRequest(*req).
		Execute()
	if err != nil || resp.StatusCode >= 400 {
		return nil, apierrors.NewUpdateAPIError(apierrors.APIResourceArgoCdDestinationClusterMapping, request.AgentClusterId, resp, err)
	}
	return newDomainArgoCdDestinationClusterMappingFromResponse(orgID, res)
}

func (a argoCdDestinationClusterMappingQoveryAPI) Delete(ctx context.Context, orgID string, agentClusterID string, argocdClusterUrl string) error {
	resp, err := a.client.ArgoCDAPI.
		DeleteArgoCdDestinationClusterMapping(ctx, orgID).
		AgentClusterId(agentClusterID).
		ArgocdClusterUrl(argocdClusterUrl).
		Execute()
	if err != nil || resp.StatusCode >= 300 {
		return apierrors.NewDeleteAPIError(apierrors.APIResourceArgoCdDestinationClusterMapping, agentClusterID, resp, err)
	}
	return nil
}

func newDomainArgoCdDestinationClusterMappingFromResponse(orgID string, res *qovery.ArgoCdDestinationClusterMappingResponse) (*argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping, error) {
	orgUUID, err := parseUUID(orgID, argoCdDestinationClusterMapping.ErrInvalidOrganizationIDParam)
	if err != nil {
		return nil, err
	}
	agentUUID, err := parseUUID(res.AgentClusterId, argoCdDestinationClusterMapping.ErrInvalidAgentClusterIDParam)
	if err != nil {
		return nil, err
	}
	clusterUUID, err := parseUUID(res.GetClusterId(), argoCdDestinationClusterMapping.ErrInvalidClusterIDParam)
	if err != nil {
		return nil, err
	}
	return &argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping{
		OrganizationID:   orgUUID,
		AgentClusterID:   agentUUID,
		ArgocdClusterUrl: res.ArgocdClusterUrl,
		ClusterID:        clusterUUID,
	}, nil
}
