package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
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
		Description: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "Id of the project.",
				Required:    true,
			},
			"cluster_id": schema.StringAttribute{
				Description: "Id of the cluster [NOTE: can't be updated after creation].",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the environment.",
				Required:    true,
			},
			"mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Mode of the environment [NOTE: can't be updated after creation].",
					clientEnumToStringArray(environment.AllowedModeValues),
					pointer.ToString(environment.DefaultMode.String()),
				),
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(clientEnumToStringArray(environment.AllowedModeValues)),
				},
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this environment.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the environment variable.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the environment variable.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable.",
							Required:    true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description: "List of environment variable aliases linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the variable to alias.",
							Required:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{

					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable override.",
							Required:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{

					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the secret.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret.",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{

					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the secret to alias.",
							Required:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this environment.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret override.",
							Required:    true,
							Sensitive:   true,
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
	tflog.Trace(ctx, "read environment", map[string]interface{}{"environment_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
