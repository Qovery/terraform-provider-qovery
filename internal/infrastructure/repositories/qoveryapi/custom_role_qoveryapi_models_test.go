//go:build unit && !integration

package qoveryapi

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

const (
	testCustomRoleOrgID     = "00000000-0000-0000-0000-00000000aaaa"
	testCustomRoleID        = "00000000-0000-0000-0000-00000000bbbb"
	testCustomRoleClusterID = "00000000-0000-0000-0000-00000000cc01"
	testCustomRoleProjectA  = "00000000-0000-0000-0000-00000000dd01"
	testCustomRoleProjectB  = "00000000-0000-0000-0000-00000000dd02"
)

func customRolePtr[T any](v T) *T { return &v }

// full server matrix: 1 cluster (VIEWER), 2 projects (NO_ACCESS everywhere)
func customRoleServerRole() *qovery.OrganizationCustomRole {
	noAccess := func() []qovery.OrganizationCustomRoleProjectPermissionsInnerPermissionsInner {
		perms := make([]qovery.OrganizationCustomRoleProjectPermissionsInnerPermissionsInner, 0, 4)
		for _, et := range []qovery.EnvironmentModeEnum{
			qovery.ENVIRONMENTMODEENUM_DEVELOPMENT, qovery.ENVIRONMENTMODEENUM_PREVIEW,
			qovery.ENVIRONMENTMODEENUM_STAGING, qovery.ENVIRONMENTMODEENUM_PRODUCTION,
		} {
			et := et
			perms = append(perms, qovery.OrganizationCustomRoleProjectPermissionsInnerPermissionsInner{
				EnvironmentType: &et,
				Permission:      customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_NO_ACCESS),
			})
		}
		return perms
	}
	return &qovery.OrganizationCustomRole{
		Id:   customRolePtr(testCustomRoleID),
		Name: customRolePtr("my-role"),
		ClusterPermissions: []qovery.OrganizationCustomRoleClusterPermissionsInner{
			{ClusterId: customRolePtr(testCustomRoleClusterID), ClusterName: customRolePtr("cluster-1"), Permission: customRolePtr(qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_VIEWER)},
		},
		ProjectPermissions: []qovery.OrganizationCustomRoleProjectPermissionsInner{
			{ProjectId: customRolePtr(testCustomRoleProjectA), ProjectName: customRolePtr("proj-a"), IsAdmin: customRolePtr(false), Permissions: noAccess()},
			{ProjectId: customRolePtr(testCustomRoleProjectB), ProjectName: customRolePtr("proj-b"), IsAdmin: customRolePtr(false), Permissions: noAccess()},
		},
	}
}

func TestNewDomainCustomRoleFromQovery(t *testing.T) {
	t.Parallel()

	t.Run("nil role returns error", func(t *testing.T) {
		role, err := newDomainCustomRoleFromQovery(testCustomRoleOrgID, nil)
		assert.Error(t, err)
		assert.Nil(t, role)
	})

	t.Run("full matrix converts, is_admin project keeps nil permissions", func(t *testing.T) {
		server := customRoleServerRole()
		server.ProjectPermissions[0].IsAdmin = customRolePtr(true)
		server.ProjectPermissions[0].Permissions = []qovery.OrganizationCustomRoleProjectPermissionsInnerPermissionsInner{}

		role, err := newDomainCustomRoleFromQovery(testCustomRoleOrgID, server)
		require.NoError(t, err)
		assert.Equal(t, testCustomRoleID, role.ID.String())
		assert.Equal(t, testCustomRoleOrgID, role.OrganizationID.String())
		assert.Equal(t, "my-role", role.Name)
		require.Len(t, role.ClusterPermissions, 1)
		assert.Equal(t, customrole.ClusterPermissionViewer, role.ClusterPermissions[0].Permission)
		require.Len(t, role.ProjectPermissions, 2)
		assert.True(t, role.ProjectPermissions[0].IsAdmin)
		assert.Empty(t, role.ProjectPermissions[0].Permissions)
		assert.False(t, role.ProjectPermissions[1].IsAdmin)
		assert.Len(t, role.ProjectPermissions[1].Permissions, 4)
	})
}

func TestNewDomainCustomRoleFromQoveryRemoteAccess(t *testing.T) {
	t.Parallel()

	server := customRoleServerRole()
	server.ProjectPermissions[1].Permissions[3].Permission = customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_DEPLOYER)
	server.ProjectPermissions[1].Permissions[3].Shell = false
	server.ProjectPermissions[1].Permissions[3].PortForward = true

	role, err := newDomainCustomRoleFromQovery(testCustomRoleOrgID, server)
	require.NoError(t, err)
	assert.Equal(t, customrole.EnvironmentPermission{
		EnvironmentType: customrole.EnvironmentTypeProduction,
		Permission:      customrole.ProjectPermissionDeployer,
		Shell:           customRolePtr(false),
		PortForward:     customRolePtr(true),
	}, role.ProjectPermissions[1].Permissions[3])
}

