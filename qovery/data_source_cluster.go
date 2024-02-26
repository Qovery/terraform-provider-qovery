package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
			"instance_type": schema.StringAttribute{
				Description: "Instance type of the cluster. I.e: For Aws `t3a.xlarge`, for Scaleway `DEV-L`",
				Computed:    true,
			},
			"disk_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"min_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters].",
					clusterMinRunningNodesMin,
					&clusterMinRunningNodesDefault,
				),
				Optional: true,
				Computed: true,
			},
			"max_running_nodes": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of nodes running for the cluster. [NOTE: have to be set to 1 in case of K3S clusters]",
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
	cluster, apiErr := d.client.GetCluster(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToCluster(ctx, cluster, data)
	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
