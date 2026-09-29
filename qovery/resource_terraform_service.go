package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &terraformServiceResource{}
	_ resource.ResourceWithImportState = terraformServiceResource{}
	_ resource.ResourceWithModifyPlan  = &terraformServiceResource{}
)

var (
	terraformActionValues  = clientEnumToStringArray(terraformservice.AllowedTerraformActionValues)
	terraformActionDefault = string(terraformservice.TerraformActionDefault)
)

type terraformServiceResource struct {
	terraformServiceService terraformservice.Service
	advancedSettingsService *advanced_settings.ServiceAdvancedSettingsService
}

func newTerraformServiceResource() resource.Resource {
	return &terraformServiceResource{}
}

func (r terraformServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_service"
}

func (r *terraformServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.terraformServiceService = provider.terraformServiceService
	r.advancedSettingsService = provider.advancedSettingsService
}

func (r terraformServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery Terraform service: Terraform or OpenTofu code from a git repository that Qovery plans and applies in its environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("Terraform service"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("Terraform service"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
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
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("Terraform service"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: descriptionDescription("Terraform service"),
				Optional:            true,
			},
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: createdFromBlueprintIDDescription("Terraform service") + blueprintIDRemovalNote,
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: q-core records the
				// blueprint only on create, so removal keeps it and a change is a plan error.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					RejectChangeAfterCreate(blueprintIDChangeReason),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				MarkdownDescription: terraformServiceAutoDeployDescription,
				Required:            true,
			},
			"terraform_action": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(
					terraformServiceActionDescription,
					terraformActionValues,
					&terraformActionDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(terraformActionDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(terraformActionValues),
				},
			},
			"git_repository": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceGitRepositoryDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: gitRepositoryURLDescription,
						Required:            true,
					},
					"branch": schema.StringAttribute{
						MarkdownDescription: terraformServiceBranchDescription,
						Optional:            true,
					},
					"root_path": schema.StringAttribute{
						MarkdownDescription: descriptions.NewStringDefaultDescription(terraformServiceRootPathDescription, terraformservice.DefaultRootPath),
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(terraformservice.DefaultRootPath),
					},
					"git_token_id": schema.StringAttribute{
						MarkdownDescription: gitRepositoryTokenIDDescription,
						Optional:            true,
					},
				},
			},
			"tfvars_files": schema.ListAttribute{
				MarkdownDescription: terraformServiceTfvarsFilesDescription + " Each path must start with `git_repository.root_path`.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"variables": schema.SetNestedAttribute{
				MarkdownDescription: terraformServiceVariablesDescription,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: terraformServiceVariableKeyDescription,
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: terraformServiceVariableValueDescription,
							Required:            true,
						},
						"is_secret": schema.BoolAttribute{
							MarkdownDescription: descriptions.NewBoolDefaultDescription(terraformServiceVariableIsSecretDescription, false),
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
			},
			"backend": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceBackendDescription + " Set exactly one of `kubernetes`, `user_provided` and `blueprint`.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"kubernetes": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceKubernetesBackendDescription + " Set it to `{}`.",
						Optional:            true,
						Attributes:          map[string]schema.Attribute{},
					},
					"user_provided": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceUserProvidedBackendDescription + " Set it to `{}`.",
						Optional:            true,
						Attributes:          map[string]schema.Attribute{},
					},
					"blueprint": schema.SingleNestedAttribute{
						MarkdownDescription: terraformServiceBlueprintBackendDescription,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: terraformServiceBlueprintBackendTypeDescription,
								Required:            true,
							},
							"config": schema.MapAttribute{
								MarkdownDescription: terraformServiceBlueprintBackendConfigDescription,
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
				},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(terraformServiceEngineDescription, terraformServiceEngines, nil),
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator([]string{"TERRAFORM", "OPEN_TOFU"}),
				},
			},
			"engine_version": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceEngineVersionDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"explicit_version": schema.StringAttribute{
						MarkdownDescription: terraformServiceExplicitVersionDescription,
						Required:            true,
					},
					"read_from_terraform_block": schema.BoolAttribute{
						MarkdownDescription: descriptions.NewBoolDefaultDescription(terraformServiceReadFromTerraformBlockDescription, false),
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"job_resources": schema.SingleNestedAttribute{
				MarkdownDescription: terraformServiceJobResourcesDescription,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cpu_milli": schema.Int64Attribute{
						MarkdownDescription: descriptions.NewInt64MinDescription(
							cpuDescription("Terraform job"),
							int64(terraformservice.MinCPU),
							toInt64Pointer(terraformservice.DefaultCPU),
						),
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(int64(terraformservice.DefaultCPU)),
						Validators: []validator.Int64{
							validators.Int64MinValidator{Min: int64(terraformservice.MinCPU)},
						},
					},
					"ram_mib": schema.Int64Attribute{
						MarkdownDescription: descriptions.NewInt64MinDescription(
							terraformServiceJobRAMDescription,
							int64(terraformservice.MinRAM),
							toInt64Pointer(terraformservice.DefaultRAM),
						),
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(int64(terraformservice.DefaultRAM)),
						Validators: []validator.Int64{
							validators.Int64MinValidator{Min: int64(terraformservice.MinRAM)},
						},
					},
					"gpu": schema.Int64Attribute{
						MarkdownDescription: descriptions.NewInt64MinDescription(
							terraformServiceJobGPUDescription,
							int64(terraformservice.MinGPU),
							toInt64Pointer(terraformservice.DefaultGPU),
						),
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(int64(terraformservice.DefaultGPU)),
						Validators: []validator.Int64{
							validators.Int64MinValidator{Min: int64(terraformservice.MinGPU)},
						},
					},
					"storage_gib": schema.Int64Attribute{
						MarkdownDescription: descriptions.NewInt64MinDescription(
							terraformServiceJobStorageDescription+" Reducing it fails at plan time.",
							int64(terraformservice.MinStorage),
							toInt64Pointer(terraformservice.DefaultStorage),
						),
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(int64(terraformservice.DefaultStorage)),
						Validators: []validator.Int64{
							validators.Int64MinValidator{Min: int64(terraformservice.MinStorage)},
						},
					},
				},
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(
					terraformServiceTimeoutDescription,
					int64(terraformservice.MinTimeoutSec),
					toInt64Pointer(terraformservice.DefaultTimeoutSec),
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(int64(terraformservice.DefaultTimeoutSec)),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: int64(terraformservice.MinTimeoutSec)},
				},
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(iconURIDescription("Terraform service"), terraformservice.DefaultIconURI),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(terraformservice.DefaultIconURI),
			},
			"use_cluster_credentials": schema.BoolAttribute{
				MarkdownDescription: descriptions.NewBoolDefaultDescription(terraformServiceUseClusterCredentialsDescription, false),
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"action_extra_arguments": schema.MapAttribute{
				MarkdownDescription: terraformServiceActionExtraArgumentsDescription,
				Optional:            true,
				ElementType:         types.ListType{ElemType: types.StringType},
			},
			"external_secrets":      externalSecretsSchemaAttribute("Terraform service"),
			"external_secret_files": externalSecretFilesSchemaAttribute("Terraform service"),
			"advanced_settings_json": schema.StringAttribute{
				MarkdownDescription: advancedSettingsJSONDescription("Terraforms/operation/getDefaultTerraformAdvancedSettings"),
				Optional:            true,
				Computed:            true,
				// Documented exception to the config-is-source-of-truth rule: the QOV-2028
				// contract described in advancedSettingsJSONDescription.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: terraformServiceCreatedAtDescription,
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: terraformServiceUpdatedAtDescription,
				Computed:            true,
			},
		},
	}
}