func TestNewQoveryCustomRoleEditRequestFromRemoteAccess(t *testing.T) {
	t.Parallel()

	type flags struct{ shell, portForward bool }
	flagsByEnvType := func(req *qovery.OrganizationCustomRoleUpdateRequest, projectID string) map[qovery.EnvironmentModeEnum]flags {
		byEnvType := map[qovery.EnvironmentModeEnum]flags{}
		for _, pp := range req.ProjectPermissions {
			if *pp.ProjectId != projectID {
				continue
			}
			for _, p := range pp.Permissions {
				byEnvType[*p.EnvironmentType] = flags{p.Shell, p.PortForward}
			}
		}
		return byEnvType
	}

	t.Run("declared values are sent, the others follow the level on a newly granted environment type", func(t *testing.T) {
		// the server role has NO_ACCESS everywhere, so nothing is granted yet
		req, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionDeployer, Shell: customRolePtr(false), PortForward: customRolePtr(true)},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionDeployer},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionViewer, Shell: customRolePtr(true)},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionNoAccess},
				}},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, map[qovery.EnvironmentModeEnum]flags{
			qovery.ENVIRONMENTMODEENUM_DEVELOPMENT: {false, true},
			qovery.ENVIRONMENTMODEENUM_PREVIEW:     {true, true},
			qovery.ENVIRONMENTMODEENUM_STAGING:     {true, false},
			qovery.ENVIRONMENTMODEENUM_PRODUCTION:  {false, false},
		}, flagsByEnvType(req, testCustomRoleProjectA))
		for _, f := range flagsByEnvType(req, testCustomRoleProjectB) {
			assert.Equal(t, flags{false, false}, f, "undeclared project is NO_ACCESS")
		}
	})

	t.Run("an attribute not set keeps the current value, so a revocation made in the console survives", func(t *testing.T) {
		// given shell revoked from the console on a DEPLOYER environment type
		server := customRoleServerRole()
		server.ProjectPermissions[0].Permissions[0].Permission = customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_DEPLOYER)
		server.ProjectPermissions[0].Permissions[0].Shell = false
		server.ProjectPermissions[0].Permissions[0].PortForward = true

		// when the level changes without shell nor port_forward in the configuration
		req, err := newQoveryCustomRoleEditRequestFrom(server, customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionManager},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionNoAccess},
				}},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, flags{false, true}, flagsByEnvType(req, testCustomRoleProjectA)[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT])
	})

	t.Run("a level change applies the default of the new level when the stored value is the default of the old one", func(t *testing.T) {
		// VIEWER with both off, DEPLOYER with both on: the stored values were never set on purpose
		server := customRoleServerRole()
		server.ProjectPermissions[0].Permissions[0].Permission = customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_VIEWER)
		server.ProjectPermissions[0].Permissions[1].Permission = customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_DEPLOYER)
		server.ProjectPermissions[0].Permissions[1].Shell = true
		server.ProjectPermissions[0].Permissions[1].PortForward = true

		req, err := newQoveryCustomRoleEditRequestFrom(server, customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionDeployer},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionViewer},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionNoAccess},
				}},
			},
		})
		require.NoError(t, err)
		byEnvType := flagsByEnvType(req, testCustomRoleProjectA)
		assert.Equal(t, flags{true, true}, byEnvType[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT], "VIEWER to DEPLOYER turns them on")
		assert.Equal(t, flags{false, false}, byEnvType[qovery.ENVIRONMENTMODEENUM_PREVIEW], "DEPLOYER to VIEWER turns them off")
	})

	t.Run("a shell granted on purpose on a VIEWER survives a level change", func(t *testing.T) {
		server := customRoleServerRole()
		server.ProjectPermissions[0].Permissions[0].Permission = customRolePtr(qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_VIEWER)
		server.ProjectPermissions[0].Permissions[0].Shell = true

		req, err := newQoveryCustomRoleEditRequestFrom(server, customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionDeployer},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionNoAccess},
				}},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, flags{true, true}, flagsByEnvType(req, testCustomRoleProjectA)[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT])
	})

	t.Run("demoting an admin project keeps shell and port_forward when not set", func(t *testing.T) {
		server := customRoleServerRole()
		server.ProjectPermissions[0].IsAdmin = customRolePtr(true)
		server.ProjectPermissions[0].Permissions = nil

		req, err := newQoveryCustomRoleEditRequestFrom(server, customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionManager},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionNoAccess},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionNoAccess},
				}},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, flags{true, true}, flagsByEnvType(req, testCustomRoleProjectA)[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT])
		assert.Equal(t, flags{false, false}, flagsByEnvType(req, testCustomRoleProjectA)[qovery.ENVIRONMENTMODEENUM_PREVIEW])
	})
}

