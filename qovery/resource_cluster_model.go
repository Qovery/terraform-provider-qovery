package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type Cluster struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	CredentialsId   types.String `tfsdk:"credentials_id"`
	Name            types.String `tfsdk:"name"`
	CloudProvider   types.String `tfsdk:"cloud_provider"`
	Region          types.String `tfsdk:"region"`
	Description     types.String `tfsdk:"description"`
	CPU             types.Int64  `tfsdk:"cpu"`
	Memory          types.Int64  `tfsdk:"memory"`
	MinRunningNodes types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes types.Int64  `tfsdk:"max_running_nodes"`
	State           types.String `tfsdk:"state"`
	//Timeouts        Timeout      `tfsdk:"timeouts"`
}

func (c Cluster) toUpsertClusterRequest() qovery.ClusterRequest {
	return qovery.ClusterRequest{
		Name:            toString(c.Name),
		CloudProvider:   toString(c.CloudProvider),
		Region:          toString(c.Region),
		Description:     toStringPointer(c.Description),
		Cpu:             toInt32Pointer(c.CPU),
		Memory:          toInt32Pointer(c.Memory),
		MinRunningNodes: toInt32Pointer(c.MinRunningNodes),
		MaxRunningNodes: toInt32Pointer(c.MaxRunningNodes),
	}
}

func (c Cluster) toUpdateClusterCloudProviderInfoRequest() qovery.ClusterCloudProviderInfoRequest {
	return qovery.ClusterCloudProviderInfoRequest{
		CloudProvider: toStringPointer(c.CloudProvider),
		Region:        toStringPointer(c.Region),
		Credentials: &qovery.ClusterCloudProviderInfoRequestCredentials{
			Id:   toStringPointer(c.CredentialsId),
			Name: toStringPointer(c.Name),
		},
	}
}

func convertResponseToCluster(cluster *qovery.ClusterResponse, clusterInfo *qovery.ClusterCloudProviderInfoResponse, plan Cluster) Cluster {
	return Cluster{
		Id:              fromString(cluster.Id),
		CredentialsId:   fromStringPointer(clusterInfo.Credentials.Id),
		OrganizationId:  plan.OrganizationId,
		Name:            fromString(cluster.Name),
		CloudProvider:   fromString(cluster.CloudProvider),
		Region:          fromString(cluster.Region),
		Description:     fromNullableString(cluster.Description),
		CPU:             fromInt32Pointer(cluster.Cpu),
		Memory:          fromInt32Pointer(cluster.Memory),
		MinRunningNodes: fromInt32Pointer(cluster.MinRunningNodes),
		MaxRunningNodes: fromInt32Pointer(cluster.MaxRunningNodes),
		State:           plan.State,
		//Timeouts:        plan.Timeouts,
	}
}
