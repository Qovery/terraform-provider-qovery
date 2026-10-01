package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

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
var (
	_ resource.ResourceWithConfigure      = &gitTokenResource{}
	_ resource.ResourceWithImportState    = gitTokenResource{}
	_ resource.ResourceWithUpgradeState   = gitTokenResource{}
	_ resource.ResourceWithValidateConfig = gitTokenResource{}
)

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
		Version:             1,
		MarkdownDescription: "Manages a Qovery git token: a token of a git provider that Qovery uses to access private repositories.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("git token"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("git token"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("git token"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				// Optional only: q-core stores no description when the request omits it, and an
				// update replaces every field, so omitting it clears the description.
				MarkdownDescription: descriptionDescription("git token"),
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(
					gitTokenTypeDescription,
					gitTokenTypes,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(gitTokenTypes),
				},
			},
			"bitbucket_workspace": schema.StringAttribute{
				MarkdownDescription: gitTokenBitbucketWorkspaceDescription + " Required when `type` is `BITBUCKET`: omitting it fails at plan time.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: gitTokenTokenDescription,
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

// ValidateConfig requires bitbucket_workspace for a BITBUCKET token at plan time: the Qovery API
// needs it to list the repositories of the token.
func (r gitTokenResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config GitToken
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateGitTokenWorkspace(config.Type, config.BitbucketWorkspace); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("bitbucket_workspace"), "Invalid git token configuration", err.Error())
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
	state := gitTokenStateFromAPI(*response, plan)
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
	if handleDomainReadNotFound(ctx, resp, err, "Error on git token read") {
		return
	}

	// Refresh state values
	state = gitTokenStateFromAPI(*response, state)
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
	state = gitTokenStateFromAPI(*response, plan)
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

// UpgradeState migrates git token states written by 0.x (schema version 0).
func (r gitTokenResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Version 0 has the same attribute types as the current schema.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	priorSchema := schemaResp.Schema

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var token GitToken
				resp.Diagnostics.Append(req.State.Get(ctx, &token)...)
				if resp.Diagnostics.HasError() {
					return
				}
				token.Description = upgradeOptionalDescriptionFrom0x(token.Description)
				resp.Diagnostics.Append(resp.State.Set(ctx, token)...)
			},
		},
	}
}