func TestNewQoveryCustomRoleEditRequestFrom(t *testing.T) {
	t.Parallel()

	t.Run("declared entries overlaid, undeclared default-filled", func(t *testing.T) {
		req, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ClusterPermissions: []customrole.ClusterRolePermission{
				{ClusterID: testCustomRoleClusterID, Permission: customrole.ClusterPermissionAdmin},
			},
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, IsAdmin: true},
			},
		})
		require.NoError(t, err)
		require.Len(t, req.ClusterPermissions, 1)
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_ADMIN, *req.ClusterPermissions[0].Permission)
		require.Len(t, req.ProjectPermissions, 2)
		byID := map[string]qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInner{}
		for _, pp := range req.ProjectPermissions {
			byID[*pp.ProjectId] = pp
		}
		assert.True(t, *byID[testCustomRoleProjectA].IsAdmin)
		assert.Empty(t, byID[testCustomRoleProjectA].Permissions)
		// undeclared project default-filled with NO_ACCESS on all 4 env types
		assert.False(t, *byID[testCustomRoleProjectB].IsAdmin)
		require.Len(t, byID[testCustomRoleProjectB].Permissions, 4)
		for _, p := range byID[testCustomRoleProjectB].Permissions {
			assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_NO_ACCESS, *p.Permission)
		}
	})

	t.Run("declared non-admin project with full env-type set overlaid verbatim", func(t *testing.T) {
		req, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, IsAdmin: false, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionManager},
					{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionManager},
					{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionDeployer},
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionViewer},
				}},
			},
		})
		require.NoError(t, err)
		require.Len(t, req.ProjectPermissions, 2)
		byID := map[string]qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInner{}
		for _, pp := range req.ProjectPermissions {
			byID[*pp.ProjectId] = pp
		}
		projA := byID[testCustomRoleProjectA]
		assert.False(t, *projA.IsAdmin)
		require.Len(t, projA.Permissions, 4)
		byEnvType := map[qovery.EnvironmentModeEnum]qovery.OrganizationCustomRoleProjectPermission{}
		for _, p := range projA.Permissions {
			byEnvType[*p.EnvironmentType] = *p.Permission
		}
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_MANAGER, byEnvType[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_MANAGER, byEnvType[qovery.ENVIRONMENTMODEENUM_PREVIEW])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_DEPLOYER, byEnvType[qovery.ENVIRONMENTMODEENUM_STAGING])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_VIEWER, byEnvType[qovery.ENVIRONMENTMODEENUM_PRODUCTION])
	})

	t.Run("declared non-admin project with partial env-type set default-fills remaining with NO_ACCESS", func(t *testing.T) {
		req, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: testCustomRoleProjectA, IsAdmin: false, Permissions: []customrole.EnvironmentPermission{
					{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionManager},
				}},
			},
		})
		require.NoError(t, err)
		require.Len(t, req.ProjectPermissions, 2)
		byID := map[string]qovery.OrganizationCustomRoleUpdateRequestProjectPermissionsInner{}
		for _, pp := range req.ProjectPermissions {
			byID[*pp.ProjectId] = pp
		}
		projA := byID[testCustomRoleProjectA]
		assert.False(t, *projA.IsAdmin)
		require.Len(t, projA.Permissions, 4)
		byEnvType := map[qovery.EnvironmentModeEnum]qovery.OrganizationCustomRoleProjectPermission{}
		for _, p := range projA.Permissions {
			byEnvType[*p.EnvironmentType] = *p.Permission
		}
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_MANAGER, byEnvType[qovery.ENVIRONMENTMODEENUM_PRODUCTION])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_NO_ACCESS, byEnvType[qovery.ENVIRONMENTMODEENUM_DEVELOPMENT])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_NO_ACCESS, byEnvType[qovery.ENVIRONMENTMODEENUM_PREVIEW])
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLEPROJECTPERMISSION_NO_ACCESS, byEnvType[qovery.ENVIRONMENTMODEENUM_STAGING])
	})

	t.Run("nothing declared produces pure default matrix", func(t *testing.T) {
		req, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{Name: "my-role"})
		require.NoError(t, err)
		require.Len(t, req.ClusterPermissions, 1)
		assert.Equal(t, qovery.ORGANIZATIONCUSTOMROLECLUSTERPERMISSION_VIEWER, *req.ClusterPermissions[0].Permission)
		require.Len(t, req.ProjectPermissions, 2)
	})

	t.Run("declared unknown cluster id errors", func(t *testing.T) {
		_, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ClusterPermissions: []customrole.ClusterRolePermission{
				{ClusterID: "00000000-0000-0000-0000-00000000ffff", Permission: customrole.ClusterPermissionAdmin},
			},
		})
		assert.ErrorContains(t, err, ErrUnknownClusterID.Error())
	})

	t.Run("declared unknown project id errors", func(t *testing.T) {
		_, err := newQoveryCustomRoleEditRequestFrom(customRoleServerRole(), customrole.UpsertRequest{
			Name: "my-role",
			ProjectPermissions: []customrole.ProjectRolePermission{
				{ProjectID: "00000000-0000-0000-0000-00000000ffff", IsAdmin: true},
			},
		})
		assert.ErrorContains(t, err, ErrUnknownProjectID.Error())
	})
}
