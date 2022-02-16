package qovery

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
	"terraform-provider-qovery/qovery/descriptions"
	"terraform-provider-qovery/qovery/modifiers"
	"terraform-provider-qovery/qovery/validators"
)

const (
	databaseAPIResource       = "database"
	databaseStatusAPIResource = "database status"
)

var (
	// Database State
	databaseStateRunning = "RUNNING"
	databaseStateStopped = "STOPPED"
	databaseStates       = []string{databaseStateRunning, databaseStateStopped}
	databaseStateDefault = databaseStateRunning

	// Database Type
	databaseTypes = []string{"POSTGRESQL", "MYSQL", "MONGODB", "REDIS"}

	// Database Mode
	databaseModes = []string{"MANAGED", "CONTAINER"}

	// Database Accessibility
	databaseAccessibilities      = []string{"PRIVATE", "PUBLIC"}
	databaseAccessibilityDefault = "PUBLIC"

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

type databaseResourceType struct{}

func (r databaseResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					validators.StringEnumValidator{Enum: databaseTypes},
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
					validators.StringEnumValidator{Enum: databaseModes},
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
					validators.StringEnumValidator{Enum: databaseAccessibilities},
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
					"Storage of the database in GB [1024MB = 1GB].",
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
			"state": {
				Description: descriptions.NewStringEnumDescription(
					"State of the database.",
					databaseStates,
					&databaseStateDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(databaseStateDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.StringEnumValidator{Enum: databaseStates},
				},
			},
		},
	}, nil
}

func (r databaseResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return databaseResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type databaseResource struct {
	client *qovery.APIClient
}

// Create qovery database resource
func (r databaseResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Retrieve values from plan
	var plan Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new database
	database, res, err := r.client.DatabasesApi.
		CreateDatabase(ctx, plan.EnvironmentId.Value).
		DatabaseRequest(plan.toCreateDatabaseRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	databaseStatus, apiErr := r.updateDatabaseState(ctx, database, plan)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToDatabase(database, databaseStatus)
	tflog.Trace(ctx, "created database", "database_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery database resource
func (r databaseResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get database from the API
	database, res, err := r.client.DatabaseMainCallsApi.
		GetDatabase(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	databaseStatus, res, err := r.client.DatabaseMainCallsApi.
		GetDatabaseStatus(ctx, database.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseStatusReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToDatabase(database, databaseStatus)
	tflog.Trace(ctx, "read database", "database_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery database resource
func (r databaseResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update database in the backend
	database, res, err := r.client.DatabaseMainCallsApi.
		EditDatabase(ctx, state.Id.Value).
		DatabaseEditRequest(plan.toUpdateDatabaseRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	databaseStatus, apiErr := r.updateDatabaseState(ctx, database, plan)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToDatabase(database, databaseStatus)
	tflog.Trace(ctx, "updated database", "database_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery database resource
func (r databaseResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete database
	res, err := r.client.DatabaseMainCallsApi.
		DeleteDatabase(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		apiErr := databaseDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted database", "database_id", state.Id.Value)

	// Remove database from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery database resource using its id
func (r databaseResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func (r databaseResource) updateDatabaseState(ctx context.Context, database *qovery.DatabaseResponse, plan Database) (*qovery.Status, *apierror.APIError) {
	databaseStatus, res, err := r.client.DatabaseMainCallsApi.
		GetDatabaseStatus(ctx, database.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, databaseStatusReadAPIError(database.Id, res, err)
	}

	if plan.State.Value == databaseStateRunning && databaseStatus.State != databaseStateRunning {
		return r.deployDatabase(ctx, database.Id, databaseStatus.State)
	}

	if plan.State.Value == databaseStateStopped && databaseStatus.State != databaseStateStopped {
		return r.stopDatabase(ctx, database.Id, databaseStatus.State)
	}
	return nil, databaseStatusReadAPIError(database.Id, res, err)
}

func (r databaseResource) deployDatabase(ctx context.Context, databaseID string, currentStatus string) (*qovery.Status, *apierror.APIError) {
	// Deploy database
	switch currentStatus {
	case "QUEUED", "DEPLOYING":
		tflog.Trace(ctx, "database is already being deployed", "database_id", databaseID)
	case "DEPLOYMENT_ERROR":
		_, res, err := r.client.DatabaseActionsApi.
			RestartDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, databaseRestartAPIError(databaseID, res, err)
		}
	default:
		_, res, err := r.client.DatabaseActionsApi.
			DeployDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, databaseDeployAPIError(databaseID, res, err)
		}
	}

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-timeout.C:
			_, res, err := r.client.DatabaseMainCallsApi.
				GetDatabaseStatus(ctx, databaseID).
				Execute()
			return nil, databaseDeployAPIError(databaseID, res, err)
		case <-ticker.C:
			status, res, err := r.client.DatabaseMainCallsApi.
				GetDatabaseStatus(ctx, databaseID).
				Execute()
			if err != nil || res.StatusCode >= 400 {
				return nil, databaseStatusReadAPIError(databaseID, res, err)
			}
			if status.State == databaseStateRunning {
				tflog.Trace(ctx, "deployed database", "database_id", databaseID)
				return status, nil
			}
		}
	}
}

func (r databaseResource) stopDatabase(ctx context.Context, databaseID string, currentStatus string) (*qovery.Status, *apierror.APIError) {
	// Stop database
	switch currentStatus {
	case "QUEUED", "STOPPING":
		tflog.Trace(ctx, "database is already being stopped", "database_id", databaseID)
	default:
		_, res, err := r.client.DatabaseActionsApi.
			StopDatabase(ctx, databaseID).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, databaseDeployAPIError(databaseID, res, err)
		}
	}

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-timeout.C:
			_, res, err := r.client.DatabaseMainCallsApi.
				GetDatabaseStatus(ctx, databaseID).
				Execute()
			return nil, databaseDeployAPIError(databaseID, res, err)
		case <-ticker.C:
			status, res, err := r.client.DatabaseMainCallsApi.
				GetDatabaseStatus(ctx, databaseID).
				Execute()
			if err != nil || res.StatusCode >= 400 {
				return nil, databaseStatusReadAPIError(databaseID, res, err)
			}
			if status.State == databaseStateStopped {
				tflog.Trace(ctx, "stopped database", "database_id", databaseID)
				return status, nil
			}
		}
	}
}

func databaseCreateAPIError(databaseName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseName, apierror.Create, res, err)
}

func databaseReadAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseID, apierror.Read, res, err)
}

func databaseUpdateAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseID, apierror.Update, res, err)
}

func databaseDeleteAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseID, apierror.Delete, res, err)
}

func databaseDeployAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseID, apierror.Deploy, res, err)
}

func databaseRestartAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseAPIResource, databaseID, apierror.Restart, res, err)
}

func databaseStatusReadAPIError(databaseID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(databaseStatusAPIResource, databaseID, apierror.Read, res, err)
}
