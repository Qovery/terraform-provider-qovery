package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"reflect"

	"github.com/qovery/terraform-provider-qovery/client"
)

const (
	featureKeyVpcSubnet = "vpc_subnet"
	featureIdVpcSubnet  = "VPC_SUBNET"
	featureKeyStaticIP  = "static_ip"
	featureIdStaticIP   = "STATIC_IP"
)

type Cluster struct {
	Id               types.String `tfsdk:"id"`
	OrganizationId   types.String `tfsdk:"organization_id"`
	CredentialsId    types.String `tfsdk:"credentials_id"`
	Name             types.String `tfsdk:"name"`
	CloudProvider    types.String `tfsdk:"cloud_provider"`
	Region           types.String `tfsdk:"region"`
	Description      types.String `tfsdk:"description"`
	KubernetesMode   types.String `tfsdk:"kubernetes_mode"`
	InstanceType     types.String `tfsdk:"instance_type"`
	MinRunningNodes  types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes  types.Int64  `tfsdk:"max_running_nodes"`
	Features         types.Object `tfsdk:"features"`
	RoutingTables    types.Set    `tfsdk:"routing_table"`
	State            types.String `tfsdk:"state"`
	AdvancedSettings types.Object `tfsdk:"advanced_settings"`
}

func (c Cluster) hasFeaturesDiff(state *Cluster) bool {
	clusterFeatures := toQoveryClusterFeatures(c.Features, c.KubernetesMode.String())
	if state == nil {
		return len(clusterFeatures) > 0
	}

	stateFeature := toQoveryClusterFeatures(state.Features, c.KubernetesMode.String())
	if len(clusterFeatures) != len(stateFeature) {
		return true
	}

	stateFeaturesByID := make(map[string]interface{})
	for _, sf := range stateFeature {
		value := sf.GetValue()
		stateFeaturesByID[sf.GetId()] = value.GetActualInstance()
	}

	for _, cf := range clusterFeatures {
		value := cf.GetValue()
		if stateValue, ok := stateFeaturesByID[cf.GetId()]; !ok || stateValue != value.GetActualInstance() {
			return true
		}
	}
	return false
}

func (c Cluster) hasRoutingTableDiff(state *Cluster) bool {
	clusterRoutes := toClusterRouteList(c.RoutingTables).toUpsertRequest().Routes
	if state == nil {
		return len(clusterRoutes) > 0
	}

	stateRoutes := toClusterRouteList(state.RoutingTables).toUpsertRequest().Routes
	if len(clusterRoutes) != len(stateRoutes) {
		return true
	}

	stateRoutesByDestination := make(map[string]ClusterRoute)
	for _, sr := range stateRoutes {
		stateRoutesByDestination[sr.Destination] = fromClusterRoute(sr)
	}

	for _, cr := range clusterRoutes {
		stateRoute, ok := stateRoutesByDestination[cr.Destination]
		if !ok {
			return true
		}

		clusterRoute := fromClusterRoute(cr)
		if stateRoute.Description != clusterRoute.Description || stateRoute.Destination != clusterRoute.Destination || stateRoute.Target != clusterRoute.Target {
			return true
		}
	}
	return false
}

func (c Cluster) hasAdvSettingsDiff(state *Cluster) (bool, error) {
	clusterSettings, clusterErr := toMapStringString(c.AdvancedSettings)
	if clusterErr != nil {
		return false, clusterErr
	}
	if state == nil {
		return len(clusterSettings) > 0, nil
	}

	stateSettings, stateErr := toMapStringString(state.AdvancedSettings)
	if stateErr != nil {
		return false, stateErr
	}
	if len(clusterSettings) != len(stateSettings) {
		return true, nil
	}

	return reflect.DeepEqual(clusterSettings, stateSettings), nil
}

