package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (d databaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("database"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("database"),
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("database"),
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: databaseTypeDescription,
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: databaseVersionDescription,
				Computed:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: databaseModeDescription,
				Computed:            true,
			},
			"accessibility": schema.StringAttribute{
				MarkdownDescription: databaseAccessibilityDescription,
				Optional:            true,
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: databaseInstanceTypeDescription,
				Optional:            true,
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: cpuDescription("database"),
				Optional:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: memoryDescription("database"),
				Optional:            true,
			},
			"storage": schema.Int64Attribute{
				MarkdownDescription: databaseStorageDescription,
				Optional:            true,
			},
			"external_host": schema.StringAttribute{
				MarkdownDescription: databaseExternalHostDescription,
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("database"),
				Computed:            true,
			},
			"deployment_stage_id": schema.StringAttribute{
				MarkdownDescription: deploymentStageIDDescription,
				Optional:            true,
				Computed:            true,
			},
			"is_skipped": schema.BoolAttribute{
				MarkdownDescription: isSkippedDescription,
				Optional:            true,
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: databasePortDescription,
				Computed:            true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: databaseLoginDescription,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: databasePasswordDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("annotations", "database"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("labels", "database"),
				Computed:            true,
				ElementType:         types.StringType,
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

	// Group ids report the API value; a data source has no plan to match, so none reads as [].
	data.AnnotationsGroupIds = emptyStringSet()
	data.LabelsGroupIds = emptyStringSet()
	state := convertResponseToDatabase(data, database)
	tflog.Trace(ctx, "read database", map[string]any{"database_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
