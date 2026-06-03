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
	_ resource.ResourceWithConfigure = &helmDeploymentResource{}
)

type helmDeploymentResource struct {
	client *client.Client
}

type HelmDeployment struct {
	Id            types.String `tfsdk:"id"`
	HelmID        types.String `tfsdk:"helm_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Version       types.String `tfsdk:"version"`
}

func newHelmDeploymentResource() resource.Resource {
	return &helmDeploymentResource{}
}

func (r helmDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm_deployment"
}

func (r *helmDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r helmDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a deployment of a single qovery_helm service and waits for it to reach DEPLOYED. " +
			"Scoped to one service — unlike qovery_deployment which acts on the whole environment. " +
			"Destroying this resource is a no-op: it does not stop or uninstall the targeted helm service.",
		MarkdownDescription: "Triggers a deployment of a single `qovery_helm` service and waits for it to reach `DEPLOYED`. " +
			"Scoped to one service — unlike `qovery_deployment` which acts on the whole environment.\n\n" +
			"~> **Note:** Destroying this resource is a no-op. It does not stop or uninstall the targeted helm service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the deployment resource (UUID, generated).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"helm_id": schema.StringAttribute{
				Description: "Identifier of the qovery_helm service to deploy.",
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
					"Pass uuid() to redeploy on every apply, or a stable value (chart version, commit sha) to redeploy only when it changes.",
				Optional: true,
			},
		},
	}
}

func (r helmDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HelmDeployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.deployAndWait(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error on helm deployment", err.Error())
		return
	}

	plan.Id = plan.HelmID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r helmDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HelmDeployment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r helmDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state HelmDeployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.deployAndWait(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error on helm redeployment", err.Error())
		return
	}

	plan.Id = state.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r helmDeploymentResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r helmDeploymentResource) deployAndWait(ctx context.Context, plan HelmDeployment) error {
	api := r.client.API()
	serviceID := ToString(plan.HelmID)
	environmentID := ToString(plan.EnvironmentID)

	deployReq := qovery.NewHelmDeployRequest()
	if _, resp, err := api.HelmActionsAPI.
		DeployHelm(ctx, serviceID).
		HelmDeployRequest(*deployReq).
		Execute(); err != nil || (resp != nil && resp.StatusCode >= 400) {
		return fmt.Errorf("failed to trigger deploy for helm service %s: %w", serviceID, err)
	}

	return waitForServiceDeployed(ctx, api, environmentID, serviceID, serviceKindHelm)
}
