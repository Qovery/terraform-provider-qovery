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

// dataSourceKarpenterNodePoolConsolidateAfterAttribute is the consolidate_after of a node pool override.
func dataSourceKarpenterNodePoolConsolidateAfterAttribute(pool string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: karpenterNodePoolDescriptions(pool).ConsolidateAfter,
		Computed:            true,
	}
}

func (r clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	stable := karpenterNodePoolDescriptions("stable")
	defaultPool := karpenterNodePoolDescriptions("default")
	cronjob := karpenterNodePoolDescriptions("cronjob")
	gpu := karpenterNodePoolDescriptions("GPU")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("cluster"),
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription,
				Required:            true,
			},
			"credentials_id": schema.StringAttribute{
				MarkdownDescription: clusterCredentialsIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("cluster"),
				Computed:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: clusterCloudProviderDescription,
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: clusterRegionDescription,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("cluster"),
				Computed:            true,
				Optional:            true,
			},
			"kubernetes_mode": schema.StringAttribute{
				MarkdownDescription: clusterKubernetesModeDescription,
				Optional:            true,
				Computed:            true,
			},
			"production": schema.BoolAttribute{
				MarkdownDescription: clusterProductionDescription,
				Optional:            true,
				Computed:            true,
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: clusterInstanceTypeDescription + dataSourceClusterInstanceTypeNote,
				Computed:            true,
			},
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: clusterDiskSizeDescription + dataSourceClusterDiskSizeNote,
				Computed:            true,
			},
			"min_running_nodes": schema.Int64Attribute{
				MarkdownDescription: clusterMinRunningNodesDescription + dataSourceClusterNodeCountNote,
				Computed:            true,
			},
			"max_running_nodes": schema.Int64Attribute{
				MarkdownDescription: clusterMaxRunningNodesDescription + dataSourceClusterNodeCountNote,
				Computed:            true,
			},
			"features": schema.SingleNestedAttribute{
				MarkdownDescription: clusterFeaturesDescription,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_subnet": schema.StringAttribute{
						MarkdownDescription: clusterVpcSubnetDescription,
						Optional:            true,
						Computed:            true,
					},
					"static_ip": schema.BoolAttribute{
						MarkdownDescription: clusterStaticIPDescription,
						Optional:            true,
						Computed:            true,
					},
					"nat_gateways": schema.SingleNestedAttribute{
						MarkdownDescription: clusterNatGatewaysDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"static_ips_enabled": schema.BoolAttribute{
								MarkdownDescription: clusterNatGatewaysStaticIPsEnabledDescription,
								Optional:            true,
								Computed:            true,
							},
							"static_ips_count": schema.Int64Attribute{
								MarkdownDescription: clusterNatGatewaysStaticIPsCountDescription,
								Optional:            true,
								Computed:            true,
							},
						},
					},
					"existing_vpc": schema.SingleNestedAttribute{
						MarkdownDescription: clusterExistingVpcDescription,
						Optional:            true,
						Computed:            false,
						Attributes: map[string]schema.Attribute{
							"aws_vpc_eks_id": schema.StringAttribute{
								MarkdownDescription: clusterExistingVpcIDDescription,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "the EKS nodes"),
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "the EKS nodes"),
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "the EKS nodes"),
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"rds_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"rds_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"rds_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon RDS databases"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon DocumentDB"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("a", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("b", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcSubnetsDescription("c", "Amazon ElastiCache"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"eks_karpenter_fargate_subnets_zone_a_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("a"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_b_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("b"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_c_ids": schema.ListAttribute{
								MarkdownDescription: existingVpcFargateSubnetsDescription("c"),
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_create_nodes_in_private_subnet": schema.BoolAttribute{
								MarkdownDescription: clusterExistingVpcPrivateNodesDescription,
								Optional:            true,
								Computed:            true,
							},
						},
					},
					"gcp_existing_vpc": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            false,
						MarkdownDescription: clusterGcpExistingVpcDescription,
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
								MarkdownDescription: clusterGcpExistingVpcPrivateNodesDescription,
								Optional:            true,
								Computed:            true,
							},
						},
					},
					"karpenter": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            false,
						MarkdownDescription: clusterKarpenterDescription,
						Attributes: map[string]schema.Attribute{
							"disk_size_in_gib": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskSizeDescription,
								Required:            true,
								Computed:            false,
							},
							"disk_iops": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskIopsDescription,
								Computed:            true,
							},
							"disk_throughput": schema.Int64Attribute{
								MarkdownDescription: karpenterDiskThroughputDescription,
								Computed:            true,
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
										MarkdownDescription: karpenterRequirementsDescription,
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
												MarkdownDescription: stable.SpotEnabled,
												Computed:            true,
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: stable.Consolidation,
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
												MarkdownDescription: stable.Limits,
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
											"consolidate_after": dataSourceKarpenterNodePoolConsolidateAfterAttribute("stable"),
										},
									},
									"default_override": schema.SingleNestedAttribute{
										MarkdownDescription: defaultPool.Override,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: defaultPool.SpotEnabled,
												Computed:            true,
											},
											"limits": schema.SingleNestedAttribute{
												MarkdownDescription: defaultPool.Limits,
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
											"consolidate_after": dataSourceKarpenterNodePoolConsolidateAfterAttribute("default"),
										},
									},
									"cronjob_override": schema.SingleNestedAttribute{
										MarkdownDescription: cronjob.Override,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: cronjob.SpotEnabled,
												Computed:            true,
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: cronjob.Consolidation,
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
												MarkdownDescription: cronjob.Limits,
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
											"consolidate_after": dataSourceKarpenterNodePoolConsolidateAfterAttribute("cronjob"),
										},
									},
									"gpu_override": schema.SingleNestedAttribute{
										MarkdownDescription: gpu.Override,
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"requirements": schema.ListNestedAttribute{
												MarkdownDescription: karpenterGpuRequirementsDescription,
												Required:            true,
												Computed:            false,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"key": schema.StringAttribute{
															MarkdownDescription: karpenterRequirementKeyDescription,
															Required:            true,
															Computed:            false,
														},
														"operator": schema.StringAttribute{
															MarkdownDescription: karpenterRequirementOperDescription,
															Required:            true,
															Computed:            false,
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
												MarkdownDescription: gpu.DiskSize,
												Computed:            true,
											},
											"disk_iops": schema.Int64Attribute{
												MarkdownDescription: gpu.DiskIops,
												Computed:            true,
											},
											"disk_throughput": schema.Int64Attribute{
												MarkdownDescription: gpu.DiskThroughput,
												Computed:            true,
											},
											"spot_enabled": schema.BoolAttribute{
												MarkdownDescription: gpu.SpotEnabled,
												Computed:            true,
											},
											"consolidation": schema.SingleNestedAttribute{
												MarkdownDescription: gpu.Consolidation,
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
														MarkdownDescription: gpu.LimitsMaxGPU,
														Computed:            true,
													},
												},
											},
											"consolidate_after": dataSourceKarpenterNodePoolConsolidateAfterAttribute("GPU"),
										},
									},
								},
							},
						},
					},
					"gke_kms_key": schema.StringAttribute{
						MarkdownDescription: clusterGkeKmsKeyDescription,
						Computed:            true,
					},
				},
			},
			"keda": schema.SingleNestedAttribute{
				MarkdownDescription: clusterKedaDescription,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: clusterKedaEnabledDescription,
						Computed:            true,
					},
				},
			},
			"routing_table": schema.SetNestedAttribute{
				MarkdownDescription: clusterRoutingTableDescription,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: clusterRouteDescriptionDescription,
							Computed:            true,
						},
						"destination": schema.StringAttribute{
							MarkdownDescription: clusterRouteDestinationDescription,
							Computed:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: clusterRouteTargetDescription,
							Computed:            true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: clusterStateDescription,
				Computed:            true,
				Optional:            true,
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("cluster"),
				Optional:            true,
				Computed:            true,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("labels", "cluster"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: clusterKubeconfigDescription + dataSourceClusterKubeconfigNote,
				Computed:            true,
				Sensitive:           true,
			},
			"infrastructure_outputs": schema.SingleNestedAttribute{
				MarkdownDescription: clusterInfrastructureOutputsDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterNameDescription,
						Computed:            true,
					},
					"cluster_arn": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterArnDescription,
						Computed:            true,
					},
					"cluster_self_link": schema.StringAttribute{
						MarkdownDescription: clusterOutputClusterSelfLinkDescription,
						Computed:            true,
					},
					"cluster_oidc_issuer": schema.StringAttribute{
						MarkdownDescription: clusterOutputOidcIssuerDescription,
						Computed:            true,
					},
					"vpc_id": schema.StringAttribute{
						MarkdownDescription: clusterOutputVpcIDDescription,
						Computed:            true,
					},
				},
			},
			"infrastructure_charts_parameters": schema.SingleNestedAttribute{
				MarkdownDescription: clusterInfraChartsDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"nginx_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterNginxParametersDescription,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"replica_count": schema.Int64Attribute{
								MarkdownDescription: clusterNginxReplicaCountDescription,
								Computed:            true,
							},
							"default_ssl_certificate": schema.StringAttribute{
								MarkdownDescription: clusterNginxDefaultSSLCertificateDescription,
								Computed:            true,
							},
							"publish_status_address": schema.StringAttribute{
								MarkdownDescription: clusterNginxPublishStatusAddressDescription,
								Computed:            true,
							},
							"annotation_metal_lb_load_balancer_ips": schema.StringAttribute{
								MarkdownDescription: clusterNginxMetalLbLoadBalancerIPsDescription,
								Computed:            true,
							},
							"annotation_external_dns_kubernetes_target": schema.StringAttribute{
								MarkdownDescription: clusterNginxExternalDNSTargetDescription,
								Computed:            true,
							},
						},
					},
					"cert_manager_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterCertManagerParametersDescription,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"kubernetes_namespace": schema.StringAttribute{
								MarkdownDescription: clusterCertManagerNamespaceDescription,
								Computed:            true,
							},
						},
					},
					"metal_lb_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterMetalLbParametersDescription,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"ip_address_pools": schema.ListAttribute{
								MarkdownDescription: clusterMetalLbIPAddressPoolsDescription,
								ElementType:         types.StringType,
								Computed:            true,
							},
						},
					},
					"eks_anywhere_parameters": schema.SingleNestedAttribute{
						MarkdownDescription: clusterEksAnywhereParametersDescription,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"yaml_file_path": schema.StringAttribute{
								MarkdownDescription: clusterEksAnywhereYAMLFilePathDescription,
								Computed:            true,
							},
							"git_repository": schema.SingleNestedAttribute{
								MarkdownDescription: clusterEksAnywhereGitRepositoryDescription,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Computed:            true,
									},
									"git_token_id": schema.StringAttribute{
										MarkdownDescription: gitRepositoryTokenIDDescription,
										Computed:            true,
									},
									"commit_id": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereCommitIDDescription,
										Computed:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereBranchDescription,
										Computed:            true,
									},
									"provider": schema.StringAttribute{
										MarkdownDescription: clusterEksAnywhereProviderDescription,
										Computed:            true,
									},
								},
							},
							"cluster_backup": schema.SingleNestedAttribute{
								MarkdownDescription: clusterEksAnywhereBackupDescription,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										MarkdownDescription: clusterEksAnywhereBackupEnabledDescription,
										Computed:            true,
									},
									"s3": schema.SingleNestedAttribute{
										MarkdownDescription: clusterEksAnywhereBackupS3Description,
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"bucket": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupBucketDescription,
												Computed:            true,
											},
											"region": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupRegionDescription,
												Computed:            true,
											},
											"role_arn": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupRoleARNDescription,
												Computed:            true,
											},
											"key_prefix": schema.StringAttribute{
												MarkdownDescription: clusterEksAnywhereBackupKeyPrefixDescription,
												Computed:            true,
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
									MarkdownDescription: secretManagerEndpointProjectIDDescription,
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
									MarkdownDescription: secretManagerAuthRoleARNDescription,
									Optional:            true,
								},
								"region": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthRegionDescription,
									Optional:            true,
								},
								"access_key": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthAccessKeyDescription,
									Optional:            true,
								},
								"secret_key": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthSecretKeyDescription,
									Optional:            true,
									Sensitive:           true,
								},
								"json_credentials": schema.StringAttribute{
									MarkdownDescription: secretManagerAuthJSONCredentialsDescription,
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

	state := convertResponseToClusterForDataSource(ctx, cluster, data)

	// The kubeconfig has its own endpoint. It fails when the cluster has no kubeconfig yet, e.g.
	// before its first deployment, or when the token may not read it: that leaves it null
	// instead of failing a data source that is mostly read for other attributes.
	kubeconfig, apiErr := d.client.GetClusterKubeconfig(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if apiErr != nil {
		tflog.Warn(ctx, "failed to fetch the cluster kubeconfig", map[string]any{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		state.Kubeconfig = types.StringNull()
	} else {
		state.Kubeconfig = types.StringValue(kubeconfig)
	}

	tflog.Trace(ctx, "read cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
