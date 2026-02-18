package qovery

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

const (
	featureKeyVpcSubnet      = "vpc_subnet"
	featureIdVpcSubnet       = "VPC_SUBNET"
	featureKeyStaticIP       = "static_ip"
	featureIdStaticIP        = "STATIC_IP"
	featureIdExistingVpc     = "EXISTING_VPC"
	featureKeyExistingVpc    = "existing_vpc"
	featureIdKarpenter       = "KARPENTER"
	featureKeyKarpenter      = "karpenter"
	featureKeyGcpExistingVpc = "gcp_existing_vpc"

	instanceTypeAutoPilot = "AUTO_PILOT"

	// Infrastructure charts parameter keys
	infraChartsNginxKey       = "nginx_parameters"
	infraChartsCertManagerKey = "cert_manager_parameters"
	infraChartsMetalLbKey     = "metal_lb_parameters"
)

type Cluster struct {
	Id                             types.String `tfsdk:"id"`
	OrganizationId                 types.String `tfsdk:"organization_id"`
	CredentialsId                  types.String `tfsdk:"credentials_id"`
	Name                           types.String `tfsdk:"name"`
	CloudProvider                  types.String `tfsdk:"cloud_provider"`
	Region                         types.String `tfsdk:"region"`
	Description                    types.String `tfsdk:"description"`
	KubernetesMode                 types.String `tfsdk:"kubernetes_mode"`
	InstanceType                   types.String `tfsdk:"instance_type"`
	DiskSize                       types.Int64  `tfsdk:"disk_size"`
	MinRunningNodes                types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes                types.Int64  `tfsdk:"max_running_nodes"`
	Production                     types.Bool   `tfsdk:"production"`
	Features                       types.Object `tfsdk:"features"`
	RoutingTables                  types.Set    `tfsdk:"routing_table"`
	State                          types.String `tfsdk:"state"`
	AdvancedSettingsJson           types.String `tfsdk:"advanced_settings_json"`
	Kubeconfig                     types.String `tfsdk:"kubeconfig"`
	InfrastructureOutputs          types.Object `tfsdk:"infrastructure_outputs"`
	InfrastructureChartsParameters types.Object `tfsdk:"infrastructure_charts_parameters"`
}

