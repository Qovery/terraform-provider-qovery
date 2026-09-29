package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &apiTokenDataSource{}

type apiTokenDataSource struct {
	service apitoken.Service
}

func newApiTokenDataSource() datasource.DataSource {
	return &apiTokenDataSource{}
}

func (d apiTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (d *apiTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.service = provider.apiTokenService
}

func (d apiTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery API token, without its value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("API token"),
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription,
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("API token"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("API token"),
				Computed:            true,
			},
			"role_id": schema.StringAttribute{
				MarkdownDescription: apiTokenRoleIDDescription,
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: apiTokenTokenDescription + " Always `null`: the API returns it only at creation.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// Read qovery api token data source
func (d apiTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data ApiToken
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get api token from API
	apiToken, err := d.service.Get(ctx, data.OrganizationId.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on api token read", err.Error())
		return
	}

	state := convertDomainApiTokenToApiToken(*apiToken, data.Token)
	tflog.Trace(ctx, "read api token", map[string]any{"api_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
