package qovery

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

type databaseData struct {
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
					"Type of the database.",
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
					"Mode of the database.",
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
}

// Read qovery database resource
func (r databaseResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}

// Update qovery database resource
func (r databaseResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete qovery database resource
func (r databaseResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}

// ImportState imports a qovery database resource using its id
func (r databaseResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStateNotImplemented(ctx, "", resp)
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
