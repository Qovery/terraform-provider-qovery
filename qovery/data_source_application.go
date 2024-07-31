package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &applicationDataSource{}

type applicationDataSource struct {
	client *client.Client
}

func newApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

func (d applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = provider.client
}

func (r applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery application resource. This can be used to create and manage Qovery applications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the application.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the application.",
				Computed:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the application.",
				Optional:    true,
				Computed:    true,
			},
			"git_repository": schema.SingleNestedAttribute{
				Description: "Git repository of the application.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the git repository.",
						Computed:    true,
					},
					"branch": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Branch of the git repository.",
							applicationGitRepositoryBranchDefault,
						),
						Optional: true,
						Computed: true,
					},
					"root_path": schema.StringAttribute{
						Description: descriptions.NewStringDefaultDescription(
							"Root path of the application.",
							applicationGitRepositoryRootPathDefault,
						),
						Optional: true,
						Computed: true,
					},
					"git_token_id": schema.StringAttribute{
						Description: "The git token ID to be used",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"build_mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Build Mode of the application.",
					applicationBuildModes,
					&applicationBuildModeDefault,
				),
				Computed: true,
			},
			"dockerfile_path": schema.StringAttribute{
				Description: "Dockerfile Path of the application.\n\t- Required if: `build_mode=\"DOCKER\"`.",
				Optional:    true,
				Computed:    true,
			},
			"buildpack_language": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Buildpack Language framework.\n\t- Required if: `build_mode=\"BUILDPACKS\"`.",
					applicationBuildPackLanguages,
					nil,
				),
				Optional: true,
				Computed: true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the application in millicores (m) [1000m = 1 CPU].",
					applicationCPUMin,
					&applicationCPUDefault,
				),
				Optional: true,
				Computed: true,
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the application in MB [1024MB = 1GB].",
					applicationMemoryMin,
					&applicationMemoryDefault,
				),
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMemoryMin},
				},
			},
			"min_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Minimum number of instances running for the application.",
					applicationMinRunningInstancesMin,
					&applicationMinRunningInstancesDefault,
				),
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: applicationMinRunningInstancesMin},
				},
			},
			"max_running_instances": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Maximum number of instances running for the application.",
					applicationMaxRunningInstancesMin,
					&applicationMaxRunningInstancesDefault,
				),
				Optional: true,
				Computed: true,
			},
			"auto_preview": schema.BoolAttribute{
				Description: descriptions.NewBoolDefaultDescription(
					"Specify if the environment preview option is activated or not for this application.",
					applicationAutoPreviewDefault,
				),
				Optional: true,
				Computed: true,
			},
			"entrypoint": schema.StringAttribute{
				Description: "Entrypoint of the application.",
				Optional:    true,
				Computed:    true,
			},
			"arguments": schema.ListAttribute{
				Description: "List of arguments of this application.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"storage": schema.SetNestedAttribute{
				Description: "List of storages linked to this application.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the storage.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Type of the storage for the application.",
								clientEnumToStringArray(storage.AllowedTypeValues),
								nil,
							),
							Computed: true,
						},
						"size": schema.Int64Attribute{
							Description: descriptions.NewInt64MinDescription(
								"Size of the storage for the application in GB [1024MB = 1GB].",
								applicationStorageSizeMin,
								nil,
							),
							Computed: true,
						},
						"mount_point": schema.StringAttribute{
							Description: "Mount point of the storage for the application.",
							Computed:    true,
						},
					},
				},
			},
			"ports": schema.SetNestedAttribute{
				Description: "List of ports linked to this application.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the port.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the port.",
							Optional:    true,
							Computed:    true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the application.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Computed: true,
						},
						"external_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"External port of the application.\n\t- Required if: `ports.publicly_accessible=true`.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Computed: true,
						},
						"publicly_accessible": schema.BoolAttribute{
							Description: "Specify if the port is exposed to the world or not for this application.",
							Computed:    true,
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the application.",
								clientEnumToStringArray(port.AllowedProtocolValues),
								pointer.ToString(port.DefaultProtocol.String()),
							),
							Optional: true,
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Description: "If this port will be used for the root domain",
							Computed:    true,
						},
					},
				},
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this application.",
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this application.",
				Optional:    true,
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
						"description": schema.StringAttribute{
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description: "List of environment variable aliases linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable alias.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the variable to alias.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the environment variable alias.",
							Computed:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the environment variable override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the environment variable override.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the environment variable override.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the environment variable override.",
							Computed:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the secret.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret.",
							Computed:    true,
							Sensitive:   true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the secret.",
							Computed:    true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret alias.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Name of the secret to alias.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the secret alias.",
							Computed:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the secret override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Name of the secret override.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the secret override.",
							Computed:    true,
							Sensitive:   true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the secret override.",
							Computed:    true,
						},
					},
				},
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"custom_domains": schema.SetNestedAttribute{
				Description: "List of custom domains linked to this application.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the custom domain.",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "Your custom domain.",
							Computed:    true,
						},
						"generate_certificate": schema.BoolAttribute{
							Description: "Qovery will generate and manage the certificate for this domain.",
							Optional:    true,
						},
						"use_cdn": schema.BoolAttribute{
							Description: "Indicates if the custom domain is behind a CDN (i.e Cloudflare).\n" +
								"This will condition the way we are checking CNAME before & during a deployment:\n" +
								" * If `true` then we only check the domain points to an IP\n" +
								" * If `false` then we check that the domain resolves to the correct service Load Balancer",
							Optional: true,
						},
						"validation_domain": schema.StringAttribute{
							Description: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the custom domain.",
							Computed:    true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The application external FQDN host [NOTE: only if your application is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The application internal host.",
				Computed:    true,
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings.",
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: " Specify if the application will be automatically updated after receiving a new image tag.",
				Optional:    true,
				Computed:    true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description: "List of deployment restrictions",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Id of the deployment restriction",
							Computed:    true,
						},
						"mode": schema.StringAttribute{
							Description: "Can be EXCLUDE or MATCH",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Currently, only PATH is accepted",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the deployment restriction",
							Computed:    true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Read qovery application data source
func (d applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Application
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get application from API
	application, apiErr := d.client.GetApplication(ctx, data.Id.ValueString(), data.AdvancedSettingsJson.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	state := convertResponseToApplication(ctx, data, application)
	tflog.Trace(ctx, "read application", map[string]interface{}{"application_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
