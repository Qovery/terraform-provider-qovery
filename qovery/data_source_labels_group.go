package qovery

import (
	"context"
	"fmt"
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &labelsGroupDataSource{}

type labelsGroupDataSource struct {
	labelsGroupService labels_group.Service
}

func newLabelsGroupDataSource() datasource.DataSource {
	return &labelsGroupDataSource{}
}

func (d labelsGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_labels_group"
}

func (d *labelsGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.labelsGroupService = provider.labelsGroupService
}

func (d labelsGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing Qovery labels group.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Qovery labels group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique identifier of the labels group (UUID format).",
				MarkdownDescription: "Unique identifier of the labels group (UUID format).",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization.",
				MarkdownDescription: "Id of the organization.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the labels group.",
				MarkdownDescription: "Name of the labels group.",
				Optional:            true,
			},
			"labels": schema.SetNestedAttribute{
				Description:         "Set of labels included in this group.",
				MarkdownDescription: "Set of labels included in this group.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description:         "Key of the label.",
							MarkdownDescription: "Key of the label.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the label.",
							MarkdownDescription: "Value of the label.",
							Required:            true,
						},
						"propagate_to_cloud_provider": schema.BoolAttribute{
							Description:         "Whether this label is propagated to the underlying cloud provider resources.",
							MarkdownDescription: "Whether this label is propagated to the underlying cloud provider resources.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (d labelsGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data LabelsGroup
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get labels Group from API
	h, err := d.labelsGroupService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group read", err.Error())
		return
	}

	state := convertResponseToLabelsGroup(ctx, data, h)
	tflog.Trace(ctx, "read label group", map[string]any{"label_group_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
