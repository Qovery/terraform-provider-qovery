package qovery

import (
	"context"
	"fmt"
	"net"
	"reflect"
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
	featureKeyNatGateways    = "nat_gateways"
	featureIdNatGateway      = "NAT_GATEWAY"
	featureIdExistingVpc     = "EXISTING_VPC"
	featureKeyExistingVpc    = "existing_vpc"
	featureIdKarpenter       = "KARPENTER"
	featureKeyKarpenter      = "karpenter"
	featureKeyGcpExistingVpc = "gcp_existing_vpc"
	featureKeyGkeKmsKey      = "gke_kms_key"
	featureIdGkeKmsKey       = "GKE_KMS_KEY"

	instanceTypeAutoPilot = "AUTO_PILOT"

	// Infrastructure charts parameter keys
	infraChartsNginxKey         = "nginx_parameters"
	infraChartsCertManagerKey   = "cert_manager_parameters"
	infraChartsMetalLbKey       = "metal_lb_parameters"
	infraChartsEksAnywhereKey   = "eks_anywhere_parameters"
	infraChartsClusterBackupKey = "cluster_backup"
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
	Keda                           types.Object `tfsdk:"keda"`
	RoutingTables                  types.Set    `tfsdk:"routing_table"`
	State                          types.String `tfsdk:"state"`
	AdvancedSettingsJson           types.String `tfsdk:"advanced_settings_json"`
	Kubeconfig                     types.String `tfsdk:"kubeconfig"`
	InfrastructureOutputs          types.Object `tfsdk:"infrastructure_outputs"`
	InfrastructureChartsParameters types.Object `tfsdk:"infrastructure_charts_parameters"`
	LabelsGroupIds                 types.Set    `tfsdk:"labels_group_ids"`
	SecretManagerAccesses          types.Set    `tfsdk:"secret_manager_accesses"`
}

// stateClusterFeatures returns the features object of the prior state, or a null object when
// there is no prior state (create).
func stateClusterFeatures(state *Cluster) types.Object {
	if state == nil {
		return types.ObjectNull(createFeaturesAttrTypes())
	}
	return state.Features
}

func (c Cluster) hasFeaturesDiff(state *Cluster) bool {
	clusterFeatures, _ := toQoveryClusterFeatures(c.Features, ToString(c.KubernetesMode), ToString(c.CloudProvider), stateClusterFeatures(state))
	if state == nil {
		return len(clusterFeatures) > 0
	}

	stateFeature, _ := toQoveryClusterFeatures(state.Features, ToString(state.KubernetesMode), ToString(state.CloudProvider), state.Features)
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
		if stateValue, ok := stateFeaturesByID[cf.GetId()]; !ok || !reflect.DeepEqual(stateValue, value.GetActualInstance()) {
			return true
		}
	}
	return false
}

