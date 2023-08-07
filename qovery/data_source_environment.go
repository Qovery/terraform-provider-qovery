package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &environmentDataSource{}

type environmentDataSource struct {
	environmentService environment.Service
}

func newEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

func (d environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.environmentService = provider.environmentService
}

func (d environmentDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Use this data source to retrieve information about an existing environment.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"project_id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
			"cluster_id": {
				Description: "Id of the cluster.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"mode": {
				Description: "Mode of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this environment.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this environment.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"environment_variable_aliases": {
				Description: "List of environment variable aliases linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable alias.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the environment variable alias.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Name of the variable to alias.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"environment_variable_overrides": {
				Description: "List of environment variable overrides linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable override.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the environment variable override.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the environment variable override.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"secrets": {
				Description: "List of secrets linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the secret [NOTE: will always be empty].",
						Type:        types.StringType,
						Computed:    true,
						Sensitive:   true,
					},
				}),
			},
			"secret_aliases": {
				Description: "List of secret aliases linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret alias.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the secret alias.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Name of the secret to alias.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"secret_overrides": {
				Description: "List of secret overrides linked to this environment.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret override.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Name of the secret override.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the secret override.",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
		},
	}, nil
}

// Read qovery environment data source
func (d environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Environment
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment from API
	env, err := d.environmentService.Get(ctx, data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on environment read", err.Error())
		return
	}

	state := convertDomainEnvironmentToEnvironment(data, env)
	tflog.Trace(ctx, "read environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
