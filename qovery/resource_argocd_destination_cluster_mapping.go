package qovery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
)

var (
	_ resource.ResourceWithConfigure   = &argoCdDestinationClusterMappingResource{}
	_ resource.ResourceWithImportState = argoCdDestinationClusterMappingResource{}
)

type argoCdDestinationClusterMappingResource struct {
	argoCdDestinationClusterMappingService argoCdDestinationClusterMapping.Service
}

func newArgoCdDestinationClusterMappingResource() resource.Resource {
	return &argoCdDestinationClusterMappingResource{}
}

func (r argoCdDestinationClusterMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_argocd_destination_cluster_mapping"
}

func (r *argoCdDestinationClusterMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.argoCdDestinationClusterMappingService = provider.argoCdDestinationClusterMappingService
}

func (r argoCdDestinationClusterMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery ArgoCD destination cluster mapping resource. This maps an ArgoCD destination cluster URL to a Qovery cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier of the mapping (agent_cluster_id:argocd_cluster_url).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_cluster_id": schema.StringAttribute{
				Description: "Id of the Qovery cluster where the ArgoCD instance is running.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"argocd_cluster_url": schema.StringAttribute{
				Description: "URL of the ArgoCD destination cluster (e.g. https://kubernetes.default.svc).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "Id of the Qovery cluster mapped to the ArgoCD destination.",
				Required:    true,
			},
		},
	}
}

func (r argoCdDestinationClusterMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ArgoCdDestinationClusterMapping
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := r.argoCdDestinationClusterMappingService.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd destination cluster mapping create", err.Error())
		return
	}

	state := convertDomainArgoCdDestinationClusterMappingToTF(mapping)
	tflog.Trace(ctx, "created argocd destination cluster mapping", map[string]any{"agent_cluster_id": state.AgentClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r argoCdDestinationClusterMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ArgoCdDestinationClusterMapping
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := r.argoCdDestinationClusterMappingService.Get(
		ctx,
		state.OrganizationId.ValueString(),
		state.AgentClusterId.ValueString(),
		state.ArgocdClusterUrl.ValueString(),
	)
	if err != nil {
		// The list API only surfaces clusters ArgoCD has actively discovered; a
		// mapping may not appear until ArgoCD has polled the destination. Preserve
		// existing state rather than surfacing a spurious error.
		if errors.Is(err, argoCdDestinationClusterMapping.ErrNotFoundInList) {
			tflog.Warn(ctx, "argocd destination cluster mapping not yet visible in live cluster list — preserving state",
				map[string]any{"agent_cluster_id": state.AgentClusterId.ValueString()})
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
		resp.Diagnostics.AddError("Error on argocd destination cluster mapping read", err.Error())
		return
	}

	state = convertDomainArgoCdDestinationClusterMappingToTF(mapping)
	tflog.Trace(ctx, "read argocd destination cluster mapping", map[string]any{"agent_cluster_id": state.AgentClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r argoCdDestinationClusterMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ArgoCdDestinationClusterMapping
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := r.argoCdDestinationClusterMappingService.Update(ctx, state.OrganizationId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd destination cluster mapping update", err.Error())
		return
	}

	state = convertDomainArgoCdDestinationClusterMappingToTF(mapping)
	tflog.Trace(ctx, "updated argocd destination cluster mapping", map[string]any{"agent_cluster_id": state.AgentClusterId.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r argoCdDestinationClusterMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ArgoCdDestinationClusterMapping
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.argoCdDestinationClusterMappingService.Delete(
		ctx,
		state.OrganizationId.ValueString(),
		state.AgentClusterId.ValueString(),
		state.ArgocdClusterUrl.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error on argocd destination cluster mapping delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted argocd destination cluster mapping", map[string]any{"agent_cluster_id": state.AgentClusterId.ValueString()})

	resp.State.RemoveResource(ctx)
}

func (r argoCdDestinationClusterMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,agent_cluster_id,argocd_cluster_url. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent_cluster_id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("argocd_cluster_url"), idParts[2])...)
}