func (c Cluster) hasFeaturesDiff(state *Cluster) bool {
	clusterFeatures, _ := toQoveryClusterFeatures(c.Features, c.KubernetesMode.String())
	if state == nil {
		return len(clusterFeatures) > 0
	}

	stateFeature, _ := toQoveryClusterFeatures(state.Features, c.KubernetesMode.String())
	if len(clusterFeatures) != len(stateFeature) {
		return true
	}

	stateFeaturesByID := make(map[string]any)
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

func (c Cluster) hasInfraChartsParamsDiff(state *Cluster) bool {
	if c.InfrastructureChartsParameters.IsNull() || c.InfrastructureChartsParameters.IsUnknown() {
		return state != nil && !state.InfrastructureChartsParameters.IsNull() && !state.InfrastructureChartsParameters.IsUnknown()
	}
	if state == nil || state.InfrastructureChartsParameters.IsNull() || state.InfrastructureChartsParameters.IsUnknown() {
		return true
	}
	// Compare the object values
	return !c.InfrastructureChartsParameters.Equal(state.InfrastructureChartsParameters)
}

func (c Cluster) toUpsertClusterRequest(state *Cluster) (*client.ClusterUpsertParams, error) {
	cloudProvider, err := qovery.NewCloudProviderEnumFromValue(ToString(c.CloudProvider))
	cloudVendor, err := qovery.NewCloudVendorEnumFromValue(ToString(c.CloudProvider))
	if err != nil {
		return nil, err
	}

	kubernetesMode, err := qovery.NewKubernetesEnumFromValue(ToString(c.KubernetesMode))
	if err != nil {
		return nil, err
	}

	routingTable := toClusterRouteList(c.RoutingTables)

	// Handle PARTIALLY_MANAGED (EKS Anywhere) mode validations
	isPartiallyManaged := kubernetesMode != nil && *kubernetesMode == qovery.KUBERNETESENUM_PARTIALLY_MANAGED

	// Convert infrastructure charts parameters
	var infraChartsParams *qovery.ClusterInfrastructureChartsParameters
	if !c.InfrastructureChartsParameters.IsNull() && !c.InfrastructureChartsParameters.IsUnknown() {
		infraChartsParams, err = toQoveryInfrastructureChartsParameters(c.InfrastructureChartsParameters)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse infrastructure_charts_parameters")
		}
	}

	// Validation for PARTIALLY_MANAGED mode
	if isPartiallyManaged {
		// kubeconfig is required for PARTIALLY_MANAGED
		if c.Kubeconfig.IsNull() || c.Kubeconfig.IsUnknown() || c.Kubeconfig.ValueString() == "" {
			return nil, errors.New("kubeconfig is required when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
		}

		// infrastructure_charts_parameters is required for PARTIALLY_MANAGED
		if infraChartsParams == nil {
			return nil, errors.New("infrastructure_charts_parameters is required when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
		}

		// Validate that metal_lb_parameters.ip_address_pools is not empty
		if infraChartsParams.MetalLbParameters == nil || len(infraChartsParams.MetalLbParameters.IpAddressPools) == 0 {
			return nil, errors.New("infrastructure_charts_parameters.metal_lb_parameters.ip_address_pools is required and must not be empty for PARTIALLY_MANAGED mode")
		}

		// Features are not allowed for PARTIALLY_MANAGED mode
		if !c.Features.IsNull() && !c.Features.IsUnknown() {
			featuresAttrs := c.Features.Attributes()
			// Check if any feature is actually set (not just defaults)
			hasNonDefaultFeatures := false

			if vpcSubnet, ok := featuresAttrs[featureKeyVpcSubnet]; ok {
				if !vpcSubnet.IsNull() && !vpcSubnet.IsUnknown() {
					vpcSubnetStr := vpcSubnet.(types.String).ValueString()
					if vpcSubnetStr != "" && vpcSubnetStr != clusterFeatureVpcSubnetDefault {
						hasNonDefaultFeatures = true
					}
				}
			}
			if staticIP, ok := featuresAttrs[featureKeyStaticIP]; ok {
				if !staticIP.IsNull() && !staticIP.IsUnknown() && staticIP.(types.Bool).ValueBool() {
					hasNonDefaultFeatures = true
				}
			}
			if existingVpc, ok := featuresAttrs[featureKeyExistingVpc]; ok {
				if !existingVpc.IsNull() && !existingVpc.IsUnknown() {
					// Check if existing_vpc has actual content (aws_vpc_eks_id is required)
					existingVpcObj := existingVpc.(types.Object)
					if !existingVpcObj.IsNull() && len(existingVpcObj.Attributes()) > 0 {
						if vpcId, hasVpcId := existingVpcObj.Attributes()["aws_vpc_eks_id"]; hasVpcId {
							if !vpcId.IsNull() && !vpcId.IsUnknown() {
								hasNonDefaultFeatures = true
							}
						}
					}
				}
			}
			if karpenter, ok := featuresAttrs[featureKeyKarpenter]; ok {
				if !karpenter.IsNull() && !karpenter.IsUnknown() {
					// Check if karpenter has actual content (spot_enabled is required)
					karpenterObj := karpenter.(types.Object)
					if !karpenterObj.IsNull() && len(karpenterObj.Attributes()) > 0 {
						if spotEnabled, hasSpot := karpenterObj.Attributes()["spot_enabled"]; hasSpot {
							if !spotEnabled.IsNull() && !spotEnabled.IsUnknown() {
								hasNonDefaultFeatures = true
							}
						}
					}
				}
			}

			if hasNonDefaultFeatures {
				return nil, errors.New("features (vpc_subnet, static_ip, existing_vpc, karpenter) are not supported when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
			}
		}
	} else {
		// infrastructure_charts_parameters should not be set for non-PARTIALLY_MANAGED modes
		if infraChartsParams != nil {
			return nil, errors.New("infrastructure_charts_parameters is only supported when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
		}
	}

	features, err := toQoveryClusterFeatures(c.Features, c.KubernetesMode.String())
	if err != nil {
		return nil, err
	}

	// For PARTIALLY_MANAGED mode, clear features to avoid sending them to API
	if isPartiallyManaged {
		features = nil
	}

	for _, f := range features {
		if f.Id != nil && *f.Id == featureIdKarpenter {
			if state != nil && !IsKarpenterAlreadyInstalled(state) {
				return nil, errors.New("It is not possible to migrate to Karpenter using terraform")
			}

			if !c.InstanceType.IsUnknown() {
				return nil, errors.New("instance_type must not be defined when Karpenter feature is enabled")
			}
			if !c.MinRunningNodes.IsUnknown() {
				return nil, errors.New("min_running_nodes must not be defined when Karpenter feature is enabled")
			}
			if !c.MaxRunningNodes.IsUnknown() {
				return nil, errors.New("max_running_nodes must not be defined when Karpenter feature is enabled")
			}
			if !c.DiskSize.IsUnknown() {
				return nil, errors.New("disk_size must not be defined when Karpenter feature is enabled")
			}
		}
	}

	// Validation: Require Karpenter for new EKS clusters
	if state == nil { // This is a new cluster creation
		isAWS := ToString(c.CloudProvider) == "AWS"
		isManaged := kubernetesMode != nil && *kubernetesMode == qovery.KUBERNETESENUM_MANAGED

		if isAWS && isManaged {
			// Check if Karpenter is enabled
			karpenterEnabled := false
			for _, f := range features {
				if f.Id != nil && *f.Id == featureIdKarpenter {
					karpenterEnabled = true
					break
				}
			}

			if !karpenterEnabled {
				return nil, errors.New("Karpenter is required for new EKS (AWS MANAGED) clusters. Please configure the Karpenter feature in the cluster configuration")
			}
		}
	}

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

	forceUpdate := c.hasFeaturesDiff(state) || c.hasRoutingTableDiff(state) || c.hasInfraChartsParamsDiff(state)

	desiredState, err := qovery.NewClusterStateEnumFromValue(ToString(c.State))
	if err != nil {
		return nil, err
	}

	return &client.ClusterUpsertParams{
		ClusterCloudProviderRequest: clusterCloudProviderRequest,
		ClusterRequest: qovery.ClusterRequest{
			Name:                           ToString(c.Name),
			CloudProvider:                  *cloudVendor,
			CloudProviderCredentials:       clusterCloudProviderRequest,
			Region:                         ToString(c.Region),
			Description:                    ToStringPointer(c.Description),
			Kubernetes:                     kubernetesMode,
			InstanceType:                   ToStringPointer(c.InstanceType),
			DiskSize:                       ToInt64Pointer(c.DiskSize),
			MinRunningNodes:                ToInt32Pointer(c.MinRunningNodes),
			MaxRunningNodes:                ToInt32Pointer(c.MaxRunningNodes),
			Production:                     ToBoolPointer(c.Production),
			Features:                       features,
			InfrastructureChartsParameters: infraChartsParams,
		},
		ClusterRoutingTable:  routingTable.toUpsertRequest(),
		AdvancedSettingsJson: ToString(c.AdvancedSettingsJson),
		ForceUpdate:          forceUpdate,
		DesiredState:         *desiredState,
	}, nil
}

func IsKarpenterAlreadyInstalled(state *Cluster) bool {
	if state == nil {
		return false
	}

	oldFeatures, _ := toQoveryClusterFeatures(state.Features, state.KubernetesMode.String())
	for _, f := range oldFeatures {
		if f.Id != nil && *f.Id == featureIdKarpenter {
			return true
		}
	}
	return false
}

func convertResponseToCluster(ctx context.Context, res *client.ClusterResponse, initialPlan Cluster) Cluster {
	routingTable := fromClusterRoutingTable(res.ClusterRoutingTable)

	// Check if cluster is PARTIALLY_MANAGED (EKS Anywhere)
	isPartiallyManaged := res.ClusterResponse.Kubernetes != nil &&
		*res.ClusterResponse.Kubernetes == qovery.KUBERNETESENUM_PARTIALLY_MANAGED

	cluster := Cluster{
		Id:                             FromString(res.ClusterResponse.Id),
		CredentialsId:                  FromStringPointer(res.ClusterInfo.Credentials.Id),
		OrganizationId:                 FromString(res.OrganizationID),
		Name:                           FromString(res.ClusterResponse.Name),
		CloudProvider:                  fromClientEnum(res.ClusterResponse.CloudProvider),
		Region:                         FromString(res.ClusterResponse.Region),
		Description:                    FromStringPointer(res.ClusterResponse.Description),
		KubernetesMode:                 fromClientEnumPointer(res.ClusterResponse.Kubernetes),
		Production:                     FromBoolPointer(res.ClusterResponse.Production),
		State:                          fromClientEnumPointer(res.ClusterResponse.Status),
		AdvancedSettingsJson:           FromString(res.AdvancedSettingsJson),
		InfrastructureChartsParameters: fromQoveryInfrastructureChartsParameters(res.ClusterResponse.InfrastructureChartsParameters),
	}

	// For PARTIALLY_MANAGED (EKS Anywhere) clusters, these fields are not applicable
	// Return null values to avoid spurious terraform plan changes
	if isPartiallyManaged {
		cluster.InstanceType = types.StringNull()
		cluster.DiskSize = types.Int64Null()
		cluster.MinRunningNodes = types.Int64Null()
		cluster.MaxRunningNodes = types.Int64Null()
		cluster.Features = types.ObjectNull(createFeaturesAttrTypes())
		cluster.RoutingTables = types.SetNull(types.ObjectType{AttrTypes: clusterRouteAttrTypes})
		cluster.InfrastructureOutputs = types.ObjectNull(map[string]attr.Type{
			"cluster_name":        types.StringType,
			"cluster_arn":         types.StringType,
			"cluster_self_link":   types.StringType,
			"cluster_oidc_issuer": types.StringType,
			"vpc_id":              types.StringType,
		})
		// Preserve kubeconfig from initialPlan - it's fetched separately via API in Read operation
		cluster.Kubeconfig = initialPlan.Kubeconfig
	} else {
		cluster.InstanceType = FromStringPointer(res.ClusterResponse.InstanceType)
		cluster.DiskSize = FromInt32Pointer(res.ClusterResponse.DiskSize)

		// GCP Autopilot: preserve plan values for node counts since the API returns sentinel values.
		isAutoPilot := res.ClusterResponse.InstanceType != nil && *res.ClusterResponse.InstanceType == instanceTypeAutoPilot
		if isAutoPilot && !initialPlan.MinRunningNodes.IsNull() && !initialPlan.MinRunningNodes.IsUnknown() {
			cluster.MinRunningNodes = initialPlan.MinRunningNodes
		} else {
			cluster.MinRunningNodes = FromInt32Pointer(res.ClusterResponse.MinRunningNodes)
		}
		if isAutoPilot && !initialPlan.MaxRunningNodes.IsNull() && !initialPlan.MaxRunningNodes.IsUnknown() {
			cluster.MaxRunningNodes = initialPlan.MaxRunningNodes
		} else {
			cluster.MaxRunningNodes = FromInt32Pointer(res.ClusterResponse.MaxRunningNodes)
		}

		cluster.Features = fromQoveryClusterFeatures(res.ClusterResponse.Features)
		cluster.RoutingTables = routingTable.toTerraformSet(ctx, initialPlan.RoutingTables)
		cluster.InfrastructureOutputs = fromQoveryClusterOutput(res.ClusterResponse.InfrastructureOutputs)
		// Kubeconfig is not applicable for non-PARTIALLY_MANAGED clusters
		cluster.Kubeconfig = types.StringNull()
	}

	return cluster
}

func fromQoveryClusterOutput(
	infrastructureOutputs *qovery.InfrastructureOutputs,
) types.Object {
	// Define the schema once for consistency
	attrTypes := map[string]attr.Type{
		"cluster_name":        types.StringType,
		"cluster_arn":         types.StringType,
		"cluster_self_link":   types.StringType,
		"cluster_oidc_issuer": types.StringType,
		"vpc_id":              types.StringType,
	}

	// Default null values
	values := map[string]attr.Value{
		"cluster_name":        types.StringNull(),
		"cluster_arn":         types.StringNull(),
		"cluster_self_link":   types.StringNull(),
		"cluster_oidc_issuer": types.StringNull(),
		"vpc_id":              types.StringNull(),
	}

	if infrastructureOutputs == nil {
		return types.ObjectValueMust(attrTypes, values)
	}

	switch {
	case infrastructureOutputs.AksInfrastructureOutputs != nil:
		out := infrastructureOutputs.AksInfrastructureOutputs
		values["cluster_name"] = types.StringValue(out.ClusterName)
		values["cluster_oidc_issuer"] = types.StringValue(out.ClusterOidcIssuer)

	case infrastructureOutputs.EksInfrastructureOutputs != nil:
		out := infrastructureOutputs.EksInfrastructureOutputs
		values["cluster_name"] = types.StringValue(out.ClusterName)
		values["cluster_arn"] = types.StringValue(out.ClusterArn)
		values["cluster_oidc_issuer"] = types.StringValue(out.ClusterOidcIssuer)
		values["vpc_id"] = types.StringValue(out.VpcId)

	case infrastructureOutputs.GkeInfrastructureOutputs != nil:
		out := infrastructureOutputs.GkeInfrastructureOutputs
		values["cluster_name"] = types.StringValue(out.ClusterName)
		values["cluster_self_link"] = types.StringValue(out.ClusterSelfLink)

	case infrastructureOutputs.KapsuleInfrastructureOutputs != nil:
		out := infrastructureOutputs.KapsuleInfrastructureOutputs
		values["cluster_name"] = types.StringValue(out.ClusterName)
	}

	return types.ObjectValueMust(attrTypes, values)
}

func fromQoveryClusterFeatures(
	clusterFeatures []qovery.ClusterFeatureResponse,
) types.Object {
	if clusterFeatures == nil {
		// Early return object null without attribute types
		return types.ObjectNull(make(map[string]attr.Type))
	}

	attributes := make(map[string]attr.Value)
	attributeTypes := make(map[string]attr.Type)
	for _, f := range clusterFeatures {
		if f.Id == nil {
			continue
		}
		switch *f.Id {
		case featureIdVpcSubnet:
			if f.GetValueObject().ClusterFeatureStringResponse != nil {
				attributes[featureKeyVpcSubnet] = FromString(f.GetValueObject().ClusterFeatureStringResponse.Value)
			} else {
				attributes[featureKeyVpcSubnet] = basetypes.NewStringNull()
			}
			attributeTypes[featureKeyVpcSubnet] = types.StringType
		case featureIdStaticIP:
			if f.GetValueObject().ClusterFeatureBooleanResponse != nil {
				attributes[featureKeyStaticIP] = FromBool(f.GetValueObject().ClusterFeatureBooleanResponse.Value)
			} else {
				attributes[featureKeyStaticIP] = basetypes.NewBoolNull()
			}
			attributeTypes[featureKeyStaticIP] = types.BoolType
		case featureIdExistingVpc:
			// GCP existing VPC — check before AWS since they share the same feature ID
			if f.GetValueObject().ClusterFeatureGcpExistingVpcResponse != nil {
				gcpVpc := &f.GetValueObject().ClusterFeatureGcpExistingVpcResponse.Value
				gcpAttrTypes := createGcpExistingVpcFeatureAttrTypes()
				gcpObj, diagnostics := types.ObjectValue(gcpAttrTypes, map[string]attr.Value{
					"vpc_name":                       FromStringPointer(&gcpVpc.VpcName),
					"vpc_project_id":                 FromNullableString(gcpVpc.VpcProjectId),
					"subnetwork_name":                FromNullableString(gcpVpc.SubnetworkName),
					"ip_range_services_name":         FromNullableString(gcpVpc.IpRangeServicesName),
					"ip_range_pods_name":             FromNullableString(gcpVpc.IpRangePodsName),
					"additional_ip_range_pods_names": fromStringArrayNullIfEmpty(gcpVpc.AdditionalIpRangePodsNames),
				})
				if diagnostics.HasError() {
					panic(fmt.Errorf("bad %s feature: %s", featureKeyGcpExistingVpc, diagnostics.Errors()))
				}
				attributes[featureKeyGcpExistingVpc] = gcpObj
				attributeTypes[featureKeyGcpExistingVpc] = types.ObjectType{AttrTypes: gcpAttrTypes}

				// Set AWS existing_vpc to null since this is a GCP cluster
				existingVpcAttrTypes := createExistingVpcFeatureAttrTypes()
				attributes[featureKeyExistingVpc] = types.ObjectNull(existingVpcAttrTypes)
				attributeTypes[featureKeyExistingVpc] = types.ObjectType{AttrTypes: existingVpcAttrTypes}

				attributes[featureKeyVpcSubnet] = FromStringPointer(&clusterFeatureVpcSubnetDefault)
				attributeTypes[featureKeyVpcSubnet] = types.StringType
				continue
			}

			// AWS existing VPC
			var v *qovery.ClusterFeatureAwsExistingVpc = nil
			if f.GetValueObject().ClusterFeatureAwsExistingVpcResponse != nil {
				v = &f.GetValueObject().ClusterFeatureAwsExistingVpcResponse.Value
			}

			attrTypes := createExistingVpcFeatureAttrTypes()

			if v == nil {
				terraformObjectValue := types.ObjectNull(attrTypes)
				attributes[featureKeyExistingVpc] = terraformObjectValue
				attributeTypes[featureKeyExistingVpc] = terraformObjectValue.Type(context.Background())
				continue
			}

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

			attrVals["eks_karpenter_fargate_subnets_zone_a_ids"] = FromStringArray(v.EksKarpenterFargateSubnetsZoneAIds)
			attrVals["eks_karpenter_fargate_subnets_zone_b_ids"] = FromStringArray(v.EksKarpenterFargateSubnetsZoneBIds)
			attrVals["eks_karpenter_fargate_subnets_zone_c_ids"] = FromStringArray(v.EksKarpenterFargateSubnetsZoneCIds)
			attrVals["eks_create_nodes_in_private_subnet"] = FromBoolPointer(v.EksCreateNodesInPrivateSubnet)

			terraformObjectValue, diagnostics := types.ObjectValue(attrTypes, attrVals)
			if diagnostics.HasError() {
				panic(fmt.Errorf("bad %s feature: %s", featureKeyExistingVpc, diagnostics.Errors()))
			}
			attributes[featureKeyExistingVpc] = terraformObjectValue
			attributeTypes[featureKeyExistingVpc] = terraformObjectValue.Type(context.Background())

			// tf has a default value for it, but the api does not return this feature , as exiting vpc super seed it
			// So set the default value to match what tf expect and not break existing clients
			attributes[featureKeyVpcSubnet] = FromStringPointer(&clusterFeatureVpcSubnetDefault)
			attributeTypes[featureKeyVpcSubnet] = types.StringType
		case featureIdKarpenter:
			var karpenterParameters *qovery.ClusterFeatureKarpenterParameters = nil
			if f.GetValueObject().ClusterFeatureKarpenterParametersResponse != nil {
				karpenterParameters = &f.GetValueObject().ClusterFeatureKarpenterParametersResponse.Value
			}

			attrTypes := createKarpenterFeatureAttrTypes()
			if karpenterParameters == nil {
				terraformObjectValue := types.ObjectNull(attrTypes)
				attributes[featureKeyKarpenter] = terraformObjectValue
				attributeTypes[featureKeyKarpenter] = terraformObjectValue.Type(context.Background())
				continue
			}

			attrVals := createKarpenterFeatureAttrValue(karpenterParameters)

			terraformObjectValue, diagnostics := types.ObjectValue(attrTypes, attrVals)
			if diagnostics.HasError() {
				panic(fmt.Errorf("bad %s feature: %s", featureKeyExistingVpc, diagnostics.Errors()))
			}
			attributes[featureKeyKarpenter] = terraformObjectValue
			attributeTypes[featureKeyKarpenter] = terraformObjectValue.Type(context.Background())
		}
	}

	// All attributes should be fill even if no feature is present.
	// This is mandatory to satisfy the terraform framework schema.

	if attributes[featureKeyVpcSubnet] == nil {
		defaultFeatureKeyVpcSubnet := ""
		attributes[featureKeyVpcSubnet] = FromStringPointer(&defaultFeatureKeyVpcSubnet)
		attributeTypes[featureKeyVpcSubnet] = types.StringType
	}

	if attributes[featureKeyStaticIP] == nil {
		defaultFeatureKeyStaticIP := false
		attributes[featureKeyStaticIP] = FromBoolPointer(&defaultFeatureKeyStaticIP)
		attributeTypes[featureKeyStaticIP] = types.BoolType
	}

	// featureKeyExistingVpc includes actually 2 entries: featureKeyExistingVpc and featureKeyVpcSubnet
	if attributes[featureKeyExistingVpc] == nil {
		existingVpcAttrTypes := createExistingVpcFeatureAttrTypes()
		attributes[featureKeyExistingVpc] = types.ObjectNull(existingVpcAttrTypes)
		attributeTypes[featureKeyExistingVpc] = attributes[featureKeyExistingVpc].Type(context.Background())
		attributes[featureKeyVpcSubnet] = FromStringPointer(&clusterFeatureVpcSubnetDefault)
		attributeTypes[featureKeyVpcSubnet] = types.StringType
	}

	// create default GCP existing VPC feature if not set yet
	if attributes[featureKeyGcpExistingVpc] == nil {
		gcpVpcAttrTypes := createGcpExistingVpcFeatureAttrTypes()
		attributes[featureKeyGcpExistingVpc] = types.ObjectNull(gcpVpcAttrTypes)
		attributeTypes[featureKeyGcpExistingVpc] = types.ObjectType{AttrTypes: gcpVpcAttrTypes}
	}

	// create default karpenter feature if not set yet
	if attributes[featureKeyKarpenter] == nil {
		attrTypes := createKarpenterFeatureAttrTypes()

		terraformObjectValue := types.ObjectNull(attrTypes)
		attributes[featureKeyKarpenter] = terraformObjectValue
		attributeTypes[featureKeyKarpenter] = terraformObjectValue.Type(context.Background())
	}

	terraformObjectValue, diagnostics := types.ObjectValue(attributeTypes, attributes)
	if diagnostics.HasError() {
		panic(fmt.Errorf("bad cluster feature: %s", diagnostics.Errors()))
	}
	return terraformObjectValue
}

func toQoveryClusterFeatures(f types.Object, mode string) ([]qovery.ClusterRequestFeaturesInner, error) {
	if f.IsNull() || f.IsUnknown() || mode == "K3S" {
		return nil, nil
	}

	features := make([]qovery.ClusterRequestFeaturesInner, 0, len(f.Attributes()))
	if _, ok := f.Attributes()[featureKeyVpcSubnet]; ok {
		value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
			String: ToStringPointer(f.Attributes()[featureKeyVpcSubnet].(types.String)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    new(featureIdVpcSubnet),
			Value: *value,
		})
	}

	if _, ok := f.Attributes()[featureKeyStaticIP]; ok {
		value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
			Bool: ToBoolPointer(f.Attributes()[featureKeyStaticIP].(types.Bool)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    new(featureIdStaticIP),
			Value: *value,
		})
	}

	if _, ok := f.Attributes()[featureKeyExistingVpc]; ok {
		v := f.Attributes()[featureKeyExistingVpc].(types.Object)
		if !v.IsNull() {
			feature := qovery.ClusterFeatureAwsExistingVpc{
				AwsVpcEksId:                        ToString(v.Attributes()["aws_vpc_eks_id"].(types.String)),
				EksSubnetsZoneAIds:                 ToStringArray(v.Attributes()["eks_subnets_zone_a_ids"].(types.List)),
				EksSubnetsZoneBIds:                 ToStringArray(v.Attributes()["eks_subnets_zone_b_ids"].(types.List)),
				EksSubnetsZoneCIds:                 ToStringArray(v.Attributes()["eks_subnets_zone_c_ids"].(types.List)),
				RdsSubnetsZoneAIds:                 ToStringArray(v.Attributes()["rds_subnets_zone_a_ids"].(types.List)),
				RdsSubnetsZoneBIds:                 ToStringArray(v.Attributes()["rds_subnets_zone_b_ids"].(types.List)),
				RdsSubnetsZoneCIds:                 ToStringArray(v.Attributes()["rds_subnets_zone_c_ids"].(types.List)),
				DocumentdbSubnetsZoneAIds:          ToStringArray(v.Attributes()["documentdb_subnets_zone_a_ids"].(types.List)),
				DocumentdbSubnetsZoneBIds:          ToStringArray(v.Attributes()["documentdb_subnets_zone_b_ids"].(types.List)),
				DocumentdbSubnetsZoneCIds:          ToStringArray(v.Attributes()["documentdb_subnets_zone_c_ids"].(types.List)),
				ElasticacheSubnetsZoneAIds:         ToStringArray(v.Attributes()["elasticache_subnets_zone_a_ids"].(types.List)),
				ElasticacheSubnetsZoneBIds:         ToStringArray(v.Attributes()["elasticache_subnets_zone_b_ids"].(types.List)),
				ElasticacheSubnetsZoneCIds:         ToStringArray(v.Attributes()["elasticache_subnets_zone_c_ids"].(types.List)),
				EksKarpenterFargateSubnetsZoneAIds: ToStringArray(v.Attributes()["eks_karpenter_fargate_subnets_zone_a_ids"].(types.List)),
				EksKarpenterFargateSubnetsZoneBIds: ToStringArray(v.Attributes()["eks_karpenter_fargate_subnets_zone_b_ids"].(types.List)),
				EksKarpenterFargateSubnetsZoneCIds: ToStringArray(v.Attributes()["eks_karpenter_fargate_subnets_zone_c_ids"].(types.List)),
				EksCreateNodesInPrivateSubnet:      ToBoolPointer(v.Attributes()["eks_create_nodes_in_private_subnet"].(types.Bool)),
			}
			value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				ClusterFeatureAwsExistingVpc: &feature,
			})

			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdExistingVpc),
				Value: *value,
			})
		}
	}

	if _, ok := f.Attributes()[featureKeyGcpExistingVpc]; ok {
		v := f.Attributes()[featureKeyGcpExistingVpc].(types.Object)
		if !v.IsNull() {
			attrs := v.Attributes()
			feature := qovery.ClusterFeatureGcpExistingVpc{
				VpcName:                    ToString(attrs["vpc_name"].(types.String)),
				VpcProjectId:               ToNullableString(attrs["vpc_project_id"].(types.String)),
				SubnetworkName:             ToNullableString(attrs["subnetwork_name"].(types.String)),
				IpRangeServicesName:        ToNullableString(attrs["ip_range_services_name"].(types.String)),
				IpRangePodsName:            ToNullableString(attrs["ip_range_pods_name"].(types.String)),
				AdditionalIpRangePodsNames: ToStringArray(attrs["additional_ip_range_pods_names"].(types.List)),
			}
			value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				ClusterFeatureGcpExistingVpc: &feature,
			})

			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdExistingVpc),
				Value: *value,
			})
		}
	}

	if _, ok := f.Attributes()[featureKeyKarpenter]; ok {
		v := f.Attributes()[featureKeyKarpenter].(types.Object)
		if !v.IsNull() {
			defaultServiceArchitecture := v.Attributes()["default_service_architecture"].(types.String).ValueString()
			arch, err := toCpuArchitectureEnum(defaultServiceArchitecture)
			if err != nil {
				return nil, err
			}

			qoveryNodePools, err := toQoveryNodePools(v)
			if err != nil {
				return nil, err
			}

			feature := qovery.ClusterFeatureKarpenterParameters{
				SpotEnabled:                ToBool(v.Attributes()["spot_enabled"].(types.Bool)),
				DiskSizeInGib:              ToInt32(v.Attributes()["disk_size_in_gib"].(types.Int64)),
				DefaultServiceArchitecture: arch,
				QoveryNodePools:            *qoveryNodePools,
			}
			value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				ClusterFeatureKarpenterParameters: &feature,
			})

			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdKarpenter),
				Value: *value,
			})
		}
	}

	return features, nil
}

