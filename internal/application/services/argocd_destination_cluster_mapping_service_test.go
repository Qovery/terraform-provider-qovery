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
	"github.com/qovery/terraform-provider-qovery/internal/domain/argoCdDestinationClusterMapping"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newArgoCdMappingUpsertRequest() argoCdDestinationClusterMapping.UpsertRequest {
	return argoCdDestinationClusterMapping.UpsertRequest{
		AgentClusterId:   gofakeit.UUID(),
		ArgocdClusterUrl: gofakeit.URL(),
		ClusterId:        gofakeit.UUID(),
	}
}

func newArgoCdMapping() *argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping {
	return &argoCdDestinationClusterMapping.ArgoCdDestinationClusterMapping{
		OrganizationID:   uuid.New(),
		AgentClusterID:   uuid.New(),
		ArgocdClusterUrl: gofakeit.URL(),
		ClusterID:        uuid.New(),
	}
}

func TestNewArgoCdDestinationClusterMappingService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Repository    argoCdDestinationClusterMapping.Repository
		ExpectedError error
	}{
		{TestName: "fail_with_nil_repository", Repository: nil, ExpectedError: services.ErrInvalidRepository},
		{TestName: "success", Repository: &mocks_test.ArgoCdDestinationClusterMappingRepository{}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			svc, err := services.NewArgoCdDestinationClusterMappingService(tc.Repository)
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

func TestArgoCdDestinationClusterMappingService_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		OrgID         string
		Request       argoCdDestinationClusterMapping.UpsertRequest
		SetupMock     func(*mocks_test.ArgoCdDestinationClusterMappingRepository)
		ErrorContains string
	}{
		{
			TestName:      "fail_with_empty_organization_id",
			OrgID:         "",
			Request:       newArgoCdMappingUpsertRequest(),
			SetupMock:     func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_organization_id",
			OrgID:         "not-a-uuid",
			Request:       newArgoCdMappingUpsertRequest(),
			SetupMock:     func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:      "fail_with_invalid_request",
			OrgID:         gofakeit.UUID(),
			Request:       argoCdDestinationClusterMapping.UpsertRequest{},
			SetupMock:     func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains: argoCdDestinationClusterMapping.ErrInvalidUpsertRequest.Error(),
		},
		{
			TestName:  "fail_with_repository_error",
			OrgID:     gofakeit.UUID(),
			Request:   newArgoCdMappingUpsertRequest(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: argoCdDestinationClusterMapping.ErrFailedToCreateArgoCdDestinationClusterMapping.Error(),
		},
		{
			TestName:  "success",
			OrgID:     gofakeit.UUID(),
			Request:   newArgoCdMappingUpsertRequest(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(newArgoCdMapping(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdDestinationClusterMappingRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdDestinationClusterMappingService(repo)
			assert.NoError(t, err)

			res, err := svc.Create(context.Background(), tc.OrgID, tc.Request)
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

func TestArgoCdDestinationClusterMappingService_Get(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrgID          string
		AgentClusterID string
		SetupMock      func(*mocks_test.ArgoCdDestinationClusterMappingRepository)
		ErrorContains  string
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrgID:          "not-a-uuid",
			AgentClusterID: gofakeit.UUID(),
			SetupMock:      func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains:  argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_agent_cluster_id",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: "not-a-uuid",
			SetupMock:      func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains:  argoCdDestinationClusterMapping.ErrInvalidAgentClusterIdParam.Error(),
		},
		{
			TestName:       "fail_with_repository_error",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("api error"))
			},
			ErrorContains: argoCdDestinationClusterMapping.ErrFailedToGetArgoCdDestinationClusterMapping.Error(),
		},
		{
			TestName:       "success",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(newArgoCdMapping(), nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdDestinationClusterMappingRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdDestinationClusterMappingService(repo)
			assert.NoError(t, err)

			res, err := svc.Get(context.Background(), tc.OrgID, tc.AgentClusterID, gofakeit.URL())
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

func TestArgoCdDestinationClusterMappingService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OrgID          string
		AgentClusterID string
		SetupMock      func(*mocks_test.ArgoCdDestinationClusterMappingRepository)
		ErrorContains  string
	}{
		{
			TestName:       "fail_with_invalid_organization_id",
			OrgID:          "not-a-uuid",
			AgentClusterID: gofakeit.UUID(),
			SetupMock:      func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains:  argoCdDestinationClusterMapping.ErrInvalidOrganizationIdParam.Error(),
		},
		{
			TestName:       "fail_with_invalid_agent_cluster_id",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: "not-a-uuid",
			SetupMock:      func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {},
			ErrorContains:  argoCdDestinationClusterMapping.ErrInvalidAgentClusterIdParam.Error(),
		},
		{
			TestName:       "fail_with_repository_error",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("api error"))
			},
			ErrorContains: argoCdDestinationClusterMapping.ErrFailedToDeleteArgoCdDestinationClusterMapping.Error(),
		},
		{
			TestName:       "success",
			OrgID:          gofakeit.UUID(),
			AgentClusterID: gofakeit.UUID(),
			SetupMock: func(m *mocks_test.ArgoCdDestinationClusterMappingRepository) {
				m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			repo := &mocks_test.ArgoCdDestinationClusterMappingRepository{}
			tc.SetupMock(repo)
			svc, err := services.NewArgoCdDestinationClusterMappingService(repo)
			assert.NoError(t, err)

			err = svc.Delete(context.Background(), tc.OrgID, tc.AgentClusterID, gofakeit.URL())
			if tc.ErrorContains != "" {
				assert.ErrorContains(t, err, tc.ErrorContains)
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}
