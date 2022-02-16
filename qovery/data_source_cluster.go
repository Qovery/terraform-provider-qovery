package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
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
			"cpu": {
				Description: "CPU of the cluster in millicores (m) [1000m = 1 CPU].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"memory": {
				Description: "RAM of the cluster in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
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
		client: p.(*provider).GetClient(),
	}, nil
}

type clusterDataSource struct {
	client *qovery.APIClient
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
	clusters, res, err := d.client.ClustersApi.
		ListOrganizationCluster(ctx, data.OrganizationId.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := clusterReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Get cluster credentials from the API
	cloudProviderInfo, res, err := d.client.ClustersApi.
		GetOrganizationCloudProviderInfo(ctx, data.OrganizationId.Value, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := cloudProviderReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	var state Cluster
	found := false
	for _, cluster := range clusters.GetResults() {
		if data.Id.Value == cluster.Id {
			found = true
			state = convertResponseToCluster(&cluster, cloudProviderInfo, data)
			break
		}
	}

	// If cluster id is not in list
	// Returning Not Found error
	if !found {
		res.StatusCode = 404
		apiErr := clusterReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
