package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/qovery/terraform-provider-qovery/client"
)

const (
	featureKeyVpcSubnet   = "vpc_subnet"
	featureIdVpcSubnet    = "VPC_SUBNET"
	featureKeyStaticIP    = "static_ip"
	featureIdStaticIP     = "STATIC_IP"
	featureIdExistingVpc  = "EXISTING_VPC"
	featureKeyExistingVpc = "existing_vpc"
)

type Cluster struct {
	Id                   types.String `tfsdk:"id"`
	OrganizationId       types.String `tfsdk:"organization_id"`
	CredentialsId        types.String `tfsdk:"credentials_id"`
	Name                 types.String `tfsdk:"name"`
	CloudProvider        types.String `tfsdk:"cloud_provider"`
	Region               types.String `tfsdk:"region"`
	Description          types.String `tfsdk:"description"`
	KubernetesMode       types.String `tfsdk:"kubernetes_mode"`
	InstanceType         types.String `tfsdk:"instance_type"`
	DiskSize             types.Int64  `tfsdk:"disk_size"`
	MinRunningNodes      types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes      types.Int64  `tfsdk:"max_running_nodes"`
	Features             types.Object `tfsdk:"features"`
	RoutingTables        types.Set    `tfsdk:"routing_table"`
	State                types.String `tfsdk:"state"`
	AdvancedSettingsJson types.String `tfsdk:"advanced_settings_json"`
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

func (c Cluster) toUpsertClusterRequest(state *Cluster) (*client.ClusterUpsertParams, error) {
	cloudProvider, err := qovery.NewCloudProviderEnumFromValue(ToString(c.CloudProvider))
	if err != nil {
		return nil, err
	}

	kubernetesMode, err := qovery.NewKubernetesEnumFromValue(ToString(c.KubernetesMode))
	if err != nil {
		return nil, err
	}

	routingTable := toClusterRouteList(c.RoutingTables)

	var clusterCloudProviderRequest *qovery.ClusterCloudProviderInfoRequest
	if state == nil || c.CredentialsId != state.CredentialsId {
		clusterCloudProviderRequest = &qovery.ClusterCloudProviderInfoRequest{
			CloudProvider: cloudProvider,
			Region:        ToStringPointer(c.Region),
			Credentials: &qovery.ClusterCloudProviderInfoCredentials{
				Id:   ToStringPointer(c.CredentialsId),
				Name: ToStringPointer(c.Name),
			},
		}
	}

	forceUpdate := c.hasFeaturesDiff(state) || c.hasRoutingTableDiff(state)

	desiredState, err := qovery.NewClusterStateEnumFromValue(ToString(c.State))
	if err != nil {
		return nil, err
	}

	return &client.ClusterUpsertParams{
		ClusterCloudProviderRequest: clusterCloudProviderRequest,
		ClusterRequest: qovery.ClusterRequest{
			Name:                     ToString(c.Name),
			CloudProvider:            *cloudProvider,
			CloudProviderCredentials: clusterCloudProviderRequest,
			Region:                   ToString(c.Region),
			Description:              ToStringPointer(c.Description),
			Kubernetes:               kubernetesMode,
			InstanceType:             ToStringPointer(c.InstanceType),
			DiskSize:                 ToInt64Pointer(c.DiskSize),
			MinRunningNodes:          ToInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes:          ToInt32Pointer(c.MaxRunningNodes),
			Features:                 toQoveryClusterFeatures(c.Features, c.KubernetesMode.String()),
		},
		ClusterRoutingTable:  routingTable.toUpsertRequest(),
		AdvancedSettingsJson: ToString(c.AdvancedSettingsJson),
		ForceUpdate:          forceUpdate,
		DesiredState:         *desiredState,
	}, nil
}

func convertResponseToCluster(ctx context.Context, res *client.ClusterResponse, initialPlan Cluster) Cluster {
	routingTable := fromClusterRoutingTable(res.ClusterRoutingTable)

	return Cluster{
		Id:                   FromString(res.ClusterResponse.Id),
		CredentialsId:        FromStringPointer(res.ClusterInfo.Credentials.Id),
		OrganizationId:       FromString(res.OrganizationID),
		Name:                 FromString(res.ClusterResponse.Name),
		CloudProvider:        fromClientEnum(res.ClusterResponse.CloudProvider),
		Region:               FromString(res.ClusterResponse.Region),
		Description:          FromStringPointer(res.ClusterResponse.Description),
		KubernetesMode:       fromClientEnumPointer(res.ClusterResponse.Kubernetes),
		InstanceType:         FromStringPointer(res.ClusterResponse.InstanceType),
		DiskSize:             FromInt32Pointer(res.ClusterResponse.DiskSize),
		MinRunningNodes:      FromInt32Pointer(res.ClusterResponse.MinRunningNodes),
		MaxRunningNodes:      FromInt32Pointer(res.ClusterResponse.MaxRunningNodes),
		Features:             fromQoveryClusterFeatures(res.ClusterResponse.Features),
		RoutingTables:        routingTable.toTerraformSet(ctx, initialPlan.RoutingTables),
		State:                fromClientEnumPointer(res.ClusterResponse.Status),
		AdvancedSettingsJson: FromString(res.AdvancedSettingsJson),
	}
}

func fromQoveryClusterFeatures(ff []qovery.ClusterFeature) types.Object {
	if ff == nil {
		// Early return object null without attribute types
		return types.ObjectNull(make(map[string]attr.Type))
	}

	attributes := make(map[string]attr.Value)
	attributeTypes := make(map[string]attr.Type)
	for _, f := range ff {
		if f.Id == nil {
			continue
		}
		switch *f.Id {
		case featureIdVpcSubnet:
			attributes[featureKeyVpcSubnet] = FromStringPointer(f.GetValue().String)
			attributeTypes[featureKeyVpcSubnet] = types.StringType
		case featureIdStaticIP:
			attributes[featureKeyStaticIP] = FromBoolPointer(f.GetValue().Bool)
			attributeTypes[featureKeyStaticIP] = types.BoolType
		case featureIdExistingVpc:
			// tf has a default value for it, but the api does not returns this feature , as exisiting vpc super seed it
			// So set the default value to match what tf expect and not break existing clients
			attributes[featureKeyVpcSubnet] = FromStringPointer(&clusterFeatureVpcSubnetDefault)
			attributeTypes[featureKeyVpcSubnet] = types.StringType

			v := f.GetValue().ClusterFeatureAwsExistingVpc
			attrTypes := make(map[string]attr.Type)
			attrTypes["aws_vpc_eks_id"] = types.StringType
			attrTypes["eks_subnets_zone_a_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["eks_subnets_zone_b_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["eks_subnets_zone_c_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["rds_subnets_zone_a_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["rds_subnets_zone_b_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["rds_subnets_zone_c_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["documentdb_subnets_zone_a_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["documentdb_subnets_zone_b_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["documentdb_subnets_zone_c_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["elasticache_subnets_zone_a_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["elasticache_subnets_zone_b_ids"] = types.ListType{ElemType: types.StringType}
			attrTypes["elasticache_subnets_zone_c_ids"] = types.ListType{ElemType: types.StringType}

			attrVals := make(map[string]attr.Value)
			attrVals["aws_vpc_eks_id"] = FromStringPointer(&v.AwsVpcEksId)
			attrVals["eks_subnets_zone_a_ids"] = FromStringArray(v.EksSubnetsZoneAIds)
			attrVals["eks_subnets_zone_b_ids"] = FromStringArray(v.EksSubnetsZoneBIds)
			attrVals["eks_subnets_zone_c_ids"] = FromStringArray(v.EksSubnetsZoneCIds)
			attrVals["rds_subnets_zone_a_ids"] = FromStringArray(v.RdsSubnetsZoneAIds)
			attrVals["rds_subnets_zone_b_ids"] = FromStringArray(v.RdsSubnetsZoneBIds)
			attrVals["rds_subnets_zone_c_ids"] = FromStringArray(v.RdsSubnetsZoneCIds)
			attrVals["documentdb_subnets_zone_a_ids"] = FromStringArray(v.DocumentdbSubnetsZoneAIds)
			attrVals["documentdb_subnets_zone_b_ids"] = FromStringArray(v.DocumentdbSubnetsZoneBIds)
			attrVals["documentdb_subnets_zone_c_ids"] = FromStringArray(v.DocumentdbSubnetsZoneCIds)
			attrVals["elasticache_subnets_zone_a_ids"] = FromStringArray(v.DocumentdbSubnetsZoneAIds)
			attrVals["elasticache_subnets_zone_b_ids"] = FromStringArray(v.DocumentdbSubnetsZoneBIds)
			attrVals["elasticache_subnets_zone_c_ids"] = FromStringArray(v.DocumentdbSubnetsZoneCIds)

			terraformObjectValue, diagnostics := types.ObjectValue(attrTypes, attrVals)
			if diagnostics.HasError() {
				panic(fmt.Errorf("bad %s feature: %s", featureKeyExistingVpc, diagnostics.Errors()))
			}
			attributes[featureKeyExistingVpc] = terraformObjectValue
			attributeTypes[featureKeyExistingVpc] = terraformObjectValue.Type(context.Background())
		}
	}

	// The object should be fill even if no feature is present, but will be mark as Null
	// (e.g SCW clusters don't have any feature)
	isNull := false
	if len(attributes) == 0 && len(attributeTypes) == 0 {
		isNull = true
		defaultFeatureKeyVpcSubnet := ""
		defaultFeatureKeyStaticIP := false
		attributes[featureKeyVpcSubnet] = FromStringPointer(&defaultFeatureKeyVpcSubnet)
		attributeTypes[featureKeyVpcSubnet] = types.StringType
		attributes[featureKeyStaticIP] = FromBoolPointer(&defaultFeatureKeyStaticIP)
		attributeTypes[featureKeyStaticIP] = types.BoolType
	}

	if isNull {
		// Early return object null
		return types.ObjectNull(attributeTypes)
	}

	terraformObjectValue, diagnostics := types.ObjectValue(attributeTypes, attributes)
	if diagnostics.HasError() {
		panic(fmt.Errorf("bad cluster feature: %s", diagnostics.Errors()))
	}
	return terraformObjectValue
}

func toQoveryClusterFeatures(f types.Object, mode string) []qovery.ClusterRequestFeaturesInner {
	if f.IsNull() || f.IsUnknown() || mode == "K3S" {
		return nil
	}

	features := make([]qovery.ClusterRequestFeaturesInner, 0, len(f.Attributes()))
	if _, ok := f.Attributes()[featureKeyVpcSubnet]; ok {
		value := qovery.NewNullableClusterFeatureValue(&qovery.ClusterFeatureValue{
			String: ToStringPointer(f.Attributes()[featureKeyVpcSubnet].(types.String)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    StringAsPointer(featureIdVpcSubnet),
			Value: *value,
		})
	}

	if _, ok := f.Attributes()[featureKeyStaticIP]; ok {
		value := qovery.NewNullableClusterFeatureValue(&qovery.ClusterFeatureValue{
			Bool: ToBoolPointer(f.Attributes()[featureKeyStaticIP].(types.Bool)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    StringAsPointer(featureIdStaticIP),
			Value: *value,
		})
	}

	if _, ok := f.Attributes()[featureKeyExistingVpc]; ok {
		v := f.Attributes()[featureKeyExistingVpc].(types.Object)
		feature := qovery.ClusterFeatureAwsExistingVpc{
			AwsVpcEksId:                ToString(v.Attributes()["aws_vpc_eks_id"].(types.String)),
			EksSubnetsZoneAIds:         ToStringArray(v.Attributes()["eks_subnets_zone_a_ids"].(types.List)),
			EksSubnetsZoneBIds:         ToStringArray(v.Attributes()["eks_subnets_zone_b_ids"].(types.List)),
			EksSubnetsZoneCIds:         ToStringArray(v.Attributes()["eks_subnets_zone_c_ids"].(types.List)),
			RdsSubnetsZoneAIds:         ToStringArray(v.Attributes()["rds_subnets_zone_a_ids"].(types.List)),
			RdsSubnetsZoneBIds:         ToStringArray(v.Attributes()["rds_subnets_zone_b_ids"].(types.List)),
			RdsSubnetsZoneCIds:         ToStringArray(v.Attributes()["rds_subnets_zone_c_ids"].(types.List)),
			DocumentdbSubnetsZoneAIds:  ToStringArray(v.Attributes()["documentdb_subnets_zone_a_ids"].(types.List)),
			DocumentdbSubnetsZoneBIds:  ToStringArray(v.Attributes()["documentdb_subnets_zone_b_ids"].(types.List)),
			DocumentdbSubnetsZoneCIds:  ToStringArray(v.Attributes()["documentdb_subnets_zone_c_ids"].(types.List)),
			ElasticacheSubnetsZoneAIds: ToStringArray(v.Attributes()["elasticache_subnets_zone_a_ids"].(types.List)),
			ElasticacheSubnetsZoneBIds: ToStringArray(v.Attributes()["elasticache_subnets_zone_b_ids"].(types.List)),
			ElasticacheSubnetsZoneCIds: ToStringArray(v.Attributes()["elasticache_subnets_zone_c_ids"].(types.List)),
		}
		value := qovery.NewNullableClusterFeatureValue(&qovery.ClusterFeatureValue{
			ClusterFeatureAwsExistingVpc: &feature,
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    StringAsPointer(featureIdExistingVpc),
			Value: *value,
		})
	}

	return features
}
