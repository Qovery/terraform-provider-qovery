package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type clusterDataSourceData struct {
	Id              types.String `tfsdk:"id"`
	OrganizationId  types.String `tfsdk:"organization_id"`
	CredentialsId   types.String `tfsdk:"credentials_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	CloudProvider   types.String `tfsdk:"cloud_provider"`
	Region          types.String `tfsdk:"region"`
	CPU             types.Int64  `tfsdk:"cpu"`
	Memory          types.Int64  `tfsdk:"memory"`
	MinRunningNodes types.Int64  `tfsdk:"min_running_nodes"`
	MaxRunningNodes types.Int64  `tfsdk:"max_running_nodes"`
}

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
	var data clusterDataSourceData
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

	var state *clusterDataSourceData
	for _, cluster := range clusters.GetResults() {
		if data.Id.Value == cluster.Id {
			state = &clusterDataSourceData{
				Id:              data.Id,
				OrganizationId:  data.OrganizationId,
				CredentialsId:   types.String{Null: true},
				Name:            types.String{Value: cluster.Name},
				Description:     types.String{Null: true},
				CloudProvider:   types.String{Value: cluster.CloudProvider},
				Region:          types.String{Value: cluster.Region},
				CPU:             types.Int64{Null: true},
				Memory:          types.Int64{Null: true},
				MinRunningNodes: types.Int64{Null: true},
				MaxRunningNodes: types.Int64{Null: true},
			}
			if cloudProviderInfo.Credentials != nil {
				state.CredentialsId = types.String{Value: *cloudProviderInfo.Credentials.Id}
			}
			if cluster.Description.Get() != nil {
				state.Description = types.String{Value: *cluster.Description.Get()}
			}
			if cluster.Cpu != nil {
				state.CPU = types.Int64{Value: int64(*cluster.Cpu)}
			}
			if cluster.Memory != nil {
				state.Memory = types.Int64{Value: int64(*cluster.Memory)}
			}
			if cluster.MinRunningNodes != nil {
				state.MinRunningNodes = types.Int64{Value: int64(*cluster.MinRunningNodes)}
			}
			if cluster.MaxRunningNodes != nil {
				state.MaxRunningNodes = types.Int64{Value: int64(*cluster.MaxRunningNodes)}
			}
			break
		}
	}

	// If cluster id is not in list
	// Returning Not Found error
	if state == nil {
		res.StatusCode = 404
		apiErr := clusterReadAPIError(state.Id.Value, res, nil)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
