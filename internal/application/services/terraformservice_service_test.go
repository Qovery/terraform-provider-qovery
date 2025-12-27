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

	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	"github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewTerraformServiceService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		Repository  terraformservice.Repository
		ExpectError bool
	}{
		{
			TestName:    "success_with_valid_repository",
			Repository:  &mocks_test.TerraformServiceRepository{},
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
			service, err := NewTerraformServiceService(tc.Repository)
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

func TestTerraformServiceService_Create(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validTerraformServiceID := gofakeit.UUID()
	invalidEnvID := "invalid-uuid"
	emptyEnvID := ""

	validRequest := terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: terraformservice.UpsertRepositoryRequest{
			Name:       gofakeit.Word(),
			AutoDeploy: true,
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/example/repo",
				Branch:   "main",
				RootPath: "/",
			},
			Backend: terraformservice.Backend{
				Kubernetes: &terraformservice.KubernetesBackend{},
			},
			Engine: terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{
				ExplicitVersion:        "1.5.0",
				ReadFromTerraformBlock: false,
			},
			JobResources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
		},
	}

	invalidRequest := terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: terraformservice.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedResult := &terraformservice.TerraformService{
		ID:            uuid.MustParse(validTerraformServiceID),
		EnvironmentID: uuid.MustParse(validEnvID),
		Name:          validRequest.TerraformServiceUpsertRequest.Name,
		AutoDeploy:    validRequest.TerraformServiceUpsertRequest.AutoDeploy,
		GitRepository: validRequest.TerraformServiceUpsertRequest.GitRepository,
		Backend:       validRequest.TerraformServiceUpsertRequest.Backend,
		Engine:        validRequest.TerraformServiceUpsertRequest.Engine,
		EngineVersion: validRequest.TerraformServiceUpsertRequest.EngineVersion,
		JobResources:  validRequest.TerraformServiceUpsertRequest.JobResources,
	}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		Request       terraformservice.UpsertServiceRequest
		SetupMock     func(*mocks_test.TerraformServiceRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvID,
			Request:       validRequest,
			SetupMock:     func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvID,
			Request:       validRequest,
			SetupMock:     func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_request",
			EnvironmentID: validEnvID,
			Request:       invalidRequest,
			SetupMock:     func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:   true,
			ErrorContains: "failed to create terraform service",
		},
		{
			TestName:      "error_repository_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.TerraformServiceUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create terraform service",
		},
		{
			TestName:      "success",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.TerraformServiceUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.TerraformServiceRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewTerraformServiceService(mockRepo)
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

func TestTerraformServiceService_Get(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validTerraformServiceID := gofakeit.UUID()
	invalidTerraformServiceID := "invalid-uuid"
	emptyTerraformServiceID := ""
	advancedSettingsJson := `{"key": "value"}`

	expectedResult := &terraformservice.TerraformService{
		ID:                   uuid.MustParse(validTerraformServiceID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		Name:                 gofakeit.Word(),
		AutoDeploy:           true,
		AdvancedSettingsJson: advancedSettingsJson,
		GitRepository: terraformservice.GitRepository{
			URL:      "https://github.com/example/repo",
			Branch:   "main",
			RootPath: "/",
		},
		Backend: terraformservice.Backend{
			Kubernetes: &terraformservice.KubernetesBackend{},
		},
		Engine: terraformservice.EngineTerraform,
		EngineVersion: terraformservice.EngineVersion{
			ExplicitVersion:        "1.5.0",
			ReadFromTerraformBlock: false,
		},
		JobResources: terraformservice.JobResources{
			CPUMilli:   1000,
			RAMMiB:     1024,
			GPU:        0,
			StorageGiB: 20,
		},
	}

	testCases := []struct {
		TestName              string
		TerraformServiceID    string
		AdvancedSettingsJson  string
		IsTriggeredFromImport bool
		SetupMock             func(*mocks_test.TerraformServiceRepository)
		ExpectError           bool
		ErrorContains         string
	}{
		{
			TestName:              "error_empty_terraform_service_id",
			TerraformServiceID:    emptyTerraformServiceID,
			AdvancedSettingsJson:  advancedSettingsJson,
			IsTriggeredFromImport: false,
			SetupMock:             func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:           true,
			ErrorContains:         "invalid terraform service id param",
		},
		{
			TestName:              "error_invalid_terraform_service_id",
			TerraformServiceID:    invalidTerraformServiceID,
			AdvancedSettingsJson:  advancedSettingsJson,
			IsTriggeredFromImport: false,
			SetupMock:             func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:           true,
			ErrorContains:         "invalid terraform service id param",
		},
		{
			TestName:              "error_repository_failure",
			TerraformServiceID:    validTerraformServiceID,
			AdvancedSettingsJson:  advancedSettingsJson,
			IsTriggeredFromImport: false,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Get(mock.Anything, validTerraformServiceID, advancedSettingsJson, false).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get terraform service",
		},
		{
			TestName:              "success",
			TerraformServiceID:    validTerraformServiceID,
			AdvancedSettingsJson:  advancedSettingsJson,
			IsTriggeredFromImport: false,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Get(mock.Anything, validTerraformServiceID, advancedSettingsJson, false).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
		{
			TestName:              "success_with_import",
			TerraformServiceID:    validTerraformServiceID,
			AdvancedSettingsJson:  "",
			IsTriggeredFromImport: true,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Get(mock.Anything, validTerraformServiceID, "", true).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.TerraformServiceRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewTerraformServiceService(mockRepo)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.TerraformServiceID, tc.AdvancedSettingsJson, tc.IsTriggeredFromImport)

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

func TestTerraformServiceService_Update(t *testing.T) {
	t.Parallel()

	validTerraformServiceID := gofakeit.UUID()
	invalidTerraformServiceID := "invalid-uuid"
	emptyTerraformServiceID := ""

	validRequest := terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: terraformservice.UpsertRepositoryRequest{
			Name:       gofakeit.Word(),
			AutoDeploy: false,
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/example/repo",
				Branch:   "develop",
				RootPath: "/terraform",
			},
			Backend: terraformservice.Backend{
				UserProvided: &terraformservice.UserProvidedBackend{},
			},
			Engine: terraformservice.EngineOpenTofu,
			EngineVersion: terraformservice.EngineVersion{
				ExplicitVersion:        "1.6.0",
				ReadFromTerraformBlock: true,
			},
			JobResources: terraformservice.JobResources{
				CPUMilli:   2000,
				RAMMiB:     2048,
				GPU:        0,
				StorageGiB: 30,
			},
		},
	}

	invalidRequest := terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: terraformservice.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedResult := &terraformservice.TerraformService{
		ID:            uuid.MustParse(validTerraformServiceID),
		Name:          validRequest.TerraformServiceUpsertRequest.Name,
		AutoDeploy:    validRequest.TerraformServiceUpsertRequest.AutoDeploy,
		GitRepository: validRequest.TerraformServiceUpsertRequest.GitRepository,
		Backend:       validRequest.TerraformServiceUpsertRequest.Backend,
		Engine:        validRequest.TerraformServiceUpsertRequest.Engine,
		EngineVersion: validRequest.TerraformServiceUpsertRequest.EngineVersion,
		JobResources:  validRequest.TerraformServiceUpsertRequest.JobResources,
	}

	testCases := []struct {
		TestName           string
		TerraformServiceID string
		Request            terraformservice.UpsertServiceRequest
		SetupMock          func(*mocks_test.TerraformServiceRepository)
		ExpectError        bool
		ErrorContains      string
	}{
		{
			TestName:           "error_empty_terraform_service_id",
			TerraformServiceID: emptyTerraformServiceID,
			Request:            validRequest,
			SetupMock:          func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid terraform service id param",
		},
		{
			TestName:           "error_invalid_terraform_service_id",
			TerraformServiceID: invalidTerraformServiceID,
			Request:            validRequest,
			SetupMock:          func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid terraform service id param",
		},
		{
			TestName:           "error_invalid_request",
			TerraformServiceID: validTerraformServiceID,
			Request:            invalidRequest,
			SetupMock:          func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:        true,
			ErrorContains:      "failed to update terraform service",
		},
		{
			TestName:           "error_repository_failure",
			TerraformServiceID: validTerraformServiceID,
			Request:            validRequest,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Update(mock.Anything, validTerraformServiceID, validRequest.TerraformServiceUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update terraform service",
		},
		{
			TestName:           "success",
			TerraformServiceID: validTerraformServiceID,
			Request:            validRequest,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Update(mock.Anything, validTerraformServiceID, validRequest.TerraformServiceUpsertRequest).
					Return(expectedResult, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.TerraformServiceRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewTerraformServiceService(mockRepo)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.TerraformServiceID, tc.Request)

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

func TestTerraformServiceService_Delete(t *testing.T) {
	t.Parallel()

	validTerraformServiceID := gofakeit.UUID()
	invalidTerraformServiceID := "invalid-uuid"
	emptyTerraformServiceID := ""

	testCases := []struct {
		TestName           string
		TerraformServiceID string
		SetupMock          func(*mocks_test.TerraformServiceRepository)
		ExpectError        bool
		ErrorContains      string
	}{
		{
			TestName:           "error_empty_terraform_service_id",
			TerraformServiceID: emptyTerraformServiceID,
			SetupMock:          func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid terraform service id param",
		},
		{
			TestName:           "error_invalid_terraform_service_id",
			TerraformServiceID: invalidTerraformServiceID,
			SetupMock:          func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:        true,
			ErrorContains:      "invalid terraform service id param",
		},
		{
			TestName:           "error_repository_failure",
			TerraformServiceID: validTerraformServiceID,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Delete(mock.Anything, validTerraformServiceID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete terraform service",
		},
		{
			TestName:           "success",
			TerraformServiceID: validTerraformServiceID,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					Delete(mock.Anything, validTerraformServiceID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.TerraformServiceRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewTerraformServiceService(mockRepo)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.TerraformServiceID)

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

func TestTerraformServiceService_List(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	invalidEnvID := "invalid-uuid"
	emptyEnvID := ""

	terraformServiceID1 := uuid.New()
	terraformServiceID2 := uuid.New()

	expectedResult := []terraformservice.TerraformService{
		{
			ID:            terraformServiceID1,
			EnvironmentID: uuid.MustParse(validEnvID),
			Name:          gofakeit.Word(),
			AutoDeploy:    true,
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/example/repo1",
				Branch:   "main",
				RootPath: "/",
			},
			Backend: terraformservice.Backend{
				Kubernetes: &terraformservice.KubernetesBackend{},
			},
			Engine: terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{
				ExplicitVersion:        "1.5.0",
				ReadFromTerraformBlock: false,
			},
			JobResources: terraformservice.JobResources{
				CPUMilli:   1000,
				RAMMiB:     1024,
				GPU:        0,
				StorageGiB: 20,
			},
		},
		{
			ID:            terraformServiceID2,
			EnvironmentID: uuid.MustParse(validEnvID),
			Name:          gofakeit.Word(),
			AutoDeploy:    false,
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/example/repo2",
				Branch:   "develop",
				RootPath: "/terraform",
			},
			Backend: terraformservice.Backend{
				UserProvided: &terraformservice.UserProvidedBackend{},
			},
			Engine: terraformservice.EngineOpenTofu,
			EngineVersion: terraformservice.EngineVersion{
				ExplicitVersion:        "1.6.0",
				ReadFromTerraformBlock: true,
			},
			JobResources: terraformservice.JobResources{
				CPUMilli:   2000,
				RAMMiB:     2048,
				GPU:        0,
				StorageGiB: 30,
			},
		},
	}

	emptyResult := []terraformservice.TerraformService{}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		SetupMock     func(*mocks_test.TerraformServiceRepository)
		ExpectError   bool
		ErrorContains string
		ExpectedCount int
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvID,
			SetupMock:     func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvID,
			SetupMock:     func(m *mocks_test.TerraformServiceRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_repository_failure",
			EnvironmentID: validEnvID,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					List(mock.Anything, validEnvID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to list terraform services",
		},
		{
			TestName:      "success_with_results",
			EnvironmentID: validEnvID,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					List(mock.Anything, validEnvID).
					Return(expectedResult, nil)
			},
			ExpectError:   false,
			ExpectedCount: 2,
		},
		{
			TestName:      "success_with_empty_results",
			EnvironmentID: validEnvID,
			SetupMock: func(m *mocks_test.TerraformServiceRepository) {
				m.EXPECT().
					List(mock.Anything, validEnvID).
					Return(emptyResult, nil)
			},
			ExpectError:   false,
			ExpectedCount: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mocks_test.TerraformServiceRepository{}
			tc.SetupMock(mockRepo)

			service, err := NewTerraformServiceService(mockRepo)
			require.NoError(t, err)

			result, err := service.List(context.Background(), tc.EnvironmentID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tc.ExpectedCount)
				if tc.ExpectedCount > 0 {
					assert.Equal(t, expectedResult[0].ID, result[0].ID)
					assert.Equal(t, expectedResult[0].Name, result[0].Name)
					if tc.ExpectedCount > 1 {
						assert.Equal(t, expectedResult[1].ID, result[1].ID)
						assert.Equal(t, expectedResult[1].Name, result[1].Name)
					}
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
