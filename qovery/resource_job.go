package qovery

import (
	"context"
	"fmt"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
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

func (r jobResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery job resource. This can be used to create and manage Qovery job registry.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the job.",
				Type:        types.StringType,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the job.",
				Type:        types.StringType,
				Required:    true,
			},
			"cpu": {
				Description: descriptions.NewInt64MinDescription(
					"CPU of the job in millicores (m) [1000m = 1 CPU].",
					int64(job.MinCPU),
					pointer.ToInt64(int64(job.DefaultCPU)),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(int64(job.DefaultCPU)),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: int64(job.MinCPU)}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"memory": {
				Description: descriptions.NewInt64MinDescription(
					"RAM of the job in MB [1024MB = 1GB].",
					int64(job.MinMemory),
					pointer.ToInt64(int64(job.DefaultMemory)),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(int64(job.DefaultMemory)),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: int64(job.MinMemory)},
				},
			},
			"max_duration_seconds": {
				Description: descriptions.NewInt64MinDescription(
					"Job's max duration in seconds.",
					int64(job.MinDurationSeconds),
					pointer.ToInt64(int64(job.DefaultMaxDurationSeconds)),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(int64(job.DefaultMaxDurationSeconds)),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: int64(job.MinDurationSeconds)}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"max_nb_restart": {
				Description: descriptions.NewInt64MinDescription(
					"Job's max number of restarts.",
					int64(job.MinNbRestart),
					pointer.ToInt64(int64(job.DefaultMaxNbRestart)),
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(int64(job.DefaultMaxNbRestart)),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: int64(job.MinNbRestart)}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"port": {
				Description: descriptions.NewInt64MinMaxDescription(
					"Job's probes port.",
					port.MinPort,
					port.MaxPort,
					nil,
				),
				Type:     types.Int64Type,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
				},
			},
			"auto_preview": {
				Description: "Specify if the environment preview option is activated or not for this job.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
			"schedule": {
				Description: "Job's schedule.",
				Required:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"on_start": {
						Description: "Job's schedule on start.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Required:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"on_stop": {
						Description: "Job's schedule on stop.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Required:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"on_delete": {
						Description: "Job's schedule on delete.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"entrypoint": {
								Description: "Entrypoint of the job.",
								Type:        types.StringType,
								Required:    true,
							},
							"arguments": {
								Description: "List of arguments of this job.",
								Optional:    true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
								PlanModifiers: tfsdk.AttributePlanModifiers{
									modifiers.NewStringSliceDefaultModifier([]string{}),
								},
							},
						}),
					},
					"cronjob": {
						Description: "Job's cron.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"schedule": {
								Description: "Job's cron string.",
								Type:        types.StringType,
								Required:    true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": {
								Description: "Job's cron command.",
								Required:    true,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"entrypoint": {
										Description: "Entrypoint of the job.",
										Type:        types.StringType,
										Optional:    true,
									},
									"arguments": {
										Description: "List of arguments of this job.",
										Optional:    true,
										Type: types.ListType{
											ElemType: types.StringType,
										},
										PlanModifiers: tfsdk.AttributePlanModifiers{
											modifiers.NewStringSliceDefaultModifier([]string{}),
										},
									},
								}),
							},
						}),
					},
				}),
			},
			"source": {
				Description: "Job's source.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"image": {
						Description: "Job's image source.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"registry_id": {
								Description: "Job's image source registry ID.",
								Type:        types.StringType,
								Required:    true,
							},
							"name": {
								Description: "Job's image source name.",
								Type:        types.StringType,
								Required:    true,
							},
							"tag": {
								Description: "Job's image source tag.",
								Type:        types.StringType,
								Required:    true,
							},
						}),
					},
					"docker": {
						Description: "Job's docker source.",
						Optional:    true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"dockerfile_path": {
								Description: "Job's docker source dockerfile path.",
								Type:        types.StringType,
								Optional:    true,
							},
							"git_repository": {
								Description: "Job's docker source git repository.",
								Required:    true,
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"url": {
										Description: "Job's docker source git repository URL.",
										Type:        types.StringType,
										Required:    true,
									},
									"branch": {
										Description: "Job's docker source git repository branch.",
										Type:        types.StringType,
										Required:    true,
									},
									"root_path": {
										Description: "Job's docker source git repository root path.",
										Type:        types.StringType,
										Optional:    true,
									},
								}),
							},
						}),
					},
				}),
			},
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this job.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
			"secrets": {
				Description: "List of secrets linked to this job.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the secret.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the secret [NOTE: will always be empty].",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
			"external_host": {
				Description: "The job external FQDN host [NOTE: only if your job is using a publicly accessible port].",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The job internal host.",
				Type:        types.StringType,
				Computed:    true,
			},
			"deployment_stage_id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}, nil
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
	cont, err := r.jobService.Create(ctx, plan.EnvironmentID.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on job create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainJobToJob(plan, cont)
	tflog.Trace(ctx, "created job", map[string]interface{}{"job_id": state.ID.Value})

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
	cont, err := r.jobService.Get(ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on job read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainJobToJob(state, cont)
	tflog.Trace(ctx, "read job", map[string]interface{}{"job_id": state.ID.Value})

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
	cont, err := r.jobService.Update(ctx, state.ID.Value, *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on job update", err.Error())
		return
	}

	// Update state values
	state = convertDomainJobToJob(plan, cont)
	tflog.Trace(ctx, "updated job", map[string]interface{}{"job_id": state.ID.Value})

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
	err := r.jobService.Delete(ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on job delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted job", map[string]interface{}{"job_id": state.ID.Value})

	// Remove job from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery job resource using its id
func (r jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
