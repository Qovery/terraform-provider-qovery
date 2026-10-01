//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

var (
	customRoleClusterObjectType = types.ObjectType{AttrTypes: customRoleClusterPermissionAttrTypes}
	customRoleProjectObjectType = types.ObjectType{AttrTypes: customRoleProjectPermissionAttrTypes}
)

func customRoleClusterSet(entries map[string]customrole.ClusterPermission) types.Set {
	elements := make([]attr.Value, 0, len(entries))
	for clusterID, permission := range entries {
		elements = append(elements, types.ObjectValueMust(customRoleClusterPermissionAttrTypes, map[string]attr.Value{
			"cluster_id": types.StringValue(clusterID),
			"permission": types.StringValue(string(permission)),
		}))
	}
	return types.SetValueMust(customRoleClusterObjectType, elements)
}

func customRoleAdminProjectSet(projectIDs ...string) types.Set {
	elements := make([]attr.Value, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		elements = append(elements, types.ObjectValueMust(customRoleProjectPermissionAttrTypes, map[string]attr.Value{
			"project_id":  types.StringValue(projectID),
			"is_admin":    types.BoolValue(true),
			"permissions": types.SetNull(types.ObjectType{AttrTypes: customRoleEnvPermissionAttrTypes}),
		}))
	}
	return types.SetValueMust(customRoleProjectObjectType, elements)
}

func noAccessProject(projectID string) customrole.ProjectRolePermission {
	permissions := make([]customrole.EnvironmentPermission, 0, len(customrole.AllowedEnvironmentTypes))
	for _, envType := range customrole.AllowedEnvironmentTypes {
		permissions = append(permissions, customrole.EnvironmentPermission{EnvironmentType: envType, Permission: customrole.ProjectPermissionNoAccess})
	}
	return customrole.ProjectRolePermission{ProjectID: projectID, Permissions: permissions}
}

// clusterIDsOf returns the cluster_id of every entry of a cluster_permissions set.
func clusterIDsOf(set types.Set) []string {
	ids := make([]string, 0, len(set.Elements()))
	for _, elem := range set.Elements() {
		ids = append(ids, elem.(types.Object).Attributes()["cluster_id"].(types.String).ValueString())
	}
	return ids
}

// projectIDsOf returns the project_id of every entry of a project_permissions set.
func projectIDsOf(set types.Set) []string {
	ids := make([]string, 0, len(set.Elements()))
	for _, elem := range set.Elements() {
		ids = append(ids, elem.(types.Object).Attributes()["project_id"].(types.String).ValueString())
	}
	return ids
}

