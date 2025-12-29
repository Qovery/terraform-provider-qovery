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

	appMocks "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	repoMocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewEnvironmentService(t *testing.T) {
	t.Parallel()

	validRepo := &repoMocks.EnvironmentRepository{}
	validDeploymentService := &appMocks.DeploymentService{}
	validVariableService := &appMocks.VariableService{}
	validSecretService := &appMocks.SecretService{}

	testCases := []struct {
		TestName          string
		Repository        environment.Repository
		DeploymentService deployment.Service
		VariableService   variable.Service
		SecretService     secret.Service
		ExpectError       bool
		ExpectedError     error
	}{
		{
			TestName:          "success_with_valid_dependencies",
			Repository:        validRepo,
			DeploymentService: validDeploymentService,
			VariableService:   validVariableService,
			SecretService:     validSecretService,
			ExpectError:       false,
		},
		{
			TestName:          "error_with_nil_repository",
			Repository:        nil,
			DeploymentService: validDeploymentService,
			VariableService:   validVariableService,
			SecretService:     validSecretService,
			ExpectError:       true,
			ExpectedError:     ErrInvalidRepository,
		},
		{
			TestName:          "error_with_nil_deployment_service",
			Repository:        validRepo,
			DeploymentService: nil,
			VariableService:   validVariableService,
			SecretService:     validSecretService,
			ExpectError:       true,
			ExpectedError:     ErrInvalidService,
		},
		{
			TestName:          "error_with_nil_variable_service",
			Repository:        validRepo,
			DeploymentService: validDeploymentService,
			VariableService:   nil,
			SecretService:     validSecretService,
			ExpectError:       true,
			ExpectedError:     ErrInvalidService,
		},
		{
			TestName:          "error_with_nil_secret_service",
			Repository:        validRepo,
			DeploymentService: validDeploymentService,
			VariableService:   validVariableService,
			SecretService:     nil,
			ExpectError:       true,
			ExpectedError:     ErrInvalidService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			service, err := NewEnvironmentService(tc.Repository, tc.DeploymentService, tc.VariableService, tc.SecretService)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tc.ExpectedError != nil {
					assert.Equal(t, tc.ExpectedError, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestEnvironmentService_Create(t *testing.T) {
	t.Parallel()

	validProjectID := gofakeit.UUID()
	validEnvironmentID := gofakeit.UUID()
	validClusterID := gofakeit.UUID()
	invalidProjectID := "invalid-uuid"
	emptyProjectID := ""

	modeDevelopment := environment.ModeDevelopment
	validRequest := environment.CreateServiceRequest{
		EnvironmentCreateRequest: environment.CreateRepositoryRequest{
			Name:      gofakeit.Word(),
			ClusterID: &validClusterID,
			Mode:      &modeDevelopment,
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
	}

	invalidRequest := environment.CreateServiceRequest{
		EnvironmentCreateRequest: environment.CreateRepositoryRequest{
			Name: "", // Invalid: empty name
		},
	}

	expectedEnvironment := &environment.Environment{
		ID:        uuid.MustParse(validEnvironmentID),
		ProjectID: uuid.MustParse(validProjectID),
		ClusterID: uuid.MustParse(validClusterID),
		Name:      validRequest.EnvironmentCreateRequest.Name,
		Mode:      modeDevelopment,
	}

	testCases := []struct {
		TestName      string
		ProjectID     string
		Request       environment.CreateServiceRequest
		SetupMocks    func(*repoMocks.EnvironmentRepository, *appMocks.VariableService, *appMocks.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_project_id",
			ProjectID:     emptyProjectID,
			Request:       validRequest,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:      "error_invalid_project_id",
			ProjectID:     invalidProjectID,
			Request:       validRequest,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:      "error_invalid_request",
			ProjectID:     validProjectID,
			Request:       invalidRequest,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "failed to create environment",
		},
		{
			TestName:  "error_repository_create_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Create(mock.Anything, validProjectID, validRequest.EnvironmentCreateRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create environment",
		},
		{
			TestName:  "error_variable_service_update_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Create(mock.Anything, validProjectID, validRequest.EnvironmentCreateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create environment",
		},
		{
			TestName:  "error_secret_service_update_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Create(mock.Anything, validProjectID, validRequest.EnvironmentCreateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create environment",
		},
		{
			TestName:  "error_variable_service_list_failure_in_refresh",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Create(mock.Anything, validProjectID, validRequest.EnvironmentCreateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, errors.New("list variables error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create environment",
		},
		{
			TestName:  "success",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Create(mock.Anything, validProjectID, validRequest.EnvironmentCreateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &repoMocks.EnvironmentRepository{}
			mockDeploymentService := &appMocks.DeploymentService{}
			mockVariableService := &appMocks.VariableService{}
			mockSecretService := &appMocks.SecretService{}

			tc.SetupMocks(mockRepo, mockVariableService, mockSecretService)

			service, err := NewEnvironmentService(mockRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.ProjectID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedEnvironment.ID, result.ID)
				assert.Equal(t, expectedEnvironment.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestEnvironmentService_Get(t *testing.T) {
	t.Parallel()

	validEnvironmentID := gofakeit.UUID()
	validProjectID := gofakeit.UUID()
	validClusterID := gofakeit.UUID()
	invalidEnvironmentID := "invalid-uuid"
	emptyEnvironmentID := ""

	expectedEnvironment := &environment.Environment{
		ID:        uuid.MustParse(validEnvironmentID),
		ProjectID: uuid.MustParse(validProjectID),
		ClusterID: uuid.MustParse(validClusterID),
		Name:      gofakeit.Word(),
		Mode:      environment.ModeDevelopment,
	}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		SetupMocks    func(*repoMocks.EnvironmentRepository, *appMocks.VariableService, *appMocks.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvironmentID,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvironmentID,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_repository_get_failure",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Get(mock.Anything, validEnvironmentID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get environment",
		},
		{
			TestName:      "error_variable_service_list_failure",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Get(mock.Anything, validEnvironmentID).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get environment",
		},
		{
			TestName:      "error_secret_service_list_failure",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Get(mock.Anything, validEnvironmentID).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(secret.Secrets{}, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get environment",
		},
		{
			TestName:      "success",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Get(mock.Anything, validEnvironmentID).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &repoMocks.EnvironmentRepository{}
			mockDeploymentService := &appMocks.DeploymentService{}
			mockVariableService := &appMocks.VariableService{}
			mockSecretService := &appMocks.SecretService{}

			tc.SetupMocks(mockRepo, mockVariableService, mockSecretService)

			service, err := NewEnvironmentService(mockRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.EnvironmentID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedEnvironment.ID, result.ID)
				assert.Equal(t, expectedEnvironment.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestEnvironmentService_Update(t *testing.T) {
	t.Parallel()

	validEnvironmentID := gofakeit.UUID()
	validProjectID := gofakeit.UUID()
	validClusterID := gofakeit.UUID()
	invalidEnvironmentID := "invalid-uuid"
	emptyEnvironmentID := ""

	updatedName := gofakeit.Word()
	validRequest := environment.UpdateServiceRequest{
		EnvironmentUpdateRequest: environment.UpdateRepositoryRequest{
			Name: &updatedName,
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
	}

	requestWithVariables := environment.UpdateServiceRequest{
		EnvironmentUpdateRequest: environment.UpdateRepositoryRequest{
			Name: &updatedName,
		},
		EnvironmentVariables: variable.DiffRequest{
			Create: []variable.DiffCreateRequest{
				{UpsertRequest: variable.UpsertRequest{Key: "TEST_VAR", Value: "value"}},
			},
		},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
	}

	expectedEnvironment := &environment.Environment{
		ID:        uuid.MustParse(validEnvironmentID),
		ProjectID: uuid.MustParse(validProjectID),
		ClusterID: uuid.MustParse(validClusterID),
		Name:      updatedName,
		Mode:      environment.ModeDevelopment,
	}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		Request       environment.UpdateServiceRequest
		SetupMocks    func(*repoMocks.EnvironmentRepository, *appMocks.DeploymentService, *appMocks.VariableService, *appMocks.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_repository_update_failure",
			EnvironmentID: validEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, validRequest.EnvironmentUpdateRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update environment",
		},
		{
			TestName:      "error_variable_service_update_failure",
			EnvironmentID: validEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, validRequest.EnvironmentUpdateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update environment",
		},
		{
			TestName:      "error_secret_service_update_failure",
			EnvironmentID: validEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, validRequest.EnvironmentUpdateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update environment",
		},
		{
			TestName:      "error_redeploy_failure_when_variables_present",
			EnvironmentID: validEnvironmentID,
			Request:       requestWithVariables,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, requestWithVariables.EnvironmentUpdateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, nil)
				d.EXPECT().
					Redeploy(mock.Anything, validEnvironmentID).
					Return(nil, errors.New("redeploy error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update environment",
		},
		{
			TestName:      "success_without_redeploy",
			EnvironmentID: validEnvironmentID,
			Request:       validRequest,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, validRequest.EnvironmentUpdateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
		{
			TestName:      "success_with_redeploy",
			EnvironmentID: validEnvironmentID,
			Request:       requestWithVariables,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService, v *appMocks.VariableService, s *appMocks.SecretService) {
				r.EXPECT().
					Update(mock.Anything, validEnvironmentID, requestWithVariables.EnvironmentUpdateRequest).
					Return(expectedEnvironment, nil)
				v.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					Update(mock.Anything, validEnvironmentID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(secret.Secrets{}, nil)
				d.EXPECT().
					Redeploy(mock.Anything, validEnvironmentID).
					Return(&status.Status{}, nil)
				v.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(variable.Variables{}, nil)
				s.EXPECT().
					List(mock.Anything, validEnvironmentID).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &repoMocks.EnvironmentRepository{}
			mockDeploymentService := &appMocks.DeploymentService{}
			mockVariableService := &appMocks.VariableService{}
			mockSecretService := &appMocks.SecretService{}

			tc.SetupMocks(mockRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewEnvironmentService(mockRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.EnvironmentID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedEnvironment.ID, result.ID)
			}

			mockRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestEnvironmentService_Delete(t *testing.T) {
	t.Parallel()

	validEnvironmentID := gofakeit.UUID()
	invalidEnvironmentID := "invalid-uuid"
	emptyEnvironmentID := ""

	testCases := []struct {
		TestName      string
		EnvironmentID string
		SetupMocks    func(*repoMocks.EnvironmentRepository, *appMocks.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvironmentID,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvironmentID,
			SetupMocks:    func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService) {},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "success_environment_not_found",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService) {
				r.EXPECT().
					Exists(mock.Anything, validEnvironmentID).
					Return(false)
			},
			ExpectError: false,
		},
		{
			TestName:      "error_repository_delete_failure",
			EnvironmentID: validEnvironmentID,
			SetupMocks: func(r *repoMocks.EnvironmentRepository, d *appMocks.DeploymentService) {
				r.EXPECT().
					Exists(mock.Anything, validEnvironmentID).
					Return(true)
				d.EXPECT().
					GetStatus(mock.Anything, validEnvironmentID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete environment",
		},
		// Note: The success case for Delete is complex to test in a unit test because it involves
		// wait() calls that depend on internal implementation details (polling GetStatus).
		// The Delete logic is better tested through integration tests where the actual
		// deployment service behavior can be verified.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockRepo := &repoMocks.EnvironmentRepository{}
			mockDeploymentService := &appMocks.DeploymentService{}
			mockVariableService := &appMocks.VariableService{}
			mockSecretService := &appMocks.SecretService{}

			tc.SetupMocks(mockRepo, mockDeploymentService)

			service, err := NewEnvironmentService(mockRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.EnvironmentID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}