// hasKedaDiff reports whether the keda configuration changed between plan and
// state. KEDA installs/removes the operator on the cluster, which only takes
// effect on a (re)deploy gated by ClusterUpsertParams.ForceUpdate — so a toggle
// must force a deploy, otherwise EditCluster saves it but it is never applied.
func (c Cluster) hasKedaDiff(state *Cluster) bool {
	if state == nil {
		return toQoveryClusterKeda(c.Keda) != nil
	}
	return !c.Keda.Equal(state.Keda)
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

func (c Cluster) hasSecretManagerAccessesDiff(state *Cluster) bool {
	// Opt-in attribute: if not set in config, don't force an update.
	if c.SecretManagerAccesses.IsNull() || c.SecretManagerAccesses.IsUnknown() {
		return false
	}
	if state == nil || state.SecretManagerAccesses.IsNull() || state.SecretManagerAccesses.IsUnknown() {
		return len(c.SecretManagerAccesses.Elements()) > 0
	}
	return !c.SecretManagerAccesses.Equal(state.SecretManagerAccesses)
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

// hasClusterSpecDiff reports whether any deploy-affecting cluster spec attribute
// changed between plan and state. Qovery persists these on EditCluster but only
// applies them to the running cluster on a (re)deploy, which is gated by
// ClusterUpsertParams.ForceUpdate — so a change here must force a deploy, otherwise
// the edit is saved but silently not applied. Metadata-only attributes (name,
// description, production) are intentionally excluded: they apply without a
// redeploy. In particular, `production` is metadata-only on edit: at create it
// derives advanced-settings defaults and domain kind (InfrastructureProviderService.kt:264,466),
// but at edit q-core's KubernetesProviderDomain.update persists it without
// re-deriving any deploy-affecting settings.
func (c Cluster) hasClusterSpecDiff(state *Cluster) bool {
	if state == nil {
		return false
	}
	return !c.InstanceType.Equal(state.InstanceType) ||
		!c.DiskSize.Equal(state.DiskSize) ||
		!c.MinRunningNodes.Equal(state.MinRunningNodes) ||
		!c.MaxRunningNodes.Equal(state.MaxRunningNodes) ||
		!c.KubernetesMode.Equal(state.KubernetesMode) ||
		!c.LabelsGroupIds.Equal(state.LabelsGroupIds)
}

func (c Cluster) toUpsertClusterRequest(state *Cluster) (*client.ClusterUpsertParams, error) {
	cloudProvider, err := qovery.NewCloudProviderEnumFromValue(ToString(c.CloudProvider))
	if err != nil {
		return nil, err
	}
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

		// keda is not applicable for PARTIALLY_MANAGED: convertResponseToCluster forces it
		// to null, so a config-set value would yield "inconsistent result after apply".
		if toQoveryClusterKeda(c.Keda) != nil {
			return nil, errors.New("keda is not supported when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
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
					// Check if karpenter has actual content. Every attribute is checked rather
					// than a single required one: spot_enabled is deprecated and now optional,
					// so a karpenter block can legitimately leave it unset.
					karpenterObj := karpenter.(types.Object)
					if !karpenterObj.IsNull() {
						for _, attribute := range karpenterObj.Attributes() {
							if !attribute.IsNull() && !attribute.IsUnknown() {
								hasNonDefaultFeatures = true
								break
							}
						}
					}
				}
			}
			if gkeKmsKey, ok := featuresAttrs[featureKeyGkeKmsKey]; ok {
				if !gkeKmsKey.IsNull() && !gkeKmsKey.IsUnknown() && gkeKmsKey.(types.String).ValueString() != "" {
					hasNonDefaultFeatures = true
				}
			}

			if hasNonDefaultFeatures {
				return nil, errors.New("features (vpc_subnet, static_ip, existing_vpc, karpenter, gke_kms_key) are not supported when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
			}
		}
	} else if infraChartsParams != nil {
		// infrastructure_charts_parameters should not be set for non-PARTIALLY_MANAGED modes
		return nil, errors.New("infrastructure_charts_parameters is only supported when kubernetes_mode is PARTIALLY_MANAGED (EKS Anywhere)")
	}

	features, err := toQoveryClusterFeatures(c.Features, ToString(c.KubernetesMode), ToString(c.CloudProvider), stateClusterFeatures(state))
	if err != nil {
		return nil, err
	}

	// For PARTIALLY_MANAGED mode, clear features to avoid sending them to API
	if isPartiallyManaged {
		features = nil
	}

	karpenterEnabled := false
	for _, f := range features {
		if f.Id != nil && *f.Id == featureIdKarpenter {
			if state != nil && !IsKarpenterAlreadyInstalled(state) {
				return nil, errors.New("It is not possible to migrate to Karpenter using terraform")
			}
			karpenterEnabled = true
			break
		}
	}

	// Validation: Require Karpenter for new EKS clusters
	if state == nil { // This is a new cluster creation
		isAWS := ToString(c.CloudProvider) == "AWS"
		isManaged := kubernetesMode != nil && *kubernetesMode == qovery.KUBERNETESENUM_MANAGED

		if isAWS && isManaged {
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

	secretManagerAccesses, err := toQoverySecretManagerAccessRequests(c.SecretManagerAccesses)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse secret_manager_accesses")
	}

	forceUpdate := c.hasFeaturesDiff(state) || c.hasRoutingTableDiff(state) || c.hasInfraChartsParamsDiff(state) || c.hasClusterSpecDiff(state) || c.hasSecretManagerAccessesDiff(state) || c.hasKedaDiff(state)

	desiredState, err := qovery.NewClusterStateEnumFromValue(ToString(c.State))
	if err != nil {
		return nil, err
	}

	// When Karpenter is enabled, these fields are managed by Karpenter — don't send them to the API.
	var instanceType *string
	var diskSize *int32
	var minRunningNodes *int32
	var maxRunningNodes *int32
	if !karpenterEnabled {
		instanceType = ToStringPointer(c.InstanceType)
		diskSize = ToInt64Pointer(c.DiskSize)
		minRunningNodes = ToInt32Pointer(c.MinRunningNodes)
		maxRunningNodes = ToInt32Pointer(c.MaxRunningNodes)
	}

	var labelsGroups []qovery.ClusterLabelsGroup
	if !c.LabelsGroupIds.IsNull() && !c.LabelsGroupIds.IsUnknown() {
		labelsGroups = make([]qovery.ClusterLabelsGroup, 0, len(c.LabelsGroupIds.Elements()))
		for _, id := range c.LabelsGroupIds.Elements() {
			idStr := id.(types.String).ValueString()
			labelsGroups = append(labelsGroups, qovery.ClusterLabelsGroup{Id: &idStr})
		}
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
			InstanceType:                   instanceType,
			DiskSize:                       diskSize,
			MinRunningNodes:                minRunningNodes,
			MaxRunningNodes:                maxRunningNodes,
			Production:                     ToBoolPointer(c.Production),
			Features:                       features,
			Keda:                           toQoveryClusterKeda(c.Keda),
			InfrastructureChartsParameters: infraChartsParams,
			LabelsGroups:                   labelsGroups,
			SecretManagerAccesses:          secretManagerAccesses,
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

	oldFeatures, _ := toQoveryClusterFeatures(state.Features, ToString(state.KubernetesMode), ToString(state.CloudProvider), state.Features)
	for _, f := range oldFeatures {
		if f.Id != nil && *f.Id == featureIdKarpenter {
			return true
		}
	}
	return false
}

// responseHasKarpenter reports whether the API cluster response has the Karpenter feature enabled.
func responseHasKarpenter(features []qovery.ClusterFeatureResponse) bool {
	for _, f := range features {
		if f.Id != nil && *f.Id == featureIdKarpenter {
			return true
		}
	}
	return false
}

// clusterReadMode distinguishes the two callers of the API -> Terraform conversion, which need
// different amounts of the API response.
//
// The resource must stay symmetric between apply and import: ImportStateVerify compares the two
// states attribute by attribute, so import may only store node pool overrides that an apply would
// also store. The data source is read-only — no plan to be consistent with and no diff to keep
// quiet — so hiding a real divergence there would be a bug, and it reports whatever the API holds.
type clusterReadMode int

const (
	clusterReadModeResource clusterReadMode = iota
	clusterReadModeDataSource
)

func convertResponseToCluster(ctx context.Context, res *client.ClusterResponse, initialPlan Cluster) Cluster {
	return convertResponseToClusterWithMode(ctx, res, initialPlan, clusterReadModeResource)
}

// convertResponseToClusterForDataSource converts an API response for the read-only cluster data
// source, which reports the full API truth rather than mirroring what an apply would store.
func convertResponseToClusterForDataSource(ctx context.Context, res *client.ClusterResponse, config Cluster) Cluster {
	return convertResponseToClusterWithMode(ctx, res, config, clusterReadModeDataSource)
}

func convertResponseToClusterWithMode(ctx context.Context, res *client.ClusterResponse, initialPlan Cluster, mode clusterReadMode) Cluster {
	routingTable := fromClusterRoutingTable(res.ClusterRoutingTable)

	// Check if cluster is PARTIALLY_MANAGED (EKS Anywhere)
	isPartiallyManaged := res.ClusterResponse.Kubernetes != nil &&
		*res.ClusterResponse.Kubernetes == qovery.KUBERNETESENUM_PARTIALLY_MANAGED

	labelsGroupIds := make([]string, 0, len(res.ClusterResponse.LabelsGroups))
	for _, lg := range res.ClusterResponse.LabelsGroups {
		if lg.Id != nil {
			labelsGroupIds = append(labelsGroupIds, *lg.Id)
		}
	}

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
		LabelsGroupIds:                 fromLabelsGroupList(ctx, initialPlan.LabelsGroupIds, labelsGroupIds),
		SecretManagerAccesses:          fromQoverySecretManagerAccesses(ctx, res.ClusterResponse.SecretManagerAccesses, initialPlan.SecretManagerAccesses),
	}

	// For PARTIALLY_MANAGED (EKS Anywhere) clusters, these fields are not applicable
	// Return null values to avoid spurious terraform plan changes
	if isPartiallyManaged {
		cluster.InstanceType = types.StringNull()
		cluster.DiskSize = types.Int64Null()
		cluster.MinRunningNodes = types.Int64Null()
		cluster.MaxRunningNodes = types.Int64Null()
		cluster.Features = types.ObjectNull(createFeaturesAttrTypes())
		cluster.Keda = types.ObjectNull(createKedaAttrTypes())
		cluster.RoutingTables = types.SetNull(types.ObjectType{AttrTypes: clusterRouteAttrTypes})
		cluster.InfrastructureOutputs = types.ObjectNull(clusterInfrastructureOutputsAttrTypes)
		// Preserve kubeconfig from initialPlan - it's fetched separately via API in Read operation
		cluster.Kubeconfig = initialPlan.Kubeconfig
	} else {
		hasKarpenter := responseHasKarpenter(res.ClusterResponse.Features)

		// When Karpenter is enabled the API rewrites instance_type to the literal
		// "KARPENTER" in the response. Preserve the plan value to avoid an
		// "inconsistent result after apply" on create and a spurious diff on every
		// subsequent plan.
		if hasKarpenter && !initialPlan.InstanceType.IsNull() && !initialPlan.InstanceType.IsUnknown() {
			cluster.InstanceType = initialPlan.InstanceType
		} else {
			cluster.InstanceType = FromStringPointer(res.ClusterResponse.InstanceType)
		}
		cluster.DiskSize = FromInt32Pointer(res.ClusterResponse.DiskSize)

		// GCP Autopilot and Karpenter manage node scaling themselves, so min/max_running_nodes
		// are not set in config and the API returns unstable sentinel values (e.g. MaxInt32 right
		// after create, a real number on later reads). Preserve the plan values to avoid a spurious
		// "inconsistent result after apply" on update.
		isAutoPilot := res.ClusterResponse.InstanceType != nil && *res.ClusterResponse.InstanceType == instanceTypeAutoPilot
		nodeCountsManaged := isAutoPilot || hasKarpenter
		if nodeCountsManaged && !initialPlan.MinRunningNodes.IsNull() && !initialPlan.MinRunningNodes.IsUnknown() {
			cluster.MinRunningNodes = initialPlan.MinRunningNodes
		} else {
			cluster.MinRunningNodes = FromInt32Pointer(res.ClusterResponse.MinRunningNodes)
		}
		if nodeCountsManaged && !initialPlan.MaxRunningNodes.IsNull() && !initialPlan.MaxRunningNodes.IsUnknown() {
			cluster.MaxRunningNodes = initialPlan.MaxRunningNodes
		} else {
			cluster.MaxRunningNodes = FromInt32Pointer(res.ClusterResponse.MaxRunningNodes)
		}

		cluster.Features = clusterFeaturesFromResponse(res.ClusterResponse.Features, initialPlan.Features, mode)
		cluster.Keda = fromQoveryClusterKeda(res.ClusterResponse.Keda)
		cluster.RoutingTables = routingTable.toTerraformSet(ctx, initialPlan.RoutingTables)
		cluster.InfrastructureOutputs = fromQoveryClusterOutput(res.ClusterResponse.InfrastructureOutputs, initialPlan.InfrastructureOutputs)
		// Kubeconfig is not applicable for non-PARTIALLY_MANAGED clusters
		cluster.Kubeconfig = types.StringNull()
	}

	return cluster
}

// createKedaAttrTypes returns the attribute types for the cluster `keda` nested object.
func createKedaAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

// toQoveryClusterKeda converts the Terraform `keda` object to the API model.
// Returns nil when the block is absent so the API applies its own default.
func toQoveryClusterKeda(k types.Object) *qovery.ClusterKeda {
	if k.IsNull() || k.IsUnknown() {
		return nil
	}

	enabled := false
	if v, ok := k.Attributes()["enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		enabled = v.ValueBool()
	}

	return qovery.NewClusterKeda(enabled)
}

// fromQoveryClusterKeda converts the API `keda` model to a Terraform object.
// The API omits `keda` when KEDA is disabled, so a nil pointer maps to the
// disabled shape `{enabled = false}` rather than null. `keda` is Optional+Computed,
// so always returning a known value keeps apply results consistent with the plan
// (a planned `{enabled = false}` must not become null after apply).
func fromQoveryClusterKeda(k *qovery.ClusterKeda) types.Object {
	enabled := false
	if k != nil {
		enabled = k.GetEnabled()
	}

	return types.ObjectValueMust(createKedaAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(enabled),
	})
}

var clusterInfrastructureOutputsAttrTypes = map[string]attr.Type{
	"cluster_name":        types.StringType,
	"cluster_arn":         types.StringType,
	"cluster_self_link":   types.StringType,
	"cluster_oidc_issuer": types.StringType,
	"vpc_id":              types.StringType,
}

// fromQoveryClusterOutput converts the API InfrastructureOutputs to a Terraform Object.
// Falls back to the prior state when the API returns nil or an unrecognized payload: these
// are read-only outputs populated by a successful deploy, and the API can temporarily omit
// them (e.g. when the cluster is in DEPLOYMENT_ERROR), which would otherwise break any
// downstream resource referencing them.
func fromQoveryClusterOutput(
	infrastructureOutputs *qovery.InfrastructureOutputs,
	priorState types.Object,
) types.Object {
	if infrastructureOutputs == nil {
		if clusterOutputHasAnyKnownValue(priorState) {
			return priorState
		}
		return types.ObjectValueMust(clusterInfrastructureOutputsAttrTypes, allNullClusterOutputValues())
	}

	values := allNullClusterOutputValues()

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

	default:
		if clusterOutputHasAnyKnownValue(priorState) {
			return priorState
		}
	}

	return types.ObjectValueMust(clusterInfrastructureOutputsAttrTypes, values)
}

