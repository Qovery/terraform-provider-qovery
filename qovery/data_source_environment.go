package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
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

func (r environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing Qovery environment.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Qovery environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Unique identifier of the environment (UUID format).",
				MarkdownDescription: "Unique identifier of the environment (UUID format).",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				Description:         "Identifier of the project containing this environment.",
				MarkdownDescription: "Identifier of the project containing this environment.",
				Computed:            true,
			},
			"cluster_id": schema.StringAttribute{
				Description:         "Identifier of the cluster where this environment is deployed.",
				MarkdownDescription: "Identifier of the cluster where this environment is deployed.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the environment.",
				MarkdownDescription: "Name of the environment.",
				Computed:            true,
			},
			"mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Mode of the environment.",
					clientEnumToStringArray(environment.AllowedModeValues),
					new(environment.DefaultMode.String()),
				),
				MarkdownDescription: descriptions.NewStringEnumDescription(
					"Mode of the environment.",
					clientEnumToStringArray(environment.AllowedModeValues),
					new(environment.DefaultMode.String()),
				),
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clientEnumToStringArray(environment.AllowedModeValues)),
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				Description:         "List of built-in environment variables linked to this environment.",
				MarkdownDescription: "List of built-in environment variables linked to this environment.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable.",
							MarkdownDescription: "Identifier of the environment variable.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Computed:            true,
						},
					},
				},
			},
			"environment_variables": schema.SetNestedAttribute{
				Description:         "Set of environment variables linked to this environment.",
				MarkdownDescription: "Set of environment variables linked to this environment.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable.",
							MarkdownDescription: "Identifier of the environment variable.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description:         "Set of environment variable aliases linked to this environment.",
				MarkdownDescription: "Set of environment variable aliases linked to this environment.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable alias.",
							MarkdownDescription: "Identifier of the environment variable alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable alias.",
							MarkdownDescription: "Name of the environment variable alias.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the variable being aliased.",
							MarkdownDescription: "Name of the variable being aliased.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable alias.",
							MarkdownDescription: "Description of the environment variable alias.",
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description:         "Set of environment variable overrides linked to this environment.",
				MarkdownDescription: "Set of environment variable overrides linked to this environment.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the environment variable override.",
							MarkdownDescription: "Identifier of the environment variable override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable override.",
							MarkdownDescription: "Name of the environment variable override.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Override value of the environment variable.",
							MarkdownDescription: "Override value of the environment variable.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable override.",
							MarkdownDescription: "Description of the environment variable override.",
							Computed:            true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description:         "Set of secrets linked to this environment.",
				MarkdownDescription: "Set of secrets linked to this environment.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret.",
							MarkdownDescription: "Identifier of the secret.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret.",
							MarkdownDescription: "Value of the secret.",
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret.",
							MarkdownDescription: "Description of the secret.",
							Computed:            true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description:         "Set of secret aliases linked to this environment.",
				MarkdownDescription: "Set of secret aliases linked to this environment.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret alias.",
							MarkdownDescription: "Identifier of the secret alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret alias.",
							MarkdownDescription: "Name of the secret alias.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the secret being aliased.",
							MarkdownDescription: "Name of the secret being aliased.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret alias.",
							MarkdownDescription: "Description of the secret alias.",
							Computed:            true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description:         "Set of secret overrides linked to this environment.",
				MarkdownDescription: "Set of secret overrides linked to this environment.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Identifier of the secret override.",
							MarkdownDescription: "Identifier of the secret override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret being overridden.",
							MarkdownDescription: "Name of the secret being overridden.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Override value of the secret.",
							MarkdownDescription: "Override value of the secret.",
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret override.",
							MarkdownDescription: "Description of the secret override.",
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_files": schema.SetNestedAttribute{
				Description:         "List of environment variable files linked to this environment.",
				MarkdownDescription: "List of environment variable files linked to this environment.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable file.",
							MarkdownDescription: "Id of the environment variable file.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable file.",
							MarkdownDescription: "Key of the environment variable file.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable file.",
							MarkdownDescription: "Value of the environment variable file.",
							Computed:            true,
						},
						"mount_path": schema.StringAttribute{
							Description:         "Mount path of the environment variable file.",
							MarkdownDescription: "Mount path of the environment variable file.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable file.",
							MarkdownDescription: "Description of the environment variable file.",
							Computed:            true,
						},
					},
				},
			},
			"secret_files": schema.SetNestedAttribute{
				Description:         "List of secret files linked to this environment.",
				MarkdownDescription: "List of secret files linked to this environment.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret file.",
							MarkdownDescription: "Id of the secret file.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret file.",
							MarkdownDescription: "Key of the secret file.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret file.",
							MarkdownDescription: "Value of the secret file.",
							Computed:            true,
							Sensitive:           true,
						},
						"mount_path": schema.StringAttribute{
							Description:         "Mount path of the secret file.",
							MarkdownDescription: "Mount path of the secret file.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret file.",
							MarkdownDescription: "Description of the secret file.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
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
	env, err := d.environmentService.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on environment read", err.Error())
		return
	}

	state := convertDomainEnvironmentToEnvironment(ctx, data, env)
	tflog.Trace(ctx, "read environment", map[string]any{"environment_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
