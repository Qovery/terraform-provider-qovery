package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

// Values q-core stores for the node group of a cluster whose request omits them (ClusterRequest
// defaults in ClusterDto.kt, InstanceType.defaultInstanceType). Only a cluster whose node group
// Qovery sizes from the request uses them, see clusterNodeSizingNodeGroup.
var clusterInstanceTypeDefaults = map[string]string{
	"AWS":   "t3.xlarge",
	"SCW":   "DEV1-L",
	"AZURE": "Standard_DS2_v2",
}

const clusterDiskSizeDefault int64 = 40

// clusterNodeSizing says who sizes the nodes of a cluster, which decides what an omitted
// instance_type, disk_size, min_running_nodes or max_running_nodes means.
type clusterNodeSizing int

const (
	// clusterNodeSizingUnknown: the plan cannot tell yet, e.g. cloud_provider comes from a
	// resource that is not created yet.
	clusterNodeSizingUnknown clusterNodeSizing = iota
	// clusterNodeSizingNodeGroup: Qovery sizes a node group from the request, and an omitted
	// value means the q-core default. This is a managed cluster on AWS without Karpenter, on
	// Scaleway or on Azure.
	clusterNodeSizingNodeGroup
	// clusterNodeSizingServerSet: Karpenter, GCP Autopilot or the cluster owner sizes the nodes,
	// and Qovery ignores these values. This is every other cluster: AWS with Karpenter, GCP,
	// self-managed and partially managed.
	clusterNodeSizingServerSet
)

func clusterNodeSizingFor(cloudProvider string, kubernetesMode string, karpenterEnabled bool) clusterNodeSizing {
	if kubernetesMode != string(qovery.KUBERNETESENUM_MANAGED) {
		return clusterNodeSizingServerSet
	}
	switch cloudProvider {
	case "AWS":
		if karpenterEnabled {
			return clusterNodeSizingServerSet
		}
		return clusterNodeSizingNodeGroup
	case "SCW", "AZURE":
		return clusterNodeSizingNodeGroup
	default:
		return clusterNodeSizingServerSet
	}
}

// plannedClusterNodeSizing classifies the planned cluster, and returns its cloud provider.
func plannedClusterNodeSizing(ctx context.Context, plan tfsdk.Plan) (clusterNodeSizing, string, diag.Diagnostics) {
	var cloudProvider, kubernetesMode types.String
	var features types.Object
	var diags diag.Diagnostics
	diags.Append(plan.GetAttribute(ctx, path.Root("cloud_provider"), &cloudProvider)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("kubernetes_mode"), &kubernetesMode)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("features"), &features)...)
	if diags.HasError() || cloudProvider.IsUnknown() || kubernetesMode.IsUnknown() {
		return clusterNodeSizingUnknown, "", diags
	}

	mode := clusterKubernetesModeDefault
	if !kubernetesMode.IsNull() {
		mode = kubernetesMode.ValueString()
	}

	karpenterEnabled := false
	if features.IsUnknown() {
		return clusterNodeSizingUnknown, "", diags
	}
	if !features.IsNull() {
		karpenter, ok := features.Attributes()[featureKeyKarpenter]
		if ok && karpenter.IsUnknown() {
			return clusterNodeSizingUnknown, "", diags
		}
		karpenterEnabled = ok && !karpenter.IsNull()
	}

	return clusterNodeSizingFor(cloudProvider.ValueString(), mode, karpenterEnabled), cloudProvider.ValueString(), diags
}

// clusterNodeSizingDefaultModifier plans instance_type, disk_size, min_running_nodes and
// max_running_nodes when the configuration omits them:
//   - on a cluster whose node group Qovery sizes from the request, the q-core default, so that
//     removing the attribute plans the reset and a value changed outside Terraform shows up in
//     the plan;
//   - on any other cluster, the recorded value: Qovery derives these values there and the API
//     reports sentinels, so the value is server-set. A configured value has no effect on such a
//     cluster, and the plan warns about it.
type clusterNodeSizingDefaultModifier struct {
	// defaultValue returns the value q-core stores when the request omits the attribute, for a
	// cloud provider whose node group Qovery sizes from the request.
	defaultValue func(cloudProvider string) attr.Value
	// description names the defaults, for the documentation.
	description string
}

func (m clusterNodeSizingDefaultModifier) Description(_ context.Context) string {
	return m.description
}

