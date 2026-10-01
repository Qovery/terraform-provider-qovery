package qovery

import (
	"context"
	"fmt"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &helmRepositoryDataSource{}

type helmRepositoryDataSource struct {
	helmRepositoryService helmRepository.Service
}

func newhelmRepositoryDataSource() datasource.DataSource {
	return &helmRepositoryDataSource{}
}

func (d helmRepositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm_repository"
}

func (d *helmRepositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.helmRepositoryService = provider.helmRepositoryService
}

func (r helmRepositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery helm repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("helm repository"),
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription,
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("helm repository"),
				Optional:            true,
				Computed:            true,
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: helmRepositoryKindDescription,
				Optional:            true,
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: helmRepositoryURLDescription,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("helm repository"),
				Optional:            true,
				Computed:            true,
			},
			"skip_tls_verification": schema.BoolAttribute{
				MarkdownDescription: helmRepositorySkipTLSVerificationDescription,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d helmRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data HelmRepositoryDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get helm repository from API
	reg, err := d.helmRepositoryService.Get(ctx, data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on container registry read", err.Error())
		return
	}

	state := convertDomainHelmRepositoryToHelmRepositoryDataSource(reg)
	tflog.Trace(ctx, "read container registry", map[string]any{"container_registry_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
