package qovery

import (
	"context"
	"fmt"

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
		Description: "Provides a Qovery job resource. This can be used to create and manage Qovery job registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the job.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the job.",
				Computed:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
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
				Optional: true,
			},
			"max_duration_seconds": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
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
				Computed: true,
				Optional: true,
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this job.",
				Optional:    true,
				Computed:    true,
			},
			"healthchecks": healthchecksSchemaAttributes(false),
			"schedule": schema.SingleNestedAttribute{
				Description: "Job's schedule.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						Description: "Job's schedule on start.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description: "Entrypoint of the job.",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description: "List of arguments of this job.",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						Description: "Job's schedule on stop.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description: "Entrypoint of the job.",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description: "List of arguments of this job.",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						Description: "Job's schedule on delete.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								Description: "Entrypoint of the job.",
								Optional:    true,
								Computed:    true,
							},
							"arguments": schema.ListAttribute{
								Description: "List of arguments of this job.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
					"cronjob": schema.SingleNestedAttribute{
						Description: "Job's cron.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								Description: "Job's cron string.",
								Computed:    true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								Description: "Job's cron command.",
								Computed:    true,
								Attributes: map[string]schema.Attribute{
									"entrypoint": schema.StringAttribute{
										Description: "Entrypoint of the job.",
										Optional:    true,
										Computed:    true,
									},
									"arguments": schema.ListAttribute{
										Description: "List of arguments of this job.",
										Optional:    true,
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				Description: "Job's source.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						Description: "Job's image source.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								Description: "Job's image source registry ID.",
								Computed:    true,
							},
							"name": schema.StringAttribute{
								Description: "Job's image source name.",
								Computed:    true,
							},
							"tag": schema.StringAttribute{
								Description: "Job's image source tag.",
								Computed:    true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						Description: "Job's docker source.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"dockerfile_path": schema.StringAttribute{
								Description: "Job's docker source dockerfile path.",
								Optional:    true,
							},
							"git_repository": schema.SingleNestedAttribute{
								Description: "Job's docker source git repository.",
								Computed:    true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description: "Job's docker source git repository URL.",
										Computed:    true,
									},
									"branch": schema.StringAttribute{
										Description: "Job's docker source git repository branch.",
										Computed:    true,
									},
									"root_path": schema.StringAttribute{
										Description: "Job's docker source git repository root path.",
										Optional:    true,
										Computed:    true,
									},
									"git_token_id": schema.StringAttribute{
										Description: "The git token ID to be used",
										Optional:    true,
										Computed:    false,
									},
								},
							},
						},
					},
				},
			},
			"built_in_environment_variables": schema.SetNestedAttribute{
				Description: "List of built-in environment variables linked to this job.",
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
			// TODO (framework-migration) Extract environment variables + secrets attributes to avoid repetition everywhere (project / env / services)
			"environment_variables": schema.SetNestedAttribute{
				Description: "List of environment variables linked to this job.",
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
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				Description: "List of environment variable aliases linked to this job.",
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
					},
				},
			},
			"environment_variable_overrides": schema.SetNestedAttribute{
				Description: "List of environment variable overrides linked to this job.",
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
					},
				},
			},
			"secrets": schema.SetNestedAttribute{
				Description: "List of secrets linked to this job.",
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
					},
				},
			},
			"secret_aliases": schema.SetNestedAttribute{
				Description: "List of secret aliases linked to this job.",
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
					},
				},
			},
			"secret_overrides": schema.SetNestedAttribute{
				Description: "List of secret overrides linked to this job.",
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
					},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The job internal host.",
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
				Description: " Specify if the job will be automatically updated after receiving a new image tag.",
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
	cont, err := d.jobService.Get(ctx, data.ID.ValueString(), data.AdvancedSettingsJson.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	state := convertDomainJobToJob(ctx, data, cont)
	tflog.Trace(ctx, "read job", map[string]interface{}{"job_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