func allNullClusterOutputValues() map[string]attr.Value {
	return map[string]attr.Value{
		"cluster_name":        types.StringNull(),
		"cluster_arn":         types.StringNull(),
		"cluster_self_link":   types.StringNull(),
		"cluster_oidc_issuer": types.StringNull(),
		"vpc_id":              types.StringNull(),
	}
}

// clusterOutputHasAnyKnownValue reports whether the cluster InfrastructureOutputs object
// holds at least one non-empty string attribute. All cluster output attributes are strings;
// any non-String attr would be a schema bug and is treated as not-yet-known.
func clusterOutputHasAnyKnownValue(o types.Object) bool {
	if o.IsNull() || o.IsUnknown() {
		return false
	}
	for _, v := range o.Attributes() {
		s, ok := v.(types.String)
		if !ok {
			continue
		}
		if !s.IsNull() && !s.IsUnknown() && s.ValueString() != "" {
			return true
		}
	}
	return false
}

// planKarpenterObject digs the karpenter feature out of a planned features object. It returns
// a null object when there is no plan to compare the API response against.
func planKarpenterObject(planFeatures types.Object) types.Object {
	if planFeatures.IsNull() || planFeatures.IsUnknown() {
		return types.ObjectNull(createKarpenterFeatureAttrTypes())
	}
	karpenter, ok := planFeatures.Attributes()[featureKeyKarpenter].(basetypes.ObjectValue)
	if !ok {
		return types.ObjectNull(createKarpenterFeatureAttrTypes())
	}
	return karpenter
}

// fromQoveryClusterFeatures converts the API features to their Terraform representation.
// planFeatures is the planned (or, on a refresh, the prior state) features object: some
// Karpenter values are only stored in state when the configuration asked for them, so that a
// value the API returns on its own does not become permanent plan noise. Pass a null object
// when no plan is available, e.g. from the data source.
func clusterFeaturesFromResponse(
	clusterFeatures []qovery.ClusterFeatureResponse,
	planFeatures types.Object,
	mode clusterReadMode,
) types.Object {
	if clusterFeatures == nil {
		// Early return object null without attribute types
		return types.ObjectNull(make(map[string]attr.Type))
	}

	attributes := make(map[string]attr.Value)
	attributeTypes := make(map[string]attr.Type)
	hasStaticIPFeature := false
	natGatewaysStaticIPEnabled := false
	var natGatewaysStaticIPCount int32
	hasGcpNatGateways := false
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
		case featureIdGkeKmsKey:
			if f.GetValueObject().ClusterFeatureStringResponse != nil {
				attributes[featureKeyGkeKmsKey] = FromString(f.GetValueObject().ClusterFeatureStringResponse.Value)
			} else {
				attributes[featureKeyGkeKmsKey] = basetypes.NewStringNull()
			}
			attributeTypes[featureKeyGkeKmsKey] = types.StringType
		case featureIdNatGateway:
			if f.GetValueObject().ClusterFeatureNatGatewayParametersResponse != nil {
				natGateways := f.GetValueObject().ClusterFeatureNatGatewayParametersResponse.GetValue()
				if gcpNatGateways := natGateways.GetNatGatewayType().ClusterFeatureNatGatewayTypeGcp; gcpNatGateways != nil {
					hasGcpNatGateways = true
					natGatewaysStaticIPEnabled = gcpNatGateways.StaticIpsEnabled
					natGatewaysStaticIPCount = gcpNatGateways.StaticIpsCount
				}
			}
		case featureIdStaticIP:
			hasStaticIPFeature = true
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
					"private_nodes":                  FromBoolPointer(gcpVpc.PrivateNodes),
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
			attrVals["elasticache_subnets_zone_a_ids"] = FromStringArray(v.ElasticacheSubnetsZoneAIds)
			attrVals["elasticache_subnets_zone_b_ids"] = FromStringArray(v.ElasticacheSubnetsZoneBIds)
			attrVals["elasticache_subnets_zone_c_ids"] = FromStringArray(v.ElasticacheSubnetsZoneCIds)

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

			attrVals := karpenterFeatureAttrValue(karpenterParameters, planKarpenterObject(planFeatures), mode)

			terraformObjectValue, diagnostics := types.ObjectValue(attrTypes, attrVals)
			if diagnostics.HasError() {
				panic(fmt.Errorf("bad %s feature: %s", featureKeyExistingVpc, diagnostics.Errors()))
			}
			attributes[featureKeyKarpenter] = terraformObjectValue
			attributeTypes[featureKeyKarpenter] = terraformObjectValue.Type(context.Background())
		}
	}

	if !hasStaticIPFeature && hasGcpNatGateways {
		// Derive static_ip from the NAT_GATEWAY feature when the STATIC_IP feature is absent.
		attributes[featureKeyStaticIP] = types.BoolValue(natGatewaysStaticIPEnabled)
		attributeTypes[featureKeyStaticIP] = types.BoolType
	}

	// Determine the final static_ip bool (arbitration: STATIC_IP feature wins when present;
	// else derive from NAT enabled flag). We need it before building nat_gateways so we can
	// decide whether to normalize remembered params or emit them verbatim.
	finalStaticIP := false
	if staticIPVal, ok := attributes[featureKeyStaticIP]; ok {
		if bv, ok2 := staticIPVal.(types.Bool); ok2 && !bv.IsNull() && !bv.IsUnknown() {
			finalStaticIP = bv.ValueBool()
		}
	}

	if hasGcpNatGateways {
		natGatewayAttrTypes := createNatGatewaysFeatureAttrTypes()
		var enabled bool
		var count int64
		if finalStaticIP {
			// Verbatim mapping: emit what the API returned.
			enabled = natGatewaysStaticIPEnabled
			count = int64(natGatewaysStaticIPCount)
		} else {
			// static_ip=false: normalize to default to preserve consistency invariant.
			enabled = false
			count = 1
		}
		natGatewayObj, diagnostics := types.ObjectValue(natGatewayAttrTypes, map[string]attr.Value{
			"static_ips_enabled": types.BoolValue(enabled),
			"static_ips_count":   types.Int64Value(count),
		})
		if diagnostics.HasError() {
			panic(fmt.Errorf("bad %s feature: %s", featureKeyNatGateways, diagnostics.Errors()))
		}
		attributes[featureKeyNatGateways] = natGatewayObj
		attributeTypes[featureKeyNatGateways] = types.ObjectType{AttrTypes: natGatewayAttrTypes}
	}

	// All attributes should be fill even if no feature is present.
	// This is mandatory to satisfy the terraform framework schema.

	if attributes[featureKeyVpcSubnet] == nil {
		attributes[featureKeyVpcSubnet] = FromStringPointer(&clusterFeatureVpcSubnetDefault)
		attributeTypes[featureKeyVpcSubnet] = types.StringType
	}

	if attributes[featureKeyStaticIP] == nil {
		defaultFeatureKeyStaticIP := false
		attributes[featureKeyStaticIP] = FromBoolPointer(&defaultFeatureKeyStaticIP)
		attributeTypes[featureKeyStaticIP] = types.BoolType
	}

	if attributes[featureKeyNatGateways] == nil {
		natGatewaysAttrTypes := createNatGatewaysFeatureAttrTypes()
		attributes[featureKeyNatGateways] = types.ObjectValueMust(natGatewaysAttrTypes, map[string]attr.Value{
			"static_ips_enabled": types.BoolValue(false),
			"static_ips_count":   types.Int64Value(1),
		})
		attributeTypes[featureKeyNatGateways] = types.ObjectType{AttrTypes: natGatewaysAttrTypes}
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

	if attributes[featureKeyGkeKmsKey] == nil {
		attributes[featureKeyGkeKmsKey] = basetypes.NewStringNull()
		attributeTypes[featureKeyGkeKmsKey] = types.StringType
	}

	terraformObjectValue, diagnostics := types.ObjectValue(attributeTypes, attributes)
	if diagnostics.HasError() {
		panic(fmt.Errorf("bad cluster feature: %s", diagnostics.Errors()))
	}
	return terraformObjectValue
}

