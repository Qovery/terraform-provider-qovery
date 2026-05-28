package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

type ArgoCdDestinationClusterMapping struct {
	Id               types.String `tfsdk:"id"`
	OrganizationId   types.String `tfsdk:"organization_id"`
	AgentClusterId   types.String `tfsdk:"agent_cluster_id"`
	ArgocdClusterUrl types.String `tfsdk:"argocd_cluster_url"`
	ClusterId        types.String `tfsdk:"cluster_id"`
}

func (a ArgoCdDestinationClusterMapping) toUpsertRequest() argoCdDestinationClusterMapping.UpsertRequest {
	return argoCdDestinationClusterMapping.UpsertRequest{
		AgentClusterId:   ToString(a.AgentClusterId),
		ArgocdClusterUrl: ToString(a.ArgocdClusterUrl),
		ClusterId:        ToString(a.ClusterId),
	}
}

func convertDomainArgoCdDestinationClusterMappingToTF(res *argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping) ArgoCdDestinationClusterMapping {
	return ArgoCdDestinationClusterMapping{
		Id:               FromString(fmt.Sprintf("%s:%s", res.AgentClusterID.String(), res.ArgocdClusterUrl)),
		OrganizationId:   FromString(res.OrganizationID.String()),
		AgentClusterId:   FromString(res.AgentClusterID.String()),
		ArgocdClusterUrl: FromString(res.ArgocdClusterUrl),
		ClusterId:        FromString(res.ClusterID.String()),
	}
}
