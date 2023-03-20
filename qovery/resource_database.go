package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &databaseResource{}
var _ resource.ResourceWithImportState = databaseResource{}

var (
	// Database State
	databaseStates = clientEnumToStringArray([]qovery.StateEnum{
		qovery.STATEENUM_RUNNING,
		qovery.STATEENUM_STOPPED,
	})
	databaseStateDefault = string(qovery.STATEENUM_RUNNING)

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

func (r databaseResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery database resource. This can be used to create and manage Qovery databases.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the database.",
				Type:        types.StringType,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the database.",
				Type:        types.StringType,
				Required:    true,
			},
			"type": {
				Description: descriptions.NewStringEnumDescription(
					"Type of the database [NOTE: can't be updated after creation].",
					databaseTypes,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(databaseTypes),
				},
			},
			"version": {
				Description: "Version of the database",
				Type:        types.StringType,
				Required:    true,
			},
			"mode": {
				Description: descriptions.NewStringEnumDescription(
					"Mode of the database [NOTE: can't be updated after creation].",
					databaseModes,
					nil,
				),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(databaseModes),
				},
			},
			"accessibility": {
				Description: descriptions.NewStringEnumDescription(
					"Accessibility of the database.",
					databaseAccessibilities,
					&databaseAccessibilityDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(databaseAccessibilityDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(databaseAccessibilities),
				},
			},
			"cpu": {
				Description: descriptions.NewInt64MinDescription(
					"CPU of the database in millicores (m) [1000m = 1 CPU].",
					databaseCPUMin,
					&databaseCPUDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(databaseCPUDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: databaseCPUMin},
				},
			},
			"memory": {
				Description: descriptions.NewInt64MinDescription(
					"RAM of the database in MB [1024MB = 1GB].",
					databaseMemoryMin,
					&databaseMemoryDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(databaseMemoryDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: databaseMemoryMin},
				},
			},
			"storage": {
				Description: descriptions.NewInt64MinDescription(
					"Storage of the database in GB [1024MB = 1GB] [NOTE: can't be updated after creation].",
					databaseStorageMin,
					&databaseStorageDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewInt64DefaultModifier(databaseStorageDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: databaseStorageMin},
				},
			},
			"external_host": {
				Description: "The database external FQDN host (only if your database is publicly accessible with ACCESSIBILITY = PUBLIC)",
				Type:        types.StringType,
				Computed:    true,
			},
			"internal_host": {
				Description: "The database internal host (Recommended for your application)",
				Type:        types.StringType,
				Computed:    true,
			},
			"port": {
				Description: "The port to connect to your database",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"login": {
				Description: "The login to connect to your database",
				Type:        types.StringType,
				Computed:    true,
			},
			"password": {
				Description: "The password to connect to your database",
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
	database, apiErr := r.client.CreateDatabase(ctx, plan.EnvironmentId.Value, request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToDatabase(database)
	tflog.Trace(ctx, "created database", map[string]interface{}{"database_id": state.Id.Value})

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
	database, apiErr := r.client.GetDatabase(ctx, state.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToDatabase(database)
	tflog.Trace(ctx, "read database", map[string]interface{}{"database_id": state.Id.Value})

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
	database, apiErr := r.client.UpdateDatabase(ctx, state.Id.Value, request)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToDatabase(database)
	tflog.Trace(ctx, "updated database", map[string]interface{}{"database_id": state.Id.Value})

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
	apiErr := r.client.DeleteDatabase(ctx, state.Id.Value)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted database", map[string]interface{}{"database_id": state.Id.Value})

	// Remove database from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery database resource using its id
func (r databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
