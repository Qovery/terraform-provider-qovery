package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &deploymentResource{}
var _ resource.ResourceWithImportState = deploymentResource{}

type deploymentResource struct {
	deploymentService newdeployment.Service
}

func newDeploymentResource() resource.Resource {
	return &deploymentResource{}
}

type NewDeploymentTerraform struct {
	EnvironmentId string   `tfsdk:"environment_id"`
	ServiceIds    []string `tfsdk:"service_ids"`
	DesiredState  string   `tfsdk:"desired_state"`
}

func (r deploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *deploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.deploymentService = provider.deploymentService
}

func (r deploymentResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Provides a Qovery deployment stage resource. This can be used to create and manage Qovery deployment stages.",
		Attributes: map[string]tfsdk.Attribute{
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"service_ids": {
				Description: "List of service ids to apply to the deployment.",
				Optional:    true,
				Computed:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					modifiers.NewStringSliceDefaultModifier([]string{}),
				},
			},
			"desired_state": {
				Description: "Desired state of the deployment.",
				Type:        types.StringType,
				Optional:    true,
			},
		},
	}, nil
}

// Create qovery deployment stage resource
func (r deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan NewDeploymentTerraform
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new deployment stage
	_, err := r.deploymentService.Create(ctx, newdeployment.NewDeploymentParams{
		EnvironmentId: plan.EnvironmentId,
		ServiceIds:    plan.ServiceIds,
		DesiredState:  plan.DesiredState,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment create", err.Error())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read qovery deployment tage resource
func (r deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state NewDeploymentTerraform
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.deploymentService.Get(ctx, newdeployment.NewDeploymentParams{
		EnvironmentId: state.EnvironmentId,
		ServiceIds:    state.ServiceIds,
		DesiredState:  state.DesiredState,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment read", err.Error())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state NewDeploymentTerraform
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.deploymentService.Update(ctx, newdeployment.NewDeploymentParams{
		EnvironmentId: plan.EnvironmentId,
		ServiceIds:    plan.ServiceIds,
		DesiredState:  plan.DesiredState,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment update", err.Error())
		return
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state NewDeploymentTerraform
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.deploymentService.Delete(ctx, newdeployment.NewDeploymentParams{
		EnvironmentId: state.EnvironmentId,
		ServiceIds:    state.ServiceIds,
		// When terraform destroys, the desired state will be "DELETED"
		DesiredState: "DELETED",
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment delete", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r deploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// No import for this resource
}
