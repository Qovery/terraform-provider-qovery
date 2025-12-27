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

	"github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	repoMocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func TestNewContainerService(t *testing.T) {
	t.Parallel()

	mockContainerRepo := &repoMocks.ContainerRepository{}
	mockDeploymentService := &mocks_test.DeploymentService{}
	mockVariableService := &mocks_test.VariableService{}
	mockSecretService := &mocks_test.SecretService{}

	testCases := []struct {
		TestName                   string
		ContainerRepository        container.Repository
		ContainerDeploymentService deployment.Service
		VariableService            variable.Service
		SecretService              secret.Service
		ExpectError                bool
		ExpectedError              error
	}{
		{
			TestName:                   "success_with_all_valid_dependencies",
			ContainerRepository:        mockContainerRepo,
			ContainerDeploymentService: mockDeploymentService,
			VariableService:            mockVariableService,
			SecretService:              mockSecretService,
			ExpectError:                false,
		},
		{
			TestName:                   "error_with_nil_container_repository",
			ContainerRepository:        nil,
			ContainerDeploymentService: mockDeploymentService,
			VariableService:            mockVariableService,
			SecretService:              mockSecretService,
			ExpectError:                true,
			ExpectedError:              ErrInvalidRepository,
		},
		{
			TestName:                   "error_with_nil_deployment_service",
			ContainerRepository:        mockContainerRepo,
			ContainerDeploymentService: nil,
			VariableService:            mockVariableService,
			SecretService:              mockSecretService,
			ExpectError:                true,
			ExpectedError:              ErrInvalidService,
		},
		{
			TestName:                   "error_with_nil_variable_service",
			ContainerRepository:        mockContainerRepo,
			ContainerDeploymentService: mockDeploymentService,
			VariableService:            nil,
			SecretService:              mockSecretService,
			ExpectError:                true,
			ExpectedError:              ErrInvalidService,
		},
		{
			TestName:                   "error_with_nil_secret_service",
			ContainerRepository:        mockContainerRepo,
			ContainerDeploymentService: mockDeploymentService,
			VariableService:            mockVariableService,
			SecretService:              nil,
			ExpectError:                true,
			ExpectedError:              ErrInvalidService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			service, err := NewContainerService(
				tc.ContainerRepository,
				tc.ContainerDeploymentService,
				tc.VariableService,
				tc.SecretService,
			)

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

func TestContainerService_Create(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validContainerID := gofakeit.UUID()
	validRegistryID := gofakeit.UUID()
	invalidEnvID := "invalid-uuid"
	emptyEnvID := ""

	cpu := int32(500)
	memory := int32(512)
	minInstances := int32(1)
	maxInstances := int32(1)

	validRequest := container.UpsertServiceRequest{
		ContainerUpsertRequest: container.UpsertRepositoryRequest{
			RegistryID:          validRegistryID,
			Name:                gofakeit.Word(),
			ImageName:           gofakeit.Word(),
			Tag:                 "latest",
			CPU:                 &cpu,
			Memory:              &memory,
			MinRunningInstances: &minInstances,
			MaxRunningInstances: &maxInstances,
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
	}

	invalidRequest := container.UpsertServiceRequest{
		ContainerUpsertRequest: container.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedContainer := &container.Container{
		ID:                   uuid.MustParse(validContainerID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		RegistryID:           uuid.MustParse(validRegistryID),
		Name:                 validRequest.ContainerUpsertRequest.Name,
		IconUri:              "app://qovery-console/container",
		ImageName:            validRequest.ContainerUpsertRequest.ImageName,
		Tag:                  validRequest.ContainerUpsertRequest.Tag,
		CPU:                  cpu,
		Memory:               memory,
		MinRunningInstances:  minInstances,
		MaxRunningInstances:  maxInstances,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
	}

	testCases := []struct {
		TestName          string
		EnvironmentID     string
		Request           container.UpsertServiceRequest
		SetupMocks        func(*repoMocks.ContainerRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError       bool
		ErrorContains     string
		ExpectedContainer *container.Container
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_request",
			EnvironmentID: validEnvID,
			Request:       invalidRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_repository_create_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_variable_service_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_secret_service_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_variable_service_list_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_secret_service_list_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "error_deployment_service_get_status_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create container",
		},
		{
			TestName:      "success",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(&status.Status{State: status.StateDeployed}, nil)
			},
			ExpectError:       false,
			ExpectedContainer: expectedContainer,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockContainerRepo := &repoMocks.ContainerRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewContainerService(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)
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
				if tc.ExpectedContainer != nil {
					assert.Equal(t, tc.ExpectedContainer.ID, result.ID)
					assert.Equal(t, tc.ExpectedContainer.Name, result.Name)
					assert.Equal(t, tc.ExpectedContainer.State, result.State)
				}
			}

			mockContainerRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestContainerService_Get(t *testing.T) {
	t.Parallel()

	validContainerID := gofakeit.UUID()
	validEnvID := gofakeit.UUID()
	validRegistryID := gofakeit.UUID()
	invalidContainerID := "invalid-uuid"
	emptyContainerID := ""
	advancedSettingsJson := `{"key": "value"}`

	expectedContainer := &container.Container{
		ID:                   uuid.MustParse(validContainerID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		RegistryID:           uuid.MustParse(validRegistryID),
		Name:                 gofakeit.Word(),
		IconUri:              "app://qovery-console/container",
		ImageName:            gofakeit.Word(),
		Tag:                  "latest",
		CPU:                  500,
		Memory:               512,
		MinRunningInstances:  1,
		MaxRunningInstances:  1,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
		AdvancedSettingsJson: advancedSettingsJson,
	}

	testCases := []struct {
		TestName                      string
		ContainerID                   string
		AdvancedSettingsJsonFromState string
		IsTriggeredFromImport         bool
		SetupMocks                    func(*repoMocks.ContainerRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError                   bool
		ErrorContains                 string
	}{
		{
			TestName:                      "error_empty_container_id",
			ContainerID:                   emptyContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:                      "error_invalid_container_id",
			ContainerID:                   invalidContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:                      "error_repository_get_failure",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, advancedSettingsJson, false).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get container",
		},
		{
			TestName:                      "error_variable_service_list_failure_in_refresh",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, advancedSettingsJson, false).
					Return(expectedContainer, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get container",
		},
		{
			TestName:                      "error_secret_service_list_failure_in_refresh",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, advancedSettingsJson, false).
					Return(expectedContainer, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get container",
		},
		{
			TestName:                      "error_deployment_service_get_status_failure_in_refresh",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, advancedSettingsJson, false).
					Return(expectedContainer, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get container",
		},
		{
			TestName:                      "success",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, advancedSettingsJson, false).
					Return(expectedContainer, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(&status.Status{State: status.StateDeployed}, nil)
			},
			ExpectError: false,
		},
		{
			TestName:                      "success_with_import",
			ContainerID:                   validContainerID,
			AdvancedSettingsJsonFromState: "",
			IsTriggeredFromImport:         true,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Get(mock.Anything, validContainerID, "", true).
					Return(expectedContainer, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(&status.Status{State: status.StateDeployed}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockContainerRepo := &repoMocks.ContainerRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewContainerService(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.ContainerID, tc.AdvancedSettingsJsonFromState, tc.IsTriggeredFromImport)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedContainer.ID, result.ID)
				assert.Equal(t, expectedContainer.Name, result.Name)
			}

			mockContainerRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestContainerService_Update(t *testing.T) {
	t.Parallel()

	validContainerID := gofakeit.UUID()
	validEnvID := gofakeit.UUID()
	validRegistryID := gofakeit.UUID()
	invalidContainerID := "invalid-uuid"
	emptyContainerID := ""

	cpu := int32(1000)
	memory := int32(1024)
	minInstances := int32(2)
	maxInstances := int32(4)

	validRequest := container.UpsertServiceRequest{
		ContainerUpsertRequest: container.UpsertRepositoryRequest{
			RegistryID:          validRegistryID,
			Name:                gofakeit.Word(),
			ImageName:           gofakeit.Word(),
			Tag:                 "v2.0",
			CPU:                 &cpu,
			Memory:              &memory,
			MinRunningInstances: &minInstances,
			MaxRunningInstances: &maxInstances,
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
	}

	invalidRequest := container.UpsertServiceRequest{
		ContainerUpsertRequest: container.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedContainer := &container.Container{
		ID:                   uuid.MustParse(validContainerID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		RegistryID:           uuid.MustParse(validRegistryID),
		Name:                 validRequest.ContainerUpsertRequest.Name,
		IconUri:              "app://qovery-console/container",
		ImageName:            validRequest.ContainerUpsertRequest.ImageName,
		Tag:                  validRequest.ContainerUpsertRequest.Tag,
		CPU:                  cpu,
		Memory:               memory,
		MinRunningInstances:  minInstances,
		MaxRunningInstances:  maxInstances,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		ContainerID   string
		Request       container.UpsertServiceRequest
		SetupMocks    func(*repoMocks.ContainerRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:    "error_empty_container_id",
			ContainerID: emptyContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:    "error_invalid_container_id",
			ContainerID: invalidContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:    "error_invalid_request",
			ContainerID: validContainerID,
			Request:     invalidRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_repository_update_failure",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_variable_service_update_failure",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_secret_service_update_failure",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_variable_service_list_failure_in_refresh",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_secret_service_list_failure_in_refresh",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "error_deployment_service_get_status_failure_in_refresh",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update container",
		},
		{
			TestName:    "success",
			ContainerID: validContainerID,
			Request:     validRequest,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				cr.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.ContainerUpsertRequest).
					Return(expectedContainer, nil)
				vs.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validContainerID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validContainerID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validContainerID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validContainerID).
					Return(&status.Status{State: status.StateDeployed}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockContainerRepo := &repoMocks.ContainerRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewContainerService(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.ContainerID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedContainer.ID, result.ID)
				assert.Equal(t, expectedContainer.Name, result.Name)
			}

			mockContainerRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestContainerService_Delete(t *testing.T) {
	t.Parallel()

	validContainerID := gofakeit.UUID()
	invalidContainerID := "invalid-uuid"
	emptyContainerID := ""

	testCases := []struct {
		TestName      string
		ContainerID   string
		SetupMocks    func(*repoMocks.ContainerRepository, *mocks_test.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:    "error_empty_container_id",
			ContainerID: emptyContainerID,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:    "error_invalid_container_id",
			ContainerID: invalidContainerID,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid container id param",
		},
		{
			TestName:    "error_repository_delete_failure",
			ContainerID: validContainerID,
			SetupMocks: func(cr *repoMocks.ContainerRepository, ds *mocks_test.DeploymentService) {
				cr.EXPECT().
					Delete(mock.Anything, validContainerID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete container",
		},
		// Note: Success case for Delete is skipped because it requires the wait function
		// which polls GetStatus expecting a proper 404 APIError. This is better tested
		// in integration tests where the full API behavior can be validated.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockContainerRepo := &repoMocks.ContainerRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockContainerRepo, mockDeploymentService)

			service, err := NewContainerService(mockContainerRepo, mockDeploymentService, mockVariableService, mockSecretService)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.ContainerID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockContainerRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}