func toQoveryNodePools(obj types.Object) (*qovery.KarpenterNodePool, error) {
	karpenterNodePool := qovery.KarpenterNodePool{}
	karpenterNodePool.Requirements = []qovery.KarpenterNodePoolRequirement{}

	// Set requirements
	requirements, err := extractRequirementsFromTypesObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to extract requirements from types.Object: %v", err)
	}

	if len(requirements) == 0 {
		return nil, fmt.Errorf("karpenter nodepool requirements are mandatory: they must be set among [InstanceFamily, InstanceSize, Arch]")
	}

	// Check that requirements are correctly set
	distinctRequirementTypes := make(map[string]bool)
	for _, requirement := range requirements {
		key, ok := requirement["key"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid key type for karpenter node pool requirement")
		}
		distinctRequirementTypes[key] = true
	}
	if len(distinctRequirementTypes) != 3 {
		return nil, fmt.Errorf("missing some karpenter nodepool requirement among [InstanceFamily, InstanceSize, Arch]")
	}

	for _, req := range requirements {
		key, ok := req["key"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid key type")
		}

		var karpenterKey qovery.KarpenterNodePoolRequirementKey
		switch key {
		case "InstanceFamily":
			karpenterKey = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY
		case "InstanceSize":
			karpenterKey = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE
		case "Arch":
			karpenterKey = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH
		default:
			return nil, fmt.Errorf("unsupported key: %s", key)
		}

		operator, ok := req["operator"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid operator type")
		}

		var karpenterOperator qovery.KarpenterNodePoolRequirementOperator
		switch operator {
		case "In":
			karpenterOperator = qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN
		default:
			return nil, fmt.Errorf("unsupported operator: %s", operator)
		}

		values, ok := req["values"].([]string)
		if !ok {
			return nil, fmt.Errorf("invalid values type")
		}

		if len(values) == 0 {
			return nil, fmt.Errorf("karpenter node pool values must not be empty")
		}

		requirement := qovery.KarpenterNodePoolRequirement{
			Key:      karpenterKey,
			Operator: karpenterOperator,
			Values:   values,
		}

		karpenterNodePool.Requirements = append(karpenterNodePool.Requirements, requirement)
	}

	// Set stable node pool override
	stableOverride, err := extractStableNodePoolOverrideFromTypesObject(obj)
	if err != nil {
		return nil, err
	}
	karpenterNodePool.StableOverride = stableOverride

	// Set default node pool override
	defaultOverride, err := extractDefaultNodePoolOverrideFromTypesObject(obj)
	if err != nil {
		return nil, err
	}
	karpenterNodePool.DefaultOverride = defaultOverride

	return &karpenterNodePool, nil
}

