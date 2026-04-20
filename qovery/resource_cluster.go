package qovery

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &clusterResource{}
	_ resource.ResourceWithImportState = clusterResource{}
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
	clusterMinRunningNodesMin     int64 = 1
	clusterMinRunningNodesDefault int64 = 3

	// Cluster Max Running Nodes
	clusterMaxRunningNodesMin     int64 = 1
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
	client *client.Client
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
}

func (r clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO (framework-migration): test if Default is OK when modifying the attribute, otherwise we'll need to use a modifier
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery cluster resource. This can be used to create and manage Qovery cluster.",
		MarkdownDescription: "Provides a Qovery cluster resource. This is used to create and manage Kubernetes clusters on your chosen cloud provider through Qovery.\n\n" +
			"Qovery supports clusters on **AWS** (EKS), **GCP** (GKE), **Scaleway** (Kapsule), and **Azure** (AKS). " +
			"Each cloud provider requires its own credentials resource (e.g., `qovery_aws_credentials`). " +
			"For AWS clusters, you can optionally enable **Karpenter** for automatic node provisioning or deploy on an **existing VPC**. " +
			"For GCP clusters, you can use **Autopilot** mode or deploy on an **existing VPC**. " +
			"AWS also supports **PARTIALLY_MANAGED** mode for EKS Anywhere on-premise clusters.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the cluster.",
				MarkdownDescription: "Unique identifier of the cluster (UUID format).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credentials_id": schema.StringAttribute{
				Description:         "Id of the credentials.",
				MarkdownDescription: "ID of the cloud provider credentials to use for this cluster. Must match the `cloud_provider` type (e.g., use `qovery_aws_credentials.id` for AWS clusters, `qovery_gcp_credentials.id` for GCP clusters).",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "ID of the Qovery organization in which to create the cluster.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the cluster.",
				MarkdownDescription: "Name of the cluster. Must be unique within the organization.",
				Required:            true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				MarkdownDescription: "Cloud provider where the cluster will be deployed.\n\n" +
					"  - `AWS` - Amazon Web Services (EKS).\n" +
					"  - `GCP` - Google Cloud Platform (GKE).\n" +
					"  - `SCW` - Scaleway (Kapsule).\n" +
					"  - `AZURE` - Microsoft Azure (AKS).\n" +
					"  - `ON_PREMISE` - On-premise infrastructure.",
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(cloudProviders),
				},
			},
			"region": schema.StringAttribute{
				Description:         "Region of the cluster.",
				MarkdownDescription: "Cloud provider region where the cluster will be deployed (e.g., `us-east-2` for AWS, `europe-west9` for GCP, `pl-waw-1` for Scaleway, `westeurope` for Azure). For PARTIALLY_MANAGED clusters, use `on-premise`.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				MarkdownDescription: "Description of the cluster. Default: `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clusterDescriptionDefault),
			},
			"kubernetes_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				MarkdownDescription: "Kubernetes management mode for the cluster. Default: `MANAGED`.\n\n" +
					"  - `MANAGED` - Fully managed Kubernetes cluster provisioned and managed by Qovery (e.g., AWS EKS, GCP GKE, Azure AKS).\n" +
					"  - `SELF_MANAGED` - Bring your own Kubernetes cluster. Qovery deploys workloads but does not manage infrastructure.\n" +
					"  - `PARTIALLY_MANAGED` - EKS Anywhere / on-premise mode. Qovery manages workloads on a user-provided Kubernetes cluster via kubeconfig. Requires `kubeconfig` and `infrastructure_charts_parameters`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(clusterKubernetesModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterKubernetesModes),
				},
			},
			"production": schema.BoolAttribute{
				Description:         "Specific flag to indicate that this cluster is a production one.",
				MarkdownDescription: "Flag to mark this cluster as a production cluster. Production clusters may have different default settings and safeguards. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"instance_type": schema.StringAttribute{
				Description:         "Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`, and not set for Karpenter-enabled clusters",
				MarkdownDescription: "Instance type for the cluster nodes. The available values depend on the cloud provider:\n\n  - **AWS**: EC2 instance types (e.g., `t3a.xlarge`, `m5.large`). Not required when Karpenter is enabled.\n  - **GCP**: Machine types or `AUTO_PILOT` for GKE Autopilot mode.\n  - **Scaleway**: Node types (e.g., `DEV1-L`, `GP1-S`).\n  - **Azure**: VM sizes (e.g., `Standard_B2s_v2`, `Standard_D4s_v3`).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disk_size": schema.Int64Attribute{
				Description:         "Disk size of the cluster nodes in GB.",
				MarkdownDescription: "Disk size of the cluster nodes in GB. The default value depends on the cloud provider and instance type.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters, and not set for Karpenter-enabled clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				MarkdownDescription: "Minimum number of nodes running for the cluster autoscaler. Must be `>= 1`. Default: `3`.\n\n" +
					"~> **Note:** Must be set to `1` for K3S clusters. Do not set this attribute when Karpenter is enabled (Karpenter manages scaling automatically).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters; and not set for Karpenter-enabled clusters]",
					clusterMaxRunningNodesMin,
					&clusterMaxRunningNodesDefault,
				),
				MarkdownDescription: "Maximum number of nodes the cluster autoscaler can scale up to. Must be `>= 1`. Default: `10`.\n\n" +
					"~> **Note:** Must be set to `1` for K3S clusters. Do not set this attribute when Karpenter is enabled (Karpenter manages scaling automatically).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"features": schema.SingleNestedAttribute{
				Description:         "Features of the cluster.",
				MarkdownDescription: "Optional cluster features configuration. Use this block to customize VPC settings, enable static IPs, deploy on an existing VPC (AWS or GCP), or enable Karpenter for AWS clusters.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"vpc_subnet": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureVpcSubnetDefault,
						),
						MarkdownDescription: "Custom VPC CIDR block for AWS clusters. This defines the IP address range for the entire VPC. Default: `10.0.0.0/16`.\n\n" +
							"~> **Warning:** This value cannot be changed after cluster creation. Changing it will require destroying and recreating the cluster.",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(clusterFeatureVpcSubnetDefault),
					},
					"static_ip": schema.BoolAttribute{
						Description: descriptions.NewBoolDefaultDescription(
							"Static IP (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureStaticIPDefault,
						),
						MarkdownDescription: "Whether to assign static/elastic IP addresses to the cluster nodes (AWS only). Useful when your services need to be allowlisted by IP. Default: `false`.\n\n" +
							"~> **Warning:** This value cannot be changed after cluster creation. Changing it will require destroying and recreating the cluster.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(clusterFeatureStaticIPDefault),
					},
					"existing_vpc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: false,
						Description: "Network configuration if you want to install qovery on an existing VPC",
						MarkdownDescription: "AWS existing VPC configuration. Use this block to deploy the Qovery cluster into an existing AWS VPC instead of creating a new one. " +
							"All EKS subnets are required, while database and cache subnets are optional.\n\n" +
							"~> **Warning:** This configuration cannot be changed after cluster creation.",
						Attributes: map[string]schema.Attribute{
							"aws_vpc_eks_id": schema.StringAttribute{
								Description:         "Aws VPC id",
								MarkdownDescription: "The ID of the existing AWS VPC (e.g., `vpc-0123456789abcdef0`).",
								Required:            true,
								Computed:             false,
							},
							"eks_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone a. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "List of subnet IDs in availability zone A for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:             false,
							},
							"eks_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone b. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "List of subnet IDs in availability zone B for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:             false,
							},
							"eks_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS zone c. Must have map_public_ip_on_launch set to true",
								MarkdownDescription: "List of subnet IDs in availability zone C for EKS worker nodes. These subnets must have `map_public_ip_on_launch` set to `true`.",
								ElementType:         types.StringType,
								Required:            true,
								Computed:             false,
							},
							"rds_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "List of subnet IDs in availability zone A for Amazon RDS databases. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"rds_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "List of subnet IDs in availability zone B for Amazon RDS databases. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"rds_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for RDS",
								MarkdownDescription: "List of subnet IDs in availability zone C for Amazon RDS databases. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"documentdb_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "List of subnet IDs in availability zone A for Amazon DocumentDB. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"documentdb_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "List of subnet IDs in availability zone B for Amazon DocumentDB. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"documentdb_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for document db",
								MarkdownDescription: "List of subnet IDs in availability zone C for Amazon DocumentDB. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"elasticache_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "List of subnet IDs in availability zone A for Amazon ElastiCache. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"elasticache_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "List of subnet IDs in availability zone B for Amazon ElastiCache. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"elasticache_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for elasticache",
								MarkdownDescription: "List of subnet IDs in availability zone C for Amazon ElastiCache. These should be private subnets.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             true,
							},
							"eks_karpenter_fargate_subnets_zone_a_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone a. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "List of private subnet IDs in availability zone A for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             false,
							},
							"eks_karpenter_fargate_subnets_zone_b_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone b. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "List of private subnet IDs in availability zone B for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             false,
							},
							"eks_karpenter_fargate_subnets_zone_c_ids": schema.ListAttribute{
								Description:         "Ids of the subnets for EKS fargate zone c. Must have to be private and connected to internet through a NAT Gateway",
								MarkdownDescription: "List of private subnet IDs in availability zone C for EKS Fargate (required when using Karpenter). These subnets must be private and connected to the internet through a NAT Gateway.",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:             false,
							},
							"eks_create_nodes_in_private_subnet": schema.BoolAttribute{
								Description:         "Whether to create EKS nodes in private subnet",
								MarkdownDescription: "Whether to create EKS worker nodes in private subnets. When `true`, nodes are not directly accessible from the internet and route traffic through a NAT Gateway.",
								Optional:            true,
								Computed:             true,
							},
						},
					},
					"gcp_existing_vpc": schema.SingleNestedAttribute{
						Optional: true,
						Computed: false,
						Description: "Network configuration if you want to install qovery on an existing GCP VPC",
						MarkdownDescription: "GCP existing VPC configuration. Use this block to deploy the Qovery GKE cluster into an existing Google Cloud VPC network instead of creating a new one.\n\n" +
							"~> **Warning:** This configuration cannot be changed after cluster creation.",
						Attributes: map[string]schema.Attribute{
							"vpc_name": schema.StringAttribute{
								Description:         "Name of the existing GCP VPC network",
								MarkdownDescription: "Name of the existing GCP VPC network to use (e.g., `my-existing-vpc`).",
								Required:            true,
							},
							"vpc_project_id": schema.StringAttribute{
								Description:         "GCP project ID that owns the VPC. Defaults to the project associated with your GCP credentials",
								MarkdownDescription: "GCP project ID that owns the VPC. If omitted, defaults to the project associated with your GCP credentials. Use this when the VPC is in a different project (Shared VPC pattern).",
								Optional:            true,
							},
							"subnetwork_name": schema.StringAttribute{
								Description:         "Name of the GCP subnetwork within the VPC",
								MarkdownDescription: "Name of the GCP subnetwork within the VPC to use for the GKE cluster nodes.",
								Optional:            true,
							},
							"ip_range_services_name": schema.StringAttribute{
								Description:         "Name of the secondary IP range for GKE services",
								MarkdownDescription: "Name of the secondary IP range in the subnetwork to use for GKE services (ClusterIP range).",
								Optional:            true,
							},
							"ip_range_pods_name": schema.StringAttribute{
								Description:         "Name of the secondary IP range for pods",
								MarkdownDescription: "Name of the primary secondary IP range in the subnetwork to use for GKE pods.",
								Optional:            true,
							},
							"additional_ip_range_pods_names": schema.ListAttribute{
								Description:         "Additional secondary IP range names for pods",
								MarkdownDescription: "Additional secondary IP range names for pods. Use this when you need multiple pod IP ranges (e.g., for multi-tenancy or large clusters).",
								ElementType:         types.StringType,
								Optional:            true,
							},
						},
					},
					"karpenter": schema.SingleNestedAttribute{
						Optional: true,
						Computed: false,
						Description: "Karpenter parameters if you want to use Karpenter on an EKS cluster",
						MarkdownDescription: "Karpenter configuration for AWS EKS clusters. [Karpenter](https://karpenter.sh/) is a Kubernetes node autoscaler that automatically provisions right-sized compute resources. " +
							"When Karpenter is enabled, do not set `instance_type`, `min_running_nodes`, or `max_running_nodes` — Karpenter manages node scaling automatically.",
						Attributes: map[string]schema.Attribute{
							"spot_enabled": schema.BoolAttribute{
								Description:         "Enable spot instances",
								MarkdownDescription: "Whether to enable EC2 Spot instances for cost savings. Spot instances can be interrupted by AWS with a 2-minute notice, so enable this only for fault-tolerant workloads.",
								Required:            true,
								Computed:             false,
							},
							"disk_size_in_gib": schema.Int64Attribute{
								Description:         "Disk size in GiB for Karpenter-provisioned nodes.",
								MarkdownDescription: "Root disk size in GiB for nodes provisioned by Karpenter (e.g., `50`).",
								Required:            true,
								Computed:             false,
							},
							"default_service_architecture": schema.StringAttribute{
								Description:         "The default architecture of service",
								MarkdownDescription: "Default CPU architecture for services deployed on this cluster. Common values: `AMD64`, `ARM64`. This determines the default node architecture when no specific architecture is requested by a service.",
								Required:            true,
								Computed:             false,
							},
							"qovery_node_pools": schema.SingleNestedAttribute{
								Description:         "Karpenter node pool configuration",
								MarkdownDescription: "Karpenter node pool configuration. Defines the requirements (instance families, sizes, architectures) and optional resource limits for Qovery-managed node pools.",
								Required:            true,
								Computed:             false,
								Attributes: map[string]schema.Attribute{
									"requirements": schema.ListNestedAttribute{
										Description:         "List of requirements for the node pool",
										MarkdownDescription: "List of node selection requirements for the Karpenter node pool. Each requirement constrains which EC2 instances Karpenter can provision. You should define at least `InstanceFamily`, `InstanceSize`, and `Arch` requirements.",
										Required:            true,
										Computed:             false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"key": schema.StringAttribute{
													Description: "The key of the requirement (e.g., InstanceFamily, InstanceSize, Arch)",
													MarkdownDescription: "The requirement key. Valid values:\n\n" +
														"  - `InstanceFamily` - EC2 instance family (e.g., `c5`, `m5`, `t3a`). Use broad families to reduce allocation issues.\n" +
														"  - `InstanceSize` - EC2 instance size (e.g., `small`, `medium`, `xlarge`, `2xlarge`).\n" +
														"  - `Arch` - CPU architecture (e.g., `AMD64`, `ARM64`).",
													Required: true,
													Computed: false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"InstanceFamily", "InstanceSize", "Arch"}),
													},
												},
												"operator": schema.StringAttribute{
													Description:         "The operator for the requirement (e.g., In)",
													MarkdownDescription: "The operator for the requirement. Currently only `In` is supported, meaning the node must match one of the specified values.",
													Required:            true,
													Computed:             false,
													Validators: []validator.String{
														validators.NewStringEnumValidator([]string{"In"}),
													},
												},
												"values": schema.ListAttribute{
													Description:         "List of values for the requirement",
													MarkdownDescription: "List of allowed values for the requirement. For example, for `InstanceFamily`: `[\"c5\", \"m5\", \"t3a\"]`, for `Arch`: `[\"AMD64\", \"ARM64\"]`.",
													Required:            true,
													Computed:             false,
													ElementType:         types.StringType,
												},
											},
										},
									},
									"stable_override": schema.SingleNestedAttribute{
										Description:         "Defines some overriden options for Qovery stable node pool",
										MarkdownDescription: "Override options for the Qovery **stable** node pool. The stable node pool runs services that require consistent availability (e.g., Qovery agents). Use this to configure consolidation windows and resource limits.",
										Optional:            true,
										Computed:             false,
										Attributes: map[string]schema.Attribute{
											"consolidation": schema.SingleNestedAttribute{
												Description:         "Specifies the period to consolidate nodes (by default, no consolidation happens)",
												MarkdownDescription: "Node consolidation schedule for the stable node pool. Consolidation replaces underutilized nodes with more cost-effective alternatives. By default, no consolidation occurs on stable nodes.",
												Optional:            true,
												Computed:             false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Whether the consolidation schedule is active.",
														MarkdownDescription: "Whether the consolidation schedule defined here is active. Set to `true` to enable scheduled consolidation.",
														Required:            true,
														Computed:             false,
													},
													"days": schema.ListAttribute{
														Description:         "Days of the week when consolidation runs.",
														MarkdownDescription: "List of days of the week when consolidation should run (e.g., `[\"Monday\", \"Tuesday\", \"Wednesday\"]`).",
														Required:            true,
														Computed:             false,
														ElementType:         types.StringType,
													},
													"start_time": schema.StringAttribute{
														Description:         "Start time for the consolidation window in ISO-8601 time format.",
														MarkdownDescription: "Start time for the consolidation window. Must follow the ISO-8601 time format: `PThh:mm` (e.g., `PT02:00` for 2:00 AM UTC).",
														Required:            true,
														Computed:             false,
													},
													"duration": schema.StringAttribute{
														Description:         "Duration of the consolidation window in ISO-8601 duration format.",
														MarkdownDescription: "Duration of the consolidation window. Must follow the ISO-8601 duration format: `PThhHmmM` (e.g., `PT04H00M` for a 4-hour window).",
														Required:            true,
														Computed:             false,
													},
												},
											},
											"limits": schema.SingleNestedAttribute{
												Description:         "Specifies the limits to apply on the stable node pool",
												MarkdownDescription: "Resource limits for the stable node pool. Use this to cap the total resources Karpenter can provision for stable workloads.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Enabled the limit",
														MarkdownDescription: "Whether to enforce resource limits on the stable node pool.",
														Required:            true,
														Computed:             false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description:         "Maximum number of vCPU cores for the stable node pool.",
														MarkdownDescription: "Maximum total vCPU cores that Karpenter can provision for the stable node pool.",
														Required:            true,
														Computed:             false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description:         "Maximum memory in GiB for the stable node pool.",
														MarkdownDescription: "Maximum total memory in GiB that Karpenter can provision for the stable node pool.",
														Required:            true,
														Computed:             false,
													},
												},
											},
										},
									},
									"default_override": schema.SingleNestedAttribute{
										Description:         "Defines some overriden options for Qovery default node pool",
										MarkdownDescription: "Override options for the Qovery **default** node pool. The default node pool runs user application workloads. Use this to set resource limits.",
										Optional:            true,
										Computed:             false,
										Attributes: map[string]schema.Attribute{
											"limits": schema.SingleNestedAttribute{
												Description:         "Specifies the limits to apply on the default node pool",
												MarkdownDescription: "Resource limits for the default node pool. Use this to cap the total resources Karpenter can provision for application workloads.",
												Optional:            true,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description:         "Enabled the limit",
														MarkdownDescription: "Whether to enforce resource limits on the default node pool.",
														Required:            true,
														Computed:             false,
													},
													"max_cpu_in_vcpu": schema.Int64Attribute{
														Description:         "Maximum number of vCPU cores for the default node pool.",
														MarkdownDescription: "Maximum total vCPU cores that Karpenter can provision for the default node pool.",
														Required:            true,
														Computed:             false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description:         "Maximum memory in GiB for the default node pool.",
														MarkdownDescription: "Maximum total memory in GiB that Karpenter can provision for the default node pool.",
														Required:            true,
														Computed:             false,
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
				MarkdownDescription: "Custom routing table entries for the cluster VPC. Use this to define network routes for traffic between the cluster and other networks (e.g., VPN, peering connections).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description:         "Description of the route.",
							MarkdownDescription: "Human-readable description of the route's purpose.",
							Required:            true,
						},
						"destination": schema.StringAttribute{
							Description:         "Destination of the route.",
							MarkdownDescription: "Destination CIDR block for the route (e.g., `10.1.0.0/16`).",
							Required:            true,
						},
						"target": schema.StringAttribute{
							Description:         "Target of the route.",
							MarkdownDescription: "Target gateway or endpoint for the route (e.g., a VPC peering connection ID or NAT gateway ID).",
							Required:            true,
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
				MarkdownDescription: "Desired state of the cluster. Default: `DEPLOYED`.\n\n" +
					"  - `DEPLOYED` - The cluster is running and ready to accept workloads.\n" +
					"  - `STOPPED` - The cluster infrastructure is stopped to save costs. All workloads will be unavailable.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(clusterStateDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterStates),
				},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description:         "Advanced settings of the cluster.",
				MarkdownDescription: "Advanced settings of the cluster as a JSON string. Use `jsonencode()` to set values. The complete list of available settings is in the [Qovery API documentation](https://api-doc.qovery.com/#tag/Clusters/operation/getDefaultClusterAdvancedSettings). Only include settings you want to override.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels_group_ids": schema.SetAttribute{
				Description:         "List of labels group ids (EKS clusters only).",
				MarkdownDescription: "List of labels group ids. Labels groups allow you to add Kubernetes labels to the cluster's resources. **Currently supported only for EKS (AWS managed Kubernetes) clusters.** See [Labels & Annotations](https://www.qovery.com/docs/configuration/organization/labels-annotations).",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"kubeconfig": schema.StringAttribute{
				Description:         "Kubeconfig for connecting to the cluster. Required for PARTIALLY_MANAGED (EKS Anywhere) clusters.",
				MarkdownDescription: "Kubeconfig YAML content for connecting to the cluster. **Required** for `PARTIALLY_MANAGED` (EKS Anywhere) clusters. This is a sensitive value and will not be displayed in plan output. Use `file()` to read from a file.",
				Optional:            true,
				Sensitive:           true,
			},
			"infrastructure_outputs": schema.SingleNestedAttribute{
				Description:         "Outputs related to the underlying Kubernetes infrastructure. These values are only available once the cluster is deployed.",
				MarkdownDescription: "Read-only outputs from the underlying Kubernetes infrastructure. These values are populated after the cluster is deployed and can be used to integrate with other infrastructure resources.",
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"cluster_name": schema.StringAttribute{
						Description:         "The name of the Kubernetes cluster. Available after deployment for all providers.",
						MarkdownDescription: "The name of the Kubernetes cluster as assigned by the cloud provider. Available after deployment for all providers.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_arn": schema.StringAttribute{
						Description:         "The ARN of the AWS cluster. Only available for AWS after deployment.",
						MarkdownDescription: "The Amazon Resource Name (ARN) of the EKS cluster. Only populated for AWS clusters after deployment.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_self_link": schema.StringAttribute{
						Description:         "The self-link of the GCP cluster. Only available for GCP after deployment.",
						MarkdownDescription: "The self-link URL of the GKE cluster. Only populated for GCP clusters after deployment.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_oidc_issuer": schema.StringAttribute{
						Description:         "The OIDC issuer URL for the cluster. Available for AWS and Azure after deployment.",
						MarkdownDescription: "The OIDC issuer URL for the cluster. Useful for configuring IAM roles for service accounts (IRSA on AWS, workload identity on Azure). Available for AWS and Azure after deployment.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"vpc_id": schema.StringAttribute{
						Description:         "The VPC ID used by the cluster. Only available for AWS after deployment.",
						MarkdownDescription: "The VPC ID used by the cluster. Only populated for AWS clusters after deployment. Useful for setting up VPC peering or other networking resources.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"infrastructure_charts_parameters": schema.SingleNestedAttribute{
				Description:         "Infrastructure charts parameters for PARTIALLY_MANAGED (EKS Anywhere) clusters. Required when kubernetes_mode is PARTIALLY_MANAGED.",
				MarkdownDescription: "Infrastructure Helm chart parameters for `PARTIALLY_MANAGED` (EKS Anywhere) clusters. **Required** when `kubernetes_mode` is `PARTIALLY_MANAGED`. These configure the core infrastructure components (ingress, TLS, load balancing) on your on-premise cluster.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"nginx_parameters": schema.SingleNestedAttribute{
						Description:         "Nginx ingress controller parameters.",
						MarkdownDescription: "Configuration for the Nginx ingress controller deployed on the cluster.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"replica_count": schema.Int64Attribute{
								Description:         "Number of Nginx replicas.",
								MarkdownDescription: "Number of Nginx ingress controller replicas. Increase for high-availability setups.",
								Optional:            true,
							},
							"default_ssl_certificate": schema.StringAttribute{
								Description:         "Default SSL certificate (e.g., 'cert-manager/letsencrypt-acme-qovery-cert').",
								MarkdownDescription: "Default SSL certificate reference in `namespace/secret-name` format (e.g., `qovery/letsencrypt-acme-qovery-cert`).",
								Optional:            true,
							},
							"publish_status_address": schema.StringAttribute{
								Description:         "Public IP address for status publishing.",
								MarkdownDescription: "Public IP address reported in the ingress status. This is the IP that external DNS will resolve to.",
								Optional:            true,
							},
							"annotation_metal_lb_load_balancer_ips": schema.StringAttribute{
								Description:         "MetalLB load balancer IP annotation.",
								MarkdownDescription: "IP address annotation for MetalLB load balancer allocation (e.g., `192.168.1.100`). Must be within a MetalLB IP address pool.",
								Optional:            true,
							},
							"annotation_external_dns_kubernetes_target": schema.StringAttribute{
								Description:         "External DNS Kubernetes target annotation.",
								MarkdownDescription: "IP address or hostname used by external-dns for DNS record creation (e.g., `192.168.1.100`).",
								Optional:            true,
							},
						},
					},
					"cert_manager_parameters": schema.SingleNestedAttribute{
						Description:         "Cert-manager parameters.",
						MarkdownDescription: "Configuration for cert-manager, used for automatic TLS certificate provisioning.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"kubernetes_namespace": schema.StringAttribute{
								Description:         "Kubernetes namespace for cert-manager (e.g., 'cert-manager').",
								MarkdownDescription: "Kubernetes namespace where cert-manager is installed (e.g., `cert-manager` or `qovery`).",
								Optional:            true,
							},
						},
					},
					"metal_lb_parameters": schema.SingleNestedAttribute{
						Description:         "MetalLB load balancer parameters. Required for PARTIALLY_MANAGED mode.",
						MarkdownDescription: "Configuration for MetalLB, a bare-metal load balancer for Kubernetes. Required for `PARTIALLY_MANAGED` clusters to expose services externally.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ip_address_pools": schema.ListAttribute{
								Description:         "List of IP address pools as single IPs or IP range format (e.g., '192.168.1.100' or '192.168.1.100-192.168.1.200').",
								MarkdownDescription: "List of IP address pools for MetalLB. Each entry can be a single IP or an IP range (e.g., `192.168.1.100` or `192.168.1.100-192.168.1.200`). These IPs must be routable on your network.",
								ElementType:         types.StringType,
								Required:            true,
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
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
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
			fmt.Sprintf("Expected import identifier with format: cluster_id,organization_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