func TestConvertDomainCustomRoleToCustomRole(t *testing.T) {
	t.Parallel()

	declaredCluster, consoleCluster, defaultCluster := uuid.NewString(), uuid.NewString(), uuid.NewString()
	declaredProject, consoleProject, defaultProject := uuid.NewString(), uuid.NewString(), uuid.NewString()

	// The API returns an entry for every cluster and project of the organization.
	role := &customrole.CustomRole{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "role",
		Description:    new("described"),
		ClusterPermissions: []customrole.ClusterRolePermission{
			{ClusterID: declaredCluster, Permission: customrole.ClusterPermissionViewer},
			{ClusterID: consoleCluster, Permission: customrole.ClusterPermissionAdmin},
			{ClusterID: defaultCluster, Permission: customrole.ClusterPermissionViewer},
		},
		ProjectPermissions: []customrole.ProjectRolePermission{
			{ProjectID: declaredProject, IsAdmin: true},
			{ProjectID: consoleProject, IsAdmin: true},
			noAccessProject(defaultProject),
		},
	}
	defaultsOnly := &customrole.CustomRole{
		ID:                 role.ID,
		OrganizationID:     role.OrganizationID,
		Name:               role.Name,
		ClusterPermissions: []customrole.ClusterRolePermission{{ClusterID: defaultCluster, Permission: customrole.ClusterPermissionViewer}},
		ProjectPermissions: []customrole.ProjectRolePermission{noAccessProject(defaultProject)},
	}

	testCases := []struct {
		TestName          string
		Role              *customrole.CustomRole
		Declared          *CustomRole
		Mode              customRoleReadMode
		ExpectClusters    []string
		ExpectProjects    []string
		ExpectNullSets    bool
		ExpectEmptySets   bool
		ExpectDescription types.String
	}{
		{
			TestName: "keeps_declared_entries_and_console_grants_on_undeclared_targets",
			Role:     role,
			Declared: &CustomRole{
				ClusterPermissions: customRoleClusterSet(map[string]customrole.ClusterPermission{declaredCluster: customrole.ClusterPermissionViewer}),
				ProjectPermissions: customRoleAdminProjectSet(declaredProject),
			},
			Mode:              customRoleReadModeDeclaredOrNonDefault,
			ExpectClusters:    []string{declaredCluster, consoleCluster},
			ExpectProjects:    []string{declaredProject, consoleProject},
			ExpectDescription: types.StringValue("described"),
		},
		{
			TestName: "console_grants_show_up_when_nothing_is_declared",
			Role:     role,
			Declared: &CustomRole{
				ClusterPermissions: types.SetNull(customRoleClusterObjectType),
				ProjectPermissions: types.SetNull(customRoleProjectObjectType),
			},
			Mode:              customRoleReadModeDeclaredOrNonDefault,
			ExpectClusters:    []string{consoleCluster},
			ExpectProjects:    []string{declaredProject, consoleProject},
			ExpectDescription: types.StringValue("described"),
		},
		{
			TestName:          "import_records_the_non_default_entries",
			Role:              role,
			Declared:          &CustomRole{},
			Mode:              customRoleReadModeDeclaredOrNonDefault,
			ExpectClusters:    []string{consoleCluster},
			ExpectProjects:    []string{declaredProject, consoleProject},
			ExpectDescription: types.StringValue("described"),
		},
		{
			TestName: "only_defaults_keep_null_sets_null",
			Role:     defaultsOnly,
			Declared: &CustomRole{
				ClusterPermissions: types.SetNull(customRoleClusterObjectType),
				ProjectPermissions: types.SetNull(customRoleProjectObjectType),
			},
			Mode:              customRoleReadModeDeclaredOrNonDefault,
			ExpectNullSets:    true,
			ExpectDescription: types.StringValue(""),
		},
		{
			TestName: "only_defaults_keep_empty_sets_empty",
			Role:     defaultsOnly,
			Declared: &CustomRole{
				ClusterPermissions: types.SetValueMust(customRoleClusterObjectType, []attr.Value{}),
				ProjectPermissions: types.SetValueMust(customRoleProjectObjectType, []attr.Value{}),
			},
			Mode:              customRoleReadModeDeclaredOrNonDefault,
			ExpectEmptySets:   true,
			ExpectDescription: types.StringValue(""),
		},
		{
			TestName:          "data_source_keeps_the_full_matrix",
			Role:              role,
			Mode:              customRoleReadModeKeepAll,
			ExpectClusters:    []string{declaredCluster, consoleCluster, defaultCluster},
			ExpectProjects:    []string{declaredProject, consoleProject, defaultProject},
			ExpectDescription: types.StringValue("described"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			got := convertDomainCustomRoleToCustomRole(tc.Role, tc.Declared, tc.Mode)

			assert.Equal(t, tc.ExpectDescription, got.Description)
			switch {
			case tc.ExpectNullSets:
				assert.True(t, got.ClusterPermissions.IsNull(), "cluster_permissions: got %s", got.ClusterPermissions)
				assert.True(t, got.ProjectPermissions.IsNull(), "project_permissions: got %s", got.ProjectPermissions)
			case tc.ExpectEmptySets:
				assert.False(t, got.ClusterPermissions.IsNull())
				assert.Empty(t, got.ClusterPermissions.Elements())
				assert.False(t, got.ProjectPermissions.IsNull())
				assert.Empty(t, got.ProjectPermissions.Elements())
			default:
				assert.ElementsMatch(t, tc.ExpectClusters, clusterIDsOf(got.ClusterPermissions))
				assert.ElementsMatch(t, tc.ExpectProjects, projectIDsOf(got.ProjectPermissions))
			}
		})
	}
}