func extractRequirementsFromTypesObject(obj types.Object) ([]map[string]any, error) {
	qoveryNodePools, exists := obj.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !exists {
		return nil, fmt.Errorf("qovery_node_pools field not found")
	}

	requirementsAttr, exists := qoveryNodePools.Attributes()["requirements"]
	if !exists {
		return nil, fmt.Errorf("requirements field not found")
	}

	requirementsList, ok := requirementsAttr.(basetypes.ListValue)
	if !ok {
		return nil, fmt.Errorf("requirements field is not a list")
	}

	result := make([]map[string]any, 0, len(requirementsList.Elements()))
	for _, reqAttr := range requirementsList.Elements() {
		reqMap, err := convertObjectToMap(reqAttr)
		if err != nil {
			return nil, err
		}
		result = append(result, reqMap)
	}

	return result, nil
}

func extractStableNodePoolOverrideFromTypesObject(obj types.Object) (*qovery.KarpenterStableNodePoolOverride, error) {
	qoveryNodePools, exists := obj.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !exists {
		return nil, fmt.Errorf("qovery_node_pools field not found")
	}

	stableOverrideAttr, exists := qoveryNodePools.Attributes()["stable_override"]
	if !exists {
		return nil, nil
	}

	if stableOverrideAttr.IsNull() {
		// It means stable_override is not defined
		// No issue as this field is optional
		return nil, nil
	}

	stableOverride, ok := stableOverrideAttr.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("stable_override field cannot be parsed to Object")
	}

	qoveryStableOverride := qovery.KarpenterStableNodePoolOverride{}

	// Set consolidation
	consolidationAttr, exists := stableOverride.Attributes()["consolidation"]

	// The consolidation is allowed to be null
	if exists && !consolidationAttr.IsNull() {
		consolidation, ok := consolidationAttr.(basetypes.ObjectValue)
		if !ok {
			return nil, fmt.Errorf("consolidation field cannot be parsed to Object")
		}

		consolidationEnabled := consolidation.Attributes()["enabled"].(basetypes.BoolValue)
		consolidationDays := consolidation.Attributes()["days"].(basetypes.ListValue)
		consolidationStartTime := consolidation.Attributes()["start_time"].(basetypes.StringValue)
		consolidationDuration := consolidation.Attributes()["duration"].(basetypes.StringValue)

		// Converts consolidation days (string) to expected enum type (WeekdayEnum)
		consolidationWeekDayEnumList := make([]qovery.WeekdayEnum, 0)
		for _, value := range consolidationDays.Elements() {
			valueAsString := value.(basetypes.StringValue).ValueString()
			fromValue, err := qovery.NewWeekdayEnumFromValue(valueAsString)
			if err != nil {
				return nil, fmt.Errorf("cannot convert '%s' to WeekdayEnum", valueAsString)
			}
			consolidationWeekDayEnumList = append(consolidationWeekDayEnumList, *fromValue)
		}

		qoveryConsolidation := qovery.NewKarpenterNodePoolConsolidation(
			consolidationEnabled.ValueBool(),
			consolidationWeekDayEnumList,
			consolidationStartTime.ValueString(),
			consolidationDuration.ValueString(),
		)
		qoveryStableOverride.Consolidation = qoveryConsolidation
	}

	// Set limits
	limitsAttr, exists := stableOverride.Attributes()["limits"]

	// The limits are allowed to be null
	if exists && !limitsAttr.IsNull() {
		limits, ok := limitsAttr.(basetypes.ObjectValue)
		if !ok {
			return nil, fmt.Errorf("limits field cannot be parsed to Object")
		}

		enabled := limits.Attributes()["enabled"].(basetypes.BoolValue)
		limitsCpu := limits.Attributes()["max_cpu_in_vcpu"].(basetypes.Int64Value)
		limitsRam := limits.Attributes()["max_memory_in_gibibytes"].(basetypes.Int64Value)

		qoveryLimits := qovery.NewKarpenterNodePoolLimits(enabled.ValueBool(), int32(limitsCpu.ValueInt64()), int32(limitsRam.ValueInt64()), 0)
		qoveryStableOverride.Limits = qoveryLimits
	}

	// To avoid over-checking conditions when converting the API response to Terraform object, forbid the stable_override block if both consolidation and limits are undefined
	if consolidationAttr.IsNull() && limitsAttr.IsNull() {
		return nil, fmt.Errorf("if `qovery_node_pools.stable_override` is defined, you must define at least its `consolidation` or its `limits`")
	}

	return &qoveryStableOverride, nil
}

