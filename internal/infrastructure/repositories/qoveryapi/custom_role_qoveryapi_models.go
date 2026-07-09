package qoveryapi

import (
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

var (
	ErrUnknownClusterID = errors.New("declared cluster_id does not exist in the organization")
	ErrUnknownProjectID = errors.New("declared project_id does not exist in the organization")
)

func newDomainCustomRoleFromQovery(organizationID string, role *qovery.OrganizationCustomRole) (*customrole.CustomRole, error) {
	if role == nil {
		return nil, customrole.ErrInvalidCustomRole
	}

	roleID, err := parseUUID(role.GetId(), customrole.ErrInvalidCustomRoleIdParam)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(organizationID, customrole.ErrInvalidOrganizationIdParam)
	if err != nil {
		return nil, err
	}

	clusterPermissions := make([]customrole.ClusterRolePermission, 0, len(role.ClusterPermissions))
	for _, cp := range role.ClusterPermissions {
		clusterPermissions = append(clusterPermissions, customrole.ClusterRolePermission{
			ClusterID:  cp.GetClusterId(),
			Permission: customrole.ClusterPermission(string(cp.GetPermission())),
		})
	}

	projectPermissions := make([]customrole.ProjectRolePermission, 0, len(role.ProjectPermissions))
	for _, pp := range role.ProjectPermissions {
		project := customrole.ProjectRolePermission{
			ProjectID: pp.GetProjectId(),
			IsAdmin:   pp.GetIsAdmin(),
		}
		if !project.IsAdmin {
			project.Permissions = make([]customrole.EnvironmentPermission, 0, len(pp.Permissions))
			for _, ep := range pp.Permissions {
				project.Permissions = append(project.Permissions, customrole.EnvironmentPermission{
					EnvironmentType: customrole.EnvironmentType(string(ep.GetEnvironmentType())),
					Permission:      customrole.ProjectPermission(string(ep.GetPermission())),
				})
			}
		}
		projectPermissions = append(projectPermissions, project)
	}

	domainRole := &customrole.CustomRole{
		ID:                 roleID,
		OrganizationID:     orgID,
		Name:               role.GetName(),
		Description:        role.Description,
		ClusterPermissions: clusterPermissions,
		ProjectPermissions: projectPermissions,
	}
	if err := domainRole.Validate(); err != nil {
		return nil, err
	}
	return domainRole, nil
}

// newQoveryCustomRoleEditRequestFrom builds the full-replace PUT payload the API requires:
// every cluster/project present in `current` (the authoritative org enumeration) appears once,
// declared entries win, everything else keeps the server defaults (VIEWER / NO_ACCESS).
func newQoveryCustomRoleEditRequestFrom(current *qovery.OrganizationCustomRole, request customrole.UpsertRequest) (*qovery.OrganizationCustomRoleUpdateRequest, error) {
	declaredClusters := make(map[string]customrole.ClusterRolePermission, len(request.ClusterPermissions))
	for _, cp := range request.ClusterPermissions {
		declaredClusters[cp.ClusterID] = cp
	}
	declaredProjects := make(map[string]customrole.ProjectRolePermission, len(request.ProjectPermissions))
	for _, pp := range request.ProjectPermissions {
		declaredProjects[pp.ProjectID] = pp
	}

	clusterPermissions := make([]qovery.OrganizationCustomRoleUpdateRequestClusterPermissionsInner, 0, len(current.ClusterPermissions))
	seenClusters := make(map[string]bool, len(current.ClusterPermissions))
	for _, cur := range current.ClusterPermissions {
		clusterID := cur.GetClusterId()
		seenClusters[clusterID] = true
		permission := qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_VIEWER
		if declared, ok := declaredClusters[clusterID]; ok {
			p, err := qovery.NewOrganizationCustomRoleClusterPermissionFromValue(string(declared.Permission))
			if err != nil {
				return nil, err
			}
			permission = *p
		}
		permissionValue := permission
		clusterIDValue := clusterID
		clusterPermissions = append(clusterPermissions, qovery.OrganizationCustomRoleUpdateRequestClusterPermissionsInner{
			ClusterId:  &clusterIDValue,
			Permission: &permissionValue,
		})
	}
	for clusterID := range declaredClusters {
		if !seenClusters[clusterID] {
			return nil, errors.Wrap(ErrUnknownClusterID, clusterID)
		}
	}

	projectPermissions := make([]qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInner, 0, len(current.ProjectPermissions))
	seenProjects := make(map[string]bool, len(current.ProjectPermissions))
	for _, cur := range current.ProjectPermissions {
		projectIDValue := cur.GetProjectId()
		seenProjects[projectIDValue] = true

		isAdmin := false
		var envPermissions []customrole.EnvironmentPermission
		if declared, ok := declaredProjects[projectIDValue]; ok {
			isAdmin = declared.IsAdmin
			envPermissions = declared.Permissions
		}

		inner := qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInner{
			ProjectId: &projectIDValue,
			IsAdmin:   &isAdmin,
		}
		if !isAdmin {
			perms := make([]qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInnerPermissionsInner, 0, len(customrole.AllowedEnvironmentTypes))
			declaredByEnvType := make(map[customrole.EnvironmentType]customrole.ProjectPermission, len(envPermissions))
			for _, ep := range envPermissions {
				declaredByEnvType[ep.EnvironmentType] = ep.Permission
			}
			for _, envType := range customrole.AllowedEnvironmentTypes {
				permissionValue := customrole.ProjectPermissionNoAccess
				if declared, ok := declaredByEnvType[envType]; ok {
					permissionValue = declared
				}
				et, err := qovery.NewEnvironmentModeEnumFromValue(string(envType))
				if err != nil {
					return nil, err
				}
				p, err := qovery.NewOrganizationCustomRoleProjectPermissionFromValue(string(permissionValue))
				if err != nil {
					return nil, err
				}
				perms = append(perms, qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInnerPermissionsInner{
					EnvironmentType: et,
					Permission:      p,
				})
			}
			inner.Permissions = perms
		}
		projectPermissions = append(projectPermissions, inner)
	}
	for projectID := range declaredProjects {
		if !seenProjects[projectID] {
			return nil, errors.Wrap(ErrUnknownProjectID, projectID)
		}
	}

	return &qovery.OrganizationCustomRoleUpdateRequest{
		Name:               request.Name,
		Description:        request.Description,
		ClusterPermissions: clusterPermissions,
		ProjectPermissions: projectPermissions,
	}, nil
}
