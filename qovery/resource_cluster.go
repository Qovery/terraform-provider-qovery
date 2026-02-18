package qovery

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credentials_id": schema.StringAttribute{
				Description: "Id of the credentials.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Cloud provider of the cluster.",
					cloudProviders,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(cloudProviders),
				},
			},
			"region": schema.StringAttribute{
				Description: "Region of the cluster.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: descriptions.NewStringDefaultDescription(
					"Description of the cluster.",
					clusterDescriptionDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(clusterDescriptionDefault),
			},
			"kubernetes_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Kubernetes mode of the cluster.",
					clusterKubernetesModes,
					&clusterKubernetesModeDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(clusterKubernetesModeDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterKubernetesModes),
				},
			},
			"production": schema.BoolAttribute{
				Description: "Specific flag to indicate that this cluster is a production one.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters; and not set for Karpenter-enabled clusters]",
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
						Default:  stringdefault.StaticString(clusterFeatureVpcSubnetDefault),
					},
					"static_ip": schema.BoolAttribute{
						Description: descriptions.NewBoolDefaultDescription(
							"Static IP (AWS only) [NOTE: can't be updated after creation].",
							clusterFeatureStaticIPDefault,
						),
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(clusterFeatureStaticIPDefault),
					},
					"existing_vpc": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    false,
						Description: "Network configuration if you want to install qovery on an existing VPC",
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
							},
							"eks_create_nodes_in_private_subnet": schema.BoolAttribute{
								Description: "Whether to create EKS nodes in private subnet",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"gcp_existing_vpc": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    false,
						Description: "Network configuration if you want to install qovery on an existing GCP VPC",
						Attributes: map[string]schema.Attribute{
							"vpc_name": schema.StringAttribute{
								Description: "Name of the existing GCP VPC network",
								Required:    true,
							},
							"vpc_project_id": schema.StringAttribute{
								Description: "GCP project ID that owns the VPC. Defaults to the project associated with your GCP credentials",
								Optional:    true,
							},
							"subnetwork_name": schema.StringAttribute{
								Description: "Name of the GCP subnetwork within the VPC",
								Optional:    true,
							},
							"ip_range_services_name": schema.StringAttribute{
								Description: "Name of the secondary IP range for GKE services",
								Optional:    true,
							},
							"ip_range_pods_name": schema.StringAttribute{
								Description: "Name of the secondary IP range for pods",
								Optional:    true,
							},
							"additional_ip_range_pods_names": schema.ListAttribute{
								Description: "Additional secondary IP range names for pods",
								ElementType: types.StringType,
								Optional:    true,
							},
						},
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
										Description: "Defines some overriden options for Qovery stable node pool",
										Optional:    true,
										Computed:    false,
										Attributes: map[string]schema.Attribute{
											"consolidation": schema.SingleNestedAttribute{
												Description: "Specifies the period to consolidate nodes (by default, no consolidation happens)",
												Optional:    true,
												Computed:    false,
												Attributes: map[string]schema.Attribute{
													"enabled": schema.BoolAttribute{
														Description: "",
														Required:    true,
														Computed:    false,
													},
													"days": schema.ListAttribute{
														Description: "",
														Required:    true,
														Computed:    false,
														ElementType: types.StringType,
													},
													"start_time": schema.StringAttribute{
														Description: "",
														Required:    true,
														Computed:    false,
													},
													"duration": schema.StringAttribute{
														Description: "",
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
														Description: "",
														Required:    true,
														Computed:    false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description: "",
														Required:    true,
														Computed:    false,
													},
												},
											},
										},
									},
									"default_override": schema.SingleNestedAttribute{
										Description: "Defines some overriden options for Qovery default node pool",
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
														Description: "",
														Required:    true,
														Computed:    false,
													},
													"max_memory_in_gibibytes": schema.Int64Attribute{
														Description: "",
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
							Required:    true,
						},
						"destination": schema.StringAttribute{
							Description: "Destination of the route.",
							Required:    true,
						},
						"target": schema.StringAttribute{
							Description: "Target of the route.",
							Required:    true,
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
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(clusterStateDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(clusterStates),
				},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings of the cluster.",
				Optional:    true,
				Computed:    true,
			},
			"kubeconfig": schema.StringAttribute{
				Description: "Kubeconfig for connecting to the cluster. Required for PARTIALLY_MANAGED (EKS Anywhere) clusters.",
				Optional:    true,
				Sensitive:   true,
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
			"infrastructure_charts_parameters": schema.SingleNestedAttribute{
				Description: "Infrastructure charts parameters for PARTIALLY_MANAGED (EKS Anywhere) clusters. Required when kubernetes_mode is PARTIALLY_MANAGED.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"nginx_parameters": schema.SingleNestedAttribute{
						Description: "Nginx ingress controller parameters.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"replica_count": schema.Int64Attribute{
								Description: "Number of Nginx replicas.",
								Optional:    true,
							},
							"default_ssl_certificate": schema.StringAttribute{
								Description: "Default SSL certificate (e.g., 'cert-manager/letsencrypt-acme-qovery-cert').",
								Optional:    true,
							},
							"publish_status_address": schema.StringAttribute{
								Description: "Public IP address for status publishing.",
								Optional:    true,
							},
							"annotation_metal_lb_load_balancer_ips": schema.StringAttribute{
								Description: "MetalLB load balancer IP annotation.",
								Optional:    true,
							},
							"annotation_external_dns_kubernetes_target": schema.StringAttribute{
								Description: "External DNS Kubernetes target annotation.",
								Optional:    true,
							},
						},
					},
					"cert_manager_parameters": schema.SingleNestedAttribute{
						Description: "Cert-manager parameters.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"kubernetes_namespace": schema.StringAttribute{
								Description: "Kubernetes namespace for cert-manager (e.g., 'cert-manager').",
								Optional:    true,
							},
						},
					},
					"metal_lb_parameters": schema.SingleNestedAttribute{
						Description: "MetalLB load balancer parameters. Required for PARTIALLY_MANAGED mode.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"ip_address_pools": schema.ListAttribute{
								Description: "List of IP address pools as single IPs or IP range format (e.g., '192.168.1.100' or '192.168.1.100-192.168.1.200').",
								ElementType: types.StringType,
								Required:    true,
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
			tflog.Warn(ctx, "failed to fetch kubeconfig after create", map[string]interface{}{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "created cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

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
			tflog.Warn(ctx, "failed to fetch kubeconfig for PARTIALLY_MANAGED cluster", map[string]interface{}{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

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
			tflog.Warn(ctx, "failed to fetch kubeconfig after update", map[string]interface{}{"cluster_id": state.Id.ValueString(), "error": apiErr.Detail()})
		} else {
			state.Kubeconfig = types.StringValue(kubeconfig)
		}
	}

	tflog.Trace(ctx, "updated cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

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

	tflog.Trace(ctx, "deleted cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

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
