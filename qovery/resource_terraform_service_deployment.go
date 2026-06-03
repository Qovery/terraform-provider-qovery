package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client"
)

var (
	_ resource.ResourceWithConfigure = &terraformServiceDeploymentResource{}
)

type terraformServiceDeploymentResource struct {
	client *client.Client
}

type TerraformServiceDeployment struct {
	Id                  types.String `tfsdk:"id"`
	TerraformServiceID  types.String `tfsdk:"terraform_service_id"`
	EnvironmentID       types.String `tfsdk:"environment_id"`
	Version             types.String `tfsdk:"version"`
}

func newTerraformServiceDeploymentResource() resource.Resource {
	return &terraformServiceDeploymentResource{}
}

func (r terraformServiceDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_service_deployment"
}

func (r *terraformServiceDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = provider.client
}

func (r terraformServiceDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a deployment of a single qovery_terraform_service and waits for it to reach DEPLOYED. " +
			"Scoped to one service — unlike qovery_deployment which acts on the whole environment. " +
			"Destroying this resource is a no-op: it does not stop or uninstall the targeted terraform service.",
		MarkdownDescription: "Triggers a deployment of a single `qovery_terraform_service` and waits for it to reach `DEPLOYED`. " +
			"Scoped to one service — unlike `qovery_deployment` which acts on the whole environment.\n\n" +
			"~> **Note:** Destroying this resource is a no-op. It does not stop or uninstall the targeted terraform service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the deployment resource (UUID, generated).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"terraform_service_id": schema.StringAttribute{
				Description: "Identifier of the qovery_terraform_service to deploy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "Identifier of the environment that contains the service (used to poll deployment status).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "Opaque token to force a redeployment when nothing else has changed. " +
					"Pass uuid() to redeploy on every apply, or a stable value (commit sha, tag) to redeploy only when it changes.",
				Optional: true,
			},
		},
	}
}

func (r terraformServiceDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TerraformServiceDeployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.deployAndWait(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error on terraform service deployment", err.Error())
		return
	}

	plan.Id = plan.TerraformServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r terraformServiceDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Status is recomputed on every apply via the version attribute; no remote read needed.
	var state TerraformServiceDeployment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r terraformServiceDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state TerraformServiceDeployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.deployAndWait(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error on terraform service redeployment", err.Error())
		return
	}

	plan.Id = state.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r terraformServiceDeploymentResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: deployment lifecycle is decoupled from the qovery_terraform_service it targets.
	resp.State.RemoveResource(ctx)
}

func (r terraformServiceDeploymentResource) deployAndWait(ctx context.Context, plan TerraformServiceDeployment) error {
	api := r.client.API()
	serviceID := ToString(plan.TerraformServiceID)
	environmentID := ToString(plan.EnvironmentID)

	deployReq := qovery.NewTerraformDeployRequest()
	if _, resp, err := api.TerraformActionsAPI.
		DeployTerraform(ctx, serviceID).
		TerraformDeployRequest(*deployReq).
		Execute(); err != nil || (resp != nil && resp.StatusCode >= 400) {
		return fmt.Errorf("failed to trigger deploy for terraform service %s: %w", serviceID, err)
	}

	return waitForServiceDeployed(ctx, api, environmentID, serviceID, serviceKindTerraform)
}