// toQoveryClusterFeatures converts the Terraform features object into the API request. Passing
// stateFeatures alongside it — the features of the prior state, or a null object on create — lets
// the deprecated global Karpenter spot_enabled fall back to the value the API last derived when
// the plan leaves it unknown; see the karpenter branch below.
func toQoveryClusterFeatures(f types.Object, mode string, cloudProvider string, stateFeatures types.Object) ([]qovery.ClusterRequestFeaturesInner, error) {
	if f.IsNull() || f.IsUnknown() || mode == "K3S" {
		return nil, nil
	}

	features := make([]qovery.ClusterRequestFeaturesInner, 0, len(f.Attributes()))
	if vpcSubnetAttr, ok := f.Attributes()[featureKeyVpcSubnet]; ok {
		vpcSubnet := vpcSubnetAttr.(types.String)
		if cloudProvider != "GCP" {
			// Normalize the legacy empty-string state value to the schema default so a
			// provider upgrade doesn't manufacture a features diff (and a forced redeploy).
			if !vpcSubnet.IsNull() && !vpcSubnet.IsUnknown() && vpcSubnet.ValueString() == "" {
				vpcSubnet = types.StringValue(clusterFeatureVpcSubnetDefault)
			}
			value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				String: ToStringPointer(vpcSubnet),
			})

			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdVpcSubnet),
				Value: *value,
			})
		} else if !vpcSubnet.IsNull() && !vpcSubnet.IsUnknown() && vpcSubnet.ValueString() != "" && vpcSubnet.ValueString() != clusterFeatureVpcSubnetDefault {
			return nil, errors.New("features.vpc_subnet is not supported for GCP clusters")
		}
	}

	if gkeKmsKeyAttr, ok := f.Attributes()[featureKeyGkeKmsKey]; ok {
		gkeKmsKey := gkeKmsKeyAttr.(types.String)
		if !gkeKmsKey.IsNull() && !gkeKmsKey.IsUnknown() && gkeKmsKey.ValueString() != "" {
			value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				String: ToStringPointer(gkeKmsKey),
			})
			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdGkeKmsKey),
				Value: *value,
			})
		}
	}

	staticIPAttr, hasStaticIP := f.Attributes()[featureKeyStaticIP]
	natGatewaysAttr, hasNatGateways := f.Attributes()[featureKeyNatGateways]
	staticIPEnabled := false
	if hasStaticIP {
		staticIPEnabled = ToBool(staticIPAttr.(types.Bool))
	}

	// Apply-time backstop: nat_gateways with enabled=true or count>1 is only valid for
	// GCP clusters with static_ip enabled. Plan-time validation (ValidateConfig) provides
	// the earlier user-facing error; this is a defensive safety net.
	if hasNatGateways && !natGatewaysAttr.IsNull() && !natGatewaysAttr.IsUnknown() {
		natGateways := natGatewaysAttr.(types.Object)
		natAttrs := natGateways.Attributes()
		blockEnabled := false
		if ev, ok := natAttrs["static_ips_enabled"]; ok && !ev.IsNull() && !ev.IsUnknown() {
			blockEnabled = ev.(types.Bool).ValueBool()
		}
		blockCount := int32(1)
		if cv, ok := natAttrs["static_ips_count"]; ok && !cv.IsNull() && !cv.IsUnknown() {
			blockCount = ToInt32(cv.(types.Int64))
		}
		if (blockEnabled || blockCount > 1) && !(cloudProvider == "GCP" && staticIPEnabled) {
			return nil, errors.New("features.nat_gateways with static_ips_enabled or static_ips_count > 1 requires a GCP cluster with features.static_ip enabled")
		}
	}

	if hasStaticIP {
		value := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
			Bool: ToBoolPointer(staticIPAttr.(types.Bool)),
		})

		features = append(features, qovery.ClusterRequestFeaturesInner{
			Id:    new(featureIdStaticIP),
			Value: *value,
		})

		if cloudProvider == "GCP" && staticIPEnabled {
			// GCP with static_ip=true: ALWAYS emit NAT_GATEWAY verbatim from the block
			// (including the disabled shape {false,1}) to keep DB params in sync with plan.
			blockEnabled := false
			blockCount := int32(1)
			if hasNatGateways && !natGatewaysAttr.IsNull() && !natGatewaysAttr.IsUnknown() {
				natGateways := natGatewaysAttr.(types.Object)
				natAttrs := natGateways.Attributes()
				if ev, ok := natAttrs["static_ips_enabled"]; ok && !ev.IsNull() && !ev.IsUnknown() {
					blockEnabled = ev.(types.Bool).ValueBool()
				}
				if cv, ok := natAttrs["static_ips_count"]; ok && !cv.IsNull() && !cv.IsUnknown() {
					blockCount = ToInt32(cv.(types.Int64))
				}
			}
			natGatewayType := qovery.ClusterFeatureNatGatewayTypeGcpAsClusterFeatureNatGatewayParametersNatGatewayType(
				qovery.NewClusterFeatureNatGatewayTypeGcp("gcp", blockEnabled, blockCount),
			)
			natGatewayParameters := qovery.ClusterFeatureNatGatewayParameters{}
			natGatewayParameters.SetNatGatewayType(natGatewayType)
			natValue := qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{
				ClusterFeatureNatGatewayParameters: &natGatewayParameters,
			})

			features = append(features, qovery.ClusterRequestFeaturesInner{
				Id:    new(featureIdNatGateway),
				Value: *natValue,
			})
		}
		// GCP && !staticIPEnabled: emit nothing (absent = not configured in q-core).
		// Non-GCP: never emit NAT_GATEWAY.
	}

	return appendRemainingQoveryClusterFeatures(features, f, stateFeatures)
}