func (m clusterNodeSizingDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// planValue returns the value to plan, and false to keep the plan as it is.
func (m clusterNodeSizingDefaultModifier) planValue(ctx context.Context, attributePath path.Path, config attr.Value, state attr.Value, stateRaw tfsdk.State, plan tfsdk.Plan, diags *diag.Diagnostics) (attr.Value, bool) {
	// Destroy: nothing is planned.
	if plan.Raw.IsNull() {
		return nil, false
	}

	sizing, cloudProvider, sizingDiags := plannedClusterNodeSizing(ctx, plan)
	diags.Append(sizingDiags...)
	if sizingDiags.HasError() {
		return nil, false
	}

	if !config.IsNull() {
		if sizing == clusterNodeSizingServerSet {
			diags.AddAttributeWarning(
				attributePath,
				fmt.Sprintf("%s has no effect on this cluster", attributePath),
				fmt.Sprintf("Karpenter, GCP Autopilot or the cluster owner sizes the nodes of Karpenter, GCP, self-managed and partially managed clusters, so Qovery ignores %s there. "+
					"Remove %s from the configuration.", attributePath, attributePath),
			)
		}
		return nil, false
	}

	switch sizing {
	case clusterNodeSizingNodeGroup:
		return m.defaultValue(cloudProvider), true
	case clusterNodeSizingServerSet:
		// Server-set: keep what the state records, including null on a partially managed
		// cluster. On a create the value stays unknown until the API reports it.
		if !stateRaw.Raw.IsNull() {
			return state, true
		}
	}
	return nil, false
}

func (m clusterNodeSizingDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if value, ok := m.planValue(ctx, req.Path, req.ConfigValue, req.StateValue, req.State, req.Plan, &resp.Diagnostics); ok {
		resp.PlanValue = value.(types.String)
	}
}

func (m clusterNodeSizingDefaultModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if value, ok := m.planValue(ctx, req.Path, req.ConfigValue, req.StateValue, req.State, req.Plan, &resp.Diagnostics); ok {
		resp.PlanValue = value.(types.Int64)
	}
}

// ClusterInstanceTypeDefault returns the plan modifier for qovery_cluster.instance_type.
func ClusterInstanceTypeDefault() planmodifier.String {
	return clusterNodeSizingDefaultModifier{
		defaultValue: func(cloudProvider string) attr.Value {
			return types.StringValue(clusterInstanceTypeDefaults[cloudProvider])
		},
		description: fmt.Sprintf("When the instance type is omitted on a cluster whose node group Qovery sizes, plans the Qovery default: %s on AWS, %s on Scaleway, %s on Azure.",
			clusterInstanceTypeDefaults["AWS"], clusterInstanceTypeDefaults["SCW"], clusterInstanceTypeDefaults["AZURE"]),
	}
}

// ClusterNodeSizingInt64Default returns the plan modifier for the disk size and node counts of
// qovery_cluster, whose q-core default does not depend on the cloud provider.
func ClusterNodeSizingInt64Default(defaultValue int64) planmodifier.Int64 {
	return clusterNodeSizingDefaultModifier{
		defaultValue: func(string) attr.Value {
			return types.Int64Value(defaultValue)
		},
		description: fmt.Sprintf("When the value is omitted on a cluster whose node group Qovery sizes, plans the Qovery default: %d.", defaultValue),
	}
}

// rejectForbiddenClusterFeatureChanges fails the plan of an update that makes a features change
// the Qovery API rejects on an existing cluster. Omitting features, or one of its attributes,
// plans the default value, so the plan names what to declare instead of letting the apply fail.
func rejectForbiddenClusterFeatureChanges(ctx context.Context, state tfsdk.State, plan tfsdk.Plan, diags *diag.Diagnostics) {
	// Creation or destroy: there is no existing cluster, or nothing is planned.
	if state.Raw.IsNull() || plan.Raw.IsNull() {
		return
	}

	var stateFeatures, planFeatures types.Object
	var cloudProvider, clusterState types.String
	diags.Append(state.GetAttribute(ctx, path.Root("features"), &stateFeatures)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("features"), &planFeatures)...)
	diags.Append(plan.GetAttribute(ctx, path.Root("cloud_provider"), &cloudProvider)...)
	diags.Append(state.GetAttribute(ctx, path.Root("state"), &clusterState)...)
	if diags.HasError() || stateFeatures.IsNull() || stateFeatures.IsUnknown() || planFeatures.IsUnknown() {
		return
	}

	rejectKarpenterDisable(stateFeatures, planFeatures, diags)
	rejectStaticIPDisable(stateFeatures, planFeatures, cloudProvider, clusterState, diags)
	rejectGkeKmsKeyChange(stateFeatures, planFeatures, diags)
}

// featureValue returns the named attribute of a features object, or nil when the object is null.
func featureValue(features types.Object, name string) attr.Value {
	if features.IsNull() || features.IsUnknown() {
		return nil
	}
	return features.Attributes()[name]
}

