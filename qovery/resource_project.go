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
)

const (
	projectAPIResource                    = "project"
	projectEnvironmentVariableAPIResource = "project environment variable"
)

type projectResourceType struct{}

func (r projectResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery project resource. This can be used to create and manage Qovery projects.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the project.",
				Type:        types.StringType,
				Computed:    true,
			},
			"organization_id": {
				Description: "Id of the organization.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the project.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: "Description of the project.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"environment_variables": {
				Description: "List of environment variables linked to this project.",
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

func (r projectResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return projectResource{
		client: p.(*provider).GetClient(),
	}, nil
}

type projectResource struct {
	client *qovery.APIClient
}

// Create qovery project resource
func (r projectResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Retrieve values from plan
	var plan Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new project
	project, res, err := r.client.ProjectsApi.
		CreateProject(ctx, plan.OrganizationId.Value).
		ProjectRequest(plan.toUpsertProjectRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectCreateAPIError(plan.Name.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	projectVariables, apiErr := r.updateProjectEnvironmentVariables(ctx, project.Id, plan.EnvironmentVariables)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Initialize state values
	state := convertResponseToProject(project, projectVariables)
	tflog.Trace(ctx, "created project", "project_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery project resource
func (r projectResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from the API
	project, res, err := r.client.ProjectMainCallsApi.
		GetProject(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	projectVariables, res, err := r.client.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, project.Id).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectEnvironmentVariableReadAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Refresh state values
	state = convertResponseToProject(project, projectVariables)
	tflog.Trace(ctx, "read project", "project_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery project resource
func (r projectResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan and current state
	var plan, state Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update project in the backend
	project, res, err := r.client.ProjectMainCallsApi.
		EditProject(ctx, state.Id.Value).
		ProjectRequest(plan.toUpsertProjectRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		apiErr := projectUpdateAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	projectVariables, apiErr := r.updateProjectEnvironmentVariables(ctx, project.Id, plan.EnvironmentVariables)
	if apiErr != nil {
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	// Update state values
	state = convertResponseToProject(project, projectVariables)
	tflog.Trace(ctx, "updated project", "project_id", state.Id.Value)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery project resource
func (r projectResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Get current state
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete project
	res, err := r.client.ProjectMainCallsApi.
		DeleteProject(ctx, state.Id.Value).
		Execute()
	if err != nil || res.StatusCode >= 300 {
		apiErr := projectDeleteAPIError(state.Id.Value, res, err)
		resp.Diagnostics.AddError(apiErr.Summary(), apiErr.Detail())
		return
	}

	tflog.Trace(ctx, "deleted project", "project_id", state.Id.Value)

	// Remove project from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery project resource using its id
func (r projectResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func (r projectResource) updateProjectEnvironmentVariables(ctx context.Context, projectID string, plan []EnvironmentVariable) (*qovery.EnvironmentVariableResponseList, *apierror.APIError) {
	projectVariables, res, err := r.client.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, projectEnvironmentVariableReadAPIError(projectID, res, err)
	}

	diff := diffEnvironmentVariables(
		convertResponseToEnvironmentVariables(projectVariables, EnvironmentVariableScopeProject),
		plan,
	)

	for _, variable := range diff.ToRemove {
		res, err := r.client.ProjectEnvironmentVariableApi.
			DeleteProjectEnvironmentVariable(ctx, projectID, variable.Id.Value).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, projectEnvironmentVariableDeleteAPIError(variable.Id.Value, res, err)
		}
	}

	for _, variable := range diff.ToUpdate {
		_, res, err := r.client.ProjectEnvironmentVariableApi.
			EditProjectEnvironmentVariable(ctx, projectID, variable.Id.Value).
			EnvironmentVariableEditRequest(variable.toUpdateRequest()).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, projectEnvironmentVariableUpdateAPIError(variable.Id.Value, res, err)
		}
	}

	for _, variable := range diff.ToCreate {
		_, res, err := r.client.ProjectEnvironmentVariableApi.
			CreateProjectEnvironmentVariable(ctx, projectID).
			EnvironmentVariableRequest(variable.toCreateRequest()).
			Execute()
		if err != nil || res.StatusCode >= 400 {
			return nil, projectEnvironmentVariableCreateAPIError(variable.Key.Value, res, err)
		}
	}

	projectVariables, res, err = r.client.ProjectEnvironmentVariableApi.
		ListProjectEnvironmentVariable(ctx, projectID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, projectEnvironmentVariableReadAPIError(projectID, res, err)
	}
	return projectVariables, nil
}

func projectCreateAPIError(projectName string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectName, apierror.Create, res, err)
}

func projectReadAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Read, res, err)
}

func projectUpdateAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Update, res, err)
}

func projectDeleteAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectAPIResource, projectID, apierror.Delete, res, err)
}

// Project Environment Variable
func projectEnvironmentVariableCreateAPIError(projectID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectEnvironmentVariableAPIResource, projectID, apierror.Create, res, err)
}

func projectEnvironmentVariableReadAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectEnvironmentVariableAPIResource, variableID, apierror.Read, res, err)
}

func projectEnvironmentVariableUpdateAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectEnvironmentVariableAPIResource, variableID, apierror.Update, res, err)
}

func projectEnvironmentVariableDeleteAPIError(variableID string, res *http.Response, err error) *apierror.APIError {
	return apierror.New(projectEnvironmentVariableAPIResource, variableID, apierror.Delete, res, err)
}