func extractDefaultNodePoolOverrideFromTypesObject(obj types.Object) (*qovery.KarpenterDefaultNodePoolOverride, error) {
	qoveryNodePools, exists := obj.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !exists {
		return nil, fmt.Errorf("qovery_node_pools field not found")
	}

	defaultOverrideAttr, exists := qoveryNodePools.Attributes()["default_override"]
	if !exists {
		return nil, nil
	}

	if defaultOverrideAttr.IsNull() {
		// It means stable_override is not defined
		// No issue as this field is optional
		return nil, nil
	}

	defaultOverride, ok := defaultOverrideAttr.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("default_override field cannot be parsed to Object")
	}

	qoveryDefaultOverride := qovery.KarpenterDefaultNodePoolOverride{}

	// Set limits
	limitsAttr, exists := defaultOverride.Attributes()["limits"]

	// To avoid over-checking conditions when converting the API response to Terraform object, forbid the default_override block if limits are not defined
	if !exists || limitsAttr.IsNull() {
		return nil, fmt.Errorf("if `qovery_node_pools.default_override` is defined, you must define its `limits`")
	}

	limits, ok := limitsAttr.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("limits field cannot be parsed to Object")
	}

	enabled := limits.Attributes()["enabled"].(basetypes.BoolValue)
	limitsCpu := limits.Attributes()["max_cpu_in_vcpu"].(basetypes.Int64Value)
	limitsRam := limits.Attributes()["max_memory_in_gibibytes"].(basetypes.Int64Value)

	qoveryLimits := qovery.NewKarpenterNodePoolLimits(enabled.ValueBool(), int32(limitsCpu.ValueInt64()), int32(limitsRam.ValueInt64()), 0)
	qoveryDefaultOverride.Limits = qoveryLimits

	return &qoveryDefaultOverride, nil
}

