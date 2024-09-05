package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &helmDataSource{}

type helmDataSource struct {
	helmService helm.Service
}

func newHelmDataSource() datasource.DataSource {
	return &helmDataSource{}
}

func (d helmDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm"
}

func (d *helmDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.helmService = provider.helmService
}

func (d helmDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery helm resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the helm.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the helm.",
				Computed:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the helm service.",
				Optional:    true,
				Computed:    true,
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this helm.",
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
				Description: "List of environment variables linked to this helm.",
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
				Description: "List of environment variable aliases linked to this helm.",
				Optional:    true,
				Computed:    true,
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
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this helm.",
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
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this helm.",
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
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this helm.",
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
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this helm.",
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
							Description: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The helm external FQDN host [NOTE: only if your helm is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The helm internal host.",
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
			"timeout_sec": schema.Int64Attribute{
				Description: "Helm timeout in seconds",
				Optional:    true,
				Computed:    true,
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this helm.",
				Optional:    true,
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: " Specify the service will be automatically updated on every new commit on the branch.",
				Optional:    true,
				Computed:    true,
			},
			"arguments": schema.SetAttribute{
				Description: "Helm arguments",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"allow_cluster_wide_resources": schema.BoolAttribute{
				Description: "Allow this chart to deploy resources outside of this environment namespace (including CRDs or non-namespaced resources)",
				Computed:    true,
			},
			"source": schema.SingleNestedAttribute{
				Description: "Helm chart from a Helm repository or from a git repository",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"helm_repository": schema.SingleNestedAttribute{
						Description: "Helm repositories can be private or public",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"helm_repository_id": schema.StringAttribute{
								Description: "helm repository id",
								Required:    true,
							},
							"chart_name": schema.StringAttribute{
								Description: "Chart name",
								Required:    true,
							},
							"chart_version": schema.StringAttribute{
								Description: "Chart version",
								Required:    true,
							},
						},
					},
					"git_repository": schema.SingleNestedAttribute{
						Description: "Git repository",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Helm's source git repository URL",
								Required:    true,
							},
							"branch": schema.StringAttribute{
								Description: "Helm's source git repository branch",
								Optional:    true,
								Computed:    true,
							},
							"root_path": schema.StringAttribute{
								Description: "Helm's source git repository root path",
								Optional:    true,
								Computed:    true,
							},
							"git_token_id": schema.StringAttribute{
								Description: "The git token ID to be used",
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				Description: "Define your own overrides to customize the helm chart behaviour.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"set": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"set_string": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"set_json": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"file": schema.SingleNestedAttribute{
						Description: "Define overrides by selecting a YAML file from a git repository (preferred) or by passing raw YAML files.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"raw": schema.MapNestedAttribute{
								Description: "Raw YAML files",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"content": schema.StringAttribute{
											Description: "content of the file",
											Required:    true,
										},
									},
								},
							},
							"git_repository": schema.SingleNestedAttribute{
								Description: "YAML file from a git repository",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description: "YAML file git repository URL",
										Required:    true,
									},
									"branch": schema.StringAttribute{
										Description: "YAML file git repository branch",
										Required:    true,
									},
									"paths": schema.SetAttribute{
										Description: "YAML files git repository paths",
										Required:    true,
										ElementType: types.StringType,
									},
									"git_token_id": schema.StringAttribute{
										Description: "The git token ID to be used",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"ports": schema.MapNestedAttribute{
				Description: "List of ports linked to this helm.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							Description: "",
							Required:    true,
						},
						"namespace": schema.StringAttribute{
							Description: "",
							Optional:    true,
						},
						"internal_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"Internal port of the container.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"external_port": schema.Int64Attribute{
							Description: descriptions.NewInt64MinMaxDescription(
								"External port of the container.\n\t- Required if: `ports.publicly_accessible=true`.",
								port.MinPort,
								port.MaxPort,
								nil,
							),
							Required: true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"protocol": schema.StringAttribute{
							Description: descriptions.NewStringEnumDescription(
								"Protocol used for the port of the container.",
								helmPortProtocols,
								pointer.ToString(helm.DefaultProtocol.String()),
							),
							Validators: []validator.String{
								validators.NewStringEnumValidator(helmPortProtocols),
							},
							Optional: true,
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Description: "If this port will be used for the root domain",
							Required:    true,
						},
					},
				},
			},
			"custom_domains": schema.SetNestedAttribute{
				Description: "List of custom domains linked to this container.",
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
						"validation_domain": schema.StringAttribute{
							Description: "URL provided by Qovery. You must create a CNAME on your DNS provider using that URL.",
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
						"status": schema.StringAttribute{
							Description: "Status of the custom domain.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read qovery helm data source
func (d helmDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Helm
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get helm from API
	h, err := d.helmService.Get(ctx, data.ID.ValueString(), data.AdvancedSettingsJson.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error on helm read", err.Error())
		return
	}

	state := convertDomainHelmToHelm(ctx, data, h)
	tflog.Trace(ctx, "read helm", map[string]interface{}{"helm_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
