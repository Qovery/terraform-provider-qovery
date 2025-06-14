package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/qovery/validators"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &clusterDataSource{}

type clusterDataSource struct {
	client *client.Client
}

func newClusterDataSource() datasource.DataSource {
	return &clusterDataSource{}
}

func (d clusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *clusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = provider.client
}

func (r clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the cluster.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"credentials_id": schema.StringAttribute{
				Description: "Id of the credentials.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Computed:    true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				Computed: true,
			},
			"region": schema.StringAttribute{
				Description: "Region of the cluster.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				Computed: true,
				Optional: true,
			},
			"kubernetes_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				Optional: true,
				Computed: true,
			},
			"production": schema.BoolAttribute{
				Description: "Specific flag to indicate that this cluster is a production one.",
				Optional:    true,
				Computed:    true,
			},
			"instance_type": schema.StringAttribute{
				Description: "Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`, and not set for Karpenter-enabled clusters",
				Optional:    true,
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				Optional: true,
				Computed: true,
			},
			"max_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters]",
					clusterMaxRunningNodesMin,
					&clusterMaxRunningNodesDefault,
				),
				Optional: true,
				Computed: true,
			},
			"features": schema.SingleNestedAttribute{
				Description: "Features of the cluster.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"vpc_subnet": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureVpcSubnetDefault,
						),
						Optional: true,
						Computed: true,
					},
					"static_ip": schema.BoolAttribute{
						Description: descriptions.NewBoolDefaultDescription(
							"Static IP (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureStaticIPDefault,
						),
						Optional: true,
						Computed: true,
					},
					"existing_vpc": schema.SingleNestedAttribute{
						Description: "Network configuration if you want to install qovery on an existing VPC",
						Optional:    true,
						Computed:    false,
						Attributes: map[string]schema.Attribute{
							"aws_vpc_eks_id": schema.StringAttribute{
								Description: "Aws VPC id",
								Required:    true,
								Computed:    false,
							},
							"eks_subnets_zone_a_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS zone a. Must have map_public_ip_on_launch set to true",
								ElementType: types.StringType,
								Required:    true,
								Computed:    false,
							},
							"eks_subnets_zone_b_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS zone b. Must have map_public_ip_on_launch set to true",
								ElementType: types.StringType,
								Required:    true,
								Computed:    false,
							},
							"eks_subnets_zone_c_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS zone c. Must have map_public_ip_on_launch set to true",
								ElementType: types.StringType,
								Required:    true,
								Computed:    false,
							},
							"rds_subnets_zone_a_ids": schema.ListAttribute{
								Description: "Ids of the subnets for RDS",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"rds_subnets_zone_b_ids": schema.ListAttribute{
								Description: "Ids of the subnets for RDS",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"rds_subnets_zone_c_ids": schema.ListAttribute{
								Description: "Ids of the subnets for RDS",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"documentdb_subnets_zone_a_ids": schema.ListAttribute{
								Description: "Ids of the subnets for document db",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"documentdb_subnets_zone_b_ids": schema.ListAttribute{
								Description: "Ids of the subnets for document db",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"documentdb_subnets_zone_c_ids": schema.ListAttribute{
								Description: "Ids of the subnets for document db",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"elasticache_subnets_zone_a_ids": schema.ListAttribute{
								Description: "Ids of the subnets for elasticache",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"elasticache_subnets_zone_b_ids": schema.ListAttribute{
								Description: "Ids of the subnets for elasticache",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"elasticache_subnets_zone_c_ids": schema.ListAttribute{
								Description: "Ids of the subnets for elasticache",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    true,
							},
							"eks_karpenter_fargate_subnets_zone_a_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS fargate zone a. Must have to be private and connected to internet through a NAT Gateway",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    false,
							},
							"eks_karpenter_fargate_subnets_zone_b_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS fargate zone b. Must have to be private and connected to internet through a NAT Gateway",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    false,
							},
							"eks_karpenter_fargate_subnets_zone_c_ids": schema.ListAttribute{
								Description: "Ids of the subnets for EKS fargate zone c. Must have to be private and connected to internet through a NAT Gateway",
								ElementType: types.StringType,
								Optional:    true,
								Computed:    false,
							}},
					},
					"karpenter": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    false,
						Description: "Karpenter parameters if you want to use Karpenter on an EKS cluster",
						Attributes: map[string]schema.Attribute{
							"spot_enabled": schema.BoolAttribute{
								Description: "Enable spot instances",
								Required:    true,
								Computed:    false,
							},
							"disk_size_in_gib": schema.Int64Attribute{
								Required: true,
								Computed: false,
							},
							"default_service_architecture": schema.StringAttribute{
								Description: "The default architecture of service",
								Required:    true,
								Computed:    false,
							},
							"qovery_node_pools": schema.SingleNestedAttribute{
								Description: "Karpenter node pool configuration",
								Required:    true,
								Computed:    false,
								Attributes: map[string]schema.Attribute{
									"requirements": schema.ListNestedAttribute{
										Description: "List of requirements for the node pool",
										Required:    true,
										Computed:    false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description: "The key of the requirement (e.g., InstanceFamily, InstanceSize, Arch)",
													Required:    true,
													Computed:    false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"InstanceFamily", "InstanceSize", "Arch"}),
													},
												},
												"operator": schema.StringAttribute{
													Description: "The operator for the requirement (e.g., In)",
													Required:    true,
													Computed:    false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"In"}),
													},
												},
												"values": schema.ListAttribute{
													Description: "List of values for the requirement",
													Required:    true,
													Computed:    false,
													ElementType: types.StringType,
												},
											},
										},
									},
									"stable_override": schema.SingleNestedAttribute{
										Description: "Defines some overridden options for Qovery stable node pool",
										Optional:    true,
										Computed:    false,
										Attributes: map[string]schema.Attribute{
											"consolidation": schema.SingleNestedAttribute{
												Description: "Specifies the period to consolidate nodes (by default, no consolidation happens)",
												Optional:    true,
												Computed:    false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description: "Indicates if the consolidation defines here must be applied",
														Required:    true,
														Computed:    false,
													},
													"days": schema.ListAttribute{
														Description: "A list of days where the consolidation must be triggered",
														Required:    true,
														Computed:    false,
														ElementType: types.StringType,
													},
													"start_time": schema.StringAttribute{
														Description: "The start time where the consolidation must begin. It must follow the ISO-8601 time format: `PThh:mm`",
														Required:    true,
														Computed:    false,
													},
													"duration": schema.StringAttribute{
														Description: "The period during the consolidation will be applied. It must follow the ISO-8601 duration format: `PThhHmmM`",
														Required:    true,
														Computed:    false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												Description: "Specifies the limits to apply on the stable node pool",
												Optional:    true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description: "Enabled the limit",
														Required:    true,
														Computed:    false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description: "The maximum number of cpu cores to be used inside the stable node pool",
														Required:    true,
														Computed:    false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description: "The maximum number of memory to be used inside the stable node pool",
														Required:    true,
														Computed:    false,
													},
												},
											},
										},
									},
									"default_override": schema.SingleNestedAttribute{
										Description: "Defines some overridden options for Qovery default node pool",
										Optional:    true,
										Computed:    false,
										Attributes: map[string]schema.Attribute{
											"limits": schema.SingleNestedAttribute{
												Description: "Specifies the limits to apply on the default node pool",
												Optional:    true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description: "Enabled the limit",
														Required:    true,
														Computed:    false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description: "The maximum number of cpu cores to be used inside the default node pool",
														Required:    true,
														Computed:    false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description: "The maximum number of memory to be used inside the default node pool",
														Required:    true,
														Computed:    false,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"routing_table": schema.SetNestedAttribute{
				Description: "List of routes of the cluster.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description: "Description of the route.",
							Computed:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination of the route.",
							Computed:    true,
						},
						"target": schema.StringAttribute{
							Description: "Target of the route.",
							Computed:    true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"State of the cluster.",
					clusterStates,
					&clusterStateDefault,
				),
				Computed: true,
				Optional: true,
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings of the cluster.",
				Optional:    true,
				Computed:    true,
			},
			"infrastructure_outputs": schema.SingleNestedAttribute{
				Description: "Outputs related to the underlying Kubernetes infrastructure. These values are only available once the cluster is deployed.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						Description: "The name of the Kubernetes cluster. Available after deployment for all providers.",
						Computed:    true,
					},
					"cluster_arn": schema.StringAttribute{
						Description: "The ARN of the AWS cluster. Only available for AWS after deployment.",
						Computed:    true,
					},
					"cluster_self_link": schema.StringAttribute{
						Description: "The self-link of the GCP cluster. Only available for GCP after deployment.",
						Computed:    true,
					},
					"cluster_oidc_issuer": schema.StringAttribute{
						Description: "The OIDC issuer URL for the cluster. Available for AWS and Azure after deployment.",
						Computed:    true,
					},
					"vpc_id": schema.StringAttribute{
						Description: "The VPC ID used by the cluster. Only available for AWS after deployment.",
						Computed:    true,
					},
				},
			},
		},
	}
}

// Read qovery cluster data source
func (d clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from the API
	cluster, apiErr := d.client.GetCluster(ctx, data.OrganizationId.ValueString(), data.Id.ValueString(), data.AdvancedSettingsJson.ValueString(), true)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToCluster(ctx, cluster, data)
	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
