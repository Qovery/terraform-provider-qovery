//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewDeploymentStageService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  deploymentstage.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.DeploymentStageRepository{},
			ExpectError: false,
		},
		{
			TestName:    "error_with_nil_repository",
			Repository:  nil,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewDeploymentStageService(tc.Repository)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				assert.Equal(t, ErrInvalidRepository, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestDeploymentStageService_Create(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validDeploymentStageID := gofakeit.UUID()

	validRequest := deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        gofakeit.Word(),
			Description: gofakeit.Sentence(5),
		},
	}

	invalidRequest := deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedResult := &deploymentstage.DeploymentStage{
		ID:            uuid.MustParse(validDeploymentStageID),
		EnvironmentID: uuid.MustParse(validEnvID),
		Name:          validRequest.DeploymentStageUpsertRequest.Name,
		Description:   validRequest.DeploymentStageUpsertRequest.Description,
	}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		Request       deploymentstage.UpsertServiceRequest
		SetupMock     func(*mocks_test.DeploymentStageRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_invalid_request",
			EnvironmentID: validEnvID,
			Request:       invalidRequest,
			SetupMock:     func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:   true,
			ErrorContains: "failed to create deployment stage",
		},
		{
			TestName:      "error_repository_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.DeploymentStageUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create deployment stage",
		},
		{
			TestName:      "success",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.DeploymentStageUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentStageRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentStageService(mockRepo)
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.EnvironmentID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentStageService_Get(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validDeploymentStageID := gofakeit.UUID()
	invalidDeploymentStageID := "invalid-uuid"
	emptyDeploymentStageID := ""

	expectedResult := &deploymentstage.DeploymentStage{
		ID:            uuid.MustParse(validDeploymentStageID),
		EnvironmentID: uuid.MustParse(validEnvID),
		Name:          gofakeit.Word(),
		Description:   gofakeit.Sentence(5),
	}

	testCases := []struct {
		TestName           string
		EnvironmentID      string
		DeploymentStageID  string
		SetupMock          func(*mocks_test.DeploymentStageRepository)
		ExpectError        bool
		ErrorContains      string
	}{
		{
			TestName:           "error_empty_deployment_stage_id",
			EnvironmentID:      validEnvID,
			DeploymentStageID:  emptyDeploymentStageID,
			SetupMock:          func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid deployment stage ID",
		},
		{
			TestName:           "error_invalid_deployment_stage_id",
			EnvironmentID:      validEnvID,
			DeploymentStageID:  invalidDeploymentStageID,
			SetupMock:          func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid deployment stage ID",
		},
		{
			TestName:           "error_repository_failure",
			EnvironmentID:      validEnvID,
			DeploymentStageID:  validDeploymentStageID,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Get(mock.Anything, validEnvID, validDeploymentStageID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get deployment stage",
		},
		{
			TestName:           "success",
			EnvironmentID:      validEnvID,
			DeploymentStageID:  validDeploymentStageID,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Get(mock.Anything, validEnvID, validDeploymentStageID).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentStageRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentStageService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.EnvironmentID, tc.DeploymentStageID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentStageService_GetAllByEnvironmentID(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	deploymentStageName := "production"
	deploymentStageID1 := uuid.New()
	deploymentStageID2 := uuid.New()

	deploymentStages := &[]deploymentstage.DeploymentStage{
		{
			ID:            deploymentStageID1,
			EnvironmentID: uuid.MustParse(validEnvID),
			Name:          "staging",
			Description:   "Staging environment",
		},
		{
			ID:            deploymentStageID2,
			EnvironmentID: uuid.MustParse(validEnvID),
			Name:          deploymentStageName,
			Description:   "Production environment",
		},
	}

	emptyDeploymentStages := &[]deploymentstage.DeploymentStage{}

	testCases := []struct {
		TestName            string
		EnvironmentID       string
		DeploymentStageName string
		SetupMock           func(*mocks_test.DeploymentStageRepository)
		ExpectError         bool
		ErrorContains       string
		ExpectedID          *uuid.UUID
	}{
		{
			TestName:            "error_repository_failure",
			EnvironmentID:       validEnvID,
			DeploymentStageName: deploymentStageName,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					GetAllByEnvironmentID(mock.Anything, validEnvID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get deployment stage",
		},
		{
			TestName:            "error_deployment_stage_not_found",
			EnvironmentID:       validEnvID,
			DeploymentStageName: "nonexistent",
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					GetAllByEnvironmentID(mock.Anything, validEnvID).
					Return(deploymentStages, nil)
			},
			ExpectError:   true,
			ErrorContains: "Cannot find deployment stage with name",
		},
		{
			TestName:            "error_empty_list",
			EnvironmentID:       validEnvID,
			DeploymentStageName: deploymentStageName,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					GetAllByEnvironmentID(mock.Anything, validEnvID).
					Return(emptyDeploymentStages, nil)
			},
			ExpectError:   true,
			ErrorContains: "Cannot find deployment stage with name",
		},
		{
			TestName:            "success",
			EnvironmentID:       validEnvID,
			DeploymentStageName: deploymentStageName,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					GetAllByEnvironmentID(mock.Anything, validEnvID).
					Return(deploymentStages, nil)
			},
			ExpectError: false,
			ExpectedID:  &deploymentStageID2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentStageRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentStageService(mockRepo)
			require.NoError(t, err)

			result, err := service.GetAllByEnvironmentID(context.Background(), tc.EnvironmentID, tc.DeploymentStageName)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.ExpectedID != nil {
					assert.Equal(t, *tc.ExpectedID, result.ID)
					assert.Equal(t, tc.DeploymentStageName, result.Name)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentStageService_Update(t *testing.T) {
	t.Parallel()

	validDeploymentStageID := gofakeit.UUID()
	invalidDeploymentStageID := "invalid-uuid"
	emptyDeploymentStageID := ""

	validRequest := deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name:        gofakeit.Word(),
			Description: gofakeit.Sentence(5),
		},
	}

	invalidRequest := deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedResult := &deploymentstage.DeploymentStage{
		ID:          uuid.MustParse(validDeploymentStageID),
		Name:        validRequest.DeploymentStageUpsertRequest.Name,
		Description: validRequest.DeploymentStageUpsertRequest.Description,
	}

	testCases := []struct {
		TestName          string
		DeploymentStageID string
		Request           deploymentstage.UpsertServiceRequest
		SetupMock         func(*mocks_test.DeploymentStageRepository)
		ExpectError       bool
		ErrorContains     string
	}{
		{
			TestName:          "error_invalid_request",
			DeploymentStageID: validDeploymentStageID,
			Request:           invalidRequest,
			SetupMock:         func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:       true,
			ErrorContains:     "failed to update deployment stage",
		},
		{
			TestName:          "error_empty_deployment_stage_id",
			DeploymentStageID: emptyDeploymentStageID,
			Request:           validRequest,
			SetupMock:         func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:       true,
			ErrorContains:     "invalid deployment stage ID",
		},
		{
			TestName:          "error_invalid_deployment_stage_id",
			DeploymentStageID: invalidDeploymentStageID,
			Request:           validRequest,
			SetupMock:         func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:       true,
			ErrorContains:     "invalid deployment stage ID",
		},
		{
			TestName:          "error_repository_failure",
			DeploymentStageID: validDeploymentStageID,
			Request:           validRequest,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Update(mock.Anything, validDeploymentStageID, validRequest.DeploymentStageUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create deployment stage",
		},
		{
			TestName:          "success",
			DeploymentStageID: validDeploymentStageID,
			Request:           validRequest,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Update(mock.Anything, validDeploymentStageID, validRequest.DeploymentStageUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentStageRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentStageService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.DeploymentStageID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedResult.ID, result.ID)
				assert.Equal(t, expectedResult.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeploymentStageService_Delete(t *testing.T) {
	t.Parallel()

	validDeploymentStageID := gofakeit.UUID()
	invalidDeploymentStageID := "invalid-uuid"
	emptyDeploymentStageID := ""

	testCases := []struct {
		TestName          string
		DeploymentStageID string
		SetupMock         func(*mocks_test.DeploymentStageRepository)
		ExpectError       bool
		ErrorContains     string
	}{
		{
			TestName:          "error_empty_deployment_stage_id",
			DeploymentStageID: emptyDeploymentStageID,
			SetupMock:         func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:       true,
			ErrorContains:     "invalid deployment stage ID",
		},
		{
			TestName:          "error_invalid_deployment_stage_id",
			DeploymentStageID: invalidDeploymentStageID,
			SetupMock:         func(m *mocks_test.DeploymentStageRepository) {},
			ExpectError:       true,
			ErrorContains:     "invalid deployment stage ID",
		},
		{
			TestName:          "error_repository_failure",
			DeploymentStageID: validDeploymentStageID,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Delete(mock.Anything, validDeploymentStageID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete deployment stage",
		},
		{
			TestName:          "success",
			DeploymentStageID: validDeploymentStageID,
			SetupMock: func(m *mocks_test.DeploymentStageRepository) {
				m.EXPECT().
					Delete(mock.Anything, validDeploymentStageID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.DeploymentStageRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewDeploymentStageService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.DeploymentStageID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
