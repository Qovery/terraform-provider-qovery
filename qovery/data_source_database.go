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
var _ datasource.DataSource = databaseDataSource{}

type databaseDataSource struct {
	client *client.Client
}

func newDatabaseDataSource() datasource.DataSource {
	return &databaseDataSource{}
}

func (d databaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *databaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d databaseDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"deployment_stage_id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}, nil
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
