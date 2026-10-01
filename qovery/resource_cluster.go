package qovery

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure      = &clusterResource{}
	_ resource.ResourceWithImportState    = clusterResource{}
	_ resource.ResourceWithValidateConfig = clusterResource{}
	_ resource.ResourceWithModifyPlan     = clusterResource{}
	_ resource.ResourceWithUpgradeState   = clusterResource{}
)

var (
	// Cluster State
	clusterStates = clientEnumToStringArray([]qovery.StateEnum{
		qovery.STATEENUM_DEPLOYED,
		qovery.STATEENUM_STOPPED,
		qovery.STATEENUM_READY,
	})
	clusterStateDefault = string(qovery.STATEENUM_DEPLOYED)

	// Cluster Description
	clusterDescriptionDefault = ""

	// Cloud Provider
	cloudProviders = clientEnumToStringArray(qovery.AllowedCloudProviderEnumEnumValues)

	// Cluster Min Running Nodes
	clusterMinRunningNodesDefault int64 = 3

	// Cluster Max Running Nodes
	clusterMaxRunningNodesDefault int64 = 10

	// Cluster Feature VPC_SUBNET
	clusterFeatureVpcSubnetDefault = "10.0.0.0/16"

	// Cluster Feature STATIC_IP
	clusterFeatureStaticIPDefault = false

	// Cluster Kubernetes Mode
	clusterKubernetesModes = clientEnumToStringArray([]qovery.KubernetesEnum{
		qovery.KUBERNETESENUM_MANAGED,
		qovery.KUBERNETESENUM_SELF_MANAGED,
		qovery.KUBERNETESENUM_PARTIALLY_MANAGED,
	})
	clusterKubernetesModeDefault = string(qovery.KUBERNETESENUM_MANAGED)
)

type clusterResource struct {
	client                         *client.Client
	clusterAdvancedSettingsService *advanced_settings.ClusterAdvancedSettingsService
}

func newClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = provider.client
	r.clusterAdvancedSettingsService = provider.clusterAdvancedSettingsService
}

// ModifyPlan warns at plan time about advanced_settings_json keys that are not recognized
// cluster advanced settings, instead of letting them silently no-op, and about Karpenter node
// pool changes that are easy to read past in a plan. It fails the plan of a features change the
// Qovery API rejects on an existing cluster.
func (r clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	warnUnknownClusterAdvancedSettings(ctx, r.clusterAdvancedSettingsService, req.Config, &resp.Diagnostics)
	warnKarpenterSpotToOnDemand(ctx, req.State, req.Plan, &resp.Diagnostics)
	warnKarpenterNodePoolRemoval(ctx, req.State, req.Plan, &resp.Diagnostics)
	rejectForbiddenClusterFeatureChanges(ctx, req.State, req.Plan, &resp.Diagnostics)
}

var karpenterPath = path.Root("features").AtName("karpenter")

// karpenterStateAndPlanViews returns the Karpenter views of the prior state and of the plan of an
// update. ok is false on a create or a destroy, or when either side has no Karpenter object.
func karpenterStateAndPlanViews(ctx context.Context, state tfsdk.State, plan tfsdk.Plan) (stateView, planView karpenterPlanView, ok bool) {
	if state.Raw.IsNull() || plan.Raw.IsNull() {
		return karpenterPlanView{}, karpenterPlanView{}, false
	}

	var stateKarpenter, planKarpenter types.Object
	if state.GetAttribute(ctx, karpenterPath, &stateKarpenter).HasError() || plan.GetAttribute(ctx, karpenterPath, &planKarpenter).HasError() {
		return karpenterPlanView{}, karpenterPlanView{}, false
	}

	stateView = newKarpenterPlanView(stateKarpenter)
	planView = newKarpenterPlanView(planKarpenter)
	return stateView, planView, stateView.available && planView.available
}

// warnKarpenterSpotToOnDemand warns when the plan moves a Karpenter node pool from spot to
// on-demand instances. A configuration upgraded from 0.x without pinning its spot pools applies
// exactly that change, and the plan only shows it as an override block being removed, which is
// easy to read past.
func warnKarpenterSpotToOnDemand(ctx context.Context, state tfsdk.State, plan tfsdk.Plan, diags *diag.Diagnostics) {
	stateView, planView, ok := karpenterStateAndPlanViews(ctx, state, plan)
	if !ok {
		return
	}

	for _, name := range karpenterNodePoolOverrideNames {
		// Removing cronjob_override or gpu_override removes the node pool altogether: that is not
		// a move to on-demand instances.
		if (name == "cronjob_override" || name == "gpu_override") && !planView.declaresOverride(name) {
			continue
		}

		wasSpot, stateKnown := stateView.spotEnabled(name)
		isSpot, planKnown := planView.spotEnabled(name)
		if !stateKnown || !planKnown || !wasSpot || isSpot {
			continue
		}

		diags.AddAttributeWarning(
			karpenterPath.AtName("qovery_node_pools").AtName(name),
			"Karpenter node pool moves to on-demand instances",
			fmt.Sprintf("The %s node pool runs on EC2 Spot instances, and this plan moves it to on-demand instances because the configuration does not set `%s.spot_enabled = true`. "+
				"A node pool without spot_enabled runs on on-demand instances. Set `spot_enabled = true` in `%s` to keep spot instances.",
				strings.TrimSuffix(name, "_override"), name, name),
		)
	}
}

