package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ datasource.DataSourceWithConfigure = &jobDataSource{}

type jobDataSource struct {
	jobService job.Service
}

func newJobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

func (d jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.jobService = provider.jobService
}

func (d jobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "job")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "job")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "job")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "job")
	secrets := variableListDescriptions("secrets", "job")
	secretAliases := variableListDescriptions("secret_aliases", "job")
	secretOverrides := variableListDescriptions("secret_overrides", "job")
	envVarFiles := variableListDescriptions("environment_variable_files", "job")
	secretFiles := variableListDescriptions("secret_files", "job")
	externalSecrets := variableListDescriptions("external_secrets", "job")
	externalSecretFiles := variableListDescriptions("external_secret_files", "job")
	restrictions := deploymentRestrictionDescriptions("job")

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Qovery job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("job"),
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("job"),
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("job"),
				Optional:            true,
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: cpuDescription("job"),
				Optional:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: memoryDescription("job"),
				Optional:            true,
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: ephemeralStorageDescription("job"),
				Computed:            true,
			},
			"max_duration_seconds": schema.Int64Attribute{
				MarkdownDescription: jobMaxDurationSecondsDescription,
				Optional:            true,
				Computed:            true,
			},
			"max_nb_restart": schema.Int64Attribute{
				MarkdownDescription: jobMaxNbRestartDescription,
				Optional:            true,
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: jobPortDescription,
				Computed:            true,
				Optional:            true,
			},
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: autoPreviewDescription("job"),
				Optional:            true,
				Computed:            true,
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: jobScheduleDescription,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnStartDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription,
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnStopDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription,
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnDeleteDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription,
								Optional:            true,
								Computed:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"lifecycle_type": schema.StringAttribute{
						MarkdownDescription: jobLifecycleTypeDescription,
						Optional:            true,
						Computed:            true,
					},
					"cronjob": schema.SingleNestedAttribute{
						MarkdownDescription: jobCronJobDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								MarkdownDescription: jobCronJobScheduleDescription,
								Computed:            true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								MarkdownDescription: jobCronJobCommandDescription,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"entrypoint": schema.StringAttribute{
										MarkdownDescription: entrypointDescription,
										Optional:            true,
										Computed:            true,
									},
									"arguments": schema.ListAttribute{
										MarkdownDescription: argumentsDescription,
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: jobSourceDescription,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						MarkdownDescription: jobSourceImageDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								MarkdownDescription: containerImageRegistryIDDescription,
								Computed:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: containerImageNameDescription,
								Computed:            true,
							},
							"tag": schema.StringAttribute{
								MarkdownDescription: containerImageTagDescription,
								Computed:            true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						MarkdownDescription: jobSourceDockerDescription,
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"dockerfile_path": schema.StringAttribute{
								MarkdownDescription: dockerfilePathDescription,
								Optional:            true,
							},
							"dockerfile_raw": schema.StringAttribute{
								MarkdownDescription: jobDockerfileRawDescription,
								Optional:            true,
							},
							"git_repository": schema.SingleNestedAttribute{
								MarkdownDescription: jobGitRepositoryDescription,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Computed:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: jobGitRepositoryBranchDescription,
										Computed:            true,
									},
									"root_path": schema.StringAttribute{
										MarkdownDescription: gitRepositoryRootPathDescription,
										Optional:            true,
										Computed:            true,
									},
									"git_token_id": schema.StringAttribute{
										MarkdownDescription: gitRepositoryTokenIDDescription,
										Optional:            true,
										Computed:            false,
									},
								},
							},
							"docker_target_build_stage": schema.StringAttribute{
								MarkdownDescription: dockerTargetBuildStageDescription,
								Optional:            true,
							},
						},
					},
				},
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
				MarkdownDescription: jobHostDescription,
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: jobHostDescription,
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
				MarkdownDescription: dataSourceAdvancedSettingsJSONDescription("job"),
				Optional:            true,
				Computed:            true,
			},
			"build_settings": buildSettingsDataSourceSchemaAttributes(),
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: jobAutoDeployDescription,
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
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("annotations", "job"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: dataSourceGroupIDsDescription("labels", "job"),
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Read qovery job data source
func (d jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var data Job
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job from API
	cont, err := d.jobService.Get(ctx, data.ID.ValueString(), data.AdvancedSettingsJson.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	// Group ids report the API value; a data source has no plan to match, so none reads as [].
	data.AnnotationsGroupIds = emptyStringSet()
	data.LabelssGroupIds = emptyStringSet()
	state := convertDomainJobToJob(ctx, data, cont)
	state.BuildSettings = buildSettingsFromQovery(cont.BuildSettings)
	tflog.Trace(ctx, "read job", map[string]any{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
