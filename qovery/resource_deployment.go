package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &deploymentResource{}
	_ resource.ResourceWithImportState = deploymentResource{}
)

// deploymentResource is an exception to the config-is-source-of-truth rule (AGENTS.md): Qovery
// stores no deployment object, since a deployment is an action on an environment. The resource
// sends that action on create and update and its refresh keeps the state, so a deploy or a stop
// made from the Console does not show up in the plan.
type deploymentResource struct {
	deploymentService newdeployment.Service
}

// default deployment states
var deploymentStates = []string{
	newdeployment.RUNNING.String(),
	newdeployment.STOPPED.String(),
	newdeployment.RESTARTED.String(),
}

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

func (r deploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the deployment of a Qovery environment: Terraform deploys, stops or redeploys all its services. " +
			"Qovery stores no deployment, so a deploy or a stop made outside Terraform does not show up in the plan.\n\n" +
			"~> **Note:** Destroying this resource deletes the environment and all its services.",
		Attributes: map[string]schema.Attribute{
			// id is a value only the provider sets: with no deployment object in Qovery, the
			// provider generates the UUID when the configuration omits it, and UseStateForUnknown
			// keeps it for the life of the resource.
			"id": schema.StringAttribute{
				MarkdownDescription: deploymentIDDescription + " A random UUID is generated when omitted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// environment_id deliberately does not force replacement: its Delete deletes the target environment.
			"environment_id": schema.StringAttribute{
				MarkdownDescription: deploymentEnvironmentIDDescription,
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: deploymentVersionDescription + " Changing it runs the `desired_state` action again: set it to `uuid()` to redeploy on every apply.",
				Optional:            true,
				Computed:            false,
			},
			"desired_state": schema.StringAttribute{
				MarkdownDescription: deploymentDesiredStateDescription + " `RUNNING` deploys all its services, `STOPPED` stops them, and `RESTARTED` redeploys them but fails at creation.",
				Required:            true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(deploymentStates),
				},
			},
		},
	}
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

	// Get calls no API: it rebuilds the deployment from the state (see deploymentResource).
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
