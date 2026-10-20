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
	"shell":            types.BoolType,
	"port_forward":     types.BoolType,
}

var customRoleProjectPermissionAttrTypes = map[string]attr.Type{
	"project_id":  types.StringType,
	"is_admin":    types.BoolType,
	"permissions": types.SetType{ElemType: types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes}},
}

type customRoleReadMode int

const (
	// keep the entries declared in the prior plan or state, plus every entry that differs from
	// the server defaults (resource Create, Read, Update and import)
	customRoleReadModeDeclaredOrNonDefault customRoleReadMode = iota
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
						Shell:           knownBoolPointer(permAttrs["shell"].(types.Bool)),
						PortForward:     knownBoolPointer(permAttrs["port_forward"].(types.Bool)),
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

// knownBoolPointer returns nil for a null or unknown value: the flag is then left to q-core.
func knownBoolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	b := value.ValueBool()
	return &b
}

type customRoleEnvKey struct {
	projectID       string
	environmentType string
}

// declaredRemoteAccess records, per project and environment type, which of shell and port_forward
// the prior plan or state sets.
type declaredRemoteAccess struct {
	shell       bool
	portForward bool
}

func declaredRemoteAccessByEnv(set types.Set) map[customRoleEnvKey]declaredRemoteAccess {
	declared := make(map[customRoleEnvKey]declaredRemoteAccess)
	if set.IsNull() || set.IsUnknown() {
		return declared
	}
	for _, elem := range set.Elements() {
		attrs := elem.(types.Object).Attributes()
		projectID := ToString(attrs["project_id"].(types.String))
		permissions, _ := attrs["permissions"].(types.Set)
		if permissions.IsNull() || permissions.IsUnknown() {
			continue
		}
		for _, permElem := range permissions.Elements() {
			permAttrs := permElem.(types.Object).Attributes()
			declared[customRoleEnvKey{projectID, ToString(permAttrs["environment_type"].(types.String))}] = declaredRemoteAccess{
				shell:       !permAttrs["shell"].(types.Bool).IsNull(),
				portForward: !permAttrs["port_forward"].(types.Bool).IsNull(),
			}
		}
	}
	return declared
}

// remoteAccessStateValue decides what the state records for a shell or port_forward flag: the server
// value when the flag is declared (or for the data source), null when it is not, so q-core keeps
// managing it. With nothing declared on the environment type (an undeclared project, or an import),
// the server value is recorded only when it differs from the q-core default for the permission.
func remoteAccessStateValue(value *bool, permission customrole.ProjectPermission, declaredFlag bool, envDeclared bool, mode customRoleReadMode) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	switch {
	case mode == customRoleReadModeKeepAll, declaredFlag:
		return types.BoolValue(*value)
	case envDeclared:
		return types.BoolNull()
	case *value != permission.GrantsRemoteAccessByDefault():
		return types.BoolValue(*value)
	default:
		return types.BoolNull()
	}
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
//
// The resource keeps a declared entry even when it equals the defaults, so the state matches the
// plan, and keeps every non-default entry on an undeclared cluster or project, so a permission
// granted from the Console shows up in the plan. The write path resets undeclared entries to the
// defaults: q-core rejects an update that omits a cluster or project and replaces the whole
// matrix, so that is the only way to remove such a grant.
func convertDomainCustomRoleToCustomRole(role *customrole.CustomRole, declared *CustomRole, mode customRoleReadMode) CustomRole {
	declaredClusters := map[string]bool{}
	declaredProjects := map[string]bool{}
	declaredRemoteAccesses := map[customRoleEnvKey]declaredRemoteAccess{}
	declaredClustersNull, declaredProjectsNull := true, true
	if declared != nil {
		declaredClusters = declaredIDSet(declared.ClusterPermissions, "cluster_id")
		declaredProjects = declaredIDSet(declared.ProjectPermissions, "project_id")
		declaredRemoteAccesses = declaredRemoteAccessByEnv(declared.ProjectPermissions)
		declaredClustersNull = declared.ClusterPermissions.IsNull()
		declaredProjectsNull = declared.ProjectPermissions.IsNull()
	}

	clusterObjects := make([]attr.Value, 0, len(role.ClusterPermissions))
	for _, cp := range role.ClusterPermissions {
		if mode == customRoleReadModeDeclaredOrNonDefault && !declaredClusters[cp.ClusterID] && isDefaultClusterPermission(cp) {
			continue
		}
		clusterObjects = append(clusterObjects, types.ObjectValueMust(customRoleClusterPermissionAttrTypes, map[string]attr.Value{
			"cluster_id": FromString(cp.ClusterID),
			"permission": FromString(string(cp.Permission)),
		}))
	}

	projectObjects := make([]attr.Value, 0, len(role.ProjectPermissions))
	for _, pp := range role.ProjectPermissions {
		if mode == customRoleReadModeDeclaredOrNonDefault && !declaredProjects[pp.ProjectID] && isDefaultProjectPermission(pp) {
			continue
		}
		permissionsValue := types.SetNull(types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes})
		if !pp.IsAdmin {
			permObjects := make([]attr.Value, 0, len(pp.Permissions))
			for _, ep := range pp.Permissions {
				declaredFlags, envDeclared := declaredRemoteAccesses[customRoleEnvKey{pp.ProjectID, string(ep.EnvironmentType)}]
				permObjects = append(permObjects, types.ObjectValueMust(customRoleEnvPermissionAttrTypes, map[string]attr.Value{
					"environment_type": FromString(string(ep.EnvironmentType)),
					"permission":       FromString(string(ep.Permission)),
					"shell":            remoteAccessStateValue(ep.Shell, ep.Permission, declaredFlags.shell, envDeclared, mode),
					"port_forward":     remoteAccessStateValue(ep.PortForward, ep.Permission, declaredFlags.portForward, envDeclared, mode),
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
		Description:        storedDescriptionFromAPI(role.Description),
		ClusterPermissions: clusterSet,
		ProjectPermissions: projectSet,
	}
}
