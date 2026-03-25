package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/qovery/qovery-client-go"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &jobResource{}
var _ resource.ResourceWithImportState = jobResource{}

type jobResource struct {
	jobService job.Service
}

func newJobResource() resource.Resource {
	return &jobResource{}
}

func (r jobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (r *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.jobService = provider.jobService
}

func (r jobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Provides a Qovery job resource. This can be used to create and manage Qovery jobs (cron jobs and lifecycle jobs).",
		MarkdownDescription: "Provides a Qovery job resource. This can be used to create and manage Qovery jobs (cron jobs and lifecycle jobs).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the job.",
				MarkdownDescription: "Id of the job.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description:         "Id of the environment.",
				MarkdownDescription: "Id of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the job.",
				MarkdownDescription: "Name of the job.",
				Required:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description:         "Icon URI representing the job.",
				MarkdownDescription: "Icon URI representing the job.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Computed: true,
				Default:  int64default.StaticInt64(job.DefaultCPU),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinCPU},
				},
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
				Computed: true,
				Default:  int64default.StaticInt64(job.DefaultMemory),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinMemory},
				},
			},
			"max_duration_seconds": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Job's max duration in seconds.",
					job.MinDurationSeconds,
					pointer.ToInt64(job.DefaultMaxDurationSeconds),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"Job's max duration in seconds.",
					job.MinDurationSeconds,
					pointer.ToInt64(job.DefaultMaxDurationSeconds),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(job.DefaultMaxDurationSeconds),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinDurationSeconds}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"max_nb_restart": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Job's max number of restarts.",
					job.MinNbRestart,
					pointer.ToInt64(job.DefaultMaxNbRestart),
				),
				MarkdownDescription: descriptions.NewInt64MinDescription(
					"Job's max number of restarts.",
					job.MinNbRestart,
					pointer.ToInt64(job.DefaultMaxNbRestart),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(job.DefaultMaxNbRestart),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinNbRestart}, // TODO(benjaminch): useless check, by design won't be < 0
				},
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
				Optional: true,
				Validators: []validator.Int64{
					validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
				},
			},
			"auto_preview": schema.BoolAttribute{
				Description:         "Specify if the environment preview option is activated or not for this job.",
				MarkdownDescription: "Specify if the environment preview option is activated or not for this job.",
				Optional:    true,
				Computed:    true,
			},
			"healthchecks": healthchecksSchemaAttributes(true),
			"schedule": schema.SingleNestedAttribute{
				Description:         "Job's schedule configuration. Use on_start, on_stop, and on_delete for lifecycle jobs, or cronjob for cron jobs.",
				MarkdownDescription: "Job's schedule configuration. Use `on_start`, `on_stop`, and `on_delete` for lifecycle jobs, or `cronjob` for cron jobs.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment starts. Define the entrypoint and arguments for this event.",
						MarkdownDescription: "Lifecycle job event: executed when the environment starts. Define the entrypoint and arguments for this event.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment stops. Define the entrypoint and arguments for this event.",
						MarkdownDescription: "Lifecycle job event: executed when the environment stops. Define the entrypoint and arguments for this event.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:    true,
								ElementType: types.StringType,
								Computed:    true,
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						Description:         "Lifecycle job event: executed when the environment is deleted. Define the entrypoint and arguments for this event.",
						MarkdownDescription: "Lifecycle job event: executed when the environment is deleted. Define the entrypoint and arguments for this event.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description:         "Entrypoint of the job (e.g. the command to execute).",
								MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description:         "List of arguments passed to the entrypoint.",
								MarkdownDescription: "List of arguments passed to the entrypoint.",
								Optional:    true,
								ElementType: types.StringType,
								Computed:    true,
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
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
						Description:         "Cron job configuration. Use this to run the job on a recurring schedule.",
						MarkdownDescription: "Cron job configuration. Use this to run the job on a recurring schedule.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								Description:         "Cron expression defining the job schedule (5-field format, e.g. */5 * * * * for every 5 minutes). See https://crontab.guru/ for help.",
								MarkdownDescription: "Cron expression defining the job schedule (5-field format, e.g. `*/5 * * * *` for every 5 minutes). See https://crontab.guru/ for help.",
								Required:    true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								Description:         "Command to execute when the cron job triggers.",
								MarkdownDescription: "Command to execute when the cron job triggers.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"entrypoint": schema.StringAttribute{
										Description:         "Entrypoint of the job (e.g. the command to execute).",
										MarkdownDescription: "Entrypoint of the job (e.g. the command to execute).",
										Optional:    true,
										Computed:    true,
									},
									"arguments": schema.ListAttribute{
										Description:         "List of arguments passed to the entrypoint.",
										MarkdownDescription: "List of arguments passed to the entrypoint.",
										Optional:    true,
										ElementType: types.StringType,
										Computed:    true,
										Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
									},
								},
							},
						},
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				Description:         "Job's source configuration. Use image to deploy from a container registry, or docker to build from a Dockerfile in a git repository.",
				MarkdownDescription: "Job's source configuration. Use `image` to deploy from a container registry, or `docker` to build from a Dockerfile in a git repository.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						Description:         "Job's image source. Use this to deploy a pre-built image from a container registry.",
						MarkdownDescription: "Job's image source. Use this to deploy a pre-built image from a container registry.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								Description:         "Job's image source registry ID (refers to a qovery_container_registry resource).",
								MarkdownDescription: "Job's image source registry ID (refers to a `qovery_container_registry` resource).",
								Required:    true,
							},
							"name": schema.StringAttribute{
								Description:         "Job's image source name.",
								MarkdownDescription: "Job's image source name.",
								Required:    true,
							},
							"tag": schema.StringAttribute{
								Description:         "Job's image source tag.",
								MarkdownDescription: "Job's image source tag.",
								Required:    true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						Description:         "Job's Docker source. Use this to build the job image from a Dockerfile in a git repository.",
						MarkdownDescription: "Job's Docker source. Use this to build the job image from a Dockerfile in a git repository.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"dockerfile_path": schema.StringAttribute{
								Description:         "Path to the Dockerfile relative to the git repository root path (e.g. Dockerfile or build/Dockerfile).",
								MarkdownDescription: "Path to the Dockerfile relative to the git repository root path (e.g. `Dockerfile` or `build/Dockerfile`).",
								Optional:    true,
							},
							"dockerfile_raw": schema.StringAttribute{
								Description:         "Inline Dockerfile content to inject for building the image. Use this instead of dockerfile_path to define the Dockerfile directly in Terraform.",
								MarkdownDescription: "Inline Dockerfile content to inject for building the image. Use this instead of `dockerfile_path` to define the Dockerfile directly in Terraform.",
								Optional:    true,
							},
							"git_repository": schema.SingleNestedAttribute{
								Description:         "Git repository containing the Dockerfile for the job.",
								MarkdownDescription: "Git repository containing the Dockerfile for the job.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description:         "Git repository URL (e.g. https://github.com/org/repo.git).",
										MarkdownDescription: "Git repository URL (e.g. `https://github.com/org/repo.git`).",
										Required:    true,
									},
									"branch": schema.StringAttribute{
										Description:         "Git branch to use for the Docker source.",
										MarkdownDescription: "Git branch to use for the Docker source.",
										Required:    true,
									},
									"root_path": schema.StringAttribute{
										Description:         "Root path in the git repository where the Dockerfile is located.",
										MarkdownDescription: "Root path in the git repository where the Dockerfile is located.",
										Optional:    true,
										Computed:    true,
									},
									"git_token_id": schema.StringAttribute{
										Description:         "Git token ID for accessing a private repository (refers to a qovery_git_token resource).",
										MarkdownDescription: "Git token ID for accessing a private repository (refers to a `qovery_git_token` resource).",
										Optional:    true,
										Computed:    false,
									},
								},
							},
							"docker_target_build_stage": schema.StringAttribute{
								Description:         "Target build stage in a multi-stage Dockerfile (e.g. production or builder).",
								MarkdownDescription: "Target build stage in a multi-stage Dockerfile (e.g. `production` or `builder`).",
								Optional:    true,
							},
						},
					},
				},
			},
			"built_in_environment_variables": schema.ListNestedAttribute{
				Description:         "List of built-in environment variables linked to this job.",
				MarkdownDescription: "List of built-in environment variables linked to this job.",
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description:         "List of environment variables linked to this job.",
				MarkdownDescription: "List of environment variables linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable.",
							MarkdownDescription: "Id of the environment variable.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the environment variable.",
							MarkdownDescription: "Key of the environment variable.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable.",
							MarkdownDescription: "Value of the environment variable.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable.",
							MarkdownDescription: "Description of the environment variable.",
							Optional:    true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description:         "List of environment variable aliases linked to this job.",
				MarkdownDescription: "List of environment variable aliases linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable alias.",
							MarkdownDescription: "Id of the environment variable alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable alias.",
							MarkdownDescription: "Name of the environment variable alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the variable to alias.",
							MarkdownDescription: "Name of the variable to alias.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable alias.",
							MarkdownDescription: "Description of the environment variable alias.",
							Optional:    true,
						},
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description:         "List of environment variable overrides linked to this job.",
				MarkdownDescription: "List of environment variable overrides linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the environment variable override.",
							MarkdownDescription: "Id of the environment variable override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the environment variable override.",
							MarkdownDescription: "Name of the environment variable override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the environment variable override.",
							MarkdownDescription: "Value of the environment variable override.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the environment variable override.",
							MarkdownDescription: "Description of the environment variable override.",
							Optional:    true,
						},
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description:         "List of secrets linked to this job.",
				MarkdownDescription: "List of secrets linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret.",
							MarkdownDescription: "Id of the secret.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Key of the secret.",
							MarkdownDescription: "Key of the secret.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret.",
							MarkdownDescription: "Value of the secret.",
							Required:    true,
							Sensitive:   true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret.",
							MarkdownDescription: "Description of the secret.",
							Optional:    true,
						},
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description:         "List of secret aliases linked to this job.",
				MarkdownDescription: "List of secret aliases linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret alias.",
							MarkdownDescription: "Id of the secret alias.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret alias.",
							MarkdownDescription: "Name of the secret alias.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Name of the secret to alias.",
							MarkdownDescription: "Name of the secret to alias.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret alias.",
							MarkdownDescription: "Description of the secret alias.",
							Optional:    true,
						},
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description:         "List of secret overrides linked to this job.",
				MarkdownDescription: "List of secret overrides linked to this job.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the secret override.",
							MarkdownDescription: "Id of the secret override.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description:         "Name of the secret override.",
							MarkdownDescription: "Name of the secret override.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the secret override.",
							MarkdownDescription: "Value of the secret override.",
							Required:    true,
							Sensitive:   true,
						},
						"description": schema.StringAttribute{
							Description:         "Description of the secret override.",
							MarkdownDescription: "Description of the secret override.",
							Optional:    true,
						},
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description:         "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				MarkdownDescription: "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_host": schema.StringAttribute{
				Description:         "The job internal host.",
				MarkdownDescription: "The job internal host.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				Description:         "Id of the deployment stage. Deployment stages allow you to control the order in which services are deployed within an environment.",
				MarkdownDescription: "Id of the deployment stage. Deployment stages allow you to control the order in which services are deployed within an environment.",
				Optional:    true,
				Computed:    true,
			},
			"is_skipped": schema.BoolAttribute{
				Description:         "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				MarkdownDescription: "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"advanced_settings_json": schema.StringAttribute{
				Description:         "Advanced settings in JSON format. See the Qovery API documentation for the full list of available settings: https://api-doc.qovery.com/#tag/Jobs/operation/getDefaultJobAdvancedSettings",
				MarkdownDescription: "Advanced settings in JSON format. See the Qovery API documentation for the full list of available settings: https://api-doc.qovery.com/#tag/Jobs/operation/getDefaultJobAdvancedSettings",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				Description:         "Specify if the job will be automatically updated after receiving a new image tag or a new commit on the branch.",
				MarkdownDescription: "Specify if the job will be automatically updated after receiving a new image tag or a new commit on the branch.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_restrictions": schema.SetNestedAttribute{
				Description:         "List of deployment restrictions. Deployment restrictions allow you to control which changes trigger a deployment based on file path patterns.",
				MarkdownDescription: "List of deployment restrictions. Deployment restrictions allow you to control which changes trigger a deployment based on file path patterns.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "Id of the deployment restriction.",
							MarkdownDescription: "Id of the deployment restriction.",
							Computed:    true,
						},
						"mode": schema.StringAttribute{
							Description:         "Deployment restriction mode. Can be: EXCLUDE, MATCH.",
							MarkdownDescription: "Deployment restriction mode.\n\t- Can be: `EXCLUDE`, `MATCH`.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description:         "Deployment restriction type. Can be: PATH.",
							MarkdownDescription: "Deployment restriction type.\n\t- Can be: `PATH`.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description:         "Value of the deployment restriction (e.g. a file path pattern like src/backend/**).",
							MarkdownDescription: "Value of the deployment restriction (e.g. a file path pattern like `src/backend/**`).",
							Required:    true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				Description:         "List of annotations group IDs to associate with this job. Annotations groups are defined using the qovery_annotations_group resource.",
				MarkdownDescription: "List of annotations group IDs to associate with this job. Annotations groups are defined using the `qovery_annotations_group` resource.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description:         "List of labels group IDs to associate with this job. Labels groups are defined using the qovery_labels_group resource.",
				MarkdownDescription: "List of labels group IDs to associate with this job. Labels groups are defined using the `qovery_labels_group` resource.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create qovery job resource
func (r jobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Job
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new job
	request, err := plan.toUpsertServiceRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError("Error on job create", err.Error())
		return
	}
	cont, err := r.jobService.Create(ctx, plan.EnvironmentID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on job create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainJobToJob(ctx, plan, cont)
	tflog.Trace(ctx, "created job", map[string]any{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery job resource
func (r jobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Job
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hack to know if this method is triggered through an import
	// EnvironmentID is always present except when importing the resource
	var isTriggeredFromImport = false
	if state.EnvironmentID.IsNull() {
		isTriggeredFromImport = true
	}

	// Get job from the API
	cont, err := r.jobService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainJobToJob(ctx, state, cont)
	tflog.Trace(ctx, "read job", map[string]any{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery job resource
func (r jobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Job
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update job in the backend
	request, err := plan.toUpsertServiceRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError("Error on job create", err.Error())
		return
	}
	cont, err := r.jobService.Update(ctx, state.ID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on job update", err.Error())
		return
	}

	// Update state values
	state = convertDomainJobToJob(ctx, plan, cont)
	tflog.Trace(ctx, "updated job", map[string]any{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery job resource
func (r jobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Job
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete job
	err := r.jobService.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on job delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted job", map[string]any{"job_id": state.ID.ValueString()})

	// Remove job from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery job resource using its id
func (r jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
