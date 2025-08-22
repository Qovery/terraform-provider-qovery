package qovery

import (
	"context"
	"fmt"
	"github.com/qovery/terraform-provider-qovery/internal/domain/custom_organization_role"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy terraform framework interfaces.
var _ resource.ResourceWithConfigure = &customOrganizationRoleResource{}
var _ resource.ResourceWithImportState = customOrganizationRoleResource{}

type customOrganizationRoleResource struct {
	customOrganizationRoleService custom_organization_role.Service
}

func newCustomOrganizationRoleResource() resource.Resource {
	return &customOrganizationRoleResource{}
}

func (r customOrganizationRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_organization_role"
}

func (r *customOrganizationRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.customOrganizationRoleService = provider.customOrganizationRoleService
}

func (r customOrganizationRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a custom organization role resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the custom role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the custom role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the custom role.",
				Optional:    true,
			},
			"cluster_permissions": schema.ListNestedAttribute{
				Description: "List of cluster permissions.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{
							Description: "ID of the cluster.",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission level for the cluster (VIEWER, ENV_CREATOR, ADMIN).",
							Required:    true,
						},
					},
				},
			},
			"project_permissions": schema.ListNestedAttribute{
				Description: "List of project permissions.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Description: "ID of the project.",
							Required:    true,
						},
						"is_admin": schema.BoolAttribute{
							Description: "Indicates if the role has admin access.",
							Optional:    true,
						},
						"permissions": schema.ListNestedAttribute{
							Description: "List of environment-specific permissions.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"environment_type": schema.StringAttribute{
										Description: "Type of environment (DEVELOPMENT, PREVIEW, STAGING, PRODUCTION).",
										Required:    true,
									},
									"permission": schema.StringAttribute{
										Description: "Permission level (NO_ACCESS, VIEWER, DEPLOYER, MANAGER, ADMIN).",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create custom organization role
func (r customOrganizationRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan CustomOrganizationRole
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new custom organization role
	request, err := plan.toUpsertServiceRequest(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Error on CustomOrganizationRole create", err.Error())
		return
	}
	newCustomOrganizationRole, err := r.customOrganizationRoleService.Create(ctx, plan.OrganizationID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on CustomOrganizationRole create", err.Error())
		return
	}

	// Initialize state values
	state := convertDomainToCustomOrganizationRole(ctx, plan, newCustomOrganizationRole)
	tflog.Trace(ctx, "created custom organization role", map[string]interface{}{"custom_organization_role_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read custom organization role
func (r customOrganizationRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state CustomOrganizationRole
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get custom organization role from the API
	newCustomOrganizationRole, err := r.customOrganizationRoleService.Get(ctx, state.OrganizationID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on custom organization role read", err.Error())
		return
	}

	// Refresh state values
	state = convertDomainToCustomOrganizationRole(ctx, state, newCustomOrganizationRole)
	tflog.Trace(ctx, "read custom organization role", map[string]interface{}{"custom_organization_role_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update custom organization role
func (r customOrganizationRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state CustomOrganizationRole
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update custom organization role in the backend
	request, err := plan.toUpsertServiceRequest(&state)
	if err != nil {
		resp.Diagnostics.AddError("Error on custom organization role create", err.Error())
		return
	}
	newCustomOrganizationRole, err := r.customOrganizationRoleService.Update(ctx, state.OrganizationID.ValueString(), state.ID.ValueString(), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on custom organization role update", err.Error())
		return
	}

	// Update state values
	state = convertDomainToCustomOrganizationRole(ctx, plan, newCustomOrganizationRole)
	tflog.Trace(ctx, "updated custom organization role", map[string]interface{}{"custom_organization_role_id": state.ID.ValueString()})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete custom organization role
func (r customOrganizationRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state CustomOrganizationRole
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete custom organization role
	err := r.customOrganizationRoleService.Delete(ctx, state.OrganizationID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error on custom organization role delete", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted custom organization role", map[string]interface{}{"custom_organization_role_id": state.ID.ValueString()})

	// Remove custom organization role from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports a custom organization role resource using its ID
func (r customOrganizationRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
