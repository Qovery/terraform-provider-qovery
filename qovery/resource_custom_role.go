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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
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
		Description: "Provides a Qovery organization custom role resource. Declare only the clusters and projects this role should have non-default access to: " +
			"any cluster not listed keeps the VIEWER permission and any project not listed keeps NO_ACCESS. " +
			"Permissions granted outside Terraform on undeclared clusters/projects are reset to those defaults on the next apply.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the custom role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Id of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					RequiresReplaceIfKnownChange(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the custom role. `owner`, `admin`, `devops`, `billing` and `viewer` are reserved built-in role names (case-insensitive).",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the custom role.",
				Optional:    true,
			},
			"cluster_permissions": schema.SetNestedAttribute{
				Description: "Cluster permissions of the custom role. Clusters not listed default to VIEWER.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{
							Description: "Id of the cluster.",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission of the role on the cluster. Can be: `VIEWER`, `ENV_CREATOR`, `ADMIN`.",
							Required:    true,
							Validators: []validator.String{
								validators.NewStringEnumValidator(clusterPermissionValues()),
							},
						},
					},
				},
			},
			"project_permissions": schema.SetNestedAttribute{
				Description: "Project permissions of the custom role. Projects not listed default to NO_ACCESS.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Description: "Id of the project.",
							Required:    true,
						},
						"is_admin": schema.BoolAttribute{
							Description: "Give full admin rights on the project (MANAGER on every environment type + manage deployment rules + delete project). Mutually exclusive with `permissions`.",
							Optional:    true,
						},
						"permissions": schema.SetNestedAttribute{
							Description: "Per-environment-type permissions. Required when `is_admin` is not true; must contain exactly one entry for each environment type (DEVELOPMENT, PREVIEW, STAGING, PRODUCTION).",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"environment_type": schema.StringAttribute{
										Description: "Environment type. Can be: `DEVELOPMENT`, `PREVIEW`, `STAGING`, `PRODUCTION`.",
										Required:    true,
										Validators: []validator.String{
											validators.NewStringEnumValidator(environmentTypeValues()),
										},
									},
									"permission": schema.StringAttribute{
										Description: "Permission of the role on the project for this environment type. Can be: `NO_ACCESS`, `VIEWER`, `DEPLOYER`, `MANAGER`.",
										Required:    true,
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

	state := convertDomainCustomRoleToCustomRole(role, &plan, customRoleReadModeFilterDeclared)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r customRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomRole
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.service.Get(ctx, ToString(state.OrganizationId), ToString(state.Id))
	if err != nil {
		resp.Diagnostics.AddError("Error on custom role read", err.Error())
		return
	}

	// After `terraform import` only id + organization_id are set (name is null): keep
	// non-default entries so the user sees what the role actually grants.
	mode := customRoleReadModeFilterDeclared
	declared := &state
	if state.Name.IsNull() {
		mode = customRoleReadModeKeepNonDefault
		declared = nil
	}

	newState := convertDomainCustomRoleToCustomRole(role, declared, mode)
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

	state := convertDomainCustomRoleToCustomRole(role, &plan, customRoleReadModeFilterDeclared)
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