// karpenterNodePoolRemovalWarnings lists the node pools that exist only while their override is
// declared, with the warning of a plan that removes the override.
var karpenterNodePoolRemovalWarnings = []struct {
	override string
	summary  string
	detail   string
}{
	{
		override: "cronjob_override",
		summary:  "Karpenter cronjob node pool will be disabled",
		detail: "This plan removes `cronjob_override`, which disables the cronjob node pool. " +
			"If the cronjob node pool was enabled outside Terraform, for example from the Qovery Console, declare `cronjob_override` in the configuration to keep it.",
	},
	{
		override: "gpu_override",
		summary:  "Karpenter GPU node pool will be deleted",
		detail: "This plan removes `gpu_override`, which deletes the GPU node pool and the nodes running on it. " +
			"If the GPU node pool was created outside Terraform, for example from the Qovery Console, declare `gpu_override` in the configuration to keep it.",
	},
}

// warnKarpenterNodePoolRemoval warns when the plan removes cronjob_override or gpu_override,
// which removes the node pool. A pool created from the Console reaches the state on refresh, so a
// configuration that never declared one sees exactly this change, shown only as a block being
// removed.
func warnKarpenterNodePoolRemoval(ctx context.Context, state tfsdk.State, plan tfsdk.Plan, diags *diag.Diagnostics) {
	stateView, planView, ok := karpenterStateAndPlanViews(ctx, state, plan)
	if !ok {
		return
	}

	for _, pool := range karpenterNodePoolRemovalWarnings {
		if !stateView.declaresOverride(pool.override) || !planView.overrideKnownAbsent(pool.override) {
			continue
		}
		diags.AddAttributeWarning(karpenterPath.AtName("qovery_node_pools").AtName(pool.override), pool.summary, pool.detail)
	}
}

// karpenterNodePoolConsolidateAfterAttribute is the consolidate_after of a node pool override.
func karpenterNodePoolConsolidateAfterAttribute(pool string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: karpenterNodePoolDescriptions(pool).ConsolidateAfter + karpenterConsolidateAfterNote,
		Optional:            true,
		Computed:            false,
		Validators: []validator.String{
			validators.NewConsolidateAfterValidator(),
		},
	}
}

