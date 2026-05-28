package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
)

var (
	_ resource.ResourceWithConfigure   = &argoCdCredentialsResource{}
	_ resource.ResourceWithImportState = argoCdCredentialsResource{}
)

type argoCdCredentialsResource struct {
	argoCdCredentialsService argoCdCredentials.Service
}

func newArgoCdCredentialsResource() resource.Resource {
	return &argoCdCredentialsResource{}
}

func (r argoCdCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_argocd_credentials"
}

func (r *argoCdCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.argoCdCredentialsService = provider.argoCdCredentialsService
}

func (r argoCdCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery ArgoCD credentials resource. This can be used to configure ArgoCD integration for a cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the ArgoCD credentials.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "Id of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"argocd_url": schema.StringAttribute{
				Description: "URL of the ArgoCD instance (e.g. https://argocd.example.com).",
				Required:    true,
			},
			"argocd_token": schema.StringAttribute{
				Description: "ArgoCD API authentication token.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r argoCdCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ArgoCdCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.argoCdCredentialsService.Create(ctx, plan.ClusterId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd credentials create", err.Error())
		return
	}

	state := convertDomainArgoCdCredentialsToTF(plan, creds)
	tflog.Trace(ctx, "created argocd credentials", map[string]any{"cluster_id": state.ClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r argoCdCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ArgoCdCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.argoCdCredentialsService.Get(ctx, state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd credentials read", err.Error())
		return
	}

	state = convertDomainArgoCdCredentialsToTF(state, creds)
	tflog.Trace(ctx, "read argocd credentials", map[string]any{"cluster_id": state.ClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r argoCdCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ArgoCdCredentials
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.argoCdCredentialsService.Update(ctx, state.ClusterId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd credentials update", err.Error())
		return
	}

	state = convertDomainArgoCdCredentialsToTF(plan, creds)
	tflog.Trace(ctx, "updated argocd credentials", map[string]any{"cluster_id": state.ClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r argoCdCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ArgoCdCredentials
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.argoCdCredentialsService.Delete(ctx, state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd credentials delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted argocd credentials", map[string]any{"cluster_id": state.ClusterId.ValueString()})

	resp.State.RemoveResource(ctx)
}

func (r argoCdCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), req.ID)...)
}
