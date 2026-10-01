package qovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
	"github.com/qovery/terraform-provider-qovery/qovery/descriptions"
	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

var (
	_ resource.ResourceWithConfigure      = &customRoleResource{}
	_ resource.ResourceWithImportState    = customRoleResource{}
	_ resource.ResourceWithValidateConfig = customRoleResource{}
)

type customRoleResource struct {
	service customrole.Service
}

func newCustomRoleResource() resource.Resource {
	return &customRoleResource{}
}

func (r customRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_role"
}

func (r *customRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	provider, ok := req.ProviderData.(*qProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *qProvider, got: %T.", req.ProviderData),
		)
		return
	}
	r.service = provider.customRoleService
}

func clusterPermissionValues() []string {
	values := make([]string, 0, len(customrole.AllowedClusterPermissions))
	for _, p := range customrole.AllowedClusterPermissions {
		values = append(values, string(p))
	}
	return values
}

func projectPermissionValues() []string {
	values := make([]string, 0, len(customrole.AllowedProjectPermissions))
	for _, p := range customrole.AllowedProjectPermissions {
		values = append(values, string(p))
	}
	return values
}

func environmentTypeValues() []string {
	values := make([]string, 0, len(customrole.AllowedEnvironmentTypes))
	for _, t := range customrole.AllowedEnvironmentTypes {
		values = append(values, string(t))
	}
	return values
}

func (r customRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Qovery custom role: an organization role with its own permissions on each cluster and project.\n\n" +
			"~> **Note:** Declare only the clusters and projects that need a permission other than the default. An import records only those, so a declared default permission shows up as a change after an import.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription("custom role"),
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: organizationIDDescription + recreatesOnChange("custom role"),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription("custom role") + " Using a built-in role name, `owner`, `admin`, `devops`, `billing` or `viewer` in any case, fails at plan time.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				// q-core stores an omitted description as "", so the Default is "": removing the
				// description from the configuration plans its reset.
				MarkdownDescription: storedDescriptionDescription("custom role"),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(storedDescriptionDefault),
			},
			"cluster_permissions": schema.SetNestedAttribute{
				MarkdownDescription: customRoleClusterPermissionsDescription + " A cluster not listed gets `VIEWER`.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{
							MarkdownDescription: customRoleClusterIDDescription,
							Required:            true,
						},
						"permission": schema.StringAttribute{
							MarkdownDescription: descriptions.NewStringEnumDescription(customRoleClusterPermissionDescription, clusterPermissionValues(), nil),
							Required:            true,
							Validators: []validator.String{
								validators.NewStringEnumValidator(clusterPermissionValues()),
							},
						},
					},
				},
			},
			"project_permissions": schema.SetNestedAttribute{
				MarkdownDescription: customRoleProjectPermissionsDescription + " A project not listed gets `NO_ACCESS` on every environment type.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							MarkdownDescription: customRoleProjectIDDescription,
							Required:            true,
						},
						"is_admin": schema.BoolAttribute{
							MarkdownDescription: descriptions.NewBoolDefaultDescription(customRoleIsAdminDescription+" When `true`, `permissions` must be omitted: even an empty set fails at plan time.", false),
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
						"permissions": schema.SetNestedAttribute{
							MarkdownDescription: customRoleEnvironmentPermissionsDescription + " Required when `is_admin` is `false`, with exactly one entry per environment type.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"environment_type": schema.StringAttribute{
										MarkdownDescription: descriptions.NewStringEnumDescription(customRoleEnvironmentTypeDescription, environmentTypeValues(), nil),
										Required:            true,
										Validators: []validator.String{
											validators.NewStringEnumValidator(environmentTypeValues()),
										},
									},
									"permission": schema.StringAttribute{
										MarkdownDescription: descriptions.NewStringEnumDescription(customRoleEnvironmentPermissionDescription, projectPermissionValues(), nil),
										Required:            true,
										Validators: []validator.String{
											validators.NewStringEnumValidator(projectPermissionValues()),
										},
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

// ValidateConfig surfaces cross-field errors (is_admin XOR permissions, 4-env-type completeness,
// reserved names) at plan time instead of apply time.
func (r customRoleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config CustomRole
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.Name.IsUnknown() || config.ClusterPermissions.IsUnknown() || config.ProjectPermissions.IsUnknown() {
		return
	}

	// An admin project must not carry a `permissions` set at all. The null-vs-empty
	// distinction is lost once toUpsertRequest folds an empty set into a nil slice (so the
	// domain's non-empty guard never fires), and convertDomainCustomRoleToCustomRole always
	// writes null for admin projects — yielding "inconsistent result after apply". Only the
	// raw config seen here can tell an explicit `permissions = []` from an omitted attribute.
	if !config.ProjectPermissions.IsNull() {
		for _, elem := range config.ProjectPermissions.Elements() {
			obj, ok := elem.(types.Object)
			if !ok {
				continue
			}
			attrs := obj.Attributes()
			isAdmin, _ := attrs["is_admin"].(types.Bool)
			permissions, _ := attrs["permissions"].(types.Set)
			if isAdmin.IsNull() || isAdmin.IsUnknown() || !isAdmin.ValueBool() {
				continue
			}
			if !permissions.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("project_permissions"),
					"Invalid custom role configuration",
					"permissions must not be set when is_admin is true, not even as an empty set",
				)
			}
		}
	}

	request := config.toUpsertRequest()
	if err := request.Validate(); err != nil {
		// ids may be unknown (references to not-yet-created resources) at plan time; only
		// fail on definite errors, not uuid-format failures from unknown values.
		if !strings.Contains(err.Error(), customrole.ErrInvalidUpsertRequest.Error()) {
			resp.Diagnostics.AddError("Invalid custom role configuration", err.Error())
		}
	}
}

func (r customRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomRole
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.toUpsertRequest()

	role, err := r.service.Create(ctx, ToString(plan.OrganizationId), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on custom role create", err.Error())
		return
	}

	state := convertDomainCustomRoleToCustomRole(role, &plan, customRoleReadModeDeclaredOrNonDefault)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r customRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomRole
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.service.Get(ctx, ToString(state.OrganizationId), ToString(state.Id))
	if handleDomainReadNotFound(ctx, resp, err, "Error on custom role read") {
		return
	}

	// After `terraform import` nothing is declared yet, so the state records the non-default entries.
	newState := convertDomainCustomRoleToCustomRole(role, &state, customRoleReadModeDeclaredOrNonDefault)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r customRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomRole
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.toUpsertRequest()

	role, err := r.service.Update(ctx, ToString(plan.OrganizationId), ToString(plan.Id), *request)
	if err != nil {
		resp.Diagnostics.AddError("Error on custom role update", err.Error())
		return
	}

	state := convertDomainCustomRoleToCustomRole(role, &plan, customRoleReadModeDeclaredOrNonDefault)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r customRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomRole
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.service.Delete(ctx, ToString(state.OrganizationId), ToString(state.Id)); err != nil {
		resp.Diagnostics.AddError("Error on custom role delete", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r customRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'organization_id,custom_role_id'. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
