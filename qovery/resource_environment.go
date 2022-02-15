package qovery

import (
	"context"
	"net/http"

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

const environmentAPIResource = "environment"

var (
	environmentModes       = []string{"PRODUCTION", "DEVELOPMENT", "STAGING", "PREVIEW"}
	environmentModeDefault = "DEVELOPMENT"
)

type environmentResourceType struct{}

func (r environmentResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery environment resource. This can be used to create and manage Qovery environments.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Computed:    true,
			},
			"project_id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Required:    true,
			},
			"cluster_id": {
				Description: "Id of the cluster [NOTE: can't be updated after creation].",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Description: "Name of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"mode": {
				Description: descriptions.NewStringEnumDescription(
					"Mode of the environment [NOTE: can't be updated after creation].",
					environmentModes,
					&environmentModeDefault,
				),
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringDefaultModifier(environmentModeDefault),
				},
				Validators: []tfsdk.AttributeValidator{
					validators.StringEnumValidator{Enum: environmentModes},
				},
			},
		},
	}, nil
}

func (r environmentResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return environmentResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type environmentResource struct {
	client *qovery.APIClient
}

// Create qovery environment resource
func (r environmentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Retrieve values from plan
	var plan Environment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new environment
	environment, res, err := r.client.EnvironmentsApi.
		CreateEnvironment(ctx, plan.ProjectId.Value).
		EnvironmentRequest(plan.toCreateEnvironmentRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToEnvironment(environment)
	tflog.Trace(ctx, "created environment", "environment_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery environment resource
func (r environmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Environment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment from the API
	environment, res, err := r.client.EnvironmentMainCallsApi.
		GetEnvironment(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToEnvironment(environment)
	tflog.Trace(ctx, "read environment", "environment_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery environment resource
func (r environmentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state Environment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update environment in the backend
	environment, res, err := r.client.EnvironmentMainCallsApi.
		EditEnvironment(ctx, state.Id.Value).
		EnvironmentEditRequest(plan.toUpdateEnvironmentRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToEnvironment(environment)
	tflog.Trace(ctx, "updated environment", "environment_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery environment resource
func (r environmentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Environment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete environment
	res, err := r.client.EnvironmentMainCallsApi.
		DeleteEnvironment(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		apiErr := environmentDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted environment", "environment_id", state.Id.Value)

	// Remove environment from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery environment resource using its id
func (r environmentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func environmentCreateAPIError(environmentName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentName, apierror.Create, res, err)
}

func environmentReadAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Read, res, err)
}

func environmentUpdateAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Update, res, err)
}

func environmentDeleteAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentAPIResource, environmentID, apierror.Delete, res, err)
}