func convertObjectToMap(obj attr.Value) (map[string]any, error) {
	reqObject, ok := obj.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("requirement is not an object")
	}

	reqMap := make(map[string]any)

	for key, attr := range reqObject.Attributes() {
		switch v := attr.(type) {
		case basetypes.StringValue:
			reqMap[key] = v.ValueString()
		case basetypes.ListValue:
			values := make([]string, len(v.Elements()))
			for i, elem := range v.Elements() {
				if strVal, ok := elem.(basetypes.StringValue); ok {
					values[i] = strVal.ValueString()
				}
			}
			reqMap[key] = values
		default:
			return nil, fmt.Errorf("unsupported attribute type for key %s", key)
		}
	}

	return reqMap, nil
}

func toCpuArchitectureEnum(arch string) (qovery.CpuArchitectureEnum, error) {
	switch arch {
	case string(qovery.CPUARCHITECTUREENUM_AMD64), string(qovery.CPUARCHITECTUREENUM_ARM64):
		return qovery.CpuArchitectureEnum(arch), nil
	default:
		return "", fmt.Errorf("invalid CPU architecture: %s", arch)
	}
}

func createKarpenterFeatureAttrTypes() map[string]attr.Type {
	attrTypes := make(map[string]attr.Type)
	attrTypes["spot_enabled"] = types.BoolType
	attrTypes["disk_size_in_gib"] = types.Int64Type
	attrTypes["default_service_architecture"] = types.StringType
	attrTypes["qovery_node_pools"] = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"requirements": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"key":      types.StringType,
						"operator": types.StringType,
						"values":   types.ListType{ElemType: types.StringType},
					},
				},
			},
			"stable_override": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"consolidation": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"enabled": types.BoolType,
							"days": types.ListType{
								ElemType: types.StringType,
							},
							"start_time": types.StringType,
							"duration":   types.StringType,
						},
					},
					"limits": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"enabled":                 types.BoolType,
							"max_cpu_in_vcpu":         types.Int64Type,
							"max_memory_in_gibibytes": types.Int64Type,
						},
					},
				},
			},
			"default_override": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"limits": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"enabled":                 types.BoolType,
							"max_cpu_in_vcpu":         types.Int64Type,
							"max_memory_in_gibibytes": types.Int64Type,
						},
					},
				},
			},
		},
	}

	return attrTypes
}

// createFeaturesAttrTypes returns the attribute types for the features object
func createFeaturesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		featureKeyVpcSubnet:      types.StringType,
		featureKeyStaticIP:       types.BoolType,
		featureKeyExistingVpc:    types.ObjectType{AttrTypes: createExistingVpcFeatureAttrTypes()},
		featureKeyGcpExistingVpc: types.ObjectType{AttrTypes: createGcpExistingVpcFeatureAttrTypes()},
		featureKeyKarpenter:      types.ObjectType{AttrTypes: createKarpenterFeatureAttrTypes()},
	}
}

func createExistingVpcFeatureAttrTypes() map[string]attr.Type {
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
	attrTypes["eks_karpenter_fargate_subnets_zone_a_ids"] = types.ListType{ElemType: types.StringType}
	attrTypes["eks_karpenter_fargate_subnets_zone_b_ids"] = types.ListType{ElemType: types.StringType}
	attrTypes["eks_karpenter_fargate_subnets_zone_c_ids"] = types.ListType{ElemType: types.StringType}
	attrTypes["eks_create_nodes_in_private_subnet"] = types.BoolType

	return attrTypes
}

// createGcpExistingVpcFeatureAttrTypes returns the attribute types for the GCP existing VPC feature.
func createGcpExistingVpcFeatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vpc_name":                       types.StringType,
		"vpc_project_id":                 types.StringType,
		"subnetwork_name":                types.StringType,
		"ip_range_services_name":         types.StringType,
		"ip_range_pods_name":             types.StringType,
		"additional_ip_range_pods_names": types.ListType{ElemType: types.StringType},
	}
}

