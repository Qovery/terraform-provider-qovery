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
	"github.com/qovery/terraform-provider-qovery/internal/domain/labels_group"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &labelsGroupResource{}
var _ resource.ResourceWithImportState = labelsGroupResource{}

type labelsGroupResource struct {
	labelsGroupService labels_group.Service
}

func newLabelsGroupResource() resource.Resource {
	return &labelsGroupResource{}
}

func (r labelsGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_labels_group"
}

func (r *labelsGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.labelsGroupService = provider.labelsGroupService
}

func (r labelsGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery labels group resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the labels group",
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
				Description: "name of the labels group",
				Required:    true,
			},
			"labels": schema.SetNestedAttribute{
				Description: "labels",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "",
							Required:    true,
						},
						"propagate_to_cloud_provider": schema.BoolAttribute{
							Description: "",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Create qovery labels group resource
func (r labelsGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LabelsGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group create", err.Error())
		return
	}
	newLabelsGroup, err := r.labelsGroupService.Create(ctx, plan.OrganizationId.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group create", err.Error())
		return
	}

	// Initialize state values
	state := convertResponseToLabelsGroup(ctx, plan, newLabelsGroup)
	tflog.Trace(ctx, "created labels group", map[string]any{"labels group": state.Name.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery labels group resource
func (r labelsGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state LabelsGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Get from the API
	labelsGroup, apiErr := r.labelsGroupService.Get(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if apiErr != nil {
		resp.Diagnostics.AddError("Error on labels group read", apiErr.Error())
		return
	}

	//Refresh state values
	state = convertResponseToLabelsGroup(ctx, state, labelsGroup)
	tflog.Trace(ctx, "read get", map[string]any{"labels_group": state.Id.ValueString()})

	//Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r labelsGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state LabelsGroup
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, err := plan.toUpsertRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group create", err.Error())
		return
	}

	labelsGroup, err := r.labelsGroupService.Update(ctx, state.OrganizationId.ValueString(), state.Id.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group update", err.Error())
		return
	}

	// Update state values
	state = convertResponseToLabelsGroup(ctx, plan, labelsGroup)
	tflog.Trace(ctx, "updated labels group", map[string]any{"label group id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r labelsGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state LabelsGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.labelsGroupService.Delete(ctx, state.OrganizationId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on labels group delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted labels group", map[string]any{"labels_group_id": state.Id.ValueString()})

	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery application resource using its id
func (r labelsGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