func appendRemainingQoveryClusterFeatures(features []qovery.ClusterRequestFeaturesInner, f types.Object, stateFeatures types.Object) ([]qovery.ClusterRequestFeaturesInner, error) {
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
				PrivateNodes:               ToBoolPointer(attrs["private_nodes"].(types.Bool)),
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

			// The deprecated global flag is planned as unknown whenever the configuration carries
			// per node pool values, because the API derives it from them (see
			// DeprecatedGlobalSpotEnabled). Unknown must not be sent as false: a node pool the
			// configuration gives no per-pool value falls back to whatever global the request
			// carries, so false would silently switch such a pool off spot midway through the
			// documented migration. Send the value the API last derived instead, which leaves those
			// pools exactly where they are. On create there is no prior state and every pool is
			// new, so false is the safe default.
			globalSpotEnabled := v.Attributes()["spot_enabled"].(types.Bool)
			if globalSpotEnabled.IsUnknown() {
				globalSpotEnabled = knownBool(planKarpenterObject(stateFeatures).Attributes()["spot_enabled"])
			}

			feature := qovery.ClusterFeatureKarpenterParameters{
				SpotEnabled:                ToBool(globalSpotEnabled),
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

	// Set cronjob node pool override
	cronjobOverride, err := extractCronjobNodePoolOverrideFromTypesObject(obj)
	if err != nil {
		return nil, err
	}
	karpenterNodePool.CronjobOverride = cronjobOverride

	return &karpenterNodePool, nil
}

// extractNodePoolOverride returns the named node pool override object when the configuration
// declares it. A null or unknown block means the block is absent, which is meaningful for the
// API: an absent block leaves the corresponding node pool on its legacy behavior.
func extractNodePoolOverride(obj types.Object, name string) (basetypes.ObjectValue, bool, error) {
	qoveryNodePools, exists := obj.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !exists {
		return basetypes.ObjectValue{}, false, fmt.Errorf("qovery_node_pools field not found")
	}

	overrideAttr, exists := qoveryNodePools.Attributes()[name]
	if !exists || overrideAttr == nil || overrideAttr.IsNull() || overrideAttr.IsUnknown() {
		return basetypes.ObjectValue{}, false, nil
	}

	override, ok := overrideAttr.(basetypes.ObjectValue)
	if !ok {
		return basetypes.ObjectValue{}, false, fmt.Errorf("%s field cannot be parsed to Object", name)
	}

	return override, true, nil
}

// extractNodePoolSpotEnabled returns the configured per node pool spot_enabled, or nil when it
// is null or unknown. Nil must be sent as an absent field: the API then applies the deprecated
// global spot_enabled to that node pool, which is the behavior of every configuration written
// before per node pool support.
func extractNodePoolSpotEnabled(override basetypes.ObjectValue) (*bool, error) {
	spotEnabledAttr, exists := override.Attributes()["spot_enabled"]
	if !exists || spotEnabledAttr == nil || spotEnabledAttr.IsNull() || spotEnabledAttr.IsUnknown() {
		return nil, nil
	}

	spotEnabled, ok := spotEnabledAttr.(basetypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("spot_enabled field cannot be parsed to Bool")
	}

	return spotEnabled.ValueBoolPointer(), nil
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
	stableOverride, declared, err := extractNodePoolOverride(obj, "stable_override")
	if err != nil {
		return nil, err
	}
	if !declared {
		// It means stable_override is not defined
		// No issue as this field is optional
		return nil, nil
	}

	qoveryStableOverride := qovery.KarpenterStableNodePoolOverride{}

	// Set spot_enabled
	spotEnabled, err := extractNodePoolSpotEnabled(stableOverride)
	if err != nil {
		return nil, err
	}
	if spotEnabled != nil {
		SetStableNodePoolSpotEnabled(&qoveryStableOverride, *spotEnabled)
	}

	// Set consolidation
	consolidationAttr, hasConsolidation := stableOverride.Attributes()["consolidation"]
	hasConsolidation = hasConsolidation && consolidationAttr != nil && !consolidationAttr.IsNull()

	// The consolidation is allowed to be null
	if hasConsolidation {
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
	limitsAttr, hasLimits := stableOverride.Attributes()["limits"]
	hasLimits = hasLimits && limitsAttr != nil && !limitsAttr.IsNull()

	// The limits are allowed to be null
	if hasLimits {
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

	// To avoid over-checking conditions when converting the API response to Terraform object, forbid an empty stable_override block
	if !hasConsolidation && !hasLimits && spotEnabled == nil {
		return nil, fmt.Errorf("if `qovery_node_pools.stable_override` is defined, you must define at least its `spot_enabled`, its `consolidation` or its `limits`")
	}

	return &qoveryStableOverride, nil
}

func extractDefaultNodePoolOverrideFromTypesObject(obj types.Object) (*qovery.KarpenterDefaultNodePoolOverride, error) {
	defaultOverride, declared, err := extractNodePoolOverride(obj, "default_override")
	if err != nil {
		return nil, err
	}
	if !declared {
		// It means default_override is not defined
		// No issue as this field is optional
		return nil, nil
	}

	qoveryDefaultOverride := qovery.KarpenterDefaultNodePoolOverride{}

	// Set spot_enabled
	spotEnabled, err := extractNodePoolSpotEnabled(defaultOverride)
	if err != nil {
		return nil, err
	}
	if spotEnabled != nil {
		SetDefaultNodePoolSpotEnabled(&qoveryDefaultOverride, *spotEnabled)
	}

	// Set limits
	limitsAttr, hasLimits := defaultOverride.Attributes()["limits"]
	hasLimits = hasLimits && limitsAttr != nil && !limitsAttr.IsNull()

	// To avoid over-checking conditions when converting the API response to Terraform object, forbid an empty default_override block
	if !hasLimits {
		if spotEnabled == nil {
			return nil, fmt.Errorf("if `qovery_node_pools.default_override` is defined, you must define at least its `spot_enabled` or its `limits`")
		}
		return &qoveryDefaultOverride, nil
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

// extractCronjobNodePoolOverrideFromTypesObject converts the cronjob_override block. The block
// is never synthesized: its mere presence enables the dedicated cronjob node pool across the
// Qovery stack, so it is sent only when the configuration declares it.
func extractCronjobNodePoolOverrideFromTypesObject(obj types.Object) (*qovery.KarpenterCronjobNodePoolOverride, error) {
	cronjobOverride, declared, err := extractNodePoolOverride(obj, "cronjob_override")
	if err != nil {
		return nil, err
	}
	if !declared {
		return nil, nil
	}

	qoveryCronjobOverride := qovery.KarpenterCronjobNodePoolOverride{}

	spotEnabled, err := extractNodePoolSpotEnabled(cronjobOverride)
	if err != nil {
		return nil, err
	}
	if spotEnabled != nil {
		SetCronjobNodePoolSpotEnabled(&qoveryCronjobOverride, *spotEnabled)
	}

	return &qoveryCronjobOverride, nil
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

// The Karpenter node pool object shapes are used by the schema conversion, by the
// TF -> API extraction and by the API -> TF injection. They live in one place each so
// the three stay in sync.

func karpenterRequirementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":      types.StringType,
		"operator": types.StringType,
		"values":   types.ListType{ElemType: types.StringType},
	}
}

func karpenterConsolidationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":    types.BoolType,
		"days":       types.ListType{ElemType: types.StringType},
		"start_time": types.StringType,
		"duration":   types.StringType,
	}
}

func karpenterLimitsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":                 types.BoolType,
		"max_cpu_in_vcpu":         types.Int64Type,
		"max_memory_in_gibibytes": types.Int64Type,
	}
}

func karpenterStableOverrideAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spot_enabled":  types.BoolType,
		"consolidation": types.ObjectType{AttrTypes: karpenterConsolidationAttrTypes()},
		"limits":        types.ObjectType{AttrTypes: karpenterLimitsAttrTypes()},
	}
}

func karpenterDefaultOverrideAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spot_enabled": types.BoolType,
		"limits":       types.ObjectType{AttrTypes: karpenterLimitsAttrTypes()},
	}
}

func karpenterCronjobOverrideAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spot_enabled": types.BoolType,
	}
}

func karpenterNodePoolsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"requirements":     types.ListType{ElemType: types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()}},
		"stable_override":  types.ObjectType{AttrTypes: karpenterStableOverrideAttrTypes()},
		"default_override": types.ObjectType{AttrTypes: karpenterDefaultOverrideAttrTypes()},
		"cronjob_override": types.ObjectType{AttrTypes: karpenterCronjobOverrideAttrTypes()},
	}
}

func createKarpenterFeatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"spot_enabled":                 types.BoolType,
		"disk_size_in_gib":             types.Int64Type,
		"default_service_architecture": types.StringType,
		"qovery_node_pools":            types.ObjectType{AttrTypes: karpenterNodePoolsAttrTypes()},
	}
}

// createFeaturesAttrTypes returns the attribute types for the features object
func createFeaturesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		featureKeyVpcSubnet:      types.StringType,
		featureKeyStaticIP:       types.BoolType,
		featureKeyNatGateways:    types.ObjectType{AttrTypes: createNatGatewaysFeatureAttrTypes()},
		featureKeyExistingVpc:    types.ObjectType{AttrTypes: createExistingVpcFeatureAttrTypes()},
		featureKeyGcpExistingVpc: types.ObjectType{AttrTypes: createGcpExistingVpcFeatureAttrTypes()},
		featureKeyKarpenter:      types.ObjectType{AttrTypes: createKarpenterFeatureAttrTypes()},
		featureKeyGkeKmsKey:      types.StringType,
	}
}

func createNatGatewaysFeatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"static_ips_enabled": types.BoolType,
		"static_ips_count":   types.Int64Type,
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
		"private_nodes":                  types.BoolType,
	}
}

// karpenterPlanView exposes the parts of the plan — or, on a refresh, of the prior state —
// that the API -> Terraform conversion needs to stay plan-consistent. It is unavailable on an
// import or a data source read, where there is no plan to be consistent with; the conversion
// then falls back to what the API response actually contains.
type karpenterPlanView struct {
	available bool
	karpenter basetypes.ObjectValue
}

func newKarpenterPlanView(planKarpenter types.Object) karpenterPlanView {
	if planKarpenter.IsNull() || planKarpenter.IsUnknown() {
		return karpenterPlanView{}
	}
	return karpenterPlanView{available: true, karpenter: planKarpenter}
}

// override returns the planned node pool override block and whether the plan declares it.
func (p karpenterPlanView) override(name string) (basetypes.ObjectValue, bool) {
	if !p.available {
		return basetypes.ObjectValue{}, false
	}
	nodePools, ok := p.karpenter.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !ok || nodePools.IsNull() || nodePools.IsUnknown() {
		return basetypes.ObjectValue{}, false
	}
	override, ok := nodePools.Attributes()[name].(basetypes.ObjectValue)
	if !ok || override.IsNull() || override.IsUnknown() {
		return basetypes.ObjectValue{}, false
	}
	return override, true
}

// declaresOverride reports whether the plan contains the named node pool override block.
func (p karpenterPlanView) declaresOverride(name string) bool {
	_, ok := p.override(name)
	return ok
}

// overrideSpotEnabled returns the planned spot_enabled of a node pool override, or a null
// bool when the plan does not know it.
func (p karpenterPlanView) overrideSpotEnabled(name string) types.Bool {
	override, ok := p.override(name)
	if !ok {
		return types.BoolNull()
	}
	return knownBool(override.Attributes()["spot_enabled"])
}

// globalSpotEnabled returns the planned deprecated global spot_enabled, or a null bool when
// the plan does not know it.
func (p karpenterPlanView) globalSpotEnabled() types.Bool {
	if !p.available {
		return types.BoolNull()
	}
	return knownBool(p.karpenter.Attributes()["spot_enabled"])
}

// declaresAnyNodePoolSpotEnabled reports whether at least one node pool override carries a
// spot_enabled that is not null. Unlike hasAnyNodePoolSpotEnabled it counts unknown values: a
// value that comes from an unresolved expression is still configured, and the caller — the plan
// modifier of the deprecated global flag — must treat it as such. Where a value cannot be seen
// because the enclosing object is itself unknown, it assumes there is one: over-reporting only
// costs a "known after apply" in the plan, while under-reporting resurrects the stale global.
func (p karpenterPlanView) declaresAnyNodePoolSpotEnabled() bool {
	if !p.available {
		return false
	}

	nodePools, ok := p.karpenter.Attributes()["qovery_node_pools"].(basetypes.ObjectValue)
	if !ok || nodePools.IsNull() {
		return false
	}
	if nodePools.IsUnknown() {
		return true
	}

	for _, name := range []string{"stable_override", "default_override", "cronjob_override"} {
		overrideAttr, exists := nodePools.Attributes()[name]
		if !exists || overrideAttr == nil || overrideAttr.IsNull() {
			continue
		}
		override, ok := overrideAttr.(basetypes.ObjectValue)
		if !ok {
			continue
		}
		if override.IsUnknown() {
			return true
		}
		if spotEnabled, exists := override.Attributes()["spot_enabled"]; exists && spotEnabled != nil && !spotEnabled.IsNull() {
			return true
		}
	}

	return false
}

// hasAnyNodePoolSpotEnabled reports whether the plan carries at least one per node pool
// spot_enabled, i.e. whether the API is going to derive the global flag rather than echo it.
func (p karpenterPlanView) hasAnyNodePoolSpotEnabled() bool {
	for _, name := range []string{"stable_override", "default_override", "cronjob_override"} {
		if !p.overrideSpotEnabled(name).IsNull() {
			return true
		}
	}
	return false
}

// knownBool narrows an attribute to a bool value, returning a null bool for anything that is
// not a known, non-null bool.
func knownBool(v attr.Value) types.Bool {
	b, ok := v.(basetypes.BoolValue)
	if !ok || b.IsNull() || b.IsUnknown() {
		return types.BoolNull()
	}
	return b
}

// resolveNodePoolSpotEnabled picks the value to store in state for a per node pool
// spot_enabled. The API value wins; the planned value is the fallback for as long as the API
// does not echo the flag back (QOV-2155), so that a configured value does not read back as
// null and fail the apply.
func resolveNodePoolSpotEnabled(apiValue *bool, planned types.Bool) types.Bool {
	if apiValue != nil {
		return types.BoolValue(*apiValue)
	}
	return planned
}

func karpenterConsolidationAttrValue(consolidation *qovery.KarpenterNodePoolConsolidation) basetypes.ObjectValue {
	if consolidation == nil {
		return types.ObjectNull(karpenterConsolidationAttrTypes())
	}

	days := make([]attr.Value, len(consolidation.Days))
	for i, day := range consolidation.Days {
		days[i] = types.StringValue(string(day))
	}

	return types.ObjectValueMust(karpenterConsolidationAttrTypes(), map[string]attr.Value{
		"enabled":    types.BoolValue(consolidation.Enabled),
		"days":       types.ListValueMust(types.StringType, days),
		"start_time": types.StringValue(consolidation.StartTime),
		"duration":   types.StringValue(consolidation.Duration),
	})
}

func karpenterLimitsAttrValue(limits *qovery.KarpenterNodePoolLimits) basetypes.ObjectValue {
	if limits == nil {
		return types.ObjectNull(karpenterLimitsAttrTypes())
	}

	return types.ObjectValueMust(karpenterLimitsAttrTypes(), map[string]attr.Value{
		"enabled":                 types.BoolValue(limits.Enabled),
		"max_cpu_in_vcpu":         types.Int64Value(int64(limits.MaxCpuInVcpu)),
		"max_memory_in_gibibytes": types.Int64Value(int64(limits.MaxMemoryInGibibytes)),
	})
}

