package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func (r deploymentStageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery deployment stage resource. This can be used to create and manage Qovery deployment stages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the deployment stage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Id of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment stage.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the deployment stage.",
				Optional:    true,
			},
			"is_after": schema.StringAttribute{
				Description: "Move the current deployment stage after the target deployment stage",
				Optional:    true,
			},
			"is_before": schema.StringAttribute{
				Description: "Move the current deployment stage before the target deployment stage",
				Optional:    true,
			},
		},
	}
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
	deploymentStage, err := r.deploymentStageService.Create(ctx, plan.EnvironmentId.ValueString(), plan.toCreateServiceRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainDeploymentStageToDeploymentStage(deploymentStage, plan.Description)
	tflog.Info(ctx, "created deployment stage", map[string]any{"deployment_stage_id": state.Id.ValueString()})

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
	deploymentStage, err := r.deploymentStageService.Get(ctx, state.EnvironmentId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage read", err.Error())
		return
	}

	// Refresh state values
	newState := convertDomainDeploymentStageToDeploymentStage(deploymentStage, state.Description)
	tflog.Trace(ctx, "read deployment stage", map[string]any{"deployment_stage_id": state.Id.ValueString()})

	// We need to keep the 'IsAfter' and 'IsBefore' properties
	newState = DeploymentStage{
		Id:            newState.Id,
		EnvironmentId: newState.EnvironmentId,
		Name:          newState.Name,
		Description:   newState.Description,
		IsAfter:       state.IsAfter,
		IsBefore:      state.IsBefore,
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
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
	deploymentStage, err := r.deploymentStageService.Update(ctx, state.Id.ValueString(), plan.toUpdateServiceRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage update", err.Error())
		return
	}

	// Update state values
	state = convertDomainDeploymentStageToDeploymentStage(deploymentStage, plan.Description)
	tflog.Trace(ctx, "updated deployment stage", map[string]any{"deployment_stage_id": state.Id.ValueString()})

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
	err := r.deploymentStageService.Delete(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment stage delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted deployment stage", map[string]any{"deployment_stage_id": state.Id.ValueString()})

	// Remove deployment stage from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery deployment stage resource using its id
func (r deploymentStageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: environment_id,deployment_stage_name. Got: %q", req.ID),
		)
		return
	}

	environmentId := idParts[0]
	deploymentStageName := idParts[1]
	deploymentStage, err := r.deploymentStageService.GetAllByEnvironmentID(ctx, environmentId, deploymentStageName)
	if err != nil {
		resp.Diagnostics.AddError("Error", err.Error())
		return
	}

	req.ID = deploymentStage.ID.String()
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), deploymentStage.ID.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environmentId)...)
}
