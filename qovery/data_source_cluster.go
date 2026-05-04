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
		Description:         "Use this data source to retrieve information about an existing Qovery cluster.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Qovery cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the cluster.",
				MarkdownDescription: "ID of the cluster to retrieve.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "ID of the organization containing the cluster.",
				Required:            true,
			},
			"credentials_id": schema.StringAttribute{
				Description:         "Id of the credentials.",
				MarkdownDescription: "ID of the cloud provider credentials associated with this cluster.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the cluster.",
				MarkdownDescription: "Name of the cluster.",
				Computed:            true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				MarkdownDescription: "Cloud provider of the cluster (`AWS`, `GCP`, `SCW`, `AZURE`, or `ON_PREMISE`).",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				Description:         "Region of the cluster.",
				MarkdownDescription: "Cloud provider region where the cluster is deployed.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				MarkdownDescription: "Description of the cluster.",
				Computed:            true,
				Optional:            true,
			},
			"kubernetes_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				MarkdownDescription: "Kubernetes management mode (`MANAGED`, `SELF_MANAGED`, or `PARTIALLY_MANAGED`).",
				Optional:            true,
				Computed:            true,
			},
			"production": schema.BoolAttribute{
				Description:         "Specific flag to indicate that this cluster is a production one.",
				MarkdownDescription: "Whether this cluster is flagged as a production cluster.",
				Optional:            true,
				Computed:            true,
			},
			"instance_type": schema.StringAttribute{
				Description:         "Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`, and not set for Karpenter-enabled clusters",
				MarkdownDescription: "Instance type of the cluster nodes (e.g., `t3a.xlarge` for AWS, `DEV1-L` for Scaleway, `AUTO_PILOT` for GCP).",
				Optional:            true,
				Computed:            true,
			},
			"disk_size": schema.Int64Attribute{
				Description:         "Disk size of the cluster nodes in GB.",
				MarkdownDescription: "Disk size of the cluster nodes in GB.",
				Optional:            true,
				Computed:            true,
			},
			"min_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				MarkdownDescription: "Minimum number of nodes for the cluster autoscaler.",
				Optional:            true,
				Computed:            true,
			},
			"max_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters]",
					clusterMaxRunningNodesMin,
					&clusterMaxRunningNodesDefault,
				),
				MarkdownDescription: "Maximum number of nodes for the cluster autoscaler.",
				Optional:            true,
				Computed:            true,
			},
			"features": schema.SingleNestedAttribute{
				Description:         "Features of the cluster.",
				MarkdownDescription: "Cluster features configuration including VPC settings, static IPs, existing VPC, and Karpenter.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"vpc_subnet": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureVpcSubnetDefault,
						),
						MarkdownDescription: "Custom VPC CIDR block (AWS only). Immutable after creation.",
						Optional:            true,
						Computed:            true,
					},
					"static_ip": schema.BoolAttribute{
						Description: descriptions.NewBoolDefaultDescription(
							"Static IP (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureStaticIPDefault,
						),
						MarkdownDescription: "Whether static/elastic IPs are enabled (AWS only). Immutable after creation.",
						Optional:            true,
						Computed:            true,
					},
					"existing_vpc": schema.SingleNestedAttribute{
						Description:         "Network configuration if you want to install qovery on an existing VPC",
						MarkdownDescription: "AWS existing VPC configuration, if the cluster is deployed on an existing VPC.",
						Optional:            true,
						Computed:            false,
						Attributes: map[string]schema.Attribute{
							"aws_vpc_eks_id": schema.StringAttribute{
								Description:         "Aws VPC id",
								MarkdownDescription: "The ID of the existing AWS VPC.",
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone a. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "Subnet IDs in availability zone A for EKS worker nodes.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone b. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "Subnet IDs in availability zone B for EKS worker nodes.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"eks_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone c. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "Subnet IDs in availability zone C for EKS worker nodes.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:            false,
							},
							"rds_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "Subnet IDs in availability zone A for Amazon RDS.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"rds_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "Subnet IDs in availability zone B for Amazon RDS.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"rds_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "Subnet IDs in availability zone C for Amazon RDS.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "Subnet IDs in availability zone A for Amazon DocumentDB.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "Subnet IDs in availability zone B for Amazon DocumentDB.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"documentdb_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "Subnet IDs in availability zone C for Amazon DocumentDB.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "Subnet IDs in availability zone A for Amazon ElastiCache.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "Subnet IDs in availability zone B for Amazon ElastiCache.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"elasticache_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "Subnet IDs in availability zone C for Amazon ElastiCache.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
							"eks_karpenter_fargate_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone a. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "Private subnet IDs in availability zone A for EKS Fargate (Karpenter).",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone b. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "Private subnet IDs in availability zone B for EKS Fargate (Karpenter).",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_karpenter_fargate_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone c. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "Private subnet IDs in availability zone C for EKS Fargate (Karpenter).",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            false,
							},
							"eks_create_nodes_in_private_subnet": schema.BoolAttribute{
								Description:         "Specifies whether to create EKS nodes in private subnets",
								MarkdownDescription: "Whether EKS worker nodes are created in private subnets.",
								Optional:            true,
								Computed:            true,
							},
						},
					},
					"gcp_existing_vpc": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            false,
						Description:         "Network configuration if you want to install qovery on an existing GCP VPC",
						MarkdownDescription: "GCP existing VPC configuration, if the cluster is deployed on an existing GCP VPC.",
						Attributes: map[string]schema.Attribute{
							"vpc_name": schema.StringAttribute{
								Description:         "Name of the existing GCP VPC network",
								MarkdownDescription: "Name of the existing GCP VPC network.",
								Required:            true,
							},
							"vpc_project_id": schema.StringAttribute{
								Description:         "GCP project ID that owns the VPC. Defaults to the project associated with your GCP credentials",
								MarkdownDescription: "GCP project ID that owns the VPC.",
								Optional:            true,
							},
							"subnetwork_name": schema.StringAttribute{
								Description:         "Name of the GCP subnetwork within the VPC",
								MarkdownDescription: "Name of the GCP subnetwork within the VPC.",
								Optional:            true,
							},
							"ip_range_services_name": schema.StringAttribute{
								Description:         "Name of the secondary IP range for GKE services",
								MarkdownDescription: "Name of the secondary IP range for GKE services.",
								Optional:            true,
							},
							"ip_range_pods_name": schema.StringAttribute{
								Description:         "Name of the secondary IP range for pods",
								MarkdownDescription: "Name of the secondary IP range for pods.",
								Optional:            true,
							},
							"additional_ip_range_pods_names": schema.ListAttribute{
								Description:         "Additional secondary IP range names for pods",
								MarkdownDescription: "Additional secondary IP range names for pods.",
								ElementType:         types.StringType,
								Optional:            true,
							},
						},
					},
					"karpenter": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            false,
						Description:         "Karpenter parameters if you want to use Karpenter on an EKS cluster",
						MarkdownDescription: "Karpenter configuration for AWS EKS clusters.",
						Attributes: map[string]schema.Attribute{
							"spot_enabled": schema.BoolAttribute{
								Description:         "Enable spot instances",
								MarkdownDescription: "Whether EC2 Spot instances are enabled.",
								Required:            true,
								Computed:            false,
							},
							"disk_size_in_gib": schema.Int64Attribute{
								Description:         "Disk size in GiB for Karpenter-provisioned nodes.",
								MarkdownDescription: "Root disk size in GiB for Karpenter-provisioned nodes.",
								Required:            true,
								Computed:            false,
							},
							"default_service_architecture": schema.StringAttribute{
								Description:         "The default architecture of service",
								MarkdownDescription: "Default CPU architecture for services (`AMD64` or `ARM64`).",
								Required:            true,
								Computed:            false,
							},
							"qovery_node_pools": schema.SingleNestedAttribute{
								Description:         "Karpenter node pool configuration",
								MarkdownDescription: "Karpenter node pool configuration with requirements and resource limits.",
								Required:            true,
								Computed:            false,
								Attributes: map[string]schema.Attribute{
									"requirements": schema.ListNestedAttribute{
										Description:         "List of requirements for the node pool",
										MarkdownDescription: "Node selection requirements for the Karpenter node pool.",
										Required:            true,
										Computed:            false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description:         "The key of the requirement (e.g., InstanceFamily, InstanceSize, Arch)",
													MarkdownDescription: "Requirement key (`InstanceFamily`, `InstanceSize`, or `Arch`).",
													Required:            true,
													Computed:            false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"InstanceFamily", "InstanceSize", "Arch"}),
													},
												},
												"operator": schema.StringAttribute{
													Description:         "The operator for the requirement (e.g., In)",
													MarkdownDescription: "Requirement operator. Currently only `In` is supported.",
													Required:            true,
													Computed:            false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"In"}),
													},
												},
												"values": schema.ListAttribute{
													Description:         "List of values for the requirement",
													MarkdownDescription: "Allowed values for the requirement.",
													Required:            true,
													Computed:            false,
													ElementType:         types.StringType,
												},
											},
										},
									},
									"stable_override": schema.SingleNestedAttribute{
										Description:         "Defines some overridden options for Qovery stable node pool",
										MarkdownDescription: "Override options for the stable node pool (consolidation and resource limits).",
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"consolidation": schema.SingleNestedAttribute{
												Description:         "Specifies the period to consolidate nodes (by default, no consolidation happens)",
												MarkdownDescription: "Node consolidation schedule for the stable node pool.",
												Optional:            true,
												Computed:            false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Indicates if the consolidation defines here must be applied",
														MarkdownDescription: "Whether the consolidation schedule is active.",
														Required:            true,
														Computed:            false,
													},
													"days": schema.ListAttribute{
														Description:         "A list of days where the consolidation must be triggered",
														MarkdownDescription: "Days of the week when consolidation runs.",
														Required:            true,
														Computed:            false,
														ElementType:         types.StringType,
													},
													"start_time": schema.StringAttribute{
														Description:         "The start time where the consolidation must begin. It must follow the ISO-8601 time format: `PThh:mm`",
														MarkdownDescription: "Start time in ISO-8601 format (`PThh:mm`).",
														Required:            true,
														Computed:            false,
													},
													"duration": schema.StringAttribute{
														Description:         "The period during the consolidation will be applied. It must follow the ISO-8601 duration format: `PThhHmmM`",
														MarkdownDescription: "Duration in ISO-8601 format (`PThhHmmM`).",
														Required:            true,
														Computed:            false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												Description:         "Specifies the limits to apply on the stable node pool",
												MarkdownDescription: "Resource limits for the stable node pool.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Enabled the limit",
														MarkdownDescription: "Whether resource limits are enforced.",
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description:         "The maximum number of cpu cores to be used inside the stable node pool",
														MarkdownDescription: "Maximum total vCPU cores for the stable node pool.",
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description:         "The maximum number of memory to be used inside the stable node pool",
														MarkdownDescription: "Maximum total memory in GiB for the stable node pool.",
														Required:            true,
														Computed:            false,
													},
												},
											},
										},
									},
									"default_override": schema.SingleNestedAttribute{
										Description:         "Defines some overridden options for Qovery default node pool",
										MarkdownDescription: "Override options for the default node pool (resource limits).",
										Optional:            true,
										Computed:            false,
										Attributes: map[string]schema.Attribute{
											"limits": schema.SingleNestedAttribute{
												Description:         "Specifies the limits to apply on the default node pool",
												MarkdownDescription: "Resource limits for the default node pool.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Enabled the limit",
														MarkdownDescription: "Whether resource limits are enforced.",
														Required:            true,
														Computed:            false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description:         "The maximum number of cpu cores to be used inside the default node pool",
														MarkdownDescription: "Maximum total vCPU cores for the default node pool.",
														Required:            true,
														Computed:            false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description:         "The maximum number of memory to be used inside the default node pool",
														MarkdownDescription: "Maximum total memory in GiB for the default node pool.",
														Required:            true,
														Computed:            false,
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
				Description:         "List of routes of the cluster.",
				MarkdownDescription: "Custom routing table entries for the cluster VPC.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description:         "Description of the route.",
							MarkdownDescription: "Description of the route.",
							Computed:            true,
						},
						"destination": schema.StringAttribute{
							Description:         "Destination of the route.",
							MarkdownDescription: "Destination CIDR block for the route.",
							Computed:            true,
						},
						"target": schema.StringAttribute{
							Description:         "Target of the route.",
							MarkdownDescription: "Target gateway or endpoint for the route.",
							Computed:            true,
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
				MarkdownDescription: "Current state of the cluster (`DEPLOYED` or `STOPPED`).",
				Computed:            true,
				Optional:            true,
			},
			"advanced_settings_json": schema.StringAttribute{
				Description:         "Advanced settings of the cluster.",
				MarkdownDescription: "Advanced settings of the cluster as a JSON string.",
				Optional:            true,
				Computed:            true,
			},
			"labels_group_ids": schema.SetAttribute{
				Description:         "List of labels group ids associated with the cluster.",
				MarkdownDescription: "List of labels group ids associated with the cluster.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"kubeconfig": schema.StringAttribute{
				Description:         "Kubeconfig for connecting to the cluster. Only available for PARTIALLY_MANAGED (EKS Anywhere) clusters.",
				MarkdownDescription: "Kubeconfig for connecting to the cluster. Only available for `PARTIALLY_MANAGED` clusters.",
				Computed:            true,
				Sensitive:           true,
			},
			"infrastructure_outputs": schema.SingleNestedAttribute{
				Description:         "Outputs related to the underlying Kubernetes infrastructure. These values are only available once the cluster is deployed.",
				MarkdownDescription: "Read-only outputs from the underlying Kubernetes infrastructure. Available after deployment.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						Description:         "The name of the Kubernetes cluster. Available after deployment for all providers.",
						MarkdownDescription: "The name of the Kubernetes cluster as assigned by the cloud provider.",
						Computed:            true,
					},
					"cluster_arn": schema.StringAttribute{
						Description:         "The ARN of the AWS cluster. Only available for AWS after deployment.",
						MarkdownDescription: "The ARN of the EKS cluster (AWS only).",
						Computed:            true,
					},
					"cluster_self_link": schema.StringAttribute{
						Description:         "The self-link of the GCP cluster. Only available for GCP after deployment.",
						MarkdownDescription: "The self-link URL of the GKE cluster (GCP only).",
						Computed:            true,
					},
					"cluster_oidc_issuer": schema.StringAttribute{
						Description:         "The OIDC issuer URL for the cluster. Available for AWS and Azure after deployment.",
						MarkdownDescription: "The OIDC issuer URL (AWS and Azure only).",
						Computed:            true,
					},
					"vpc_id": schema.StringAttribute{
						Description:         "The VPC ID used by the cluster. Only available for AWS after deployment.",
						MarkdownDescription: "The VPC ID used by the cluster (AWS only).",
						Computed:            true,
					},
				},
			},
			"infrastructure_charts_parameters": schema.SingleNestedAttribute{
				Description:         "Infrastructure charts parameters for PARTIALLY_MANAGED (EKS Anywhere) clusters.",
				MarkdownDescription: "Infrastructure Helm chart parameters for `PARTIALLY_MANAGED` clusters.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"nginx_parameters": schema.SingleNestedAttribute{
						Description:         "Nginx ingress controller parameters.",
						MarkdownDescription: "Nginx ingress controller configuration.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"replica_count": schema.Int64Attribute{
								Description:         "Number of Nginx replicas.",
								MarkdownDescription: "Number of Nginx ingress controller replicas.",
								Computed:            true,
							},
							"default_ssl_certificate": schema.StringAttribute{
								Description:         "Default SSL certificate.",
								MarkdownDescription: "Default SSL certificate reference.",
								Computed:            true,
							},
							"publish_status_address": schema.StringAttribute{
								Description:         "Public IP address for status publishing.",
								MarkdownDescription: "Public IP address for ingress status publishing.",
								Computed:            true,
							},
							"annotation_metal_lb_load_balancer_ips": schema.StringAttribute{
								Description:         "MetalLB load balancer IP annotation.",
								MarkdownDescription: "MetalLB load balancer IP annotation.",
								Computed:            true,
							},
							"annotation_external_dns_kubernetes_target": schema.StringAttribute{
								Description:         "External DNS Kubernetes target annotation.",
								MarkdownDescription: "External DNS Kubernetes target annotation.",
								Computed:            true,
							},
						},
					},
					"cert_manager_parameters": schema.SingleNestedAttribute{
						Description:         "Cert-manager parameters.",
						MarkdownDescription: "Cert-manager configuration.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"kubernetes_namespace": schema.StringAttribute{
								Description:         "Kubernetes namespace for cert-manager.",
								MarkdownDescription: "Kubernetes namespace where cert-manager is installed.",
								Computed:            true,
							},
						},
					},
					"metal_lb_parameters": schema.SingleNestedAttribute{
						Description:         "MetalLB load balancer parameters.",
						MarkdownDescription: "MetalLB load balancer configuration.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"ip_address_pools": schema.ListAttribute{
								Description:         "List of IP address pools.",
								MarkdownDescription: "List of IP address pools for MetalLB.",
								ElementType:         types.StringType,
								Computed:            true,
							},
						},
					},
					"eks_anywhere_parameters": schema.SingleNestedAttribute{
						Description:         "EKS Anywhere GitOps parameters.",
						MarkdownDescription: "EKS Anywhere GitOps parameters.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"yaml_file_path": schema.StringAttribute{
								Description:         "Path to the EKS Anywhere cluster YAML file in the Git repository.",
								MarkdownDescription: "Path to the EKS Anywhere cluster YAML file in the Git repository.",
								Computed:            true,
							},
							"git_repository": schema.SingleNestedAttribute{
								Description:         "Git repository settings used for EKS Anywhere.",
								MarkdownDescription: "Git repository settings used for EKS Anywhere.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description:         "Git repository URL.",
										MarkdownDescription: "Git repository URL.",
										Computed:            true,
									},
									"git_token_id": schema.StringAttribute{
										Description:         "Qovery Git token ID used to access the repository.",
										MarkdownDescription: "Qovery Git token ID used to access the repository.",
										Computed:            true,
									},
									"commit_id": schema.StringAttribute{
										Description:         "Optional git commit SHA to pin EKS Anywhere configuration.",
										MarkdownDescription: "Optional git commit SHA to pin EKS Anywhere configuration to a specific revision.",
										Computed:            true,
									},
									"branch": schema.StringAttribute{
										Description:         "Repository branch name.",
										MarkdownDescription: "Repository branch name.",
										Computed:            true,
									},
									"provider": schema.StringAttribute{
										Description:         "Git provider (`BITBUCKET`, `GITHUB`, `GITLAB`).",
										MarkdownDescription: "Git provider (`BITBUCKET`, `GITHUB`, `GITLAB`).",
										Computed:            true,
									},
								},
							},
							"cluster_backup": schema.SingleNestedAttribute{
								Description:         "EKS Anywhere cluster backup parameters.",
								MarkdownDescription: "EKS Anywhere cluster backup parameters.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description:         "Enable or disable EKS Anywhere cluster backup.",
										MarkdownDescription: "Enable or disable EKS Anywhere cluster backup.",
										Computed:            true,
									},
									"s3": schema.SingleNestedAttribute{
										Description:         "S3 settings used to store backup artifacts.",
										MarkdownDescription: "S3 settings used to store backup artifacts.",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"bucket": schema.StringAttribute{
												Description:         "S3 bucket name used to store EKS Anywhere backup artifacts.",
												MarkdownDescription: "S3 bucket name used to store EKS Anywhere backup artifacts.",
												Computed:            true,
											},
											"region": schema.StringAttribute{
												Description:         "AWS region where the backup bucket is hosted.",
												MarkdownDescription: "AWS region where the backup bucket is hosted.",
												Computed:            true,
											},
											"role_arn": schema.StringAttribute{
												Description:         "IAM role ARN assumed to upload backup artifacts.",
												MarkdownDescription: "IAM role ARN assumed to upload backup artifacts.",
												Computed:            true,
											},
											"key_prefix": schema.StringAttribute{
												Description:         "Optional S3 key prefix used for backup object keys.",
												MarkdownDescription: "Optional S3 key prefix used for backup object keys.",
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
	tflog.Trace(ctx, "read cluster", map[string]any{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
