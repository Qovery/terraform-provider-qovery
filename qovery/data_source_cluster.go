package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
)

type clusterDataSourceType struct{}

func (t clusterDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
		},
	}, nil
}

func (t clusterDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return clusterDataSource{
		client: p.(*provider).client,
	}, nil
}

type clusterDataSource struct {
	client *client.Client
}

// Read qovery cluster data source
func (d clusterDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
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