func (c Cluster) toUpsertClusterRequest(state *Cluster) (*client.ClusterUpsertParams, error) {
	cloudProvider, err := qovery.NewCloudProviderEnumFromValue(toString(c.CloudProvider))
	if err != nil {
		return nil, err
	}

	kubernetesMode, err := qovery.NewKubernetesEnumFromValue(toString(c.KubernetesMode))
	if err != nil {
		return nil, err
	}

	routingTable := toClusterRouteList(c.RoutingTables)

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

	// NOTE: force update clusters if features or routing table have changed
	advSettingsDiff, diffErr := c.hasAdvSettingsDiff(state)
	if diffErr != nil {
		return nil, diffErr
	}
	forceUpdate := c.hasFeaturesDiff(state) || c.hasRoutingTableDiff(state) || advSettingsDiff

	desiredState, err := qovery.NewStateEnumFromValue(toString(c.State))
	if err != nil {
		return nil, err
	}

	advSettings, parseErr := toMapStringString(c.AdvancedSettings)
	if parseErr != nil {
		return nil, parseErr
	}

	return &client.ClusterUpsertParams{
		ClusterCloudProviderRequest: clusterCloudProviderRequest,
		ClusterRequest: qovery.ClusterRequest{
			Name:            toString(c.Name),
			CloudProvider:   *cloudProvider,
			Region:          toString(c.Region),
			Description:     toStringPointer(c.Description),
			Kubernetes:      kubernetesMode,
			InstanceType:    toStringPointer(c.InstanceType),
			MinRunningNodes: toInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes: toInt32Pointer(c.MaxRunningNodes),
			Features:        toQoveryClusterFeatures(c.Features, c.KubernetesMode.String()),
		},
		ClusterRoutingTable:     routingTable.toUpsertRequest(),
		ClusterAdvancedSettings: advSettings,
		ForceUpdate:             forceUpdate,
		DesiredState:            *desiredState,
	}, nil
}

func convertResponseToCluster(res *client.ClusterResponse) Cluster {
	routingTable := fromClusterRoutingTable(res.ClusterRoutingTable)

	return Cluster{
		Id:               fromString(res.ClusterResponse.Id),
		CredentialsId:    fromStringPointer(res.ClusterInfo.Credentials.Id),
		OrganizationId:   fromString(res.OrganizationID),
		Name:             fromString(res.ClusterResponse.Name),
		CloudProvider:    fromClientEnum(res.ClusterResponse.CloudProvider),
		Region:           fromString(res.ClusterResponse.Region),
		Description:      fromStringPointer(res.ClusterResponse.Description),
		KubernetesMode:   fromClientEnumPointer(res.ClusterResponse.Kubernetes),
		InstanceType:     fromStringPointer(res.ClusterResponse.InstanceType),
		MinRunningNodes:  fromInt32Pointer(res.ClusterResponse.MinRunningNodes),
		MaxRunningNodes:  fromInt32Pointer(res.ClusterResponse.MaxRunningNodes),
		Features:         fromQoveryClusterFeatures(res.ClusterResponse.Features),
		RoutingTables:    routingTable.toTerraformSet(),
		State:            fromClientEnumPointer(res.ClusterResponse.Status),
		AdvancedSettings: fromStringMap(res.ClusterAdvancedSetting),
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
		case featureIdStaticIP:
			attrs[featureKeyStaticIP] = fromBoolPointer(f.GetValue().Bool)
			attrTypes[featureKeyStaticIP] = types.BoolType
		}
	}

	// The object should be fill even if no feature is present, but will be mark as Null
	// (e.g SCW clusters don't have any feature)
	isNull := false
	if len(attrs) == 0 && len(attrTypes) == 0 {
		isNull = true
		defaultFeatureKeyVpcSubnet := ""
		defaultFeatureKeyStaticIP := false
		attrs[featureKeyVpcSubnet] = fromStringPointer(&defaultFeatureKeyVpcSubnet)
		attrTypes[featureKeyVpcSubnet] = types.StringType
		attrs[featureKeyStaticIP] = fromBoolPointer(&defaultFeatureKeyStaticIP)
		attrTypes[featureKeyStaticIP] = types.BoolType
	}

	return types.Object{
		Attrs:     attrs,
		AttrTypes: attrTypes,
		Null:      isNull,
	}
}

func toQoveryClusterFeatures(f types.Object, mode string) []qovery.ClusterRequestFeaturesInner {
	if f.Null || f.Unknown || mode == "K3S" {
		return nil
	}

	features := make([]qovery.ClusterRequestFeaturesInner, 0, len(f.Attrs))
	if _, ok := f.Attrs[featureKeyVpcSubnet]; ok {
		value := qovery.NewNullableClusterFeatureValue(&qovery.ClusterFeatureValue{
			String: toStringPointer(f.Attrs[featureKeyVpcSubnet].(types.String)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    stringAsPointer(featureIdVpcSubnet),
			Value: *value,
		})
	}

	if _, ok := f.Attrs[featureKeyStaticIP]; ok {
		value := qovery.NewNullableClusterFeatureValue(&qovery.ClusterFeatureValue{
			Bool: toBoolPointer(f.Attrs[featureKeyStaticIP].(types.Bool)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    stringAsPointer(featureIdStaticIP),
			Value: *value,
		})
	}

	return features
}