func (r clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stable := karpenterNodePoolDescriptions("stable")
	defaultPool := karpenterNodePoolDescriptions("default")
	cronjob := karpenterNodePoolDescriptions("cronjob")
	gpu := karpenterNodePoolDescriptions("GPU")

	// TODO (framework-migration): test if Default is OK when modifying the attribute, otherwise we'll need to use a modifier
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Qovery cluster: the Kubernetes cluster Qovery deploys services to, on AWS (EKS), GCP (GKE), Scaleway (Kapsule), Azure (AKS) or EKS Anywhere.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("cluster"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credentials_id": schema.StringAttribute{
				MarkdownDescription: clusterCredentialsIDDescription,
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("cluster"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("cluster"),
				Required:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(clusterCloudProviderDescription, cloudProviders, nil),
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(cloudProviders),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: clusterRegionDescription,
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("cluster"),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clusterDescriptionDefault),
			},
			"kubernetes_mode": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(clusterKubernetesModeDescription, clusterKubernetesModeDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clusterKubernetesModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterKubernetesModes),
				},
			},
			"production": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterProductionDescription, false),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: clusterInstanceTypeDescription + clusterNodeSizingNote + clusterInstanceTypeDefaultNote,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					ClusterInstanceTypeDefault(),
				},
			},
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(clusterDiskSizeDescription+clusterNodeSizingNote, clusterDiskSizeDefault),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					ClusterNodeSizingInt64Default(clusterDiskSizeDefault),
				},
			},
			"min_running_nodes": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(clusterMinRunningNodesDescription+clusterNodeSizingNote, clusterMinRunningNodesDefault),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					ClusterNodeSizingInt64Default(clusterMinRunningNodesDefault),
				},
			},
			"max_running_nodes": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(clusterMaxRunningNodesDescription+clusterNodeSizingNote, clusterMaxRunningNodesDefault),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					ClusterNodeSizingInt64Default(clusterMaxRunningNodesDefault),
				},
			},
			"features": schema.SingleNestedAttribute{
				MarkdownDescription: clusterFeaturesDescription,
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(clusterFeaturesDefault()),
				Attributes: map[string]schema.Attribute{
					"vpc_subnet": schema.StringAttribute{
						MarkdownDescription: descriptions.NewStringDefaultDescription(clusterVpcSubnetDescription+recreatesOnChange("cluster"), clusterFeatureVpcSubnetDefault),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(clusterFeatureVpcSubnetDefault),
						PlanModifiers: []planmodifier.String{
							// Treat a legacy state value of "" as the default so a provider
							// upgrade doesn't manufacture a phantom replacement.
							RequiresReplaceIfKnownChangeTreatingEmptyAs(clusterFeatureVpcSubnetDefault),
						},
					},
					"static_ip": schema.BoolAttribute{
						MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterStaticIPDescription+clusterStaticIPNote, clusterFeatureStaticIPDefault),
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(clusterFeatureStaticIPDefault),
					},
					"nat_gateways": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: clusterNatGatewaysDescription,
						Default:             objectdefault.StaticValue(clusterNatGatewaysDefault()),
						Attributes: map[string]schema.Attribute{
							"static_ips_enabled": schema.BoolAttribute{
								MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterNatGatewaysStaticIPsEnabledDescription+clusterNatGatewaysStaticIPsEnabledNote, false),
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
							"static_ips_count": schema.Int64Attribute{
								MarkdownDescription: descriptions.NewInt64MinDescription(clusterNatGatewaysStaticIPsCountDescription, 1, new(int64(1))),
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								Validators: []validator.Int64{
									validators.Int64MinValidator{Min: 1},
								},
							},
						},
					},
					"existing_vpc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: false,
						PlanModifiers: []planmodifier.Object{
							RejectExistingVpcChange(),
						},
						MarkdownDescription: clusterExistingVpcDescription + existingVpcImmutableNote,
						Attributes: map[string]schema.Attribute{
							"aws_vpc_eks_id": schema.StringAttribute{
								MarkdownDescription: clusterExistingVpcIDDescription,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "the EKS nodes") + clusterExistingVpcEksSubnetsNote,
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "the EKS nodes") + clusterExistingVpcEksSubnetsNote,
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "the EKS nodes") + clusterExistingVpcEksSubnetsNote,
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"rds_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"rds_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"rds_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"documentdb_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"documentdb_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"documentdb_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"elasticache_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"elasticache_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"elasticache_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
							},
							"eks_karpenter_fargate_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("a") + clusterExistingVpcFargateSubnetsNote,
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("b") + clusterExistingVpcFargateSubnetsNote,
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("c") + clusterExistingVpcFargateSubnetsNote,
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_create_nodes_in_private_subnet": schema.BoolAttribute{
								MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterExistingVpcPrivateNodesDescription, false),
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
						},
					},
					"gcp_existing_vpc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: false,
						PlanModifiers: []planmodifier.Object{
							RejectExistingVpcChange(),
						},
						MarkdownDescription: clusterGcpExistingVpcDescription + existingVpcImmutableNote,
						Attributes: map[string]schema.Attribute{
							"vpc_name": schema.StringAttribute{
								MarkdownDescription: clusterGcpExistingVpcNameDescription,
								Required:            true,
							},
							"vpc_project_id": schema.StringAttribute{
								MarkdownDescription: clusterGcpExistingVpcProjectIDDescription,
								Optional:            true,
							},
							"subnetwork_name": schema.StringAttribute{
								MarkdownDescription: clusterGcpExistingVpcSubnetworkDescription,
								Optional:            true,
							},
							"ip_range_services_name": schema.StringAttribute{
								MarkdownDescription: clusterGcpExistingVpcServicesRangeDescription,
								Optional:            true,
							},
							"ip_range_pods_name": schema.StringAttribute{
								MarkdownDescription: clusterGcpExistingVpcPodsRangeDescription,
								Optional:            true,
							},
							"additional_ip_range_pods_names": schema.ListAttribute{
								MarkdownDescription: clusterGcpExistingVpcExtraPodsRangesDescription,
								ElementType:         types.StringType,
								Optional:            true,
							},
							"private_nodes": schema.BoolAttribute{
								MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterGcpExistingVpcPrivateNodesDescription, false),
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
						},
					},
					"karpenter": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            false,
						MarkdownDescription: clusterKarpenterDescription + clusterKarpenterNote,
						Attributes: map[string]schema.Attribute{
							"disk_size_in_gib": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskSizeDescription + karpenterDiskSizeMinNote,
								Required:            true,
								Computed:            false,
							},
							"disk_iops": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskIopsDescription,
								Optional:            true,
								Computed:            false,
							},
							"disk_throughput": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskThroughputDescription,
								Optional:            true,
								Computed:            false,
							},
							"default_service_architecture": schema.StringAttribute{
								MarkdownDescription: karpenterDefaultServiceArchitectureDescription,
								Required:            true,
								Computed:            false,
							},
							"qovery_node_pools": schema.SingleNestedAttribute{
								MarkdownDescription: karpenterNodePoolsDescription,
								Required:            true,
								Computed:            false,
								Attributes: map[string]schema.Attribute{
									"requirements": schema.ListNestedAttribute{
										MarkdownDescription: karpenterRequirementsDescription + karpenterRequirementsNote,
										Required:            true,
										Computed:            false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													MarkdownDescription: karpenterRequirementKeyDescription,
													Required:            true,
													Computed:            false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"InstanceFamily", "InstanceSize", "Arch"}),
													},
												},
												"operator": schema.StringAttribute{
													MarkdownDescription: karpenterRequirementOperDescription,
													Required:            true,
													Computed:            false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"In"}),
													},
												},
												"values": schema.ListAttribute{
													MarkdownDescription: karpenterRequirementValueDescription,
													Required:            true,
													Computed:            false,
													ElementType:         types.StringType,
												},
											},
										},
									},
									"stable_override": schema.SingleNestedAttribute{
										MarkdownDescription: stable.Override,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: descriptions.NewBoolDefaultDescription(stable.SpotEnabled, false),
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(false),
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: stable.Consolidation + karpenterConsolidationOmittedNote,
												Optional:            true,
												Computed:            false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: stable.ConsolidationEnabled,
														Required:            true,
														Computed:            false,
													},
													"days": schema.ListAttribute{
														MarkdownDescription: stable.ConsolidationDays,
														Required:            true,
														Computed:            false,
														ElementType:         types.StringType,
													},
													"start_time": schema.StringAttribute{
														MarkdownDescription: stable.ConsolidationStartTime,
														Required:            true,
														Computed:            false,
													},
													"duration": schema.StringAttribute{
														MarkdownDescription: stable.ConsolidationDuration,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												MarkdownDescription: stable.Limits + karpenterNodePoolLimitsMinNote,
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: stable.LimitsEnabled,
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														MarkdownDescription: stable.LimitsMaxCPU,
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														MarkdownDescription: stable.LimitsMaxMemory,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"consolidate_after": karpenterNodePoolConsolidateAfterAttribute("stable"),
										},
									},
									"default_override": schema.SingleNestedAttribute{
										MarkdownDescription: defaultPool.Override,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: descriptions.NewBoolDefaultDescription(defaultPool.SpotEnabled, false),
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(false),
											},
											"limits": schema.SingleNestedAttribute{
												MarkdownDescription: defaultPool.Limits + karpenterNodePoolLimitsMinNote,
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: defaultPool.LimitsEnabled,
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														MarkdownDescription: defaultPool.LimitsMaxCPU,
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														MarkdownDescription: defaultPool.LimitsMaxMemory,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"consolidate_after": karpenterNodePoolConsolidateAfterAttribute("default"),
										},
									},
									"cronjob_override": schema.SingleNestedAttribute{
										MarkdownDescription: cronjob.Override + karpenterCronjobNodePoolNote,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: descriptions.NewBoolDefaultDescription(cronjob.SpotEnabled, false),
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(false),
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: cronjob.Consolidation + karpenterConsolidationOmittedNote,
												Optional:            true,
												Computed:            false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: cronjob.ConsolidationEnabled,
														Required:            true,
														Computed:            false,
													},
													"days": schema.ListAttribute{
														MarkdownDescription: cronjob.ConsolidationDays,
														Required:            true,
														Computed:            false,
														ElementType:         types.StringType,
													},
													"start_time": schema.StringAttribute{
														MarkdownDescription: cronjob.ConsolidationStartTime,
														Required:            true,
														Computed:            false,
													},
													"duration": schema.StringAttribute{
														MarkdownDescription: cronjob.ConsolidationDuration,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												MarkdownDescription: cronjob.Limits + karpenterNodePoolLimitsMinNote,
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: cronjob.LimitsEnabled,
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														MarkdownDescription: cronjob.LimitsMaxCPU,
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														MarkdownDescription: cronjob.LimitsMaxMemory,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"consolidate_after": karpenterNodePoolConsolidateAfterAttribute("cronjob"),
										},
									},
									"gpu_override": schema.SingleNestedAttribute{
										MarkdownDescription: gpu.Override + karpenterGpuNodePoolNote,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"requirements": schema.ListNestedAttribute{
												MarkdownDescription: karpenterGpuRequirementsDescription + karpenterRequirementsNote,
												Required:            true,
												Computed:            false,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"key": schema.StringAttribute{
															MarkdownDescription: karpenterRequirementKeyDescription,
															Required:            true,
															Computed:            false,
															Validators: []validator.String{
																validators.NewStringEnumValidator([]string{"InstanceFamily", "InstanceSize", "Arch"}),
															},
														},
														"operator": schema.StringAttribute{
															MarkdownDescription: karpenterRequirementOperDescription,
															Required:            true,
															Computed:            false,
															Validators: []validator.String{
																validators.NewStringEnumValidator([]string{"In"}),
															},
														},
														"values": schema.ListAttribute{
															MarkdownDescription: karpenterRequirementValueDescription,
															Required:            true,
															Computed:            false,
															ElementType:         types.StringType,
														},
													},
												},
											},
											"disk_size_in_gib": schema.Int64Attribute{
												MarkdownDescription: gpu.DiskSize + karpenterDiskSizeMinNote,
												Required:            true,
												Computed:            false,
											},
											"disk_iops": schema.Int64Attribute{
												MarkdownDescription: gpu.DiskIops,
												Optional:            true,
												Computed:            false,
											},
											"disk_throughput": schema.Int64Attribute{
												MarkdownDescription: gpu.DiskThroughput,
												Optional:            true,
												Computed:            false,
											},
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: descriptions.NewBoolDefaultDescription(gpu.SpotEnabled, false),
												Optional:            true,
												Computed:            true,
												Default:             booldefault.StaticBool(false),
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: gpu.Consolidation + karpenterConsolidationOmittedNote,
												Optional:            true,
												Computed:            false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: gpu.ConsolidationEnabled,
														Required:            true,
														Computed:            false,
													},
													"days": schema.ListAttribute{
														MarkdownDescription: gpu.ConsolidationDays,
														Required:            true,
														Computed:            false,
														ElementType:         types.StringType,
													},
													"start_time": schema.StringAttribute{
														MarkdownDescription: gpu.ConsolidationStartTime,
														Required:            true,
														Computed:            false,
													},
													"duration": schema.StringAttribute{
														MarkdownDescription: gpu.ConsolidationDuration,
														Required:            true,
														Computed:            false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												MarkdownDescription: gpu.Limits,
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														MarkdownDescription: gpu.LimitsEnabled,
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														MarkdownDescription: gpu.LimitsMaxCPU,
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														MarkdownDescription: gpu.LimitsMaxMemory,
														Required:            true,
														Computed:            false,
													},
													"max_gpu": schema.Int64Attribute{
														MarkdownDescription: descriptions.NewInt64DefaultDescription(gpu.LimitsMaxGPU, 0),
														Optional:            true,
														Computed:            true,
														Default:             int64default.StaticInt64(0),
													},
												},
											},
											"consolidate_after": karpenterNodePoolConsolidateAfterAttribute("GPU"),
										},
									},
								},
							},
						},
					},
					"gke_kms_key": schema.StringAttribute{
						MarkdownDescription: clusterGkeKmsKeyDescription + clusterGkeKmsKeyNote,
						Optional:            true,
					},
				},
			},
			"keda": schema.SingleNestedAttribute{
				MarkdownDescription: clusterKedaDescription + clusterKedaNote,
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(clusterKedaDefault()),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: descriptions.NewBoolDefaultDescription(clusterKedaEnabledDescription, false),
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"routing_table": schema.SetNestedAttribute{
				MarkdownDescription: clusterRoutingTableDescription + clusterRoutingTableNote,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: clusterRouteDescriptionDescription,
							Required:            true,
						},
						"destination": schema.StringAttribute{
							MarkdownDescription: clusterRouteDestinationDescription,
							Required:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: clusterRouteTargetDescription,
							Required:            true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(clusterStateDescription+clusterStateValuesNote, clusterStateDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clusterStateDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterStates),
				},
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: advancedSettingsJSONDescription("Clusters/operation/getDefaultClusterAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("labels", "the cluster's resources") + clusterLabelsGroupIDsNote,
				Optional:            true,
				ElementType:         types.StringType,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: clusterKubeconfigDescription + clusterPartiallyManagedRequiredNote,
				Optional:            true,
				Sensitive:           true,
			},
			"infrastructure_outputs": schema.SingleNestedAttribute{
				MarkdownDescription: clusterInfrastructureOutputsDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterNameDescription,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_arn": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterArnDescription,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_self_link": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterSelfLinkDescription,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_oidc_issuer": schema.StringAttribute{
						MarkdownDescription: clusterOutputOidcIssuerDescription,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vpc_id": schema.StringAttribute{
						MarkdownDescription: clusterOutputVpcIDDescription,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"infrastructure_charts_parameters": schema.SingleNestedAttribute{
				MarkdownDescription: clusterInfraChartsDescription + clusterPartiallyManagedRequiredNote,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"nginx_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterNginxParametersDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"replica_count": schema.Int64Attribute{
								MarkdownDescription: clusterNginxReplicaCountDescription,
								Optional:            true,
							},
							"default_ssl_certificate": schema.StringAttribute{
								MarkdownDescription: clusterNginxDefaultSSLCertificateDescription,
								Optional:            true,
							},
							"publish_status_address": schema.StringAttribute{
								MarkdownDescription: clusterNginxPublishStatusAddressDescription,
								Optional:            true,
							},
							"annotation_metal_lb_load_balancer_ips": schema.StringAttribute{
								MarkdownDescription: clusterNginxMetalLbLoadBalancerIPsDescription + clusterNginxMetalLbLoadBalancerIPsNote,
								Optional:            true,
							},
							"annotation_external_dns_kubernetes_target": schema.StringAttribute{
								MarkdownDescription: clusterNginxExternalDNSTargetDescription,
								Optional:            true,
							},
						},
					},
					"cert_manager_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterCertManagerParametersDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"kubernetes_namespace": schema.StringAttribute{
								MarkdownDescription: clusterCertManagerNamespaceDescription,
								Optional:            true,
							},
						},
					},
					"metal_lb_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterMetalLbParametersDescription + clusterPartiallyManagedRequiredNote,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ip_address_pools": schema.ListAttribute{
								MarkdownDescription: clusterMetalLbIPAddressPoolsDescription,
								ElementType:         types.StringType,
								Required:            true,
							},
						},
					},
					"eks_anywhere_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterEksAnywhereParametersDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"yaml_file_path": schema.StringAttribute{
								MarkdownDescription: clusterEksAnywhereYAMLFilePathDescription,
								Required:            true,
							},
							"git_repository": schema.SingleNestedAttribute{
								MarkdownDescription: clusterEksAnywhereGitRepositoryDescription,
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Required:            true,
									},
									"git_token_id": schema.StringAttribute{
										MarkdownDescription: gitRepositoryTokenIDDescription,
										Required:            true,
									},
									"commit_id": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereCommitIDDescription,
										Optional:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereBranchDescription,
										Optional:            true,
									},
									"provider": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereProviderDescription,
										Optional:            true,
										Validators: []validator.String{
											validators.NewStringEnumValidator([]string{"BITBUCKET", "GITHUB", "GITLAB"}),
										},
									},
								},
							},
							"cluster_backup": schema.SingleNestedAttribute{
								MarkdownDescription: clusterEksAnywhereBackupDescription,
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										MarkdownDescription: clusterEksAnywhereBackupEnabledDescription,
										Optional:            true,
									},
									"s3": schema.SingleNestedAttribute{
										MarkdownDescription: clusterEksAnywhereBackupS3Description,
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"bucket": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupBucketDescription,
												Required:            true,
											},
											"region": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupRegionDescription,
												Required:            true,
											},
											"role_arn": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupRoleARNDescription,
												Required:            true,
											},
											"key_prefix": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupKeyPrefixDescription,
												Optional:            true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"secret_manager_accesses": schema.SetNestedAttribute{
				MarkdownDescription: clusterSecretManagerAccessesDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretManagerAccessIDDescription,
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: secretManagerAccessNameDescription,
							Required:            true,
						},
						"endpoint": schema.SingleNestedAttribute{
							MarkdownDescription: secretManagerEndpointDescription,
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: secretManagerEndpointTypeDescription,
									Required:            true,
									Validators: []validator.String{
										validators.NewStringEnumValidator([]string{"AWS_PARAMETER_STORE", "AWS_SECRET_MANAGER", "GCP_SECRET_MANAGER"}),
									},
								},
								"region": schema.StringAttribute{
									MarkdownDescription: secretManagerEndpointRegionDescription,
									Required:            true,
								},
								"project_id": schema.StringAttribute{
									MarkdownDescription: secretManagerEndpointProjectIDDescription + requiredWhenType("GCP_SECRET_MANAGER"),
									Optional:            true,
								},
							},
						},
						"authentication": schema.SingleNestedAttribute{
							MarkdownDescription: secretManagerAuthenticationDescription,
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthTypeDescription,
									Required:            true,
									Validators: []validator.String{
										validators.NewStringEnumValidator([]string{"AUTOMATICALLY_CONFIGURED", "AWS_ROLE_ARN", "AWS_STATIC_CREDENTIALS", "GCP_JSON_CREDENTIALS"}),
									},
								},
								"role_arn": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthRoleARNDescription + requiredWhenType("AWS_ROLE_ARN"),
									Optional:            true,
								},
								"region": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthRegionDescription + requiredWhenType("AWS_STATIC_CREDENTIALS"),
									Optional:            true,
								},
								"access_key": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthAccessKeyDescription + requiredWhenType("AWS_STATIC_CREDENTIALS"),
									Optional:            true,
								},
								"secret_key": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthSecretKeyDescription + requiredWhenType("AWS_STATIC_CREDENTIALS"),
									Optional:            true,
									Sensitive:           true,
								},
								"json_credentials": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthJSONCredentialsDescription + requiredWhenType("GCP_JSON_CREDENTIALS"),
									Optional:            true,
									Sensitive:           true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create qovery cluster resource
