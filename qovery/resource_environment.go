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

const (
	environmentAPIResource                    = "environment"
	environmentEnvironmentVariableAPIResource = "environment environment variable"
)

var (
	// Environment Mode
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
			"environment_variables": {
				Description: "List of environment variables linked to this environment.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
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

	environmentVariables, apiErr := r.updateEnvironmentEnvironmentVariables(ctx, environment.Id, plan.EnvironmentVariables)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToEnvironment(environment, environmentVariables)
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

	environmentVariables, res, err := r.client.EnvironmentVariableApi.
		ListEnvironmentEnvironmentVariable(ctx, environment.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := environmentEnvironmentVariableReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToEnvironment(environment, environmentVariables)
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

	environmentVariables, apiErr := r.updateEnvironmentEnvironmentVariables(ctx, environment.Id, plan.EnvironmentVariables)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// TODO restart the whole environment if env vars have been changed

	// Update state values
	state = convertResponseToEnvironment(environment, environmentVariables)
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

func (r environmentResource) updateEnvironmentEnvironmentVariables(ctx context.Context, environmentID string, plan []EnvironmentVariable) (*qovery.EnvironmentVariableResponseList, *apierror.APIError) {
	environmentVariables, res, err := r.client.EnvironmentVariableApi.
		ListEnvironmentEnvironmentVariable(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, environmentEnvironmentVariableReadAPIError(environmentID, res, err)
	}

	diff := diffEnvironmentVariables(
		convertResponseToEnvironmentVariables(environmentVariables, EnvironmentVariableScopeEnvironment),
		plan,
	)

	for _, variable := range diff.ToRemove {
		res, err := r.client.EnvironmentVariableApi.
			DeleteEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id.Value).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, environmentEnvironmentVariableDeleteAPIError(variable.Id.Value, res, err)
		}
	}

	for _, variable := range diff.ToUpdate {
		_, res, err := r.client.EnvironmentVariableApi.
			EditEnvironmentEnvironmentVariable(ctx, environmentID, variable.Id.Value).
			EnvironmentVariableEditRequest(variable.toUpdateRequest()).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, environmentEnvironmentVariableUpdateAPIError(variable.Id.Value, res, err)
		}
	}

	for _, variable := range diff.ToCreate {
		_, res, err := r.client.EnvironmentVariableApi.
			CreateEnvironmentEnvironmentVariable(ctx, environmentID).
			EnvironmentVariableRequest(variable.toCreateRequest()).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, environmentEnvironmentVariableCreateAPIError(variable.Key.Value, res, err)
		}
	}

	environmentVariables, res, err = r.client.EnvironmentVariableApi.
		ListEnvironmentEnvironmentVariable(ctx, environmentID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, environmentEnvironmentVariableReadAPIError(environmentID, res, err)
	}
	return environmentVariables, nil
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

// Environment Environment Variable
func environmentEnvironmentVariableCreateAPIError(environmentID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentEnvironmentVariableAPIResource, environmentID, apierror.Create, res, err)
}

func environmentEnvironmentVariableReadAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentEnvironmentVariableAPIResource, variableID, apierror.Read, res, err)
}

func environmentEnvironmentVariableUpdateAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentEnvironmentVariableAPIResource, variableID, apierror.Update, res, err)
}

func environmentEnvironmentVariableDeleteAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(environmentEnvironmentVariableAPIResource, variableID, apierror.Delete, res, err)
}
