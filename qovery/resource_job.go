package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		Description: "Provides a Qovery job resource. This can be used to create and manage Qovery job registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the job.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the job.",
				Required:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
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
				Optional: true,
				Validators: []validator.Int64{
					validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
				},
			},
			"auto_preview": schema.BoolAttribute{
				Description: "Specify if the environment preview option is activated or not for this job.",
				Optional:    true,
				Computed:    true,
			},
			"healthchecks": healthchecksSchemaAttributes(true),
			"schedule": schema.SingleNestedAttribute{
				Description: "Job's schedule.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						Description: "Job's schedule on start.",
						Optional:    true,
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
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						Description: "Job's schedule on stop.",
						Optional:    true,
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
								Computed:    true,
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						Description: "Job's schedule on delete.",
						Optional:    true,
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
								Computed:    true,
								Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"cronjob": schema.SingleNestedAttribute{
						Description: "Job's cron.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								Description: "Job's cron string.",
								Required:    true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								Description: "Job's cron command.",
								Required:    true,
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
				Description: "Job's source.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						Description: "Job's image source.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								Description: "Job's image source registry ID.",
								Required:    true,
							},
							"name": schema.StringAttribute{
								Description: "Job's image source name.",
								Required:    true,
							},
							"tag": schema.StringAttribute{
								Description: "Job's image source tag.",
								Required:    true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						Description: "Job's docker source.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"dockerfile_path": schema.StringAttribute{
								Description: "Job's docker source dockerfile path.",
								Optional:    true,
							},
							"git_repository": schema.SingleNestedAttribute{
								Description: "Job's docker source git repository.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Description: "Job's docker source git repository URL.",
										Required:    true,
									},
									"branch": schema.StringAttribute{
										Description: "Job's docker source git repository branch.",
										Required:    true,
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
				Description: "List of environment variable aliases linked to this job.",
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
	tflog.Trace(ctx, "created job", map[string]interface{}{"job_id": state.ID.ValueString()})

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

	// Get job from the API
	cont, err := r.jobService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainJobToJob(ctx, state, cont)
	tflog.Trace(ctx, "read job", map[string]interface{}{"job_id": state.ID.ValueString()})

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
	tflog.Trace(ctx, "updated job", map[string]interface{}{"job_id": state.ID.ValueString()})

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

	tflog.Trace(ctx, "deleted job", map[string]interface{}{"job_id": state.ID.ValueString()})

	// Remove job from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery job resource using its id
func (r jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
