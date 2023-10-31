package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
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

func (d gitTokenDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing git token.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the git token.",
				Type:        types.StringType,
				Computed:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the git token.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: "Description of the git token.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description: descriptions.NewStringEnumDescription(
					"Type of the git token.",
					gitTokenTypes,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(gitTokenTypes),
				},
			},
			"bitbucket_workspace": {
				Description: "(Mandatory only for Bitbucket git token) Workspace where the token has permissions .",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"token": {
				Description: "Value of the git token.",
				Type:        types.StringType,
				Required:    true,
				Sensitive:   true,
			},
		},
	}, nil
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
	response, err := d.service.Get(ctx, data.OrganizationId.Value, data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on git token read", err.Error())
		return
	}

	state := toTerraformObject(data.OrganizationId.Value, data.Token.Value, *response)
	tflog.Trace(ctx, "read git token", map[string]interface{}{"git_token_id": state.ID.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
