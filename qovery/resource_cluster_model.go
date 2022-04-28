package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

const (
	featureIdVpcSubnet = "VPC_SUBNET"
)

type ClusterFeatures struct {
	VpcSubnet types.String `tfsdk:"vpc_subnet"`
}

func newClusterFeatures(ff []qovery.ClusterFeature) *ClusterFeatures {
	var features ClusterFeatures
	for _, f := range ff {
		if f.Id == nil {
			continue
		}
		switch *f.Id {
		case featureIdVpcSubnet:
			features.VpcSubnet = fromString(f.Value.(string))
		}
	}
	return &features

}

func (f ClusterFeatures) ToQoveryFeatures() []qovery.ClusterFeatureRequestFeatures {
	features := make([]qovery.ClusterFeatureRequestFeatures, 0, 1)
	if !f.VpcSubnet.Null && !f.VpcSubnet.Unknown {
		features = append(features, qovery.ClusterFeatureRequestFeatures{
			Id:    stringAsPointer(featureIdVpcSubnet),
			Value: *qovery.NewNullableString(toStringPointer(f.VpcSubnet)),
		})
	}
	return features
}

type Cluster struct {
	Id              types.String     `tfsdk:"id"`
	OrganizationId  types.String     `tfsdk:"organization_id"`
	CredentialsId   types.String     `tfsdk:"credentials_id"`
	Name            types.String     `tfsdk:"name"`
	CloudProvider   types.String     `tfsdk:"cloud_provider"`
	Region          types.String     `tfsdk:"region"`
	Description     types.String     `tfsdk:"description"`
	CPU             types.Int64      `tfsdk:"cpu"`
	Memory          types.Int64      `tfsdk:"memory"`
	MinRunningNodes types.Int64      `tfsdk:"min_running_nodes"`
	MaxRunningNodes types.Int64      `tfsdk:"max_running_nodes"`
	Features        *ClusterFeatures `tfsdk:"features"`
	State           types.String     `tfsdk:"state"`
}

func (c Cluster) toUpsertClusterRequest(state *Cluster) (*client.ClusterUpsertParams, error) {
	cloudProvider, err := qovery.NewCloudProviderEnumFromValue(toString(c.CloudProvider))
	if err != nil {
		return nil, err
	}

	var clusterCloudProviderRequest *qovery.ClusterCloudProviderInfoRequest
	if state == nil || c.CredentialsId != state.CredentialsId {
		clusterCloudProviderRequest = &qovery.ClusterCloudProviderInfoRequest{
			CloudProvider: cloudProvider,
			Region:        toStringPointer(c.Region),
			Credentials: &qovery.ClusterCloudProviderInfoRequestCredentials{
				Id:   toStringPointer(c.CredentialsId),
				Name: toStringPointer(c.Name),
			},
		}
	}

	var features []qovery.ClusterFeatureRequestFeatures
	if c.Features != nil {
		features = c.Features.ToQoveryFeatures()
	}

	desiredState, err := qovery.NewStateEnumFromValue(toString(c.State))
	if err != nil {
		return nil, err
	}

	return &client.ClusterUpsertParams{
		ClusterCloudProviderRequest: clusterCloudProviderRequest,
		ClusterRequest: qovery.ClusterRequest{
			Name:            toString(c.Name),
			CloudProvider:   *cloudProvider,
			Region:          toString(c.Region),
			Description:     toStringPointer(c.Description),
			Cpu:             toInt32Pointer(c.CPU),
			Memory:          toInt32Pointer(c.Memory),
			MinRunningNodes: toInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes: toInt32Pointer(c.MaxRunningNodes),
			Features:        features,
		},
		DesiredState: *desiredState,
	}, nil
}

func convertResponseToCluster(res *client.ClusterResponse) Cluster {
	return Cluster{
		Id:              fromString(res.ClusterResponse.Id),
		CredentialsId:   fromStringPointer(res.ClusterInfo.Credentials.Id),
		OrganizationId:  fromString(res.OrganizationID),
		Name:            fromString(res.ClusterResponse.Name),
		CloudProvider:   fromClientEnum(res.ClusterResponse.CloudProvider),
		Region:          fromString(res.ClusterResponse.Region),
		Description:     fromStringPointer(res.ClusterResponse.Description),
		CPU:             fromInt32Pointer(res.ClusterResponse.Cpu),
		Memory:          fromInt32Pointer(res.ClusterResponse.Memory),
		MinRunningNodes: fromInt32Pointer(res.ClusterResponse.MinRunningNodes),
		MaxRunningNodes: fromInt32Pointer(res.ClusterResponse.MaxRunningNodes),
		Features:        newClusterFeatures(res.ClusterResponse.Features),
		State:           fromClientEnumPointer(res.ClusterResponse.Status),
	}
}
