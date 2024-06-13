package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/annotations_group"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &annotationsGroupResource{}
var _ resource.ResourceWithImportState = annotationsGroupResource{}

type annotationsGroupResource struct {
	annotationsGroupService annotations_group.Service
}

func newAnnotationsGroupResource() resource.Resource {
	return &annotationsGroupResource{}
}

func (r annotationsGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotations_group"
}

func (r *annotationsGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.annotationsGroupService = provider.annotationsGroupService
}

func (r annotationsGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery annotations group resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the annotations group",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "name of the annotations group",
				Required:    true,
			},
			"annotations": schema.MapAttribute{
				Description: "annotations",
				Required:    true,
				ElementType: types.StringType,
			},
			"scopes": schema.SetAttribute{
				Description: "scopes of the annotations group",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create qovery annotations group resource
func (r annotationsGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AnnotationsGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group create", err.Error())
		return
	}
	newAnnotationsGroup, err := r.annotationsGroupService.Create(ctx, plan.OrganizationId.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group create", err.Error())
		return
	}

	// Initialize state values
	state := convertResponseToAnnotationsGroup(ctx, plan, newAnnotationsGroup)
	tflog.Trace(ctx, "created annotations group", map[string]interface{}{"annotations group": state.Name.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery annotations group resource
func (r annotationsGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state AnnotationsGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get from the API
	annotationsGroup, apiErr := r.annotationsGroupService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError("Error on annotations group read", apiErr.Error())
		return
	}

	// Refresh state values
	state = convertResponseToAnnotationsGroup(ctx, state, annotationsGroup)
	tflog.Trace(ctx, "read get", map[string]interface{}{"annotations_group": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r annotationsGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state AnnotationsGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group create", err.Error())
		return
	}

	annotationsGroup, err := r.annotationsGroupService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group update", err.Error())
		return
	}

	// Update state values
	state = convertResponseToAnnotationsGroup(ctx, plan, annotationsGroup)
	tflog.Trace(ctx, "updated annotations group", map[string]interface{}{"annotation group id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r annotationsGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AnnotationsGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.annotationsGroupService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on annotations group delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted annotations group", map[string]interface{}{"annotations_group_id": state.Id.ValueString()})

	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery application resource using its id
func (r annotationsGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
