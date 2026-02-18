package qovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &gitTokenResource{}
var _ resource.ResourceWithImportState = gitTokenResource{}

var gitTokenTypes = clientEnumToStringArray(gittoken.AllowedGitTokenTypeValues)

type gitTokenResource struct {
	service gittoken.Service
}

func newGitTokenResource() resource.Resource {
	return &gitTokenResource{}
}

func (r gitTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_token"
}

func (r *gitTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.service = provider.gitTokenService
}

func (r gitTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery git token resource. This can be used to create and manage Qovery git token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the git token.",
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
				Description: "Name of the git token.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the git token.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: descriptions.NewStringEnumDescription(
					"Type of the git token.",
					gitTokenTypes,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(gitTokenTypes),
				},
			},
			"bitbucket_workspace": schema.StringAttribute{
				Description: "(Mandatory only for Bitbucket git token) Workspace where the token has permissions .",
				Optional:    true,
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "Value of the git token.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

// Create qovery git token resource
func (r gitTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan GitToken
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new git token
	response, err := r.service.Create(ctx, plan.OrganizationId.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on git token create", err.Error())
		return
	}

	// Initialize state values
	state := toTerraformObject(plan.OrganizationId.ValueString(), plan.Token.ValueString(), *response)
	tflog.Trace(ctx, "created git token", map[string]any{"git_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery git token resource
func (r gitTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state GitToken
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get git token from the API
	response, err := r.service.Get(ctx, state.OrganizationId.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on git token read", err.Error())
		return
	}

	// Refresh state values
	state = toTerraformObject(state.OrganizationId.ValueString(), state.Token.ValueString(), *response)
	tflog.Trace(ctx, "read git token", map[string]any{"git_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery git token resource
func (r gitTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state GitToken
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update git token in the backend
	response, err := r.service.Update(ctx, state.OrganizationId.ValueString(), state.ID.ValueString(), plan.toUpsertRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on git token update", err.Error())
		return
	}

	// Update state values
	state = toTerraformObject(plan.OrganizationId.ValueString(), plan.Token.ValueString(), *response)
	tflog.Trace(ctx, "updated git token", map[string]any{"git_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery git token resource
func (r gitTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state GitToken
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete git token
	err := r.service.Delete(ctx, state.OrganizationId.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on git token delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted git token", map[string]any{"git_token_id": state.ID.ValueString()})

	// Remove git token from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery git token resource using its id
func (r gitTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,git_token_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
