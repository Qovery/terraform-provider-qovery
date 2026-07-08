//go:build unit && !integration

package customrole_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
)

func validRequest() customrole.UpsertRequest {
	return customrole.UpsertRequest{
		Name: "project-admin",
		ClusterPermissions: []customrole.ClusterRolePermission{
			{ClusterID: "6c1f4b3e-1b1a-4b0e-8b0a-000000000001", Permission: customrole.ClusterPermissionEnvCreator},
		},
		ProjectPermissions: []customrole.ProjectRolePermission{
			{ProjectID: "6c1f4b3e-1b1a-4b0e-8b0a-000000000002", IsAdmin: true},
			{ProjectID: "6c1f4b3e-1b1a-4b0e-8b0a-000000000003", Permissions: []customrole.EnvironmentPermission{
				{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionManager},
				{EnvironmentType: customrole.EnvironmentTypePreview, Permission: customrole.ProjectPermissionManager},
				{EnvironmentType: customrole.EnvironmentTypeStaging, Permission: customrole.ProjectPermissionDeployer},
				{EnvironmentType: customrole.EnvironmentTypeProduction, Permission: customrole.ProjectPermissionViewer},
			}},
		},
	}
}

func TestUpsertRequestValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		description string
		mutate      func(r *customrole.UpsertRequest)
		expectedErr error
	}{
		{description: "valid request", mutate: func(r *customrole.UpsertRequest) {}, expectedErr: nil},
		{description: "empty declared permissions is valid", mutate: func(r *customrole.UpsertRequest) {
			r.ClusterPermissions = nil
			r.ProjectPermissions = nil
		}, expectedErr: nil},
		{description: "blank name", mutate: func(r *customrole.UpsertRequest) { r.Name = "  " }, expectedErr: customrole.ErrInvalidName},
		{description: "untrimmed name", mutate: func(r *customrole.UpsertRequest) { r.Name = "role " }, expectedErr: customrole.ErrInvalidName},
		{description: "reserved name case-insensitive", mutate: func(r *customrole.UpsertRequest) { r.Name = "Admin" }, expectedErr: customrole.ErrReservedName},
		{description: "reserved name owner", mutate: func(r *customrole.UpsertRequest) { r.Name = "owner" }, expectedErr: customrole.ErrReservedName},
		{description: "invalid cluster permission", mutate: func(r *customrole.UpsertRequest) {
			r.ClusterPermissions[0].Permission = "SUPERUSER"
		}, expectedErr: customrole.ErrInvalidClusterPermission},
		{description: "invalid cluster id", mutate: func(r *customrole.UpsertRequest) {
			r.ClusterPermissions[0].ClusterID = "not-a-uuid"
		}, expectedErr: customrole.ErrInvalidUpsertRequest},
		{description: "duplicate cluster id", mutate: func(r *customrole.UpsertRequest) {
			r.ClusterPermissions = append(r.ClusterPermissions, r.ClusterPermissions[0])
		}, expectedErr: customrole.ErrDuplicateClusterID},
		{description: "duplicate project id", mutate: func(r *customrole.UpsertRequest) {
			r.ProjectPermissions = append(r.ProjectPermissions, r.ProjectPermissions[0])
		}, expectedErr: customrole.ErrDuplicateProjectID},
		{description: "is_admin with permissions", mutate: func(r *customrole.UpsertRequest) {
			r.ProjectPermissions[0].Permissions = []customrole.EnvironmentPermission{
				{EnvironmentType: customrole.EnvironmentTypeDevelopment, Permission: customrole.ProjectPermissionViewer},
			}
		}, expectedErr: customrole.ErrAdminWithPermissions},
		{description: "missing env types", mutate: func(r *customrole.UpsertRequest) {
			r.ProjectPermissions[1].Permissions = r.ProjectPermissions[1].Permissions[:3]
		}, expectedErr: customrole.ErrIncompleteEnvironmentTypes},
		{description: "duplicate env type", mutate: func(r *customrole.UpsertRequest) {
			r.ProjectPermissions[1].Permissions[3].EnvironmentType = customrole.EnvironmentTypeDevelopment
		}, expectedErr: customrole.ErrDuplicateEnvironmentType},
		{description: "invalid project permission", mutate: func(r *customrole.UpsertRequest) {
			r.ProjectPermissions[1].Permissions[0].Permission = "GOD"
		}, expectedErr: customrole.ErrInvalidProjectPermission},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			req := validRequest()
			tc.mutate(&req)
			err := req.Validate()
			if tc.expectedErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}
