package qovery

import (
	"context"
	"fmt"
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
var _ resource.ResourceWithConfigure = &databaseResource{}
var _ resource.ResourceWithImportState = databaseResource{}

var (
	// Database State
	databaseStates = clientEnumToStringArray([]qovery.StateEnum{
		qovery.STATEENUM_DEPLOYED,
		qovery.STATEENUM_STOPPED,
	})
	databaseStateDefault = string(qovery.STATEENUM_DEPLOYED)

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

	// Database Instance Type
	databaseInstanceTypeDefault *string = nil
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
		Description: "Provides a Qovery database resource. This can be used to create and manage Qovery databases.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the database.",
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
			"name": schema.StringAttribute{
				Description: "Name of the database.",
				Required:    true,
			},
			"icon_uri": schema.StringAttribute{
				Description: "Icon URI representing the database.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Type of the database [NOTE: can't be updated after creation].",
					databaseTypes,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseTypes),
				},
			},
			"version": schema.StringAttribute{
				Description: "Version of the database",
				Required:    true,
			},
			"mode": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Mode of the database [NOTE: can't be updated after creation].",
					databaseModes,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseModes),
				},
			},
			"accessibility": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Accessibility of the database.",
					databaseAccessibilities,
					&databaseAccessibilityDefault,
				),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(databaseAccessibilityDefault),
				Validators: []validator.String{
					validators.NewStringEnumValidator(databaseAccessibilities),
				},
			},
			"instance_type": schema.StringAttribute{
				Description: "Instance type of the database.",
				Optional:    true,
				Computed:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"CPU of the database in millicores (m) [1000m = 1 CPU].",
					databaseCPUMin,
					&databaseCPUDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(databaseCPUDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseCPUMin},
				},
			},
			"memory": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"RAM of the database in MB [1024MB = 1GB].",
					databaseMemoryMin,
					&databaseMemoryDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(databaseMemoryDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseMemoryMin},
				},
			},
			"storage": schema.Int64Attribute{
				Description: descriptions.NewInt64MinDescription(
					"Storage of the database in GB [1024MB = 1GB] [NOTE: can't be updated after creation].",
					databaseStorageMin,
					&databaseStorageDefault,
				),
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(databaseStorageDefault),
				Validators: []validator.Int64{
					validators.Int64MinValidator{Min: databaseStorageMin},
				},
			},
			"external_host": schema.StringAttribute{
				Description: "The database external FQDN host [NOTE: only if your container is using a publicly accessible port].",
				Computed:    true,
			},
			"internal_host": schema.StringAttribute{
				Description: "The database internal host (Recommended for your application)",
				Computed:    true,
			},
			"deployment_stage_id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Optional:    true,
				Computed:    true,
			},
			"is_skipped": schema.BoolAttribute{
				Description: "If true, the service is excluded from environment-level bulk deployments while remaining assigned to its deployment stage.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"port": schema.Int64Attribute{
				Description: "The port to connect to your database",
				Computed:    true,
			},
			"login": schema.StringAttribute{
				Description: "The login to connect to your database",
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password to connect to your database",
				Computed:    true,
			},
			"annotations_group_ids": schema.SetAttribute{
				Description: "List of annotations group ids",
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_group_ids": schema.SetAttribute{
				Description: "List of labels group ids",
				Optional:    true,
				ElementType: types.StringType,
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
	state := convertResponseToDatabase(ctx, plan, database)
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
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToDatabase(ctx, state, database)
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
	state = convertResponseToDatabase(ctx, plan, database)
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
