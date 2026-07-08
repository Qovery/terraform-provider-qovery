//go:build unit && !integration

package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/customrole"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewCustomRoleService(t *testing.T) {
	t.Parallel()
	svc, err := services.NewCustomRoleService(nil)
	assert.Error(t, err)
	assert.Nil(t, svc)

	svc, err = services.NewCustomRoleService(mocks_test.NewCustomRoleRepository(t))
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestCustomRoleServiceCreate(t *testing.T) {
	t.Parallel()
	orgID := uuid.NewString()
	validReq := customrole.UpsertRequest{Name: "my-role"}

	t.Run("invalid organization id", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Create(context.Background(), "not-a-uuid", validReq)
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrFailedToCreateCustomRole.Error())
	})

	t.Run("invalid request", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Create(context.Background(), orgID, customrole.UpsertRequest{Name: "admin"})
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrReservedName.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		expected := &customrole.CustomRole{ID: uuid.New(), OrganizationID: uuid.MustParse(orgID), Name: "my-role"}
		repo.EXPECT().Create(context.Background(), orgID, validReq).Return(expected, nil)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Create(context.Background(), orgID, validReq)
		require.NoError(t, err)
		assert.Equal(t, expected, role)
	})
}

func TestCustomRoleServiceGet(t *testing.T) {
	t.Parallel()
	orgID, roleID := uuid.NewString(), uuid.NewString()

	t.Run("invalid role id", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Get(context.Background(), orgID, "nope")
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrFailedToGetCustomRole.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		expected := &customrole.CustomRole{ID: uuid.MustParse(roleID), OrganizationID: uuid.MustParse(orgID), Name: "my-role"}
		repo.EXPECT().Get(context.Background(), orgID, roleID).Return(expected, nil)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Get(context.Background(), orgID, roleID)
		require.NoError(t, err)
		assert.Equal(t, expected, role)
	})
}

func TestCustomRoleServiceUpdate(t *testing.T) {
	t.Parallel()
	orgID, roleID := uuid.NewString(), uuid.NewString()
	validReq := customrole.UpsertRequest{Name: "my-role"}

	t.Run("invalid organization id", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Update(context.Background(), "not-a-uuid", roleID, validReq)
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrFailedToUpdateCustomRole.Error())
	})

	t.Run("invalid role id", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Update(context.Background(), orgID, "nope", validReq)
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrFailedToUpdateCustomRole.Error())
	})

	t.Run("invalid request", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Update(context.Background(), orgID, roleID, customrole.UpsertRequest{Name: "admin"})
		assert.Nil(t, role)
		assert.ErrorContains(t, err, customrole.ErrReservedName.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		expected := &customrole.CustomRole{ID: uuid.MustParse(roleID), OrganizationID: uuid.MustParse(orgID), Name: "my-role"}
		repo.EXPECT().Update(context.Background(), orgID, roleID, validReq).Return(expected, nil)
		svc, _ := services.NewCustomRoleService(repo)
		role, err := svc.Update(context.Background(), orgID, roleID, validReq)
		require.NoError(t, err)
		assert.Equal(t, expected, role)
	})
}

func TestCustomRoleServiceDelete(t *testing.T) {
	t.Parallel()
	orgID, roleID := uuid.NewString(), uuid.NewString()

	t.Run("repository success", func(t *testing.T) {
		repo := mocks_test.NewCustomRoleRepository(t)
		repo.EXPECT().Delete(context.Background(), orgID, roleID).Return(nil)
		svc, _ := services.NewCustomRoleService(repo)
		assert.NoError(t, svc.Delete(context.Background(), orgID, roleID))
	})
}