func karpenterFeatureAttrValue(karpenterParameters *qovery.ClusterFeatureKarpenterParameters, planKarpenter types.Object, mode clusterReadMode) map[string]attr.Value {
	attrVals := make(map[string]attr.Value)
	var diags diag.Diagnostics

	if karpenterParameters == nil {
		return attrVals
	}

	plan := newKarpenterPlanView(planKarpenter)

	// The global spot_enabled is derived by the API: it is recomputed on every write as the OR
	// of the per node pool values. As soon as the configuration carries per node pool values the
	// response therefore stops matching what Terraform planned, which would fail the apply with
	// "provider produced inconsistent result after apply" — keep the planned value in that case.
	// Without per node pool values the API value is authoritative and drift is reported as before.
	//
	// Reading the deprecated field is deliberate and cannot be avoided: features.karpenter
	// .spot_enabled is still a supported (deprecated) attribute, so its value has to come from
	// somewhere, and the generated getter carries the same deprecation marker. Drop the
	// suppression when the attribute itself is removed from the schema.
	//nolint:staticcheck // SA1019: the deprecated global flag is still surfaced for legacy configurations
	attrVals["spot_enabled"] = types.BoolValue(karpenterParameters.SpotEnabled)
	if plannedGlobalSpot := plan.globalSpotEnabled(); !plannedGlobalSpot.IsNull() && plan.hasAnyNodePoolSpotEnabled() {
		attrVals["spot_enabled"] = plannedGlobalSpot
	}
	attrVals["disk_size_in_gib"] = FromInt32(karpenterParameters.DiskSizeInGib)
	attrVals["default_service_architecture"] = FromString(string(karpenterParameters.DefaultServiceArchitecture))

	// Inject requirements
	nodePools := karpenterParameters.QoveryNodePools
	requirementsAttrList := make([]attr.Value, len(nodePools.Requirements))

	for i, req := range nodePools.Requirements {
		valuesAttrList := make([]attr.Value, len(req.Values))
		for j, val := range req.Values {
			valuesAttrList[j] = types.StringValue(val)
		}
		values, diags := types.ListValue(types.StringType, valuesAttrList)
		if diags.HasError() {
			return nil
		}

		reqObjectValue, diags := types.ObjectValue(karpenterRequirementAttrTypes(), map[string]attr.Value{
			"key":      types.StringValue(string(req.Key)),
			"operator": types.StringValue(string(req.Operator)),
			"values":   values,
		})
		if diags.HasError() {
			return nil
		}

		requirementsAttrList[i] = reqObjectValue
	}

	qoveryNodePoolsAttrVals := make(map[string]attr.Value)
	qoveryNodePoolsAttrVals["requirements"], diags = types.ListValue(types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()}, requirementsAttrList)
	if diags.HasError() {
		return nil
	}

	// Inject stable_override.
	// A node pool override block is stored in state when the API returns actual content for it
	// (consolidation or limits) or when the configuration declares the block. An override that
	// the API returns only because every Karpenter cluster got its per node pool spot_enabled
	// backfilled is dropped on purpose: injecting it would add a block to the state of every
	// configuration that never declared one, i.e. permanent plan noise.
	//
	// The content rule holds on the no-plan path (import, data source) too, and deliberately so.
	// The API returns a present-but-empty stable_override for Karpenter clusters, so injecting on
	// presence alone made import store a 3-attribute object where the apply path stores null, and
	// ImportStateVerify failed with `+ "…stable_override.%": "3"`. Once the spot backfill ships
	// every cluster's overrides carry spot_enabled, so keying off presence — or off spot_enabled
	// alone — would break the same way again and hand every legacy importer a spurious
	// block-removal diff on their first plan. The trade-off is accepted: importing a cluster whose
	// only divergence is a spot-only override loses that value in state until the configuration's
	// first apply re-establishes it.
	// The data source has no apply to stay symmetric with, so a spot-only override is real
	// information there and is reported rather than dropped. A fully empty block still carries
	// nothing and is dropped in both modes.
	stableOverride := nodePools.StableOverride
	stableHasContent := stableOverride != nil && (stableOverride.Consolidation != nil || stableOverride.Limits != nil)
	if mode == clusterReadModeDataSource {
		stableHasContent = stableHasContent || GetStableNodePoolSpotEnabled(stableOverride) != nil
	}
	if plan.declaresOverride("stable_override") || stableHasContent {
		var spotEnabled *bool
		var consolidation *qovery.KarpenterNodePoolConsolidation
		var limits *qovery.KarpenterNodePoolLimits
		if stableOverride != nil {
			spotEnabled = GetStableNodePoolSpotEnabled(stableOverride)
			consolidation = stableOverride.Consolidation
			limits = stableOverride.Limits
		}

		qoveryNodePoolsAttrVals["stable_override"] = types.ObjectValueMust(karpenterStableOverrideAttrTypes(), map[string]attr.Value{
			"spot_enabled":  resolveNodePoolSpotEnabled(spotEnabled, plan.overrideSpotEnabled("stable_override")),
			"consolidation": karpenterConsolidationAttrValue(consolidation),
			"limits":        karpenterLimitsAttrValue(limits),
		})
	} else {
		qoveryNodePoolsAttrVals["stable_override"] = types.ObjectNull(karpenterStableOverrideAttrTypes())
	}

	// Inject default_override — same content rule as stable_override above.
	defaultOverride := nodePools.DefaultOverride
	defaultHasContent := defaultOverride != nil && defaultOverride.Limits != nil
	if mode == clusterReadModeDataSource {
		defaultHasContent = defaultHasContent || GetDefaultNodePoolSpotEnabled(defaultOverride) != nil
	}
	if plan.declaresOverride("default_override") || defaultHasContent {
		var spotEnabled *bool
		var limits *qovery.KarpenterNodePoolLimits
		if defaultOverride != nil {
			spotEnabled = GetDefaultNodePoolSpotEnabled(defaultOverride)
			limits = defaultOverride.Limits
		}

		qoveryNodePoolsAttrVals["default_override"] = types.ObjectValueMust(karpenterDefaultOverrideAttrTypes(), map[string]attr.Value{
			"spot_enabled": resolveNodePoolSpotEnabled(spotEnabled, plan.overrideSpotEnabled("default_override")),
			"limits":       karpenterLimitsAttrValue(limits),
		})
	} else {
		qoveryNodePoolsAttrVals["default_override"] = types.ObjectNull(karpenterDefaultOverrideAttrTypes())
	}

	// Inject cronjob_override.
	// The presence of this block is what enables the dedicated cronjob node pool, so it is
	// injected only when the configuration declares it — never on the API response alone.
	cronjobOverride := nodePools.CronjobOverride
	if plan.declaresOverride("cronjob_override") || (!plan.available && cronjobOverride != nil) {
		qoveryNodePoolsAttrVals["cronjob_override"] = types.ObjectValueMust(karpenterCronjobOverrideAttrTypes(), map[string]attr.Value{
			"spot_enabled": resolveNodePoolSpotEnabled(GetCronjobNodePoolSpotEnabled(cronjobOverride), plan.overrideSpotEnabled("cronjob_override")),
		})
	} else {
		qoveryNodePoolsAttrVals["cronjob_override"] = types.ObjectNull(karpenterCronjobOverrideAttrTypes())
	}

	// Inject qovery_node_pools
	attrVals["qovery_node_pools"], diags = types.ObjectValue(karpenterNodePoolsAttrTypes(), qoveryNodePoolsAttrVals)
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
		infraChartsEksAnywhereKey: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"yaml_file_path": types.StringType,
				"git_repository": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"url":          types.StringType,
						"git_token_id": types.StringType,
						"commit_id":    types.StringType,
						"branch":       types.StringType,
						"provider":     types.StringType,
					},
				},
				infraChartsClusterBackupKey: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"enabled": types.BoolType,
						"s3": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bucket":     types.StringType,
								"region":     types.StringType,
								"role_arn":   types.StringType,
								"key_prefix": types.StringType,
							},
						},
					},
				},
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

	// Parse EKS Anywhere parameters
	if eksAnywhereAttr, ok := obj.Attributes()[infraChartsEksAnywhereKey]; ok && !eksAnywhereAttr.IsNull() {
		eksAnywhereObj := eksAnywhereAttr.(types.Object)
		eksAnywhereAttrs := eksAnywhereObj.Attributes()

		yamlFilePathAttr, ok := eksAnywhereAttrs["yaml_file_path"]
		if !ok || yamlFilePathAttr.IsNull() || yamlFilePathAttr.IsUnknown() {
			return nil, fmt.Errorf("eks_anywhere_parameters.yaml_file_path is required")
		}
		yamlFilePath := yamlFilePathAttr.(types.String).ValueString()

		gitRepositoryAttr, ok := eksAnywhereAttrs["git_repository"]
		if !ok || gitRepositoryAttr.IsNull() || gitRepositoryAttr.IsUnknown() {
			return nil, fmt.Errorf("eks_anywhere_parameters.git_repository is required")
		}
		gitRepositoryObj := gitRepositoryAttr.(types.Object)
		gitRepositoryAttrs := gitRepositoryObj.Attributes()

		urlAttr, ok := gitRepositoryAttrs["url"]
		if !ok || urlAttr.IsNull() || urlAttr.IsUnknown() {
			return nil, fmt.Errorf("eks_anywhere_parameters.git_repository.url is required")
		}
		url := urlAttr.(types.String).ValueString()

		gitTokenIDAttr, ok := gitRepositoryAttrs["git_token_id"]
		if !ok || gitTokenIDAttr.IsNull() || gitTokenIDAttr.IsUnknown() {
			return nil, fmt.Errorf("eks_anywhere_parameters.git_repository.git_token_id is required")
		}
		gitTokenID := gitTokenIDAttr.(types.String).ValueString()

		gitRepository := qovery.NewClusterEksAnywhereGitRepository(url, gitTokenID)

		if commitIDAttr, ok := gitRepositoryAttrs["commit_id"]; ok && !commitIDAttr.IsNull() && !commitIDAttr.IsUnknown() {
			commitID := commitIDAttr.(types.String).ValueString()
			gitRepository.CommitId = &commitID
		}

		if branchAttr, ok := gitRepositoryAttrs["branch"]; ok && !branchAttr.IsNull() && !branchAttr.IsUnknown() {
			branch := branchAttr.(types.String).ValueString()
			gitRepository.Branch = &branch
		}

		if providerAttr, ok := gitRepositoryAttrs["provider"]; ok && !providerAttr.IsNull() && !providerAttr.IsUnknown() {
			providerValue := providerAttr.(types.String).ValueString()
			provider, err := qovery.NewGitProviderEnumFromValue(providerValue)
			if err != nil {
				return nil, fmt.Errorf("invalid eks_anywhere_parameters.git_repository.provider: %w", err)
			}
			gitRepository.Provider = provider
		}

		eksAnywhereParams := qovery.NewClusterInfrastructureEksAnywhereParameters(*gitRepository, yamlFilePath)

		if clusterBackupAttr, ok := eksAnywhereAttrs[infraChartsClusterBackupKey]; ok && !clusterBackupAttr.IsNull() && !clusterBackupAttr.IsUnknown() {
			clusterBackupObj := clusterBackupAttr.(types.Object)
			clusterBackupAttrs := clusterBackupObj.Attributes()

			s3Attr, ok := clusterBackupAttrs["s3"]
			if !ok || s3Attr.IsNull() || s3Attr.IsUnknown() {
				return nil, fmt.Errorf("eks_anywhere_parameters.cluster_backup.s3 is required")
			}
			s3Obj := s3Attr.(types.Object)
			s3Attrs := s3Obj.Attributes()

			bucketAttr, ok := s3Attrs["bucket"]
			if !ok || bucketAttr.IsNull() || bucketAttr.IsUnknown() {
				return nil, fmt.Errorf("eks_anywhere_parameters.cluster_backup.s3.bucket is required")
			}
			regionAttr, ok := s3Attrs["region"]
			if !ok || regionAttr.IsNull() || regionAttr.IsUnknown() {
				return nil, fmt.Errorf("eks_anywhere_parameters.cluster_backup.s3.region is required")
			}
			roleArnAttr, ok := s3Attrs["role_arn"]
			if !ok || roleArnAttr.IsNull() || roleArnAttr.IsUnknown() {
				return nil, fmt.Errorf("eks_anywhere_parameters.cluster_backup.s3.role_arn is required")
			}

			clusterBackupS3 := qovery.ClusterInfrastructureEksAnywhereBackupS3Parameters{
				Bucket:  bucketAttr.(types.String).ValueString(),
				Region:  regionAttr.(types.String).ValueString(),
				RoleArn: roleArnAttr.(types.String).ValueString(),
			}
			if keyPrefixAttr, ok := s3Attrs["key_prefix"]; ok && !keyPrefixAttr.IsNull() && !keyPrefixAttr.IsUnknown() {
				keyPrefix := keyPrefixAttr.(types.String).ValueString()
				clusterBackupS3.KeyPrefix = &keyPrefix
			}

			clusterBackup := qovery.ClusterInfrastructureEksAnywhereBackupParameters{
				S3: clusterBackupS3,
			}
			if enabledAttr, ok := clusterBackupAttrs["enabled"]; ok && !enabledAttr.IsNull() && !enabledAttr.IsUnknown() {
				enabled := enabledAttr.(types.Bool).ValueBool()
				clusterBackup.Enabled = &enabled
			}

			eksAnywhereParams.ClusterBackup = &clusterBackup
		}

		params.EksAnywhereParameters = eksAnywhereParams
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

	// Convert EKS Anywhere parameters
	eksAnywhereAttrTypes := attrTypes[infraChartsEksAnywhereKey].(types.ObjectType).AttrTypes
	if params.EksAnywhereParameters != nil {
		eksAnywhere := params.EksAnywhereParameters

		gitRepositoryAttrTypes := eksAnywhereAttrTypes["git_repository"].(types.ObjectType).AttrTypes
		gitRepositoryVals := map[string]attr.Value{
			"url":          types.StringValue(eksAnywhere.GitRepository.Url),
			"git_token_id": types.StringValue(eksAnywhere.GitRepository.GitTokenId),
			"commit_id":    types.StringNull(),
			"branch":       types.StringNull(),
			"provider":     types.StringNull(),
		}
		if eksAnywhere.GitRepository.CommitId != nil {
			gitRepositoryVals["commit_id"] = types.StringValue(*eksAnywhere.GitRepository.CommitId)
		}
		if eksAnywhere.GitRepository.Branch != nil {
			gitRepositoryVals["branch"] = types.StringValue(*eksAnywhere.GitRepository.Branch)
		}
		if eksAnywhere.GitRepository.Provider != nil {
			gitRepositoryVals["provider"] = types.StringValue(string(*eksAnywhere.GitRepository.Provider))
		}

		eksAnywhereVals := map[string]attr.Value{
			"yaml_file_path": types.StringValue(eksAnywhere.YamlFilePath),
			"git_repository": types.ObjectValueMust(gitRepositoryAttrTypes, gitRepositoryVals),
			infraChartsClusterBackupKey: types.ObjectNull(
				eksAnywhereAttrTypes[infraChartsClusterBackupKey].(types.ObjectType).AttrTypes,
			),
		}

		clusterBackupAttrTypes := eksAnywhereAttrTypes[infraChartsClusterBackupKey].(types.ObjectType).AttrTypes
		clusterBackupS3AttrTypes := clusterBackupAttrTypes["s3"].(types.ObjectType).AttrTypes
		if eksAnywhere.ClusterBackup != nil {
			clusterBackupVals := map[string]attr.Value{
				"enabled": types.BoolNull(),
				"s3":      types.ObjectNull(clusterBackupS3AttrTypes),
			}

			if eksAnywhere.ClusterBackup.Enabled != nil {
				clusterBackupVals["enabled"] = types.BoolValue(*eksAnywhere.ClusterBackup.Enabled)
			}

			s3 := eksAnywhere.ClusterBackup.S3
			s3Vals := map[string]attr.Value{
				"bucket":     types.StringValue(s3.Bucket),
				"region":     types.StringValue(s3.Region),
				"role_arn":   types.StringValue(s3.RoleArn),
				"key_prefix": types.StringNull(),
			}
			if s3.KeyPrefix != nil {
				s3Vals["key_prefix"] = types.StringValue(*s3.KeyPrefix)
			}
			clusterBackupVals["s3"] = types.ObjectValueMust(clusterBackupS3AttrTypes, s3Vals)

			eksAnywhereVals[infraChartsClusterBackupKey] = types.ObjectValueMust(clusterBackupAttrTypes, clusterBackupVals)
		} else if rawClusterBackup, ok := eksAnywhere.AdditionalProperties[infraChartsClusterBackupKey]; ok {
			if clusterBackupRawMap, ok := rawClusterBackup.(map[string]interface{}); ok {
				clusterBackupVals := map[string]attr.Value{
					"enabled": types.BoolNull(),
					"s3":      types.ObjectNull(clusterBackupS3AttrTypes),
				}

				if enabledRaw, ok := clusterBackupRawMap["enabled"].(bool); ok {
					clusterBackupVals["enabled"] = types.BoolValue(enabledRaw)
				}

				if s3Raw, ok := clusterBackupRawMap["s3"].(map[string]interface{}); ok {
					s3Vals := map[string]attr.Value{
						"bucket":     types.StringNull(),
						"region":     types.StringNull(),
						"role_arn":   types.StringNull(),
						"key_prefix": types.StringNull(),
					}
					if bucketRaw, ok := s3Raw["bucket"].(string); ok {
						s3Vals["bucket"] = types.StringValue(bucketRaw)
					}
					if regionRaw, ok := s3Raw["region"].(string); ok {
						s3Vals["region"] = types.StringValue(regionRaw)
					}
					if roleArnRaw, ok := s3Raw["role_arn"].(string); ok {
						s3Vals["role_arn"] = types.StringValue(roleArnRaw)
					}
					if keyPrefixRaw, ok := s3Raw["key_prefix"].(string); ok {
						s3Vals["key_prefix"] = types.StringValue(keyPrefixRaw)
					}
					clusterBackupVals["s3"] = types.ObjectValueMust(clusterBackupS3AttrTypes, s3Vals)
				}

				eksAnywhereVals[infraChartsClusterBackupKey] = types.ObjectValueMust(clusterBackupAttrTypes, clusterBackupVals)
			}
		}

		attrVals[infraChartsEksAnywhereKey] = types.ObjectValueMust(eksAnywhereAttrTypes, eksAnywhereVals)
	} else {
		attrVals[infraChartsEksAnywhereKey] = types.ObjectNull(eksAnywhereAttrTypes)
	}

	return types.ObjectValueMust(attrTypes, attrVals)
}
