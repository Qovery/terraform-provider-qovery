package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &gitTokenDataSource{}

type gitTokenDataSource struct {
	service gittoken.Service
}

func newGitTokenDataSource() datasource.DataSource {
	return &gitTokenDataSource{}
}
func (d gitTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_token"
}

func (d *gitTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.service = provider.gitTokenService
}

func (r gitTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery git token resource. This can be used to create and manage Qovery git token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the git token.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the git token.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the git token.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Type of the git token.",
					gitTokenTypes,
					nil,
				),
				Computed: true,
			},
			"bitbucket_workspace": schema.StringAttribute{
				Description: "(Mandatory only for Bitbucket git token) Workspace where the token has permissions .",
				Optional:    true,
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "Value of the git token.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Read qovery git token data source
func (d gitTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data GitToken
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get git token from API
	response, err := d.service.Get(ctx, data.OrganizationId.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on git token read", err.Error())
		return
	}

	state := toTerraformObject(data.OrganizationId.ValueString(), data.Token.ValueString(), *response)
	tflog.Trace(ctx, "read git token", map[string]interface{}{"git_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
