package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
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
	envVars := variableListDescriptions("environment_variables", "Helm service")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "Helm service")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "Helm service")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "Helm service")
	secrets := variableListDescriptions("secrets", "Helm service")
	secretAliases := variableListDescriptions("secret_aliases", "Helm service")
	secretOverrides := variableListDescriptions("secret_overrides", "Helm service")
	envVarFiles := variableListDescriptions("environment_variable_files", "Helm service")
	secretFiles := variableListDescriptions("secret_files", "Helm service")
	externalSecrets := variableListDescriptions("external_secrets", "Helm service")
	externalSecretFiles := variableListDescriptions("external_secret_files", "Helm service")
	ports := portDescriptions("Helm service")
	customDomains := customDomainDescriptions("Helm service")
	restrictions := deploymentRestrictionDescriptions("Helm service")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery Helm service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("Helm service"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("Helm service"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("Helm service"),
				Computed:            true,
			},
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: createdFromBlueprintIDDescription("Helm service"),
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("Helm service"),
				Optional:            true,
				Computed:            true,
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: builtInEnvVars.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: builtInEnvVars.Description,
							Computed:            true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				MarkdownDescription: envVars.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVars.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVars.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVars.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVars.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				MarkdownDescription: envVarAliases.List,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarAliases.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				MarkdownDescription: envVarOverrides.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Description,
							Computed:            true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				MarkdownDescription: secrets.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secrets.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secrets.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secrets.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secrets.Description,
							Computed:            true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				MarkdownDescription: secretAliases.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretAliases.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretAliases.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretAliases.Value,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretAliases.Description,
							Computed:            true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				MarkdownDescription: secretOverrides.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretOverrides.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Description,
							Computed:            true,
						},
					},
				},
			},
			"environment_variable_files": schema.SetNestedAttribute{
				MarkdownDescription: envVarFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarFiles.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Value,
							Computed:            true,
						},
						"mount_path": schema.StringAttribute{
							MarkdownDescription: envVarFiles.MountPath,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarFiles.Description,
							Computed:            true,
						},
					},
				},
			},
			"secret_files": schema.SetNestedAttribute{
				MarkdownDescription: secretFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: secretFiles.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: secretFiles.Key,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretFiles.Value,
							Computed:            true,
							Sensitive:           true,
						},
						"mount_path": schema.StringAttribute{
							MarkdownDescription: secretFiles.MountPath,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretFiles.Description,
							Computed:            true,
						},
					},
				},
			},
			"external_secrets": schema.SetNestedAttribute{
				MarkdownDescription: externalSecrets.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{MarkdownDescription: externalSecrets.ID, Computed: true},
						"key":                      schema.StringAttribute{MarkdownDescription: externalSecrets.Key, Computed: true},
						"description":              schema.StringAttribute{MarkdownDescription: externalSecrets.Description, Computed: true},
						"reference":                schema.StringAttribute{MarkdownDescription: externalSecrets.Reference, Computed: true},
						"secret_manager_access_id": schema.StringAttribute{MarkdownDescription: externalSecrets.SecretManagerAccessID, Computed: true},
					},
				},
			},
			"external_secret_files": schema.SetNestedAttribute{
				MarkdownDescription: externalSecretFiles.List,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{MarkdownDescription: externalSecretFiles.ID, Computed: true},
						"key":                      schema.StringAttribute{MarkdownDescription: externalSecretFiles.Key, Computed: true},
						"description":              schema.StringAttribute{MarkdownDescription: externalSecretFiles.Description, Computed: true},
						"mount_path":               schema.StringAttribute{MarkdownDescription: externalSecretFiles.MountPath, Computed: true},
						"reference":                schema.StringAttribute{MarkdownDescription: externalSecretFiles.Reference, Computed: true},
						"secret_manager_access_id": schema.StringAttribute{MarkdownDescription: externalSecretFiles.SecretManagerAccessID, Computed: true},
					},
				},
			},
			"external_host": schema.StringAttribute{
				MarkdownDescription: externalHostDescription("Helm service"),
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("Helm service"),
				Computed:            true,
			},
			"deployment_stage_id": schema.StringAttribute{
				MarkdownDescription: deploymentStageIDDescription,
				Optional:            true,
				Computed:            true,
			},
			"is_skipped": schema.BoolAttribute{
				MarkdownDescription: isSkippedDescription,
				Optional:            true,
				Computed:            true,
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("Helm service"),
				Optional:            true,
				Computed:            true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				MarkdownDescription: restrictions.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: restrictions.ID,
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: restrictions.Mode,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: restrictions.Type,
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: restrictions.Value,
							Computed:            true,
						},
					},
				},
			},
			"timeout_sec": schema.Int64Attribute{
				MarkdownDescription: helmTimeoutSecDescription,
				Optional:            true,
				Computed:            true,
			},
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: autoPreviewDescription("Helm service"),
				Optional:            true,
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: helmAutoDeployDescription,
				Optional:            true,
				Computed:            true,
			},
			"arguments": schema.ListAttribute{
				MarkdownDescription: helmArgumentsDescription,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"allow_cluster_wide_resources": schema.BoolAttribute{
				MarkdownDescription: helmAllowClusterWideResourcesDescription,
				Computed:            true,
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: helmSourceDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"helm_repository": schema.SingleNestedAttribute{
						MarkdownDescription: helmSourceHelmRepositoryDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"helm_repository_id": schema.StringAttribute{
								MarkdownDescription: helmSourceHelmRepositoryIDDesc,
								Required:            true,
							},
							"chart_name": schema.StringAttribute{
								MarkdownDescription: helmSourceChartNameDescription,
								Required:            true,
							},
							"chart_version": schema.StringAttribute{
								MarkdownDescription: helmSourceChartVersionDescription,
								Required:            true,
							},
						},
					},
					"git_repository": schema.SingleNestedAttribute{
						MarkdownDescription: helmSourceGitRepositoryDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								MarkdownDescription: gitRepositoryURLDescription,
								Required:            true,
							},
							"branch": schema.StringAttribute{
								MarkdownDescription: helmSourceBranchDescription,
								Optional:            true,
								Computed:            true,
							},
							"root_path": schema.StringAttribute{
								MarkdownDescription: helmSourceRootPathDescription,
								Optional:            true,
								Computed:            true,
							},
							"git_token_id": schema.StringAttribute{
								MarkdownDescription: gitRepositoryTokenIDDescription,
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
			"values_override": schema.SingleNestedAttribute{
				MarkdownDescription: helmValuesOverrideDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"set": schema.MapAttribute{
						MarkdownDescription: helmValuesSetDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_string": schema.MapAttribute{
						MarkdownDescription: helmValuesSetStringDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"set_json": schema.MapAttribute{
						MarkdownDescription: helmValuesSetJSONDescription,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"file": schema.SingleNestedAttribute{
						MarkdownDescription: helmValuesFileDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"raw": schema.MapNestedAttribute{
								MarkdownDescription: helmValuesRawDescription,
								Optional:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"content": schema.StringAttribute{
											MarkdownDescription: helmValuesRawContentDescription,
											Required:            true,
										},
									},
								},
							},
							"git_repository": schema.SingleNestedAttribute{
								MarkdownDescription: helmValuesGitRepositoryDescription,
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Required:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: helmValuesGitBranchDescription,
										Required:            true,
									},
									"paths": schema.SetAttribute{
										MarkdownDescription: helmValuesGitPathsDescription,
										Required:            true,
										ElementType:         types.StringType,
									},
									"git_token_id": schema.StringAttribute{
										MarkdownDescription: gitRepositoryTokenIDDescription,
										Optional:            true,
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
			"ports": schema.MapNestedAttribute{
				MarkdownDescription: helmPortsDescription,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_name": schema.StringAttribute{
							MarkdownDescription: helmPortServiceNameDescription,
							Required:            true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: helmPortNamespaceDescription,
							Optional:            true,
						},
						"internal_port": schema.Int64Attribute{
							MarkdownDescription: helmPortInternalPortDescription,
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"external_port": schema.Int64Attribute{
							MarkdownDescription: helmPortExternalPortDescription,
							Required:            true,
							Validators: []validator.Int64{
								validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
							},
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: ports.Protocol,
							Validators: []validator.String{
								validators.NewStringEnumValidator(helmPortProtocols),
							},
							Optional: true,
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: dataSourcePortIsDefaultDescription("Helm service"),
							Required:            true,
						},
					},
				},
			},
			"custom_domains": schema.SetNestedAttribute{
				MarkdownDescription: customDomains.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: customDomains.ID,
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: customDomains.Domain,
							Computed:            true,
						},
						"validation_domain": schema.StringAttribute{
							MarkdownDescription: customDomains.ValidationDomain,
							Computed:            true,
						},
						"generate_certificate": schema.BoolAttribute{
							MarkdownDescription: customDomains.GenerateCertificate,
							Optional:            true,
						},
						"use_cdn": schema.BoolAttribute{
							MarkdownDescription: customDomains.UseCDN,
							Optional:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: customDomains.Status,
							Computed:            true,
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
	tflog.Trace(ctx, "read helm", map[string]any{"helm_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
