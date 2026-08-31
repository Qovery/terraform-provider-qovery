package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &organizationMemberResource{}
	_ resource.ResourceWithImportState = organizationMemberResource{}
)

type organizationMemberResource struct {
	service member.Service
}

func newOrganizationMemberResource() resource.Resource {
	return &organizationMemberResource{}
}

func (r organizationMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *organizationMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.service = provider.organizationMemberService
}

func (r organizationMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery organization member resource. This can be used to invite members to a Qovery organization and manage their role." +
			" Creating the resource sends an invitation; the invitee becomes an active member once they accept it (out-of-band)." +
			" An expired invitation stays in the state with invitation_status EXPIRED: re-send it with terraform apply -replace." +
			" The invitee must accept with the invited email address, otherwise Terraform loses track of the membership.",
		MarkdownDescription: "Provides a Qovery organization member resource. This can be used to invite members to a Qovery organization and manage their role." +
			" Creating the resource sends an invitation; the invitee becomes an active member once they accept it (out-of-band)." +
			" An expired invitation stays in the state with `invitation_status = \"EXPIRED\"`: re-send it with `terraform apply -replace=qovery_organization_member.<name>`." +
			" The invitee must accept with the invited email address, otherwise Terraform loses track of the membership.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the member. While the invitation is pending this is the invitation id; once accepted it becomes the user id. It also changes when the role of a pending invitation is updated (the invitation is re-sent).",
				MarkdownDescription: "Id of the member. While the invitation is pending this is the invitation id; once accepted it becomes the user id. It also changes when the role of a pending invitation is updated (the invitation is re-sent).",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Id of the organization. **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"email": schema.StringAttribute{
				Description:         "Email of the member. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Email of the member. **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"role_id": schema.StringAttribute{
				Description:         "Id of the role to assign to the member (built-in or custom role). Updating the role of a pending invitation re-sends the invitation.",
				MarkdownDescription: "Id of the role to assign to the member (built-in or custom role). Updating the role of a pending invitation re-sends the invitation.",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				Description:         "User id of the member. Null until the invitation is accepted.",
				MarkdownDescription: "User id of the member. Null until the invitation is accepted.",
				Computed:            true,
			},
			"invitation_status": schema.StringAttribute{
				Description:         "Status of the invitation: PENDING, EXPIRED or ACCEPTED.",
				MarkdownDescription: "Status of the invitation: `PENDING`, `EXPIRED` or `ACCEPTED`.",
				Computed:            true,
			},
		},
	}
}

// Create qovery organization member resource
func (r organizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan OrganizationMember
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Invite the member
	domainMember, err := r.service.Create(ctx, plan.OrganizationId.ValueString(), plan.toInviteRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization member create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainMemberToOrganizationMember(*domainMember, plan.Email, plan.RoleId)
	tflog.Trace(ctx, "created organization member", map[string]any{"organization_member_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery organization member resource
func (r organizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state OrganizationMember
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get member from the API
	domainMember, err := r.service.Get(ctx, state.OrganizationId.ValueString(), state.Email.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on organization member read") {
		return
	}

	state = convertDomainMemberToOrganizationMember(*domainMember, state.Email, state.RoleId)
	tflog.Trace(ctx, "read organization member", map[string]any{"organization_member_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery organization member resource
func (r organizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state OrganizationMember
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the member role
	domainMember, err := r.service.Update(ctx, state.OrganizationId.ValueString(), state.Email.ValueString(), plan.toUpdateRoleRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization member update", err.Error())
		return
	}

	state = convertDomainMemberToOrganizationMember(*domainMember, plan.Email, plan.RoleId)
	tflog.Trace(ctx, "updated organization member", map[string]any{"organization_member_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete qovery organization member resource
func (r organizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state OrganizationMember
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete member (cancels the invitation when still pending)
	err := r.service.Delete(ctx, state.OrganizationId.ValueString(), state.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization member delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted organization member", map[string]any{"organization_member_id": state.ID.ValueString()})

	// Remove member from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery organization member using its organization id and email.
func (r organizationMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,email. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), idParts[1])...)
}
