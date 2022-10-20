package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSource = environmentDataSource{}

type environmentDataSource struct {
	client *client.Client
}

func NewEnvironmentDataSource(apiClient *client.Client) func() datasource.DataSource {
	return func() datasource.DataSource {
		return environmentDataSource{
			client: apiClient,
		}
	}
}

func (d environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
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
	environment, apiErr := d.client.GetEnvironment(ctx, data.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToEnvironment(data, environment)
	tflog.Trace(ctx, "read environment", map[string]interface{}{"environment_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
