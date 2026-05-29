//go:build unit && !integration

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdCredentials"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newArgoCdCredentialsUpsertRequest() argoCdCredentials.UpsertRequest {
	return argoCdCredentials.UpsertRequest{
		ArgocdUrl:   gofakeit.URL(),
		ArgocdToken: gofakeit.UUID(),
	}
}

func newArgoCdCredentials() *argoCdCredentials.ArgoCdCredentials {
	return &argoCdCredentials.ArgoCdCredentials{
		ID:          uuid.New(),
		ClusterID:   uuid.New(),
		ArgocdUrl:   gofakeit.URL(),
		ArgocdToken: gofakeit.UUID(),
	}
}

func TestNewArgoCdCredentialsService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    argoCdCredentials.Repository
		ExpectedError error
	}{
		{TestName: "fail_with_nil_repository", Repository: nil, ExpectedError: services.ErrInvalidRepository},
		{TestName: "success", Repository: &mocks_test.ArgoCdCredentialsRepository{}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			svc, err := services.NewArgoCdCredentialsService(tc.Repository)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, svc)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, svc)
		})
	}
}

func TestArgoCdCredentialsService_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ClusterID     string
		Request       argoCdCredentials.UpsertRequest
		SetupMock     func(*mocks_test.ArgoCdCredentialsRepository)
		ErrorContains string
	}{
		{
			TestName:      "fail_with_empty_cluster_id",
			ClusterID:     "",
			Request:       newArgoCdCredentialsUpsertRequest(),
			SetupMock:     func(m *mocks_test.ArgoCdCredentialsRepository) {},
			ErrorContains: argoCdCredentials.ErrInvalidClusterIdParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_cluster_id",
			ClusterID:     "not-a-uuid",
			Request:       newArgoCdCredentialsUpsertRequest(),
			SetupMock:     func(m *mocks_test.ArgoCdCredentialsRepository) {},
			ErrorContains: argoCdCredentials.ErrInvalidClusterIdParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_request",
			ClusterID:     gofakeit.UUID(),
			Request:       argoCdCredentials.UpsertRequest{},
			SetupMock:     func(m *mocks_test.ArgoCdCredentialsRepository) {},
			ErrorContains: argoCdCredentials.ErrInvalidUpsertRequest.Error(),
		},
		{
			TestName:  "fail_with_repository_error",
			ClusterID: gofakeit.UUID(),
			Request:   newArgoCdCredentialsUpsertRequest(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: argoCdCredentials.ErrFailedToCreateArgoCdCredentials.Error(),
		},
		{
			TestName:  "success",
			ClusterID: gofakeit.UUID(),
			Request:   newArgoCdCredentialsUpsertRequest(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(newArgoCdCredentials(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdCredentialsRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdCredentialsService(repo)
			assert.NoError(t, err)

			res, err := svc.Create(context.Background(), tc.ClusterID, tc.Request)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			repo.AssertExpectations(t)
		})
	}
}

func TestArgoCdCredentialsService_Get(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ClusterID     string
		SetupMock     func(*mocks_test.ArgoCdCredentialsRepository)
		ErrorContains string
	}{
		{
			TestName:      "fail_with_invalid_cluster_id",
			ClusterID:     "not-a-uuid",
			SetupMock:     func(m *mocks_test.ArgoCdCredentialsRepository) {},
			ErrorContains: argoCdCredentials.ErrInvalidClusterIdParam.Error(),
		},
		{
			TestName:  "fail_with_repository_error",
			ClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: argoCdCredentials.ErrFailedToGetArgoCdCredentials.Error(),
		},
		{
			TestName:  "success",
			ClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything).Return(newArgoCdCredentials(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdCredentialsRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdCredentialsService(repo)
			assert.NoError(t, err)

			res, err := svc.Get(context.Background(), tc.ClusterID)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, res)
			repo.AssertExpectations(t)
		})
	}
}

func TestArgoCdCredentialsService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ClusterID     string
		SetupMock     func(*mocks_test.ArgoCdCredentialsRepository)
		ErrorContains string
	}{
		{
			TestName:      "fail_with_invalid_cluster_id",
			ClusterID:     "not-a-uuid",
			SetupMock:     func(m *mocks_test.ArgoCdCredentialsRepository) {},
			ErrorContains: argoCdCredentials.ErrInvalidClusterIdParam.Error(),
		},
		{
			TestName:  "fail_with_repository_error",
			ClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.New("api error"))
			},
			ErrorContains: argoCdCredentials.ErrFailedToDeleteArgoCdCredentials.Error(),
		},
		{
			TestName:  "success",
			ClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdCredentialsRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdCredentialsRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdCredentialsService(repo)
			assert.NoError(t, err)

			err = svc.Delete(context.Background(), tc.ClusterID)
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}
