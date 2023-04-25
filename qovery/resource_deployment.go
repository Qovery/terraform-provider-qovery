package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &deploymentResource{}
var _ resource.ResourceWithImportState = deploymentResource{}

type deploymentResource struct {
	deploymentService newdeployment.Service
}

var (
	// default deployment states
	deploymentStates = []string{
		newdeployment.DEPLOYED.String(),
		newdeployment.STOPPED.String(),
		newdeployment.RESTARTED.String(),
	}
)

func newDeploymentResource() resource.Resource {
	return &deploymentResource{}
}

type NewDeploymentTerraform struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Version       types.String `tfsdk:"version"`
	DesiredState  types.String `tfsdk:"desired_state"`
}

func newDeploymentTerraformFromDomain(domain *newdeployment.Deployment) NewDeploymentTerraform {
	var version *string = nil
	if domain.Version != nil {
		versionToString := domain.Version.String()
		version = &versionToString
	}
	return NewDeploymentTerraform{
		Id:            FromString(domain.ID.String()),
		EnvironmentId: FromString(domain.EnvironmentID.String()),
		Version:       FromStringPointer(version),
		DesiredState:  FromString(domain.DesiredState.String()),
	}
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
			"id": {
				Description: "Id of the deployment",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"environment_id": {
				Description: "Id of the environment.",
				Type:        types.StringType,
				Required:    true,
			},
			"version": {
				Description: "Version to force trigger a deployment when desired_state doesn't change (e.g redeploy a deployment having the 'DEPLOYED' state)",
				Type:        types.StringType,
				Optional:    true,
				Computed:    false,
			},
			"desired_state": {
				Description: descriptions.NewStringEnumDescription(
					"Desired state of the deployment.",
					deploymentStates,
					nil),
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					validators.NewStringEnumValidator(deploymentStates),
				},
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
	deployment, err := r.deploymentService.Create(ctx, newdeployment.NewDeploymentParams{
		ID:            ToStringPointer(plan.Id),
		EnvironmentID: ToString(plan.EnvironmentId),
		Version:       ToStringPointer(plan.Version),
		DesiredState:  ToString(plan.DesiredState),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment create", err.Error())
		return
	}

	newState := newDeploymentTerraformFromDomain(deployment)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Read qovery deployment tage resource
func (r deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state NewDeploymentTerraform
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.deploymentService.Get(ctx, newdeployment.NewDeploymentParams{
		ID:            ToStringPointer(state.Id),
		EnvironmentID: ToString(state.EnvironmentId),
		Version:       ToStringPointer(state.Version),
		DesiredState:  ToString(state.DesiredState),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment read", err.Error())
		return
	}

	newState := newDeploymentTerraformFromDomain(deployment)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state NewDeploymentTerraform
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.deploymentService.Update(ctx, newdeployment.NewDeploymentParams{
		ID:            ToStringPointer(state.Id),
		EnvironmentID: ToString(plan.EnvironmentId),
		Version:       ToStringPointer(plan.Version),
		DesiredState:  ToString(plan.DesiredState),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error on deployment update", err.Error())
		return
	}
	newState := newDeploymentTerraformFromDomain(deployment)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state NewDeploymentTerraform
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.deploymentService.Delete(ctx, newdeployment.NewDeploymentParams{
		EnvironmentID: ToString(state.EnvironmentId),
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
