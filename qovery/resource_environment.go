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

const environmentAPIResource = "environment"

var (
	environmentModes       = []string{"PRODUCTION", "DEVELOPMENT", "STAGING", "PREVIEW"}
	environmentModeDefault = "DEVELOPMENT"
)

type environmentResourceData struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	ClusterId types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
	Mode      types.String `tfsdk:"mode"`
}

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
	var plan environmentResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new environment
	payload := qovery.EnvironmentRequest{
		Name: plan.Name.Value,
	}
	if !plan.ClusterId.Null && !plan.ClusterId.Unknown {
		payload.Cluster = &plan.ClusterId.Value
	}
	if !plan.Mode.Null && !plan.Mode.Unknown {
		payload.Mode = &plan.Mode.Value
	}
	environment, res, err := r.client.EnvironmentsApi.
		CreateEnvironment(ctx, plan.ProjectId.Value).
		EnvironmentRequest(payload).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := environmentResourceData{
		Id:        types.String{Value: environment.Id},
		ProjectId: types.String{Value: environment.Project.Id},
		ClusterId: types.String{Value: environment.ClusterId},
		Name:      types.String{Value: environment.Name},
		Mode:      types.String{Value: environment.Mode},
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery environment resource
func (r environmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state environmentResourceData
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

	toRefresh := &environmentResourceData{
		ProjectId: types.String{Value: environment.Project.Id},
		ClusterId: types.String{Value: environment.ClusterId},
		Name:      types.String{Value: environment.Name},
		Mode:      types.String{Value: environment.Mode},
	}

	// Refresh state values
	state.ProjectId = toRefresh.ProjectId
	state.ClusterId = toRefresh.ClusterId
	state.Name = toRefresh.Name
	state.Mode = toRefresh.Mode

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery environment resource
func (r environmentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state environmentResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update environment in the backend
	payload := qovery.EnvironmentEditRequest{
		Name: &plan.Name.Value,
	}
	environment, res, err := r.client.EnvironmentMainCallsApi.
		EditEnvironment(ctx, state.Id.Value).
		EnvironmentEditRequest(payload).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	toRefresh := &environmentResourceData{
		ProjectId: types.String{Value: environment.Project.Id},
		ClusterId: types.String{Value: environment.ClusterId},
		Name:      types.String{Value: environment.Name},
		Mode:      types.String{Value: environment.Mode},
	}

	// Refresh state values
	state.ProjectId = toRefresh.ProjectId
	state.ClusterId = toRefresh.ClusterId
	state.Name = toRefresh.Name
	state.Mode = toRefresh.Mode

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery environment resource
func (r environmentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state environmentResourceData
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
