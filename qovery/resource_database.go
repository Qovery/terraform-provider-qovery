package qovery

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"

	"terraform-provider-qovery/qovery/apierror"
	"terraform-provider-qovery/qovery/descriptions"
	"terraform-provider-qovery/qovery/validators"
)

const databaseAPIResource = "database"

var (
	// Database Type
	databaseTypes = []string{"POSTGRESQL", "MYSQL", "MONGODB", "REDIS"}

	// Database Mode
	databaseModes = []string{"MANAGED", "CONTAINER"}

	// Database Accessibility
	databaseAccessibilities      = []string{"PRIVATE", "PUBLIC"}
	databaseAccessibilityDefault = "PRIVATE"

	// Database CPU
	databaseCPUMin     int64 = 250
	databaseCPUDefault int64 = 250

	// Database Memory
	databaseMemoryMin     int64 = 100
	databaseMemoryDefault int64 = 256

	// Database Storage
	databaseStorageMin     int64 = 10240
	databaseStorageDefault int64 = 10240
)

type databaseResourceData struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	Mode          types.String `tfsdk:"mode"`
	Accessibility types.String `tfsdk:"accessibility"`
	CPU           types.Int64  `tfsdk:"cpu"`
	Memory        types.Int64  `tfsdk:"memory"`
	Storage       types.Int64  `tfsdk:"storage"`
}

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
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: databaseMemoryMin},
				},
			},
			"storage": {
				Description: descriptions.NewInt64MinDescription(
					"Storage of the database in MB [1024MB = 1GB].",
					databaseStorageMin,
					&databaseStorageDefault,
				),
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
				Validators: []tfsdk.AttributeValidator{
					validators.Int64MinValidator{Min: databaseStorageMin},
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
	var plan databaseResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new database
	payload := qovery.DatabaseRequest{
		Name:    plan.Name.Value,
		Type:    plan.Type.Value,
		Version: plan.Version.Value,
		Mode:    plan.Mode.Value,
	}
	if !plan.Accessibility.Null && !plan.Accessibility.Unknown {
		payload.Accessibility = &plan.Accessibility.Value
	}
	if !plan.CPU.Null && !plan.CPU.Unknown {
		payload.Cpu = int32ToInt32Ptr(int32(plan.CPU.Value))
	}
	if !plan.Memory.Null && !plan.Memory.Unknown {
		payload.Memory = int32ToInt32Ptr(int32(plan.Memory.Value))
	}
	if !plan.Storage.Null && !plan.Storage.Unknown {
		payload.Storage = int32ToInt32Ptr(int32(plan.Storage.Value))
	}
	database, res, err := r.client.DatabasesApi.
		CreateDatabase(ctx, plan.EnvironmentId.Value).
		DatabaseRequest(payload).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := databaseResourceData{
		Id:            types.String{Value: database.Id},
		EnvironmentId: plan.EnvironmentId,
		Name:          types.String{Value: database.Name},
		Type:          types.String{Value: database.Type},
		Version:       types.String{Value: database.Version},
		Mode:          types.String{Value: database.Mode},
		Accessibility: types.String{Value: *database.Accessibility},
		CPU:           types.Int64{Value: int64(*database.Cpu)},
		Memory:        types.Int64{Value: int64(*database.Memory)},
		Storage:       types.Int64{Value: int64(*database.Storage)},
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery database resource
func (r databaseResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state databaseResourceData
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

	toRefresh := &databaseResourceData{
		EnvironmentId: types.String{Value: database.Environment.Id},
		Name:          types.String{Value: database.Name},
		Type:          types.String{Value: database.Type},
		Version:       types.String{Value: database.Version},
		Mode:          types.String{Value: database.Mode},
		Accessibility: types.String{Value: *database.Accessibility},
		CPU:           types.Int64{Value: int64(*database.Cpu)},
		Memory:        types.Int64{Value: int64(*database.Memory)},
		Storage:       types.Int64{Value: int64(*database.Storage)},
	}

	// Refresh state values
	state.EnvironmentId = toRefresh.EnvironmentId
	state.Name = toRefresh.Name
	state.Type = toRefresh.Type
	state.Version = toRefresh.Version
	state.Mode = toRefresh.Mode
	state.Accessibility = toRefresh.Accessibility
	state.CPU = toRefresh.CPU
	state.Memory = toRefresh.Memory
	state.Storage = toRefresh.Storage

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery database resource
func (r databaseResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state databaseResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update cluster in the backend
	payload := qovery.DatabaseEditRequest{
		Name:    &plan.Name.Value,
		Version: &plan.Version.Value,
	}
	if !plan.Accessibility.Null && !plan.Accessibility.Unknown {
		payload.Accessibility = &plan.Accessibility.Value
	}
	if !plan.CPU.Null && !plan.CPU.Unknown {
		payload.Cpu = int32ToInt32Ptr(int32(plan.CPU.Value))
	}
	if !plan.Memory.Null && !plan.Memory.Unknown {
		payload.Memory = int32ToInt32Ptr(int32(plan.Memory.Value))
	}
	if !plan.Storage.Null && !plan.Storage.Unknown {
		payload.Storage = int32ToInt32Ptr(int32(plan.Storage.Value))
	}

	database, res, err := r.client.DatabaseMainCallsApi.
		EditDatabase(ctx, state.Id.Value).
		DatabaseEditRequest(payload).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := databaseUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	toUpdate := &databaseResourceData{
		EnvironmentId: types.String{Value: database.Environment.Id},
		Name:          types.String{Value: database.Name},
		Type:          types.String{Value: database.Type},
		Version:       types.String{Value: database.Version},
		Mode:          types.String{Value: database.Mode},
		Accessibility: types.String{Value: *database.Accessibility},
		CPU:           types.Int64{Value: int64(*database.Cpu)},
		Memory:        types.Int64{Value: int64(*database.Memory)},
		Storage:       types.Int64{Value: int64(*database.Storage)},
	}

	// Update state values
	state.EnvironmentId = toUpdate.EnvironmentId
	state.Name = toUpdate.Name
	state.Type = toUpdate.Type
	state.Version = toUpdate.Version
	state.Mode = toUpdate.Mode
	state.Accessibility = toUpdate.Accessibility
	state.CPU = toUpdate.CPU
	state.Memory = toUpdate.Memory
	state.Storage = toUpdate.Storage

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery database resource
func (r databaseResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state databaseResourceData
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

	// Remove database from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery database resource using its id
func (r databaseResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
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