// rejectKarpenterDisable rejects a plan that removes features.karpenter from a cluster that runs
// Karpenter: q-core cannot disable Karpenter on an existing cluster (KarpenterDomain.kt).
func rejectKarpenterDisable(stateFeatures, planFeatures types.Object, diags *diag.Diagnostics) {
	stateKarpenter := featureValue(stateFeatures, featureKeyKarpenter)
	planKarpenter := featureValue(planFeatures, featureKeyKarpenter)
	if stateKarpenter == nil || stateKarpenter.IsNull() || stateKarpenter.IsUnknown() {
		return
	}
	if planKarpenter != nil && !planKarpenter.IsNull() {
		return
	}

	diags.AddAttributeError(
		path.Root("features").AtName(featureKeyKarpenter),
		"Cannot disable Karpenter",
		"Karpenter is enabled on this cluster, and the Qovery API cannot disable it on an existing cluster. "+
			"Declare the features.karpenter block in the configuration to keep this cluster: the Terraform state holds its current values.",
	)
}

// clusterStaticIPLockedCloudProviders are the cloud providers on which q-core rejects any change
// of static_ip once the cluster has been deployed (isStaticIpUpdateForbiddenOnDeployedCluster).
var clusterStaticIPLockedCloudProviders = map[string]bool{"AWS": true, "GCP": true, "AZURE": true}

// rejectStaticIPDisable rejects a plan that turns features.static_ip from true to false on a
// cluster where q-core rejects it: a deployed AWS, GCP or Azure cluster. Omitting static_ip plans
// false, so this names the value to declare. A cluster still in the READY state has never been
// deployed and accepts the change.
func rejectStaticIPDisable(stateFeatures, planFeatures types.Object, cloudProvider, clusterState types.String, diags *diag.Diagnostics) {
	stateStaticIP, ok := featureValue(stateFeatures, featureKeyStaticIP).(types.Bool)
	if !ok || !stateStaticIP.ValueBool() {
		return
	}
	planStaticIP, ok := featureValue(planFeatures, featureKeyStaticIP).(types.Bool)
	if ok && (planStaticIP.IsUnknown() || planStaticIP.ValueBool()) {
		return
	}
	if cloudProvider.IsUnknown() || !clusterStaticIPLockedCloudProviders[cloudProvider.ValueString()] {
		return
	}
	if clusterState.ValueString() == string(qovery.STATEENUM_READY) {
		return
	}

	diags.AddAttributeError(
		path.Root("features").AtName(featureKeyStaticIP),
		"Cannot disable static_ip on a deployed cluster",
		"This plan changes features.static_ip from true to false, and the Qovery API rejects enabling or disabling static IPs once an AWS, GCP or Azure cluster has been deployed. "+
			"Omitting features or static_ip plans false: declare `static_ip = true` in the features block to keep this cluster. "+
			"To disable static IPs, recreate the cluster explicitly, e.g. `terraform destroy -target=<cluster resource>` followed by `terraform apply`.",
	)
}

// rejectGkeKmsKeyChange rejects a plan that sets, changes or removes features.gke_kms_key after
// creation: q-core only takes the key when it creates the cluster. Like RejectExistingVpcChange,
// it fails the plan instead of replacing the cluster.
func rejectGkeKmsKeyChange(stateFeatures, planFeatures types.Object, diags *diag.Diagnostics) {
	stateKey, _ := featureValue(stateFeatures, featureKeyGkeKmsKey).(types.String)
	planKey, _ := featureValue(planFeatures, featureKeyGkeKmsKey).(types.String)
	if planKey.IsUnknown() {
		return
	}
	if stringIsNullOrEmpty(stateKey) && stringIsNullOrEmpty(planKey) {
		return
	}
	if planKey.Equal(stateKey) {
		return
	}

	keep := "remove features.gke_kms_key from the configuration, as the Terraform state holds none"
	if !stringIsNullOrEmpty(stateKey) {
		keep = fmt.Sprintf("set features.gke_kms_key to the value in the Terraform state (%q)", stateKey.ValueString())
	}
	diags.AddAttributeError(
		path.Root("features").AtName(featureKeyGkeKmsKey),
		"Cannot change features.gke_kms_key after creation",
		"The Qovery API sets the GKE KMS key only when it creates the cluster. "+
			"To keep this cluster, "+keep+". "+
			"To use a different key, recreate the cluster explicitly, e.g. `terraform destroy -target=<cluster resource>` followed by `terraform apply`. "+
			"Note that `terraform apply -replace=...` cannot be used here because the plan is rejected before the replacement is applied.",
	)
}
