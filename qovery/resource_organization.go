package qovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var (
	_ resource.ResourceWithConfigure    = &organizationResource{}
	_ resource.ResourceWithImportState  = organizationResource{}
	_ resource.ResourceWithUpgradeState = organizationResource{}
)

var organizationPlans = clientEnumToStringArray(organization.AllowedPlanValues)

type organizationResource struct {
	organizationService organization.Service
}

func newOrganizationResource() resource.Resource {
	return &organizationResource{}
}

func (r organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.organizationService = provider.organizationService
}

func (r organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		MarkdownDescription: "Manages a Qovery organization.\n\n" +
			"~> **Note:** Terraform cannot create or delete an organization. Import an existing one, and stop managing it with a `removed` block instead of destroying it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("organization"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("organization"),
				Required:            true,
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: descriptions.NewStringEnumDescription(
					organizationPlanDescription,
					organizationPlans,
					nil,
				),
				Required: true,
				Validators: []validator.String{
					validators.NewStringEnumValidator(organizationPlans),
				},
			},
			"description": schema.StringAttribute{
				// Optional only: q-core stores no description when the request omits it, and an
				// update replaces every field, so omitting it clears the description.
				MarkdownDescription: descriptionDescription("organization"),
				Optional:            true,
			},
		},
	}
}

// Create qovery organization resource

func (r organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Error on organization create", "Organization creation is not allowed using terraform.")
}

// Read qovery organization resource
func (r organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Organization
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization from API
	orga, err := r.organizationService.Get(ctx, state.Id.ValueString())
	if handleDomainReadNotFound(ctx, resp, err, "Error on organization read") {
		return
	}

	// Refresh state values
	state = organizationStateFromAPI(orga, state)
	tflog.Trace(ctx, "read organization", map[string]any{"organization_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update qovery organization resource
func (r organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state Organization
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update organization in backend
	orga, err := r.organizationService.Update(ctx, state.Id.ValueString(), plan.toOrganizationUpdateRequest())
	if err != nil {
		resp.Diagnostics.AddError("Error on organization update", err.Error())
		return
	}

	// Update state values
	state = organizationStateFromAPI(orga, plan)
	tflog.Trace(ctx, "updated organization", map[string]any{"organization_id": state.Id.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete qovery organization resource
func (r organizationResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Error on organization delete", "Organization deletion is not allowed using terraform.")
}

// ImportState imports a qovery organization resource using its id
func (r organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// UpgradeState migrates organization states written by 0.x (schema version 0).
func (r organizationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Version 0 has the same attribute types as the current schema.
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	priorSchema := schemaResp.Schema

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &priorSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var orga Organization
				resp.Diagnostics.Append(req.State.Get(ctx, &orga)...)
				if resp.Diagnostics.HasError() {
					return
				}
				orga.Description = upgradeOptionalDescriptionFrom0x(orga.Description)
				resp.Diagnostics.Append(resp.State.Set(ctx, orga)...)
			},
		},
	}
}
