package qovery

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"

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
var _ resource.ResourceWithConfigure = &clusterResource{}
var _ resource.ResourceWithImportState = clusterResource{}

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
		qovery.KUBERNETESENUM_K3_S,
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

	// Initialize state values
	state := convertResponseToCluster(ctx, cluster, plan)
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

	// Get cluster from the API
	cluster, apiErr := r.client.GetCluster(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state = convertResponseToCluster(ctx, cluster, state)
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
	// Update state values
	state = convertResponseToCluster(ctx, cluster, plan)
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