func (r terraformServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan TerraformService
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request from plan
	request, err := plan.toUpsertServiceRequest(nil)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service create", err.Error())
		return
	}

	// Create new terraform service
	terraformSvc, err := r.terraformServiceService.Create(ctx, ToString(plan.EnvironmentID), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service create", err.Error())
		if terraformSvc == nil {
			return
		}
		// The terraform service was created in Qovery and only a follow-up call failed. Save
		// it so Terraform tracks it and taints it for replacement, instead of leaving an
		// orphan that makes every later apply fail with "a terraform named X already exists".
		resp.Diagnostics.Append(resp.State.Set(ctx, convertDomainTerraformServiceToTerraformService(ctx, plan, terraformSvc))...)
		return
	}

	// Convert domain entity to Terraform state
	state := convertDomainTerraformServiceToTerraformService(ctx, plan, terraformSvc)
	tflog.Trace(ctx, "created terraform service", map[string]any{"terraform_service_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r terraformServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve current state
	var state TerraformService
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get terraform service from API
	// Detect import: during import, EnvironmentID is null since only ID is provided
	isTriggeredFromImport := state.EnvironmentID.IsNull()
	terraformSvc, err := r.terraformServiceService.Get(
		ctx,
		ToString(state.ID),
		ToString(state.AdvancedSettingsJson),
		isTriggeredFromImport,
	)
	if handleDomainReadNotFound(ctx, resp, err, "Error on terraform service read") {
		return
	}

	// Convert domain entity to Terraform state
	state = convertDomainTerraformServiceToTerraformService(ctx, state, terraformSvc)
	tflog.Trace(ctx, "read terraform service", map[string]any{"terraform_service_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r terraformServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan and current state
	var plan, state TerraformService
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request from plan
	request, err := plan.toUpsertServiceRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service update", err.Error())
		return
	}

	// Update terraform service
	terraformSvc, err := r.terraformServiceService.Update(ctx, ToString(state.ID), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service update", err.Error())
		return
	}

	// Convert domain entity to Terraform state
	state = convertDomainTerraformServiceToTerraformService(ctx, plan, terraformSvc)
	tflog.Trace(ctx, "updated terraform service", map[string]any{"terraform_service_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r terraformServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve current state
	var state TerraformService
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete terraform service
	err := r.terraformServiceService.Delete(ctx, ToString(state.ID))
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted terraform service", map[string]any{"terraform_service_id": state.ID.ValueString()})
}

func (r terraformServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r terraformServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	warnUnknownAdvancedSettings(ctx, r.advancedSettingsService, domain.TERRAFORM, req.Config, &resp.Diagnostics)
	// Prevent storage reduction
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state TerraformService
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.JobResources != nil && state.JobResources != nil {
		planStorage := plan.JobResources.StorageGiB
		stateStorage := state.JobResources.StorageGiB

		if !planStorage.IsNull() && !stateStorage.IsNull() {
			if ToInt32(planStorage) < ToInt32(stateStorage) {
				resp.Diagnostics.AddError(
					"Storage cannot be reduced",
					fmt.Sprintf("Storage cannot be reduced from %d GiB to %d GiB. Current: %d GiB, Planned: %d GiB",
						ToInt32(stateStorage),
						ToInt32(planStorage),
						ToInt32(stateStorage),
						ToInt32(planStorage),
					),
				)
			}
		}
	}
}

// Helper function for descriptions
func toInt64Pointer(i int32) *int64 {
	i64 := int64(i)
	return &i64
}
