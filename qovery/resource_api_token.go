package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apitoken"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &apiTokenResource{}
	_ resource.ResourceWithImportState = apiTokenResource{}
)

type apiTokenResource struct {
	service apitoken.Service
}

func newApiTokenResource() resource.Resource {
	return &apiTokenResource{}
}

func (r apiTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (r *apiTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.service = provider.apiTokenService
}

func (r apiTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Qovery API token resource. This can be used to create and manage Qovery organization API tokens." +
			" The token value is only returned at creation time and is stored in the Terraform state: use an encrypted remote state with restricted access." +
			" The API does not support updating a token, so every attribute change forces a replacement (rotation).",
		MarkdownDescription: "Provides a Qovery API token resource. This can be used to create and manage Qovery organization API tokens." +
			" The token value is only returned at creation time and is **stored in the Terraform state**: use an encrypted remote state with restricted access." +
			" The API does not support updating a token, so every attribute change forces a replacement (rotation). Rotate a token explicitly with `terraform apply -replace=qovery_api_token.<name>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "Id of the API token.",
				MarkdownDescription: "Id of the API token.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description:         "Id of the organization. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Id of the organization. **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "Name of the API token. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Name of the API token. **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"description": schema.StringAttribute{
				Description:         "Description of the API token. Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Description of the API token. **Cannot be changed after creation** (forces resource replacement).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"role_id": schema.StringAttribute{
				Description:         "Id of the role to associate with the API token (built-in or custom role). Cannot be changed after creation (forces resource replacement).",
				MarkdownDescription: "Id of the role to associate with the API token (built-in or custom role). **Cannot be changed after creation** (forces resource replacement).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"token": schema.StringAttribute{
				Description:         "Value of the API token. Only returned at creation time and stored in the Terraform state; it cannot be retrieved afterwards.",
				MarkdownDescription: "Value of the API token. Only returned at creation time and stored in the Terraform state; it cannot be retrieved afterwards.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create qovery api token resource
func (r apiTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan ApiToken
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new api token
	apiToken, err := r.service.Create(ctx, plan.OrganizationId.ValueString(), plan.toCreateRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on api token create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainApiTokenToApiToken(*apiToken, plan.Token)
	tflog.Trace(ctx, "created api token", map[string]any{"api_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read qovery api token resource
func (r apiTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ApiToken
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get api token from the API
	apiToken, err := r.service.Get(ctx, state.OrganizationId.ValueString(), state.ID.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on api token read") {
		return
	}

	// Refresh state values, preserving the token value from the state since the API never returns it again
	state = convertDomainApiTokenToApiToken(*apiToken, state.Token)
	tflog.Trace(ctx, "read api token", map[string]any{"api_token_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery api token resource
// The API exposes no update endpoint: every attribute is marked RequiresReplace, so this method is unreachable.
func (r apiTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Qovery API tokens cannot be updated: every attribute change forces a replacement. Please report this issue to the provider developers.",
	)
}

// Delete qovery api token resource
func (r apiTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ApiToken
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete api token
	err := r.service.Delete(ctx, state.OrganizationId.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on api token delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted api token", map[string]any{"api_token_id": state.ID.ValueString()})

	// Remove api token from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a qovery api token resource using its organization id and token id.
// The token value cannot be retrieved from the API, so it stays null in the state after an import.
func (r apiTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: organization_id,api_token_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
}
