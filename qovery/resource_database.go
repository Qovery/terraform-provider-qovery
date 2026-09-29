package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure      = &databaseResource{}
	_ resource.ResourceWithImportState    = databaseResource{}
	_ resource.ResourceWithValidateConfig = databaseResource{}
)

var (
	// Database Type
	databaseTypes = clientEnumToStringArray(qovery.AllowedDatabaseTypeEnumEnumValues)

	// Database Mode
	databaseModes = clientEnumToStringArray(qovery.AllowedDatabaseModeEnumEnumValues)

	// Database Accessibility
	databaseAccessibilities      = clientEnumToStringArray(qovery.AllowedDatabaseAccessibilityEnumEnumValues)
	databaseAccessibilityDefault = string(qovery.DATABASEACCESSIBILITYENUM_PUBLIC)

	// Database CPU
	databaseCPUMin     int64 = 250
	databaseCPUDefault int64 = 250

	// Database Memory
	databaseMemoryMin     int64 = 100
	databaseMemoryDefault int64 = 256

	// Database Storage
	databaseStorageMin     int64 = 10
	databaseStorageDefault int64 = 10
)

type databaseResource struct {
	client *client.Client
}

func newDatabaseResource() resource.Resource {
	return &databaseResource{}
}

func (r databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *databaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = provider.client
}

func (r databaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery database: PostgreSQL, MySQL, MongoDB or Redis, run as a container on the cluster or as a managed service of the cloud provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("database"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: environmentIDDescription + recreatesOnChange("database"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("database"),
				Required:            true,
			},
			"icon_uri": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(iconURIDescription("database"), databaseIconURIDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(databaseIconURIDefault),
			},
			"type": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(databaseTypeDescription+databaseCannotChangeNote, databaseTypes, nil),
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseTypes),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: databaseVersionDescription + " The available versions depend on `type` and `mode`.",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: databaseModeDescription + databaseCannotChangeNote,
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseModes),
				},
			},
			"accessibility": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringDefaultDescription(databaseAccessibilityDescription, databaseAccessibilityDefault),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(databaseAccessibilityDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseAccessibilities),
				},
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: databaseInstanceTypeDescription + " Required when `mode` is `MANAGED`; setting it when `mode` is `CONTAINER` raises a warning.",
				Optional:            true,
				// Computed because the Qovery API derives the value of a CONTAINER database;
				// ValidateConfig requires it for MANAGED.
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(cpuDescription("database")+databaseContainerOnlyNote, databaseCPUMin, &databaseCPUDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(databaseCPUDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseCPUMin},
				},
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(memoryDescription("database")+databaseContainerOnlyNote, databaseMemoryMin, &databaseMemoryDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(databaseMemoryDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseMemoryMin},
				},
			},
			"storage": schema.Int64Attribute{
				MarkdownDescription: descriptions.NewInt64MinDescription(databaseStorageDescription, databaseStorageMin, &databaseStorageDefault),
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(databaseStorageDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseStorageMin},
				},
			},
			"external_host": schema.StringAttribute{
				MarkdownDescription: databaseExternalHostDescription,
				Computed:            true,
			},
			"internal_host": schema.StringAttribute{
				MarkdownDescription: internalHostDescription("database"),
				Computed:            true,
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
			"port": schema.Int64Attribute{
				MarkdownDescription: databasePortDescription,
				Computed:            true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: databaseLoginDescription,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: databasePasswordDescription,
				Computed:            true,
				Sensitive:           true,
			},
			"annotations_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("annotations", "the pods of a `CONTAINER` database"),
				Optional:            true,
				ElementType:         types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				MarkdownDescription: groupIDsDescription("labels", "the pods of a `CONTAINER` database"),
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Create qovery database resource
func (r databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new database
	request, err := plan.toCreateDatabaseRequest()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	database, apiErr := r.client.CreateDatabase(ctx, plan.EnvironmentId.ValueString(), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToDatabase(plan, database)
	tflog.Trace(ctx, "created database", map[string]any{"database_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery database resource
func (r databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get database from the API
	database, apiErr := r.client.GetDatabase(ctx, state.Id.ValueString())
	if handleReadNotFound(ctx, resp, apiErr) {
		return
	}

	// Refresh state values
	state = convertResponseToDatabase(state, database)
	tflog.Trace(ctx, "read database", map[string]any{"database_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery database resource
func (r databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update database in the backend
	request, err := plan.toUpdateDatabaseRequest()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	database, apiErr := r.client.UpdateDatabase(ctx, state.Id.ValueString(), request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToDatabase(plan, database)
	tflog.Trace(ctx, "updated database", map[string]any{"database_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery database resource
func (r databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete database
	apiErr := r.client.DeleteDatabase(ctx, state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted database", map[string]any{"database_id": state.Id.ValueString()})

	// Remove database from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery database resource using its id
func (r databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig checks instance_type against the mode: q-core requires it for a MANAGED
// database and ignores it for a CONTAINER database, whose type it derives.
func (r databaseResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var mode, instanceType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("mode"), &mode)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("instance_type"), &instanceType)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateDatabaseInstanceType(mode, instanceType)...)
}

func validateDatabaseInstanceType(mode, instanceType types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	if mode.IsNull() || mode.IsUnknown() {
		return diags
	}

	switch mode.ValueString() {
	case string(qovery.DATABASEMODEENUM_MANAGED):
		if instanceType.IsNull() {
			diags.AddAttributeError(
				path.Root("instance_type"),
				"Missing instance_type",
				"A MANAGED database requires instance_type: the Qovery API rejects a MANAGED database without one. "+
					"Set it to an instance type of your cloud provider, for example db.t3.micro on AWS.",
			)
		}
	case string(qovery.DATABASEMODEENUM_CONTAINER):
		if !instanceType.IsNull() {
			diags.AddAttributeWarning(
				path.Root("instance_type"),
				"instance_type is ignored for a CONTAINER database",
				"The Qovery API ignores instance_type for a CONTAINER database and reports the type it derives from the cluster. "+
					"Remove instance_type from the configuration.",
			)
		}
	}
	return diags
}
