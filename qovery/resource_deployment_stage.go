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

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &deploymentStageResource{}
var _ resource.ResourceWithImportState = deploymentStageResource{}

type deploymentStageResource struct {
	deploymentStageService deploymentstage.Service
}

func newDeploymentStageResource() resource.Resource {
	return &deploymentStageResource{}
}

func (r deploymentStageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_stage"
}

func (r *deploymentStageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.deploymentStageService = provider.deploymentStageService
}

func (r deploymentStageResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery deployment stage resource. This can be used to create and manage Qovery deployment stages.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Id of the deployment stage.",
				Type:        types.StringType,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the deployment stage.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: "Description of the deployment stage.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}, nil
}

// Create qovery deployment stage resource
func (r deploymentStageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan DeploymentStage
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new deployment stage
	deploymentStage, err := r.deploymentStageService.Create(ctx, plan.EnvironmentId.Value, plan.toCreateServiceRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainDeploymentStageToDeploymentStage(deploymentStage)
	tflog.Info(ctx, "created deployment stage", map[string]interface{}{"deployment_stage_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery deployment tage resource
func (r deploymentStageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state DeploymentStage
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get deployment stage from the API
	deploymentStage, err := r.deploymentStageService.Get(ctx, state.EnvironmentId.Value, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainDeploymentStageToDeploymentStage(deploymentStage)
	tflog.Trace(ctx, "read deployment stage", map[string]interface{}{"deployment_stage_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery deployment stage resource
func (r deploymentStageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state DeploymentStage
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update deployment stage in the backend
	deploymentStage, err := r.deploymentStageService.Update(ctx, state.Id.Value, plan.toUpdateServiceRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage update", err.Error())
		return
	}

	// Update state values
	state = convertDomainDeploymentStageToDeploymentStage(deploymentStage)
	tflog.Trace(ctx, "updated deployment stage", map[string]interface{}{"deployment_stage_id": state.Id.Value})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery deployment stage resource
func (r deploymentStageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state DeploymentStage
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete deployment stage
	err := r.deploymentStageService.Delete(ctx, state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted deployment stage", map[string]interface{}{"deployment_stage_id": state.Id.Value})

	// Remove deployment stage from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery deployment stage resource using its id
func (r deploymentStageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