func createKarpenterFeatureAttrValue(karpenterParameters *qovery.ClusterFeatureKarpenterParameters) map[string]attr.Value {
	attrVals := make(map[string]attr.Value)
	var diags diag.Diagnostics

	if karpenterParameters == nil {
		return attrVals
	}

	attrVals["spot_enabled"] = FromBoolPointer(&karpenterParameters.SpotEnabled)
	attrVals["disk_size_in_gib"] = FromInt32(karpenterParameters.DiskSizeInGib)
	attrVals["default_service_architecture"] = FromString(string(karpenterParameters.DefaultServiceArchitecture))

	// Inject requirements
	requirementsAttrList := make([]attr.Value, len(karpenterParameters.QoveryNodePools.Requirements))

	for i, req := range karpenterParameters.QoveryNodePools.Requirements {
		reqAttrVals := make(map[string]attr.Value)

		reqAttrVals["key"] = types.StringValue(string(req.Key))
		reqAttrVals["operator"] = types.StringValue(string(req.Operator))

		valuesAttrList := make([]attr.Value, len(req.Values))
		for j, val := range req.Values {
			valuesAttrList[j] = types.StringValue(val)
		}
		reqAttrVals["values"], diags = types.ListValue(types.StringType, valuesAttrList)
		if diags.HasError() {
			return nil
		}

		reqObjectValue, diag := types.ObjectValue(map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"values":   types.ListType{ElemType: types.StringType},
		}, reqAttrVals)

		if diag.HasError() {
			return nil
		}

		requirementsAttrList[i] = reqObjectValue
	}

	qoveryNodePoolsAttrVals := make(map[string]attr.Value)
	qoveryNodePoolsAttrVals["requirements"], diags = types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"values":   types.ListType{ElemType: types.StringType},
		},
	}, requirementsAttrList)

	if diags.HasError() {
		return nil
	}

	// Inject stable_override
	// Set non null stable_override only if the api returns a non null consolidation or a non null limits
	if karpenterParameters.QoveryNodePools.StableOverride != nil &&
		(karpenterParameters.QoveryNodePools.StableOverride.Consolidation != nil || karpenterParameters.QoveryNodePools.StableOverride.Limits != nil) {
		var stableOverrideConsolidationAttr basetypes.ObjectValue
		var stableOverrideLimitsAttr basetypes.ObjectValue
		consolidation := karpenterParameters.QoveryNodePools.StableOverride.Consolidation
		limits := karpenterParameters.QoveryNodePools.StableOverride.Limits

		if consolidation != nil {
			daysAttr := make([]attr.Value, len(consolidation.Days))
			for i, day := range consolidation.Days {
				daysAttr[i] = types.StringValue(string(day))
			}
			stableOverrideConsolidationAttr = types.ObjectValueMust(
				map[string]attr.Type{
					"enabled": types.BoolType,
					"days": types.ListType{
						ElemType: types.StringType,
					},
					"start_time": types.StringType,
					"duration":   types.StringType,
				},
				map[string]attr.Value{
					"enabled":    types.BoolValue(consolidation.Enabled),
					"days":       types.ListValueMust(types.StringType, daysAttr),
					"start_time": types.StringValue(consolidation.StartTime),
					"duration":   types.StringValue(consolidation.Duration),
				},
			)
		} else {
			stableOverrideConsolidationAttr = types.ObjectNull(map[string]attr.Type{
				"enabled": types.BoolType,
				"days": types.ListType{
					ElemType: types.StringType,
				},
				"start_time": types.StringType,
				"duration":   types.StringType,
			})
		}

		if limits != nil {
			stableOverrideLimitsAttr = types.ObjectValueMust(
				map[string]attr.Type{
					"enabled":                 types.BoolType,
					"max_cpu_in_vcpu":         types.Int64Type,
					"max_memory_in_gibibytes": types.Int64Type,
				},
				map[string]attr.Value{
					"enabled":                 types.BoolValue(limits.Enabled),
					"max_cpu_in_vcpu":         types.Int64Value(int64(limits.MaxCpuInVcpu)),
					"max_memory_in_gibibytes": types.Int64Value(int64(limits.MaxMemoryInGibibytes)),
				},
			)
		} else {
			stableOverrideLimitsAttr = types.ObjectNull(map[string]attr.Type{
				"enabled":                 types.BoolType,
				"max_cpu_in_vcpu":         types.Int64Type,
				"max_memory_in_gibibytes": types.Int64Type,
			})
		}

		stableOverrideAttr := types.ObjectValueMust(
			map[string]attr.Type{
				"consolidation": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled": types.BoolType,
						"days": types.ListType{
							ElemType: types.StringType,
						},
						"start_time": types.StringType,
						"duration":   types.StringType,
					},
				},
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
			map[string]attr.Value{
				"consolidation": stableOverrideConsolidationAttr,
				"limits":        stableOverrideLimitsAttr,
			},
		)

		qoveryNodePoolsAttrVals["stable_override"] = stableOverrideAttr
	} else {
		qoveryNodePoolsAttrVals["stable_override"] = types.ObjectNull(
			map[string]attr.Type{
				"consolidation": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled": types.BoolType,
						"days": types.ListType{
							ElemType: types.StringType,
						},
						"start_time": types.StringType,
						"duration":   types.StringType,
					},
				},
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
		)
	}

	// Inject default override
	var defaultOverrideLimitsAttr basetypes.ObjectValue

	if karpenterParameters.QoveryNodePools.DefaultOverride != nil {
		limits := karpenterParameters.QoveryNodePools.DefaultOverride.Limits
		defaultOverrideLimitsAttr = types.ObjectValueMust(
			map[string]attr.Type{
				"enabled":                 types.BoolType,
				"max_cpu_in_vcpu":         types.Int64Type,
				"max_memory_in_gibibytes": types.Int64Type,
			},
			map[string]attr.Value{
				"enabled":                 types.BoolValue(limits.Enabled),
				"max_cpu_in_vcpu":         types.Int64Value(int64(limits.MaxCpuInVcpu)),
				"max_memory_in_gibibytes": types.Int64Value(int64(limits.MaxMemoryInGibibytes)),
			},
		)

		defaultOverrideAttr := types.ObjectValueMust(
			map[string]attr.Type{
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
			map[string]attr.Value{
				"limits": defaultOverrideLimitsAttr,
			},
		)
		qoveryNodePoolsAttrVals["default_override"] = defaultOverrideAttr
	} else {
		qoveryNodePoolsAttrVals["default_override"] = types.ObjectNull(
			map[string]attr.Type{
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
		)
	}

	// Inject qovery_node_pools
	attrVals["qovery_node_pools"], diags = types.ObjectValue(map[string]attr.Type{
		"requirements": types.ListType{ElemType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key":      types.StringType,
				"operator": types.StringType,
				"values":   types.ListType{ElemType: types.StringType},
			},
		}},
		"stable_override": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"consolidation": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled": types.BoolType,
						"days": types.ListType{
							ElemType: types.StringType,
						},
						"start_time": types.StringType,
						"duration":   types.StringType,
					},
				},
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
		},
		"default_override": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"limits": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled":                 types.BoolType,
						"max_cpu_in_vcpu":         types.Int64Type,
						"max_memory_in_gibibytes": types.Int64Type,
					},
				},
			},
		},
	}, qoveryNodePoolsAttrVals)

	if diags.HasError() {
		return nil
	}

	return attrVals
}

// Infrastructure Charts Parameters helper functions

// validateIPAddressPool validates IP address pool format (single IP or IP-IP range)
func validateIPAddressPool(pool string) error {
	pool = strings.TrimSpace(pool)
	if pool == "" {
		return fmt.Errorf("IP address pool cannot be empty")
	}

	// Check if it's a range (IP-IP format)
	if strings.Contains(pool, "-") {
		parts := strings.Split(pool, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid IP range format '%s': expected 'IP-IP'", pool)
		}
		startIP := net.ParseIP(strings.TrimSpace(parts[0]))
		endIP := net.ParseIP(strings.TrimSpace(parts[1]))
		if startIP == nil {
			return fmt.Errorf("invalid start IP in range '%s'", pool)
		}
		if endIP == nil {
			return fmt.Errorf("invalid end IP in range '%s'", pool)
		}
		// Ensure both are IPv4
		if startIP.To4() == nil {
			return fmt.Errorf("start IP '%s' is not a valid IPv4 address", parts[0])
		}
		if endIP.To4() == nil {
			return fmt.Errorf("end IP '%s' is not a valid IPv4 address", parts[1])
		}
	} else {
		// Single IP
		ip := net.ParseIP(pool)
		if ip == nil {
			return fmt.Errorf("invalid IP address '%s'", pool)
		}
		if ip.To4() == nil {
			return fmt.Errorf("IP '%s' is not a valid IPv4 address", pool)
		}
	}
	return nil
}

