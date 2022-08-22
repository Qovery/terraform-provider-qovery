package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ provider.DataSourceType = databaseDataSourceType{}
var _ datasource.DataSource = databaseDataSource{}

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
				Description: "CPU of the database in milli-cores (m) [1000m = 1 CPU].",
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
			"external_host": {
				Description: "The database external FQDN host (only if your database is publicly accessible with ACCESSIBILITY = PUBLIC)",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The database internal host (Recommended for your application)",
				Type:        types.StringType,
				Computed:    true,
			},
			"port": {
				Description: "The port to connect to your database",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"login": {
				Description: "The login to connect to your database",
				Type:        types.StringType,
				Computed:    true,
			},
			"password": {
				Description: "The password to connect to your database",
				Type:        types.StringType,
				Computed:    true,
			},
			"state": {
				Description: "State of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (t databaseDataSourceType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return databaseDataSource{
		client: p.(*qProvider).client,
	}, nil
}

type databaseDataSource struct {
	client *client.Client
}

// Read qovery database data source
func (d databaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Database
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get database from API
	database, apiErr := d.client.GetDatabase(ctx, data.Id.Value)
	if apiErr != nil {
		return
	}

	state := convertResponseToDatabase(database)
	tflog.Trace(ctx, "read database", map[string]interface{}{"database_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
