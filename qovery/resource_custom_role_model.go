package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

type CustomRole struct {
	Id                 types.String `tfsdk:"id"`
	OrganizationId     types.String `tfsdk:"organization_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ClusterPermissions types.Set    `tfsdk:"cluster_permissions"`
	ProjectPermissions types.Set    `tfsdk:"project_permissions"`
}

var customRoleClusterPermissionAttrTypes = map[string]attr.Type{
	"cluster_id": types.StringType,
	"permission": types.StringType,
}

var customRoleEnvPermissionAttrTypes = map[string]attr.Type{
	"environment_type": types.StringType,
	"permission":       types.StringType,
}

var customRoleProjectPermissionAttrTypes = map[string]attr.Type{
	"project_id":  types.StringType,
	"is_admin":    types.BoolType,
	"permissions": types.SetType{ElemType: types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes}},
}

type customRoleReadMode int

const (
	// keep only entries declared in prior state/plan (normal Read/Create/Update)
	customRoleReadModeFilterDeclared customRoleReadMode = iota
	// keep only entries that differ from server defaults (import)
	customRoleReadModeKeepNonDefault
	// keep the full matrix (data source)
	customRoleReadModeKeepAll
)

// toUpsertRequest converts the Terraform plan/config into the domain request (declared entries only).
func (r CustomRole) toUpsertRequest() *customrole.UpsertRequest {
	clusterPermissions := make([]customrole.ClusterRolePermission, 0, len(r.ClusterPermissions.Elements()))
	if !r.ClusterPermissions.IsNull() && !r.ClusterPermissions.IsUnknown() {
		for _, elem := range r.ClusterPermissions.Elements() {
			obj := elem.(types.Object)
			attrs := obj.Attributes()
			clusterPermissions = append(clusterPermissions, customrole.ClusterRolePermission{
				ClusterID:  ToString(attrs["cluster_id"].(types.String)),
				Permission: customrole.ClusterPermission(ToString(attrs["permission"].(types.String))),
			})
		}
	}

	projectPermissions := make([]customrole.ProjectRolePermission, 0, len(r.ProjectPermissions.Elements()))
	if !r.ProjectPermissions.IsNull() && !r.ProjectPermissions.IsUnknown() {
		for _, elem := range r.ProjectPermissions.Elements() {
			obj := elem.(types.Object)
			attrs := obj.Attributes()
			project := customrole.ProjectRolePermission{
				ProjectID: ToString(attrs["project_id"].(types.String)),
			}
			isAdmin := attrs["is_admin"].(types.Bool)
			if !isAdmin.IsNull() && !isAdmin.IsUnknown() {
				project.IsAdmin = isAdmin.ValueBool()
			}
			permissionsSet := attrs["permissions"].(types.Set)
			if !permissionsSet.IsNull() && !permissionsSet.IsUnknown() {
				for _, permElem := range permissionsSet.Elements() {
					permAttrs := permElem.(types.Object).Attributes()
					project.Permissions = append(project.Permissions, customrole.EnvironmentPermission{
						EnvironmentType: customrole.EnvironmentType(ToString(permAttrs["environment_type"].(types.String))),
						Permission:      customrole.ProjectPermission(ToString(permAttrs["permission"].(types.String))),
					})
				}
			}
			projectPermissions = append(projectPermissions, project)
		}
	}

	return &customrole.UpsertRequest{
		Name:               ToString(r.Name),
		Description:        ToStringPointer(r.Description),
		ClusterPermissions: clusterPermissions,
		ProjectPermissions: projectPermissions,
	}
}

func isDefaultClusterPermission(p customrole.ClusterRolePermission) bool {
	return p.Permission == customrole.ClusterPermissionViewer
}

func isDefaultProjectPermission(p customrole.ProjectRolePermission) bool {
	if p.IsAdmin {
		return false
	}
	for _, ep := range p.Permissions {
		if ep.Permission != customrole.ProjectPermissionNoAccess {
			return false
		}
	}
	return true
}

func declaredIDSet(set types.Set, idAttr string) map[string]bool {
	ids := make(map[string]bool)
	if set.IsNull() || set.IsUnknown() {
		return ids
	}
	for _, elem := range set.Elements() {
		attrs := elem.(types.Object).Attributes()
		ids[ToString(attrs[idAttr].(types.String))] = true
	}
	return ids
}

// convertDomainCustomRoleToCustomRole converts the full server matrix into Terraform state.
// The server returns an entry for EVERY cluster and project of the org; storing that raw would
// produce perpetual diffs, so entries are filtered according to mode.
func convertDomainCustomRoleToCustomRole(role *customrole.CustomRole, declared *CustomRole, mode customRoleReadMode) CustomRole {
	declaredClusters := map[string]bool{}
	declaredProjects := map[string]bool{}
	declaredClustersNull, declaredProjectsNull := true, true
	if declared != nil {
		declaredClusters = declaredIDSet(declared.ClusterPermissions, "cluster_id")
		declaredProjects = declaredIDSet(declared.ProjectPermissions, "project_id")
		declaredClustersNull = declared.ClusterPermissions.IsNull()
		declaredProjectsNull = declared.ProjectPermissions.IsNull()
	}

	clusterObjects := make([]attr.Value, 0, len(role.ClusterPermissions))
	for _, cp := range role.ClusterPermissions {
		switch mode {
		case customRoleReadModeFilterDeclared:
			if !declaredClusters[cp.ClusterID] {
				continue
			}
		case customRoleReadModeKeepNonDefault:
			if isDefaultClusterPermission(cp) {
				continue
			}
		}
		clusterObjects = append(clusterObjects, types.ObjectValueMust(customRoleClusterPermissionAttrTypes, map[string]attr.Value{
			"cluster_id": FromString(cp.ClusterID),
			"permission": FromString(string(cp.Permission)),
		}))
	}

	projectObjects := make([]attr.Value, 0, len(role.ProjectPermissions))
	for _, pp := range role.ProjectPermissions {
		switch mode {
		case customRoleReadModeFilterDeclared:
			if !declaredProjects[pp.ProjectID] {
				continue
			}
		case customRoleReadModeKeepNonDefault:
			if isDefaultProjectPermission(pp) {
				continue
			}
		}
		permissionsValue := types.SetNull(types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes})
		if !pp.IsAdmin {
			permObjects := make([]attr.Value, 0, len(pp.Permissions))
			for _, ep := range pp.Permissions {
				permObjects = append(permObjects, types.ObjectValueMust(customRoleEnvPermissionAttrTypes, map[string]attr.Value{
					"environment_type": FromString(string(ep.EnvironmentType)),
					"permission":       FromString(string(ep.Permission)),
				}))
			}
			permissionsValue = types.SetValueMust(types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes}, permObjects)
		}
		projectObjects = append(projectObjects, types.ObjectValueMust(customRoleProjectPermissionAttrTypes, map[string]attr.Value{
			"project_id":  FromString(pp.ProjectID),
			"is_admin":    types.BoolValue(pp.IsAdmin),
			"permissions": permissionsValue,
		}))
	}

	clusterObjectType := types.ObjectType{AttrTypes: customRoleClusterPermissionAttrTypes}
	projectObjectType := types.ObjectType{AttrTypes: customRoleProjectPermissionAttrTypes}

	// null-vs-empty: an attribute the practitioner never set must stay null, not become an
	// empty set, or Terraform reports "inconsistent result after apply".
	clusterSet := types.SetValueMust(clusterObjectType, clusterObjects)
	if len(clusterObjects) == 0 && (mode != customRoleReadModeKeepAll) && declaredClustersNull {
		clusterSet = types.SetNull(clusterObjectType)
	}
	projectSet := types.SetValueMust(projectObjectType, projectObjects)
	if len(projectObjects) == 0 && (mode != customRoleReadModeKeepAll) && declaredProjectsNull {
		projectSet = types.SetNull(projectObjectType)
	}

	return CustomRole{
		Id:                 FromString(role.ID.String()),
		OrganizationId:     FromString(role.OrganizationID.String()),
		Name:               FromString(role.Name),
		Description:        FromStringPointer(role.Description),
		ClusterPermissions: clusterSet,
		ProjectPermissions: projectSet,
	}
}
