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

func (d databaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery database resource. This can be used to create and manage Qovery databases.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the database.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the database.",
				Computed:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the database.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Type of the database [NOTE: can't be updated after creation].",
					databaseTypes,
					nil,
				),
				Computed: true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the database",
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Mode of the database [NOTE: can't be updated after creation].",
					databaseModes,
					nil,
				),
				Computed: true,
			},
			"accessibility": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Accessibility of the database.",
					databaseAccessibilities,
					&databaseAccessibilityDefault,
				),
				Optional: true,
			},
			"instance_type": schema.StringAttribute{
				Description: "Instance type of the database.",
				Optional:    true,
				Computed:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the database in millicores (m) [1000m = 1 CPU].",
					databaseCPUMin,
					&databaseCPUDefault,
				),
				Optional: true,
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the database in MB [1024MB = 1GB].",
					databaseMemoryMin,
					&databaseMemoryDefault,
				),
				Optional: true,
			},
			"storage": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Storage of the database in GB [1024MB = 1GB] [NOTE: can't be updated after creation].",
					databaseStorageMin,
					&databaseStorageDefault,
				),
				Optional: true,
			},
			"external_host": schema.StringAttribute{
				Description: "The database external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The database internal host (Recommended for your application)",
				Computed:    true,
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port to connect to your database",
				Computed:    true,
			},
			"login": schema.StringAttribute{
				Description: "The login to connect to your database",
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password to connect to your database",
				Computed:    true,
			},
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
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
	database, apiErr := d.client.GetDatabase(ctx, data.Id.ValueString())
	if apiErr != nil {
		return
	}

	state := convertResponseToDatabase(ctx, data, database)
	tflog.Trace(ctx, "read database", map[string]interface{}{"database_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
