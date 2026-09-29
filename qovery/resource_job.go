package qovery

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &jobResource{}
	_ resource.ResourceWithImportState = jobResource{}
	_ resource.ResourceWithModifyPlan  = jobResource{}
)

type jobResource struct {
	jobService              job.Service
	advancedSettingsService *advanced_settings.ServiceAdvancedSettingsService
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
	r.advancedSettingsService = provider.advancedSettingsService
}

func (r jobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	envVars := variableListDescriptions("environment_variables", "job")
	builtInEnvVars := variableListDescriptions("built_in_environment_variables", "job")
	envVarAliases := variableListDescriptions("environment_variable_aliases", "job")
	envVarOverrides := variableListDescriptions("environment_variable_overrides", "job")
	secrets := variableListDescriptions("secrets", "job")
	secretAliases := variableListDescriptions("secret_aliases", "job")
	secretOverrides := variableListDescriptions("secret_overrides", "job")
	restrictions := deploymentRestrictionDescriptions("job")

	resp.Schema = schema.Schema{
		MarkdownDescription: jobResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("job"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("job"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("job"),
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: iconURIDescription("job") + jobIconURIDefaultNote,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					JobIconUriDefault(),
				},
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(cpuDescription("job"), job.MinCPU, pointer.ToInt64(job.DefaultCPU)),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(job.DefaultCPU),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinCPU},
				},
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(memoryDescription("job"), job.MinMemory, pointer.ToInt64(job.DefaultMemory)),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(job.DefaultMemory),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinMemory},
				},
			},
			"ephemeral_storage": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64DefaultDescription(ephemeralStorageDescription("job"), serviceEphemeralStorageDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(serviceEphemeralStorageDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: 0},
				},
			},
			"max_duration_seconds": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(jobMaxDurationSecondsDescription, job.MinDurationSeconds, pointer.ToInt64(job.DefaultMaxDurationSeconds)),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(job.DefaultMaxDurationSeconds),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinDurationSeconds}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"max_nb_restart": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(jobMaxNbRestartDescription, job.MinNbRestart, pointer.ToInt64(job.DefaultMaxNbRestart)),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(job.DefaultMaxNbRestart),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: job.MinNbRestart}, // TODO(benjaminch): useless check, by design won't be < 0
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinMaxDescription(jobPortDescription, port.MinPort, port.MaxPort, nil),
				Optional:            true,
				Validators: []validator.Int64{
					validators.Int64MinMaxValidator{Min: port.MinPort, Max: port.MaxPort},
				},
			},
			"auto_preview": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(autoPreviewDescription("job"), serviceAutoPreviewDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoPreviewDefault),
			},
			"healthchecks": healthchecksSchemaAttributes(true),
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: jobScheduleDescription + jobScheduleWriteNote,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"on_start": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnStartDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
								Optional:            true,
								Computed:            true,
								ElementType:         types.StringType,
								Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_stop": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnStopDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
								Optional:            true,
								ElementType:         types.StringType,
								Computed:            true,
								Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"on_delete": schema.SingleNestedAttribute{
						MarkdownDescription: jobOnDeleteDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"entrypoint": schema.StringAttribute{
								MarkdownDescription: entrypointDescription,
								Optional:            true,
							},
							"arguments": schema.ListAttribute{
								MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
								Optional:            true,
								ElementType:         types.StringType,
								Computed:            true,
								Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
							},
						},
					},
					"lifecycle_type": schema.StringAttribute{
						MarkdownDescription: descriptions.NewStringEnumDescription(
							jobLifecycleTypeDescription+jobLifecycleTypeWriteNote,
							clientEnumToStringArray(qovery.AllowedJobLifecycleTypeEnumEnumValues),
							nil,
						) + jobLifecycleTypeDefaultNote,
						Optional: true,
						Computed: true,
						// q-core rejects a change of the lifecycle type of an existing job, so a
						// planned change, including the reset of an omitted type, is a plan error.
						PlanModifiers: []planmodifier.String{
							JobLifecycleTypeDefault(),
							RejectChangeAfterCreate(lifecycleTypeChangeReason),
						},
					},
					"cronjob": schema.SingleNestedAttribute{
						MarkdownDescription: jobCronJobDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"schedule": schema.StringAttribute{
								MarkdownDescription: jobCronJobScheduleDescription,
								Required:            true,
								// TODO(benjaminch): introduce a cron string validator
							},
							"command": schema.SingleNestedAttribute{
								MarkdownDescription: jobCronJobCommandDescription,
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"entrypoint": schema.StringAttribute{
										MarkdownDescription: entrypointDescription,
										Optional:            true,
									},
									"arguments": schema.ListAttribute{
										MarkdownDescription: argumentsDescription + omittedSetsNoneNote,
										Optional:            true,
										ElementType:         types.StringType,
										Computed:            true,
										Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
									},
								},
							},
						},
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: jobSourceDescription + jobSourceWriteNote,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						MarkdownDescription: jobSourceImageDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"registry_id": schema.StringAttribute{
								MarkdownDescription: containerImageRegistryIDDescription,
								Required:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: containerImageNameDescription,
								Required:            true,
							},
							"tag": schema.StringAttribute{
								MarkdownDescription: containerImageTagDescription,
								Required:            true,
							},
						},
					},
					"docker": schema.SingleNestedAttribute{
						MarkdownDescription: jobSourceDockerDescription,
						Optional:            true,
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
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										MarkdownDescription: gitRepositoryURLDescription,
										Required:            true,
									},
									"branch": schema.StringAttribute{
										MarkdownDescription: jobGitRepositoryBranchDescription,
										Required:            true,
									},
									"root_path": schema.StringAttribute{
										MarkdownDescription: descriptions.NewStringDefaultDescription(gitRepositoryRootPathDescription+jobGitRepositoryRootPathWriteNote, jobRootPathDefault),
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(jobRootPathDefault),
										// The read maps an API "" to "/" (jobRootPathFromAPI), so "" in the
										// configuration would never match the state.
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
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
				PlanModifiers: []planmodifier.List{
					UseStateUnlessNameChanges(),
				},
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVars.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVars.Description,
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_aliases": schema.SetNestedAttribute{
				MarkdownDescription: envVarAliases.List,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: envVarAliases.ID,
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Key,
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarAliases.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: envVarOverrides.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secrets.Value,
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secrets.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretAliases.Value,
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretAliases.Description,
							Optional:            true,
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
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Value,
							Required:            true,
							Sensitive:           true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: secretOverrides.Description,
							Optional:            true,
						},
					},
				},
			},
			"environment_variable_files": environmentVariableFilesSchemaAttribute("job"),
			"secret_files":               secretFilesSchemaAttribute("job"),
			"external_secrets":           externalSecretsSchemaAttribute("job"),
			"external_secret_files":      externalSecretFilesSchemaAttribute("job"),
			"external_host": schema.StringAttribute{
				MarkdownDescription: jobHostDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: jobHostDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				MarkdownDescription: deploymentStageIDDescription + deploymentStageIDRemovalNote,
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: q-core attaches
				// every service to a stage and has no detach, so removal keeps the current stage.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_skipped": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(isSkippedDescription, false),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: advancedSettingsJSONDescription("Jobs/operation/getDefaultJobAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: the QOV-2028
				// contract described in advancedSettingsJSONDescription.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_settings": buildSettingsResourceSchemaAttributes(),
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(jobAutoDeployDescription, serviceAutoDeployDefault),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(serviceAutoDeployDefault),
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
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: restrictions.Type,
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: restrictions.Value,
							Required:            true,
						},
					},
				},
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("annotations", "the job's pods"),
				Optional:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("labels", "the job's pods"),
				Optional:            true,
				ElementType:         types.StringType,
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
		if cont == nil {
			return
		}
		// The job was created in Qovery and only a follow-up call failed. Save it so
		// Terraform tracks it and taints it for replacement, instead of leaving an orphan
		// that makes every later apply fail with "a job named X already exists".
		resp.Diagnostics.Append(resp.State.Set(ctx, convertDomainJobToJob(ctx, plan, cont))...)
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
	isTriggeredFromImport := false
	if state.EnvironmentID.IsNull() {
		isTriggeredFromImport = true
	}

	// Get job from the API
	cont, err := r.jobService.Get(ctx, state.ID.ValueString(), state.AdvancedSettingsJson.ValueString(), isTriggeredFromImport)
	if handleDomainReadNotFound(ctx, resp, err, "Error on job read") {
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

func (r jobResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	warnUnknownAdvancedSettings(ctx, r.advancedSettingsService, domain.JOB, req.Config, &resp.Diagnostics)
	modifyBuildSettingsPlan(ctx, req, resp)
}