// createInfrastructureChartsParametersAttrTypes returns the attribute types for infrastructure charts parameters
func createInfrastructureChartsParametersAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		infraChartsNginxKey: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"replica_count":                             types.Int64Type,
				"default_ssl_certificate":                   types.StringType,
				"publish_status_address":                    types.StringType,
				"annotation_metal_lb_load_balancer_ips":     types.StringType,
				"annotation_external_dns_kubernetes_target": types.StringType,
			},
		},
		infraChartsCertManagerKey: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"kubernetes_namespace": types.StringType,
			},
		},
		infraChartsMetalLbKey: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"ip_address_pools": types.ListType{ElemType: types.StringType},
			},
		},
	}
}

// toQoveryInfrastructureChartsParameters converts Terraform infrastructure charts parameters to Qovery API format
func toQoveryInfrastructureChartsParameters(obj types.Object) (*qovery.ClusterInfrastructureChartsParameters, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	params := qovery.NewClusterInfrastructureChartsParameters()

	// Parse nginx parameters
	if nginxAttr, ok := obj.Attributes()[infraChartsNginxKey]; ok && !nginxAttr.IsNull() {
		nginxObj := nginxAttr.(types.Object)
		nginxParams := qovery.NewClusterInfrastructureNginxChartParameters()

		if v, ok := nginxObj.Attributes()["replica_count"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := int32(v.(types.Int64).ValueInt64())
			nginxParams.ReplicaCount = &val
		}
		if v, ok := nginxObj.Attributes()["default_ssl_certificate"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := v.(types.String).ValueString()
			nginxParams.DefaultSslCertificate = &val
		}
		if v, ok := nginxObj.Attributes()["publish_status_address"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := v.(types.String).ValueString()
			nginxParams.PublishStatusAddress = &val
		}
		if v, ok := nginxObj.Attributes()["annotation_metal_lb_load_balancer_ips"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := v.(types.String).ValueString()
			nginxParams.AnnotationMetalLbLoadBalancerIps = &val
		}
		if v, ok := nginxObj.Attributes()["annotation_external_dns_kubernetes_target"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := v.(types.String).ValueString()
			nginxParams.AnnotationExternalDnsKubernetesTarget = &val
		}
		params.NginxParameters = nginxParams
	}

	// Parse cert manager parameters
	if certManagerAttr, ok := obj.Attributes()[infraChartsCertManagerKey]; ok && !certManagerAttr.IsNull() {
		certManagerObj := certManagerAttr.(types.Object)
		certManagerParams := qovery.NewClusterInfrastructureCertManagerChartParameters()

		if v, ok := certManagerObj.Attributes()["kubernetes_namespace"]; ok && !v.IsNull() && !v.IsUnknown() {
			val := v.(types.String).ValueString()
			certManagerParams.KubernetesNamespace = &val
		}
		params.CertManagerParameters = certManagerParams
	}

	// Parse metal LB parameters
	if metalLbAttr, ok := obj.Attributes()[infraChartsMetalLbKey]; ok && !metalLbAttr.IsNull() {
		metalLbObj := metalLbAttr.(types.Object)
		metalLbParams := qovery.NewClusterInfrastructureMetalLbChartParameters()

		if v, ok := metalLbObj.Attributes()["ip_address_pools"]; ok && !v.IsNull() && !v.IsUnknown() {
			poolsList := v.(types.List)
			pools := make([]string, 0, len(poolsList.Elements()))
			for _, elem := range poolsList.Elements() {
				pool := elem.(types.String).ValueString()
				// Validate IP pool format
				if err := validateIPAddressPool(pool); err != nil {
					return nil, fmt.Errorf("invalid ip_address_pools: %w", err)
				}
				pools = append(pools, pool)
			}
			metalLbParams.IpAddressPools = pools
		}
		params.MetalLbParameters = metalLbParams
	}

	return params, nil
}

// fromQoveryInfrastructureChartsParameters converts Qovery API infrastructure charts parameters to Terraform format
func fromQoveryInfrastructureChartsParameters(params *qovery.ClusterInfrastructureChartsParameters) types.Object {
	attrTypes := createInfrastructureChartsParametersAttrTypes()

	if params == nil {
		return types.ObjectNull(attrTypes)
	}

	attrVals := make(map[string]attr.Value)

	// Convert nginx parameters
	nginxAttrTypes := attrTypes[infraChartsNginxKey].(types.ObjectType).AttrTypes
	if params.NginxParameters != nil {
		nginx := params.NginxParameters
		nginxVals := map[string]attr.Value{
			"replica_count":                             types.Int64Null(),
			"default_ssl_certificate":                   types.StringNull(),
			"publish_status_address":                    types.StringNull(),
			"annotation_metal_lb_load_balancer_ips":     types.StringNull(),
			"annotation_external_dns_kubernetes_target": types.StringNull(),
		}
		if nginx.ReplicaCount != nil {
			nginxVals["replica_count"] = types.Int64Value(int64(*nginx.ReplicaCount))
		}
		if nginx.DefaultSslCertificate != nil {
			nginxVals["default_ssl_certificate"] = types.StringValue(*nginx.DefaultSslCertificate)
		}
		if nginx.PublishStatusAddress != nil {
			nginxVals["publish_status_address"] = types.StringValue(*nginx.PublishStatusAddress)
		}
		if nginx.AnnotationMetalLbLoadBalancerIps != nil {
			nginxVals["annotation_metal_lb_load_balancer_ips"] = types.StringValue(*nginx.AnnotationMetalLbLoadBalancerIps)
		}
		if nginx.AnnotationExternalDnsKubernetesTarget != nil {
			nginxVals["annotation_external_dns_kubernetes_target"] = types.StringValue(*nginx.AnnotationExternalDnsKubernetesTarget)
		}
		attrVals[infraChartsNginxKey] = types.ObjectValueMust(nginxAttrTypes, nginxVals)
	} else {
		attrVals[infraChartsNginxKey] = types.ObjectNull(nginxAttrTypes)
	}

	// Convert cert manager parameters
	certManagerAttrTypes := attrTypes[infraChartsCertManagerKey].(types.ObjectType).AttrTypes
	if params.CertManagerParameters != nil {
		certManager := params.CertManagerParameters
		certManagerVals := map[string]attr.Value{
			"kubernetes_namespace": types.StringNull(),
		}
		if certManager.KubernetesNamespace != nil {
			certManagerVals["kubernetes_namespace"] = types.StringValue(*certManager.KubernetesNamespace)
		}
		attrVals[infraChartsCertManagerKey] = types.ObjectValueMust(certManagerAttrTypes, certManagerVals)
	} else {
		attrVals[infraChartsCertManagerKey] = types.ObjectNull(certManagerAttrTypes)
	}

	// Convert metal LB parameters
	metalLbAttrTypes := attrTypes[infraChartsMetalLbKey].(types.ObjectType).AttrTypes
	if params.MetalLbParameters != nil {
		metalLb := params.MetalLbParameters
		metalLbVals := map[string]attr.Value{
			"ip_address_pools": types.ListNull(types.StringType),
		}
		if len(metalLb.IpAddressPools) > 0 {
			poolVals := make([]attr.Value, len(metalLb.IpAddressPools))
			for i, pool := range metalLb.IpAddressPools {
				poolVals[i] = types.StringValue(pool)
			}
			metalLbVals["ip_address_pools"] = types.ListValueMust(types.StringType, poolVals)
		}
		attrVals[infraChartsMetalLbKey] = types.ObjectValueMust(metalLbAttrTypes, metalLbVals)
	} else {
		attrVals[infraChartsMetalLbKey] = types.ObjectNull(metalLbAttrTypes)
	}

	return types.ObjectValueMust(attrTypes, attrVals)
}
