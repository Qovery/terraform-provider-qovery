package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (d clusterDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"credentials_id": {
				Description: "Id of the credentials.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cloud_provider": {
				Description: "Cloud provider of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"region": {
				Description: "Region of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"kubernetes_mode": {
				Description: "Kubernetes mode of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"instance_type": {
				Description: "Instance type of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"min_running_nodes": {
				Description: "Minimum number of nodes running for the cluster.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"max_running_nodes": {
				Description: "Maximum number of nodes running for the cluster.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"features": {
				Description: "Features of the cluster.",
				Computed:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"vpc_subnet": {
						Description: "Custom VPC subnet (AWS only) [NOTE: can't be updated after creation].",
						Type:        types.StringType,
						Computed:    true,
					},
					"static_ip": {
						Description: "Static IP (AWS only) [NOTE: can't be updated after creation].",
						Type:        types.BoolType,
						Computed:    true,
					},
				}),
			},
			"routing_table": {
				Description: "List of routes of the cluster.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"description": {
						Description: "Description of the route.",
						Type:        types.StringType,
						Computed:    true,
					},
					"destination": {
						Description: "Destination of the route.",
						Type:        types.StringType,
						Computed:    true,
					},
					"target": {
						Description: "Target of the route.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"state": {
				Description: "State of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"advanced_settings_json": {
				Description: "Advanced settings of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
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
	cluster, apiErr := d.client.GetCluster(ctx, data.OrganizationId.Value, data.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToCluster(cluster)
	tflog.Trace(ctx, "read cluster", map[string]interface{}{"cluster_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
