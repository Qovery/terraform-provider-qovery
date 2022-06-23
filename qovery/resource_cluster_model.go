package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

const (
	featureKeyVpcSubnet = "vpc_subnet"
	featureIdVpcSubnet  = "VPC_SUBNET"
)

type Cluster struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	CredentialsId   types.String `tfsdk:"credentials_id"`
	Name            types.String `tfsdk:"name"`
	CloudProvider   types.String `tfsdk:"cloud_provider"`
	Region          types.String `tfsdk:"region"`
	Description     types.String `tfsdk:"description"`
	InstanceType    types.String `tfsdk:"instance_type"`
	MinRunningNodes types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes types.Int64  `tfsdk:"max_running_nodes"`
	Features        types.Object `tfsdk:"features"`
	State           types.String `tfsdk:"state"`
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
			Credentials: &qovery.ClusterCloudProviderInfoCredentials{
				Id:   toStringPointer(c.CredentialsId),
				Name: toStringPointer(c.Name),
			},
		}
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
			InstanceType:    toStringPointer(c.InstanceType),
			MinRunningNodes: toInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes: toInt32Pointer(c.MaxRunningNodes),
			Features:        toQoveryClusterFeatures(c.Features),
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
		InstanceType:    fromStringPointer(res.ClusterResponse.InstanceType),
		MinRunningNodes: fromInt32Pointer(res.ClusterResponse.MinRunningNodes),
		MaxRunningNodes: fromInt32Pointer(res.ClusterResponse.MaxRunningNodes),
		Features:        fromQoveryClusterFeatures(res.ClusterResponse.Features),
		State:           fromClientEnumPointer(res.ClusterResponse.Status),
	}
}

func fromQoveryClusterFeatures(ff []qovery.ClusterFeature) types.Object {
	if ff == nil {
		return types.Object{Null: true}
	}

	attrs := make(map[string]attr.Value)
	attrTypes := make(map[string]attr.Type)
	for _, f := range ff {
		if f.Id == nil {
			continue
		}
		switch *f.Id {
		case featureIdVpcSubnet:
			attrs[featureKeyVpcSubnet] = fromStringPointer(f.GetValue().String)
			attrTypes[featureKeyVpcSubnet] = types.StringType
		}
	}

	if len(attrs) == 0 && len(attrTypes) == 0 {
		return types.Object{Unknown: true}
	}

	return types.Object{
		Attrs:     attrs,
		AttrTypes: attrTypes,
	}
}

func toQoveryClusterFeatures(f types.Object) *qovery.ClusterRequestFeatures {
	if f.Null || f.Unknown {
		return nil
	}

	features := make([]qovery.ClusterRequestFeaturesFeaturesInner, 0, len(f.Attrs))
	if _, ok := f.Attrs[featureKeyVpcSubnet]; ok {
		features = append(features, qovery.ClusterRequestFeaturesFeaturesInner{
			Id:    stringAsPointer(featureIdVpcSubnet),
			Value: *qovery.NewNullableString(toStringPointer(f.Attrs[featureKeyVpcSubnet].(types.String))),
		})
	}

	req := qovery.NewClusterRequestFeatures()
	req.SetFeatures(features)
	return req
}
