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

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &terraformServiceResource{}
var _ resource.ResourceWithImportState = terraformServiceResource{}
var _ resource.ResourceWithModifyPlan = &terraformServiceResource{}

type terraformServiceResource struct {
	terraformServiceService terraformservice.Service
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
}

func (r terraformServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery Terraform service resource. This can be used to create and manage Qovery terraform services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the terraform service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the terraform service.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the terraform service.",
				Optional:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Specify if the terraform service will be automatically updated on every new commit.",
				Required:    true,
			},
			"git_repository": schema.SingleNestedAttribute{
				Description: "Terraform service git repository configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "Git repository URL.",
						Required:    true,
					},
					"branch": schema.StringAttribute{
						Description: "Git branch.",
						Optional:    true,
					},
					"root_path": schema.StringAttribute{
						Description: "Git root path.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(terraformservice.DefaultRootPath),
					},
					"git_token_id": schema.StringAttribute{
						Description: "Git token ID for private repositories.",
						Optional:    true,
					},
				},
			},
			"tfvars_files": schema.ListAttribute{
				Description: "List of .tfvars file paths relative to the root path.",
				Required:    true,
				ElementType: types.StringType,
			},
			"variables": schema.SetNestedAttribute{
				Description: "Terraform variables.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Variable key.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Variable value.",
							Required:    true,
						},
						"is_secret": schema.BoolAttribute{
							Description: "Is this variable a secret.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"backend": schema.SingleNestedAttribute{
				Description: "Terraform backend configuration. Exactly one backend type must be specified.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"kubernetes": schema.SingleNestedAttribute{
						Description: "Use Kubernetes backend for state management.",
						Optional:    true,
						Attributes:  map[string]schema.Attribute{},
					},
					"user_provided": schema.SingleNestedAttribute{
						Description: "Use user-provided backend configuration (configured in Terraform code).",
						Optional:    true,
						Attributes:  map[string]schema.Attribute{},
					},
				},
			},
			"engine": schema.StringAttribute{
				Description: "Terraform engine to use (TERRAFORM or OPEN_TOFU).",
				Required:    true,
				Validators: []validator.String{
					validators.NewStringEnumValidator([]string{"TERRAFORM", "OPEN_TOFU"}),
				},
			},
			"engine_version": schema.SingleNestedAttribute{
				Description: "Terraform/OpenTofu engine version configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"explicit_version": schema.StringAttribute{
						Description: "Explicit version to use for the Terraform/OpenTofu binary.",
						Required:    true,
					},
					"read_from_terraform_block": schema.BoolAttribute{
						Description: "Whether to read the version from the terraform block in the code.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"job_resources": schema.SingleNestedAttribute{
				Description: "Resource allocation for the Terraform job.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"cpu_milli": schema.Int64Attribute{
						Description: descriptions.NewInt64MinDescription(
							"CPU of the terraform job in millicores (m) [1000m = 1 CPU].",
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
						Description: descriptions.NewInt64MinDescription(
							"RAM of the terraform job in MiB [1024 MiB = 1GiB].",
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
						Description: descriptions.NewInt64MinDescription(
							"Number of GPUs for the terraform job.",
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
						Description: descriptions.NewInt64MinDescription(
							"Storage of the terraform job in GiB [1 GiB = 1024 MiB]. WARNING: Cannot be reduced after creation.",
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
				Description: descriptions.NewInt64MinDescription(
					"Timeout in seconds for Terraform operations.",
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
				Description: "Icon URI representing the terraform service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(terraformservice.DefaultIconURI),
			},
			"use_cluster_credentials": schema.BoolAttribute{
				Description: "Use cluster credentials for cloud provider authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"action_extra_arguments": schema.MapAttribute{
				Description: "Extra CLI arguments for specific Terraform actions (plan, apply, destroy).",
				Optional:    true,
				ElementType: types.ListType{ElemType: types.StringType},
			},
			"advanced_settings_json": schema.StringAttribute{
				Description: "Advanced settings in JSON format.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation date of the terraform service.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update date of the terraform service.",
				Computed:    true,
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
	var isTriggeredFromImport = state.EnvironmentID.IsNull()
	terraformSvc, err := r.terraformServiceService.Get(
		ctx,
		ToString(state.ID),
		ToString(state.AdvancedSettingsJson),
		isTriggeredFromImport,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error on terraform service read", err.Error())
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
