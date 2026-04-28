package qovery

import (
	"context"
	"fmt"

	"github.com/qovery/qovery-client-go"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"

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
	resp.Schema = schema.Schema{
		Description:         "Provides a Qovery job data source. This can be used to read existing Qovery jobs (cron jobs and lifecycle jobs).",
		MarkdownDescription: "Provides a Qovery job data source. This can be used to read existing Qovery jobs (cron jobs and lifecycle jobs).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the job.",
				MarkdownDescription: "Id of the job.",
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				Description:         "Id of the environment.",
				MarkdownDescription: "Id of the environment.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         "Name of the job.",
				MarkdownDescription: "Name of the job.",
				Computed:            true,
			},
			"icon_uri": schema.StringAttribute{
				Description:         "Icon URI representing the job.",
				MarkdownDescription: "Icon URI representing the job.",
				Optional:            true,
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the job in millicores (m) [1000m = 1 CPU].",
					job.MinCPU,
					pointer.ToInt64(job.DefaultCPU),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"CPU of the job in millicores (m) [1000m = 1 CPU].",
					job.MinCPU,
					pointer.ToInt64(job.DefaultCPU),
				),
				Optional: true,
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the job in MB [1024MB = 1GB].",
					job.MinMemory,
					pointer.ToInt64(job.DefaultMemory),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"RAM of the job in MB [1024MB = 1GB].",
					job.MinMemory,
					pointer.ToInt64(job.DefaultMemory),
				),
				Optional: true,
			},
			"max_duration_seconds": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Job's max duration in seconds.",
					int64(job.MinDurationSeconds),
					pointer.ToInt64(int64(job.DefaultMaxDurationSeconds)),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"Job's max duration in seconds.",
					int64(job.MinDurationSeconds),
					pointer.ToInt64(int64(job.DefaultMaxDurationSeconds)),
				),
				Optional: true,
				Computed: true,
			},
			"max_nb_restart": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Job's max number of restarts.",
					int64(job.MinNbRestart),
					pointer.ToInt64(int64(job.DefaultMaxNbRestart)),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"Job's max number of restarts.",
					int64(job.MinNbRestart),
					pointer.ToInt64(int64(job.DefaultMaxNbRestart)),
				),
				Optional: true,
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Description: descriptions.NewInt64MinMaxDescription(
					"Job's probes port.",
					port.MinPort,
					port.MaxPort,
					nil,
				),
				MarkdownDescription: descriptions.NewInt64MinMaxDescription(
					"Job's probes port.",
					port.MinPort,
					port.MaxPort,
					nil,
				),
				Computed: true,
				Optional: true,
			},
			"auto_preview": schema.BoolAttribute{
				Description:         "Specify if the environment preview option is activated or not for this job.",
				MarkdownDescription: "Specify if the environment preview option is activated or not for this job.",
				Optional:            true,
				Computed:            true,
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"schedule": schema.SingleNestedAttribute{
				Description:         "Job's schedule configuration. Use on_start, on_stop, and on_delete for lifecycle jobs, or cronjob for cron jobs.",
				MarkdownDescription: "Job's schedule configuration. Use `on_start`, `on_stop`, and `on_delete` for lifecycle jobs, or `cronjob` for cron jobs.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment starts.",
						MarkdownDescription: "Lifecycle job event: executed when the environment starts.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment stops.",
						MarkdownDescription: "Lifecycle job event: executed when the environment stops.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment is deleted.",
						MarkdownDescription: "Lifecycle job event: executed when the environment is deleted.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:            true,
								Computed:            true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:            true,
								Computed:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"lifecycle_type": schema.StringAttribute{
						Description: descriptions.NewStringEnumDescription(
							"Type of the lifecycle job.",
							clientEnumToStringArray(qovery.AllowedJobLifecycleTypeEnumEnumValues),
							nil,
						),
						MarkdownDescription: descriptions.NewStringEnumDescription(
							"Type of the lifecycle job.",
							clientEnumToStringArray(qovery.AllowedJobLifecycleTypeEnumEnumValues),
							nil,
						),
						Optional: true,
						Computed: true,
					},
					"cronjob": schema.SingleNestedAttribute{
						Description:         "Cron job configuration.",
						MarkdownDescription: "Cron job configuration.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								Description:         "Cron expression defining the job schedule (5-field format, e.g. */5 * * * *).",
								MarkdownDescription: "Cron expression defining the job schedule (5-field format, e.g. `*/5 * * * *`).",
								Computed:            true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								Description:         "Command to execute when the cron job triggers.",
								MarkdownDescription: "Command to execute when the cron job triggers.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"entrypoint": schema.StringAttribute{
										Description:         "Entrypoint of the job (e.g. the command to execute).",
										MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
										Optional:            true,
										Computed:            true,
									},
									"arguments": schema.ListAttribute{
										Description:         "List of arguments passed to the entrypoint.",
										MarkdownDescription: "List of arguments passed to the entrypoint.",
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
				Description:         "Job's source configuration. Use image for container registry, or docker for building from a Dockerfile.",
				MarkdownDescription: "Job's source configuration. Use `image` for container registry, or `docker` for building from a Dockerfile.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						Description:         "Job's image source from a container registry.",
						MarkdownDescription: "Job's image source from a container registry.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								Description:         "Job's image source registry ID.",
								MarkdownDescription: "Job's image source registry ID.",
								Computed:            true,
							},
							"name": schema.StringAttribute{
								Description:         "Job's image source name.",
								MarkdownDescription: "Job's image source name.",
								Computed:            true,
							},
							"tag": schema.StringAttribute{
								Description:         "Job's image source tag.",
								MarkdownDescription: "Job's image source tag.",
								Computed:            true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						Description:         "Job's Docker source. Build the image from a Dockerfile in a git repository.",
						MarkdownDescription: "Job's Docker source. Build the image from a Dockerfile in a git repository.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"dockerfile_path": schema.StringAttribute{
								Description:         "Path to the Dockerfile relative to the git repository root path.",
								MarkdownDescription: "Path to the Dockerfile relative to the git repository root path.",
								Optional:            true,
							},
							"dockerfile_raw": schema.StringAttribute{
								Description:         "Inline Dockerfile content for building the image.",
								MarkdownDescription: "Inline Dockerfile content for building the image.",
								Optional:            true,
							},
							"git_repository": schema.SingleNestedAttribute{
								Description:         "Git repository containing the Dockerfile.",
								MarkdownDescription: "Git repository containing the Dockerfile.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description:         "Git repository URL.",
										MarkdownDescription: "Git repository URL.",
										Computed:            true,
									},
									"branch": schema.StringAttribute{
										Description:         "Git branch to use.",
										MarkdownDescription: "Git branch to use.",
										Computed:            true,
									},
									"root_path": schema.StringAttribute{
										Description:         "Root path in the git repository.",
										MarkdownDescription: "Root path in the git repository.",
										Optional:            true,
										Computed:            true,
									},
									"git_token_id": schema.StringAttribute{
										Description:         "Git token ID for accessing a private repository.",
										MarkdownDescription: "Git token ID for accessing a private repository.",
										Optional:            true,
										Computed:            false,
									},
								},
							},
							"docker_target_build_stage": schema.StringAttribute{
								Description:         "Target build stage in a multi-stage Dockerfile.",
								MarkdownDescription: "Target build stage in a multi-stage Dockerfile.",
								Optional:            true,
							},
						},
					},
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				Description:         "List of built-in environment variables linked to this job.",
				MarkdownDescription: "List of built-in environment variables linked to this job.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
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
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description:         "List of environment variables linked to this job.",
				MarkdownDescription: "List of environment variables linked to this job.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
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
				Description:         "List of environment variable aliases linked to this job.",
				MarkdownDescription: "List of environment variable aliases linked to this job.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable alias.",
							MarkdownDescription: "Id of the environment variable alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable alias.",
							MarkdownDescription: "Name of the environment variable alias.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the variable to alias.",
							MarkdownDescription: "Name of the variable to alias.",
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
				Description:         "List of environment variable overrides linked to this job.",
				MarkdownDescription: "List of environment variable overrides linked to this job.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable override.",
							MarkdownDescription: "Id of the environment variable override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable override.",
							MarkdownDescription: "Name of the environment variable override.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable override.",
							MarkdownDescription: "Value of the environment variable override.",
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
				Description:         "List of secrets linked to this job.",
				MarkdownDescription: "List of secrets linked to this job.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret.",
							MarkdownDescription: "Id of the secret.",
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
				Description:         "List of secret aliases linked to this job.",
				MarkdownDescription: "List of secret aliases linked to this job.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret alias.",
							MarkdownDescription: "Id of the secret alias.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret alias.",
							MarkdownDescription: "Name of the secret alias.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the secret to alias.",
							MarkdownDescription: "Name of the secret to alias.",
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
				Description:         "List of secret overrides linked to this job.",
				MarkdownDescription: "List of secret overrides linked to this job.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret override.",
							MarkdownDescription: "Id of the secret override.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret override.",
							MarkdownDescription: "Name of the secret override.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret override.",
							MarkdownDescription: "Value of the secret override.",
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
				Description:         "List of environment variable files linked to this job.",
				MarkdownDescription: "List of environment variable files linked to this job.",
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
				Description:         "List of secret files linked to this job.",
				MarkdownDescription: "List of secret files linked to this job.",
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
			"external_host": schema.StringAttribute{
				Description:         "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				MarkdownDescription: "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				Description:         "The job internal host.",
				MarkdownDescription: "The job internal host.",
				Computed:            true,
			},
			"deployment_stage_id": schema.StringAttribute{
				Description:         "Id of the deployment stage. Controls the order of service deployment.",
				MarkdownDescription: "Id of the deployment stage. Controls the order of service deployment.",
				Optional:            true,
				Computed:            true,
			},
			"is_skipped": schema.BoolAttribute{
				Description:         "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				MarkdownDescription: "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				Optional:            true,
				Computed:            true,
			},
			"advanced_settings_json": schema.StringAttribute{
				Description:         "Advanced settings in JSON format.",
				MarkdownDescription: "Advanced settings in JSON format.",
				Optional:            true,
				Computed:            true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description:         "Specify if the job will be automatically updated after receiving a new image tag or a new commit on the branch.",
				MarkdownDescription: "Specify if the job will be automatically updated after receiving a new image tag or a new commit on the branch.",
				Optional:            true,
				Computed:            true,
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description:         "List of deployment restrictions.",
				MarkdownDescription: "List of deployment restrictions.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the deployment restriction.",
							MarkdownDescription: "Id of the deployment restriction",
							Computed:            true,
						},
						"mode": schema.StringAttribute{
							Description:         "Deployment restriction mode. Can be: EXCLUDE, MATCH.",
							MarkdownDescription: "Deployment restriction mode.\n\t- Can be: `EXCLUDE`, `MATCH`.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							Description:         "Deployment restriction type. Can be: PATH.",
							MarkdownDescription: "Deployment restriction type.\n\t- Can be: `PATH`.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the deployment restriction (e.g. a file path pattern).",
							MarkdownDescription: "Value of the deployment restriction (e.g. a file path pattern).",
							Computed:            true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				Description:         "List of annotations group IDs.",
				MarkdownDescription: "List of annotations group IDs.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description:         "List of labels group IDs.",
				MarkdownDescription: "List of labels group IDs.",
				Optional:            true,
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

	state := convertDomainJobToJob(ctx, data, cont)
	tflog.Trace(ctx, "read job", map[string]any{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
