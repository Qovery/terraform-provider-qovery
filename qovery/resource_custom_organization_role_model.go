package qovery

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/internal/domain/custom_organization_role"
)

type CustomOrganizationRole struct {
	ID                 types.String        `tfsdk:"id"`
	OrganizationID     types.String        `tfsdk:"organization_id"`
	Name               types.String        `tfsdk:"name"`
	Description        types.String        `tfsdk:"description"`
	ClusterPermissions []ClusterPermission `tfsdk:"cluster_permissions"`
	ProjectPermissions []ProjectPermission `tfsdk:"project_permissions"`
}

type ClusterPermission struct {
	ClusterID  types.String `tfsdk:"cluster_id"`
	Permission types.String `tfsdk:"permission"` // e.g., "VIEWER", "ENV_CREATOR", "ADMIN"
}

type ProjectPermission struct {
	ProjectID   types.String            `tfsdk:"project_id"`
	IsAdmin     types.Bool              `tfsdk:"is_admin"`
	Permissions []EnvironmentPermission `tfsdk:"permissions"`
}

type EnvironmentPermission struct {
	EnvironmentType types.String `tfsdk:"environment_type"` // e.g., "DEVELOPMENT", "PRODUCTION"
	Permission      types.String `tfsdk:"permission"`       // e.g., "NO_ACCESS", "VIEWER", "DEPLOYER", "MANAGER", "ADMIN"
}

func (role CustomOrganizationRole) toUpsertServiceRequest(state *CustomOrganizationRole) (*custom_organization_role.UpsertServiceRequest, error) {

	return &custom_organization_role.UpsertServiceRequest{
		CustomOrganizationRoleUpsertRequest: custom_organization_role.UpsertRequest{
			Name:        ToString(state.Name),
			Description: ToStringPointer(state.Description),
		},
	}, nil
}

func convertDomainToCustomOrganizationRole(ctx context.Context, state CustomOrganizationRole, customOrganizationRole *custom_organization_role.CustomOrganizationRole) CustomOrganizationRole {
	return CustomOrganizationRole{}
}
