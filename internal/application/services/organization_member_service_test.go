//go:build unit && !integration

package services_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/member"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

const (
	testMemberOrganizationID = "01234567-8901-2345-6789-012345678901"
	testMemberRoleID         = "11234567-8901-2345-6789-012345678901"
	testMemberEmail          = "dev@company.com"
)

func testDomainMember() *member.Member {
	roleID := testMemberRoleID
	return &member.Member{
		ID:               "21234567-8901-2345-6789-012345678901",
		Email:            testMemberEmail,
		RoleID:           &roleID,
		InvitationStatus: member.StatusPending,
	}
}

func TestNewOrganizationMemberService(t *testing.T) {
	t.Parallel()

	svc, err := services.NewOrganizationMemberService(nil)
	assert.Nil(t, svc)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())

	svc, err = services.NewOrganizationMemberService(mocks_test.NewOrganizationMemberRepository(t))
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestOrganizationMemberServiceCreate(t *testing.T) {
	t.Parallel()

	validRequest := member.InviteRequest{Email: testMemberEmail, RoleID: testMemberRoleID}

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Create(context.Background(), "invalid", validRequest)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToInviteMember.Error())
	})

	t.Run("invalid request", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Create(context.Background(), testMemberOrganizationID, member.InviteRequest{Email: "nope", RoleID: testMemberRoleID})
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidInviteRequest.Error())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Create(mock.Anything, testMemberOrganizationID, validRequest).Return(nil, errors.New("api error"))
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Create(context.Background(), testMemberOrganizationID, validRequest)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToInviteMember.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		t.Parallel()
		expected := testDomainMember()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Create(mock.Anything, testMemberOrganizationID, validRequest).Return(expected, nil)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Create(context.Background(), testMemberOrganizationID, validRequest)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestOrganizationMemberServiceGet(t *testing.T) {
	t.Parallel()

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Get(context.Background(), "invalid", testMemberEmail)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToGetMember.Error())
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Get(context.Background(), testMemberOrganizationID, "nope")
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidEmailParam.Error())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Get(mock.Anything, testMemberOrganizationID, testMemberEmail).Return(nil, errors.New("api error"))
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Get(context.Background(), testMemberOrganizationID, testMemberEmail)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToGetMember.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		t.Parallel()
		expected := testDomainMember()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Get(mock.Anything, testMemberOrganizationID, testMemberEmail).Return(expected, nil)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Get(context.Background(), testMemberOrganizationID, testMemberEmail)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestOrganizationMemberServiceUpdate(t *testing.T) {
	t.Parallel()

	validRequest := member.UpdateRoleRequest{RoleID: testMemberRoleID}

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Update(context.Background(), "invalid", testMemberEmail, validRequest)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToUpdateMember.Error())
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Update(context.Background(), testMemberOrganizationID, "nope", validRequest)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidEmailParam.Error())
	})

	t.Run("invalid request", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Update(context.Background(), testMemberOrganizationID, testMemberEmail, member.UpdateRoleRequest{RoleID: "nope"})
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrInvalidUpdateRoleRequest.Error())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Update(mock.Anything, testMemberOrganizationID, testMemberEmail, validRequest).Return(nil, errors.New("api error"))
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Update(context.Background(), testMemberOrganizationID, testMemberEmail, validRequest)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, member.ErrFailedToUpdateMember.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		t.Parallel()
		expected := testDomainMember()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Update(mock.Anything, testMemberOrganizationID, testMemberEmail, validRequest).Return(expected, nil)
		svc, _ := services.NewOrganizationMemberService(repo)
		res, err := svc.Update(context.Background(), testMemberOrganizationID, testMemberEmail, validRequest)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestOrganizationMemberServiceDelete(t *testing.T) {
	t.Parallel()

	t.Run("invalid organization id", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		err := svc.Delete(context.Background(), "invalid", testMemberEmail)
		assert.ErrorContains(t, err, member.ErrFailedToDeleteMember.Error())
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		svc, _ := services.NewOrganizationMemberService(repo)
		err := svc.Delete(context.Background(), testMemberOrganizationID, "nope")
		assert.ErrorContains(t, err, member.ErrInvalidEmailParam.Error())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Delete(mock.Anything, testMemberOrganizationID, testMemberEmail).Return(errors.New("api error"))
		svc, _ := services.NewOrganizationMemberService(repo)
		err := svc.Delete(context.Background(), testMemberOrganizationID, testMemberEmail)
		assert.ErrorContains(t, err, member.ErrFailedToDeleteMember.Error())
	})

	t.Run("repository success", func(t *testing.T) {
		t.Parallel()
		repo := mocks_test.NewOrganizationMemberRepository(t)
		repo.EXPECT().Delete(mock.Anything, testMemberOrganizationID, testMemberEmail).Return(nil)
		svc, _ := services.NewOrganizationMemberService(repo)
		err := svc.Delete(context.Background(), testMemberOrganizationID, testMemberEmail)
		assert.NoError(t, err)
	})
}
