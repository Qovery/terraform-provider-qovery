package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"
)

type databaseDataSourceType struct{}

func (t databaseDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing database.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the database.",
				Type:        types.StringType,
				Required:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"type": {
				Description: "Type of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"version": {
				Description: "Version of the database",
				Type:        types.StringType,
				Computed:    true,
			},
			"mode": {
				Description: "Mode of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"accessibility": {
				Description: "Accessibility of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cpu": {
				Description: "CPU of the database in millicores (m) [1000m = 1 CPU].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"memory": {
				Description: "RAM of the database in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"storage": {
				Description: "Storage of the database in MB [1024MB = 1GB].",
				Type:        types.Int64Type,
				Computed:    true,
			},
		},
	}, nil
}

func (t databaseDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return databaseDataSource{
		client: p.(*provider).GetClient(),
	}, nil
}

type databaseDataSource struct {
	client *qovery.APIClient
}

// Read qovery database data source
func (d databaseDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Get current state
	var data Database
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get database from API
	database, res, err := d.client.DatabaseMainCallsApi.
		GetDatabase(ctx, data.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseReadAPIError(data.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToDatabase(database)
	tflog.Trace(ctx, "read database", "database_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
