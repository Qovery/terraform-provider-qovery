package qovery

import (
	"context"
	"fmt"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &annotationsGroupDataSource{}

type annotationsGroupDataSource struct {
	annotationsGroupService annotations_group.Service
}

func newAnnotationsGroupDataSource() datasource.DataSource {
	return &annotationsGroupDataSource{}
}

func (d annotationsGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_group"
}

func (d *annotationsGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.annotationsGroupService = provider.annotationsGroupService
}

func (d annotationsGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery annotations group resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the annotations group",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "name of the annotations group",
				Optional:    true,
			},
			"annotations": schema.MapAttribute{
				Description: "annotations",
				Optional:    true,
				ElementType: types.StringType,
			},
			"scopes": schema.SetAttribute{
				Description: "scopes of the annotations group",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d annotationsGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data AnnotationsGroup
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get annotations Group from API
	h, err := d.annotationsGroupService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group read", err.Error())
		return
	}

	state := convertResponseToAnnotationsGroup(ctx, data, h)
	tflog.Trace(ctx, "read annotation group", map[string]any{"annotation_group_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
