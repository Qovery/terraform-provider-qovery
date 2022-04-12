package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
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

func (c Cluster) toUpsertClusterRequest(state *Cluster) client.ClusterUpsertParams {
	var clusterCloudProviderRequest *qovery.ClusterCloudProviderInfoRequest
	if state == nil || c.CredentialsId != state.CredentialsId {
		clusterCloudProviderRequest = &qovery.ClusterCloudProviderInfoRequest{
			CloudProvider: toStringPointer(c.CloudProvider),
			Region:        toStringPointer(c.Region),
			Credentials: &qovery.ClusterCloudProviderInfoRequestCredentials{
				Id:   toStringPointer(c.CredentialsId),
				Name: toStringPointer(c.Name),
			},
		}
	}

	return client.ClusterUpsertParams{
		ClusterCloudProviderRequest: clusterCloudProviderRequest,
		ClusterRequest: qovery.ClusterRequest{
			Name:            toString(c.Name),
			CloudProvider:   toString(c.CloudProvider),
			Region:          toString(c.Region),
			Description:     toStringPointer(c.Description),
			Cpu:             toInt32Pointer(c.CPU),
			Memory:          toInt32Pointer(c.Memory),
			MinRunningNodes: toInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes: toInt32Pointer(c.MaxRunningNodes),
		},
		DesiredState: toString(c.State),
	}
}

func convertResponseToCluster(res *client.ClusterResponse) Cluster {
	return Cluster{
		Id:              fromString(res.ClusterResponse.Id),
		CredentialsId:   fromStringPointer(res.ClusterInfo.Credentials.Id),
		OrganizationId:  fromString(res.OrganizationID),
		Name:            fromString(res.ClusterResponse.Name),
		CloudProvider:   fromString(res.ClusterResponse.CloudProvider),
		Region:          fromString(res.ClusterResponse.Region),
		Description:     fromNullableString(res.ClusterResponse.Description),
		CPU:             fromInt32Pointer(res.ClusterResponse.Cpu),
		Memory:          fromInt32Pointer(res.ClusterResponse.Memory),
		MinRunningNodes: fromInt32Pointer(res.ClusterResponse.MinRunningNodes),
		MaxRunningNodes: fromInt32Pointer(res.ClusterResponse.MaxRunningNodes),
		State:           fromString(res.ClusterResponse.GetStatus()),
		//Timeouts:        plan.Timeouts,
	}
}
