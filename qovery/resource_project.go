package qovery

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.Resource = projectResource{}
var _ resource.ResourceWithImportState = projectResource{}

type projectResource struct {
	projectService project.Service
}

func NewProjectResource(service project.Service) func() resource.Resource {
	return func() resource.Resource {
		return projectResource{
			projectService: service,
		}
	}
}

func (r projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r projectResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"built_in_environment_variables": {
				Description: "List of built-in environment variables linked to this project.",
				Computed:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
					"value": {
						Description: "Value of the environment variable.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"environment_variables": {
				Description: "List of environment variables linked to this project.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
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
				}),
			},
			"secrets": {
				Description: "List of secrets linked to this project.",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Id of the secret.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "Key of the secret.",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Value of the secret.",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
		},
	}, nil
}

// Create qovery project resource
func (r projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new project

	proj, err := r.projectService.Create(ctx, plan.OrganizationId.Value, plan.toCreateServiceRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on project create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainProjectToProject(plan, proj)
	tflog.Trace(ctx, "created project", map[string]interface{}{"project_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery project resource
func (r projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from the API
	proj, err := r.projectService.Get(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on project read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainProjectToProject(state, proj)
	tflog.Trace(ctx, "read project", map[string]interface{}{"project_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery project resource
func (r projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Project
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update project in the backend
	proj, err := r.projectService.Update(ctx, state.Id.Value, plan.toUpdateServiceRequest(state))
	if err != nil {
		resp.Diagnostics.AddError("Error on project update", err.Error())
		return
	}

	// Update state values
	state = convertDomainProjectToProject(plan, proj)
	tflog.Trace(ctx, "updated project", map[string]interface{}{"project_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery project resource
func (r projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Project
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete project
	err := r.projectService.Delete(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on project delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted project", map[string]interface{}{"project_id": state.Id.Value})

	// Remove project from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery project resource using its id
func (r projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