func (r clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new cluster
	request, err := plan.toUpsertClusterRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	cluster, apiErr := r.client.CreateCluster(ctx, plan.OrganizationId.ValueString(), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// For PARTIALLY_MANAGED clusters, set the kubeconfig
	if plan.KubernetesMode.ValueString() == "PARTIALLY_MANAGED" && !plan.Kubeconfig.IsNull() && plan.Kubeconfig.ValueString() != "" {
		apiErr = r.client.SetClusterKubeconfig(ctx, plan.OrganizationId.ValueString(), cluster.ClusterResponse.Id, plan.Kubeconfig.ValueString())
		if apiErr != nil {
			resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
			return
		}
	}

	// Initialize state values
	state := convertResponseToCluster(ctx, cluster, plan)

	// For PARTIALLY_MANAGED clusters, fetch the kubeconfig from API to ensure state matches
	if plan.KubernetesMode.ValueString() == "PARTIALLY_MANAGED" {
		kubeconfig, apiErr := r.client.GetClusterKubeconfig(ctx, plan.OrganizationId.ValueString(), cluster.ClusterResponse.Id)
		if apiErr != nil {
			tflog.Warn(ctx, "failed to fetch kubeconfig after create", map[string]any{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "created cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery cluster resource
func (r clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hack to know if this method is triggered through an import
	// CredentialsId is always present except when importing the resource
	isTriggeredFromImport := false
	if state.CredentialsId.IsNull() {
		isTriggeredFromImport = true
	}

	// Get cluster from the API
	cluster, apiErr := r.client.GetCluster(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if handleReadNotFound(ctx, resp, apiErr) {
		return
	}

	state = convertResponseToCluster(ctx, cluster, state)

	// For PARTIALLY_MANAGED clusters, fetch the kubeconfig
	if cluster.ClusterResponse.Kubernetes != nil && *cluster.ClusterResponse.Kubernetes == qovery.KUBERNETESENUM_PARTIALLY_MANAGED {
		kubeconfig, apiErr := r.client.GetClusterKubeconfig(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
		if apiErr != nil {
			// Log warning but don't fail - kubeconfig might not be set yet
			tflog.Warn(ctx, "failed to fetch kubeconfig for PARTIALLY_MANAGED cluster", map[string]any{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "read cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery cluster resource
func (r clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Cluster
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update cluster in the backend
	request, err := plan.toUpsertClusterRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	cluster, apiErr := r.client.UpdateCluster(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// For PARTIALLY_MANAGED clusters, update the kubeconfig if changed
	if plan.KubernetesMode.ValueString() == "PARTIALLY_MANAGED" && !plan.Kubeconfig.IsNull() && plan.Kubeconfig.ValueString() != "" {
		// Only update if kubeconfig has changed
		if state.Kubeconfig.IsNull() || plan.Kubeconfig.ValueString() != state.Kubeconfig.ValueString() {
			apiErr = r.client.SetClusterKubeconfig(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), plan.Kubeconfig.ValueString())
			if apiErr != nil {
				resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
				return
			}
		}
	}

	// Update state values
	state = convertResponseToCluster(ctx, cluster, plan)

	// For PARTIALLY_MANAGED clusters, fetch the kubeconfig from API to ensure state matches
	if plan.KubernetesMode.ValueString() == "PARTIALLY_MANAGED" {
		kubeconfig, apiErr := r.client.GetClusterKubeconfig(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
		if apiErr != nil {
			tflog.Warn(ctx, "failed to fetch kubeconfig after update", map[string]any{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "updated cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery cluster resource
func (r clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete cluster
	apiErr := r.client.DeleteCluster(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Remove cluster from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery cluster resource using its id
func (r clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,cluster_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}

// UpgradeState migrates cluster states written by 0.x (schema version 0).
func (r clusterResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Version 0 has the same attribute types as the current schema; attributes removed since
	// then are skipped when the framework decodes the prior state.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	priorSchema := schemaResp.Schema

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &priorSchema,
			StateUpgrader: upgradeClusterStateV0ToV1,
		},
	}
}

// upgradeClusterStateV0ToV1 turns the empty values that 0.x stored for attributes the
// configuration omitted into null: routing_table, the database and cache subnet lists of
// features.existing_vpc, and the "" features.gke_kms_key of a GCP cluster without a key. They
// were Optional + Computed in 0.x and are Optional only since 1.0, so keeping the empty value
// would plan a change to null on every cluster that omits them.
func upgradeClusterStateV0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var cluster Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !cluster.RoutingTables.IsNull() && len(cluster.RoutingTables.Elements()) == 0 {
		cluster.RoutingTables = types.SetNull(types.ObjectType{AttrTypes: clusterRouteAttrTypes})
	}
	cluster.Features = upgradeGkeKmsKeyFrom0x(upgradeExistingVpcSubnetListsFrom0x(cluster.Features))

	resp.Diagnostics.Append(resp.State.Set(ctx, cluster)...)
}

// upgradeGkeKmsKeyFrom0x turns the "" features.gke_kms_key that 0.x stored for a GCP cluster
// without a KMS key into null, which the read now reports for it.
func upgradeGkeKmsKeyFrom0x(features types.Object) types.Object {
	gkeKmsKey, ok := featureValue(features, featureKeyGkeKmsKey).(types.String)
	if !ok || gkeKmsKey.IsNull() || gkeKmsKey.IsUnknown() || gkeKmsKey.ValueString() != "" {
		return features
	}

	featuresAttributes := maps.Clone(features.Attributes())
	featuresAttributes[featureKeyGkeKmsKey] = types.StringNull()
	return types.ObjectValueMust(createFeaturesAttrTypes(), featuresAttributes)
}

// existingVpcOptionalSubnetLists are the database and cache subnet lists of
// features.existing_vpc, which the API reports as [] when none is set.
var existingVpcOptionalSubnetLists = []string{
	"rds_subnets_zone_a_ids", "rds_subnets_zone_b_ids", "rds_subnets_zone_c_ids",
	"documentdb_subnets_zone_a_ids", "documentdb_subnets_zone_b_ids", "documentdb_subnets_zone_c_ids",
	"elasticache_subnets_zone_a_ids", "elasticache_subnets_zone_b_ids", "elasticache_subnets_zone_c_ids",
}

// upgradeExistingVpcSubnetListsFrom0x turns the empty existing_vpc subnet lists of a 0.x state
// into null. The read keeps the shape of the prior value for an empty list, so without it the
// [] would stay in state and plan a change forever.
func upgradeExistingVpcSubnetListsFrom0x(features types.Object) types.Object {
	existingVpc, ok := featureValue(features, featureKeyExistingVpc).(types.Object)
	if !ok || existingVpc.IsNull() || existingVpc.IsUnknown() {
		return features
	}

	existingVpcAttributes := maps.Clone(existingVpc.Attributes())
	changed := false
	for _, name := range existingVpcOptionalSubnetLists {
		if list, ok := existingVpcAttributes[name].(types.List); ok && !list.IsNull() && !list.IsUnknown() && len(list.Elements()) == 0 {
			existingVpcAttributes[name] = types.ListNull(types.StringType)
			changed = true
		}
	}
	if !changed {
		return features
	}

	featuresAttributes := maps.Clone(features.Attributes())
	featuresAttributes[featureKeyExistingVpc] = types.ObjectValueMust(createExistingVpcFeatureAttrTypes(), existingVpcAttributes)
	return types.ObjectValueMust(createFeaturesAttrTypes(), featuresAttributes)
}

// ValidateConfig performs plan-time cross-attribute validation for the cluster resource.
func (r clusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateNatGatewaysConfig(config.CloudProvider, config.Features)...)
	resp.Diagnostics.Append(validateGkeKmsKeyConfig(config.CloudProvider, config.Features)...)
}

// validateGkeKmsKeyConfig validates that gke_kms_key is only set on GCP clusters.
func validateGkeKmsKeyConfig(cloudProvider types.String, features types.Object) diag.Diagnostics {
	var diags diag.Diagnostics

	if features.IsNull() || features.IsUnknown() {
		return diags
	}

	gkeKmsKeyAttr, ok := features.Attributes()[featureKeyGkeKmsKey]
	if !ok || gkeKmsKeyAttr.IsNull() || gkeKmsKeyAttr.IsUnknown() {
		return diags
	}

	if gkeKmsKeyAttr.(types.String).ValueString() == "" {
		return diags
	}

	if cloudProvider.IsNull() || cloudProvider.IsUnknown() {
		return diags
	}

	if cloudProvider.ValueString() != "GCP" {
		diags.AddAttributeError(
			path.Root("features").AtName(featureKeyGkeKmsKey),
			"Invalid gke_kms_key",
			"features.gke_kms_key is only supported for GCP clusters.",
		)
	}

	return diags
}

// validateNatGatewaysConfig encapsulates the cross-attribute nat_gateways validation
// rules so they can be unit-tested without constructing a full tfsdk.Config.
//
// v3 rules (enabled-driven):
//   - Rule A (error): static_ips_enabled=true OR static_ips_count>1, on a non-GCP cluster →
//     "features.nat_gateways is only supported for GCP clusters."
//   - Rule B (error): static_ips_enabled=true AND static_ip is null-or-known-false →
//     "features.nat_gateways.static_ips_enabled requires features.static_ip to be true."
//     Skipped when static_ip is unknown (plan-time variable, benefit of the doubt).
//   - Rule B2 (warning): static_ips_count>1 AND static_ips_enabled is known-false →
//     "static_ips_count has no effect while static_ips_enabled is false."
//   - Rule C (warning): block explicitly non-null on non-GCP → ignored-on-this-provider.
//
// Checks are skipped when cloudProvider or the relevant nested values are unknown
// (variable-driven, not yet resolved at plan time).
func validateNatGatewaysConfig(cloudProvider types.String, features types.Object) diag.Diagnostics {
	var diags diag.Diagnostics

	// Skip when features itself is null or unknown.
	if features.IsNull() || features.IsUnknown() {
		return diags
	}

	// Extract nat_gateways from features.
	natGatewaysAttr, hasNatGateways := features.Attributes()[featureKeyNatGateways]
	if !hasNatGateways {
		return diags
	}

	// nat_gateways is null → omitted in config, defaults apply, nothing to validate.
	if natGatewaysAttr.IsNull() {
		return diags
	}

	// nat_gateways is unknown → can't validate yet.
	if natGatewaysAttr.IsUnknown() {
		return diags
	}

	natGatewaysObj := natGatewaysAttr.(types.Object)
	natAttrs := natGatewaysObj.Attributes()

	// Extract static_ips_enabled (may be null/unknown if not resolved yet).
	enabledKnownTrue := false
	enabledKnownFalse := false
	if ev, ok := natAttrs["static_ips_enabled"]; ok && !ev.IsNull() && !ev.IsUnknown() {
		v := ev.(types.Bool).ValueBool()
		enabledKnownTrue = v
		enabledKnownFalse = !v
	}

	// Extract static_ips_count (may be null/unknown).
	var count int64
	countKnown := false
	if cv, ok := natAttrs["static_ips_count"]; ok && !cv.IsNull() && !cv.IsUnknown() {
		count = cv.(types.Int64).ValueInt64()
		countKnown = true
	}

	// Rule C: block explicitly non-null on non-GCP → warning (always, regardless of values).
	if !cloudProvider.IsNull() && !cloudProvider.IsUnknown() && cloudProvider.ValueString() != "GCP" {
		diags.AddAttributeWarning(
			path.Root("features").AtName(featureKeyNatGateways),
			"nat_gateways ignored on non-GCP cluster",
			"features.nat_gateways is ignored on non-GCP clusters; only the default value is accepted.",
		)
	}

	// Rule A (error): enabled=true OR count>1 on non-GCP → error.
	if !cloudProvider.IsNull() && !cloudProvider.IsUnknown() && cloudProvider.ValueString() != "GCP" {
		if enabledKnownTrue || (countKnown && count > 1) {
			diags.AddAttributeError(
				path.Root("features").AtName(featureKeyNatGateways),
				"Invalid nat_gateways",
				"features.nat_gateways is only supported for GCP clusters.",
			)
			return diags
		}
		return diags
	}

	// From here on, cloud_provider is either unknown or GCP.
	// Rules B and B2 are only validated when cloud_provider is known to be GCP — when
	// unknown, we give the benefit of the doubt (variable-driven at plan time).
	if cloudProvider.IsNull() || cloudProvider.IsUnknown() {
		return diags
	}

	// Rule B (error): enabled=true AND static_ip is null-or-known-false → error.
	// Skip when static_ip is unknown (variable-driven at plan time).
	if enabledKnownTrue {
		staticIPAttr, hasStaticIP := features.Attributes()[featureKeyStaticIP]
		if !hasStaticIP || !staticIPAttr.IsUnknown() {
			// Treat an absent key or a null value as the config-default false.
			staticIPFalse := !hasStaticIP || staticIPAttr.IsNull() || !staticIPAttr.(types.Bool).ValueBool()
			if staticIPFalse {
				diags.AddAttributeError(
					path.Root("features").AtName(featureKeyNatGateways),
					"Invalid nat_gateways",
					"features.nat_gateways.static_ips_enabled requires features.static_ip to be true.",
				)
			}
		}
	}

	// Rule B2 (warning): count>1 AND enabled is known false → count has no effect.
	if countKnown && count > 1 && enabledKnownFalse {
		diags.AddAttributeWarning(
			path.Root("features").AtName(featureKeyNatGateways),
			"static_ips_count has no effect",
			"static_ips_count has no effect while static_ips_enabled is false.",
		)
	}

	return diags
}
