//go:build unit && !integration
// +build unit,!integration

package services

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	repoMocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

// testHelmDeploymentRestrictionService creates a test-only deployment restriction service
// that returns empty deployment restrictions to avoid nil pointer panics in unit tests
func testHelmDeploymentRestrictionService() deploymentrestriction.DeploymentRestrictionService {
	// Create an empty API client configuration
	cfg := qovery.NewConfiguration()
	apiClient := qovery.NewAPIClient(cfg)

	service, _ := deploymentrestriction.NewDeploymentRestrictionService(*apiClient)
	return service
}

func TestNewHelmService(t *testing.T) {
	t.Parallel()

	mockHelmRepo := &repoMocks.HelmRepository{}
	mockDeploymentService := &mocks_test.DeploymentService{}
	mockVariableService := &mocks_test.VariableService{}
	mockSecretService := &mocks_test.SecretService{}
	mockDeploymentRestrictionService := deploymentrestriction.DeploymentRestrictionService{}

	testCases := []struct {
		TestName                         string
		HelmRepository                   helm.Repository
		HelmDeploymentService            mocks_test.DeploymentService
		UseNilDeploymentService          bool
		VariableService                  variable.Service
		SecretService                    secret.Service
		DeploymentRestrictionService     deploymentrestriction.DeploymentRestrictionService
		ExpectError                      bool
		ExpectedError                    error
	}{
		{
			TestName:                         "success_with_all_valid_dependencies",
			HelmRepository:                   mockHelmRepo,
			HelmDeploymentService:            *mockDeploymentService,
			VariableService:                  mockVariableService,
			SecretService:                    mockSecretService,
			DeploymentRestrictionService:     mockDeploymentRestrictionService,
			ExpectError:                      false,
		},
		{
			TestName:                         "error_with_nil_helm_repository",
			HelmRepository:                   nil,
			HelmDeploymentService:            *mockDeploymentService,
			VariableService:                  mockVariableService,
			SecretService:                    mockSecretService,
			DeploymentRestrictionService:     mockDeploymentRestrictionService,
			ExpectError:                      true,
			ExpectedError:                    ErrInvalidRepository,
		},
		{
			TestName:                         "error_with_nil_deployment_service",
			HelmRepository:                   mockHelmRepo,
			UseNilDeploymentService:          true,
			VariableService:                  mockVariableService,
			SecretService:                    mockSecretService,
			DeploymentRestrictionService:     mockDeploymentRestrictionService,
			ExpectError:                      true,
			ExpectedError:                    ErrInvalidService,
		},
		{
			TestName:                         "error_with_nil_variable_service",
			HelmRepository:                   mockHelmRepo,
			HelmDeploymentService:            *mockDeploymentService,
			VariableService:                  nil,
			SecretService:                    mockSecretService,
			DeploymentRestrictionService:     mockDeploymentRestrictionService,
			ExpectError:                      true,
			ExpectedError:                    ErrInvalidService,
		},
		{
			TestName:                         "error_with_nil_secret_service",
			HelmRepository:                   mockHelmRepo,
			HelmDeploymentService:            *mockDeploymentService,
			VariableService:                  mockVariableService,
			SecretService:                    nil,
			DeploymentRestrictionService:     mockDeploymentRestrictionService,
			ExpectError:                      true,
			ExpectedError:                    ErrInvalidService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			var deploymentSvc deployment.Service
			if !tc.UseNilDeploymentService {
				deploymentSvc = &tc.HelmDeploymentService
			}

			service, err := NewHelmService(
				tc.HelmRepository,
				deploymentSvc,
				tc.VariableService,
				tc.SecretService,
				tc.DeploymentRestrictionService,
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

func TestHelmService_Create(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	validHelmID := gofakeit.UUID()
	invalidEnvID := "invalid-uuid"
	emptyEnvID := ""

	chartName := gofakeit.Word()
	chartVersion := "1.0.0"
	timeoutSec := int32(600)

	validRequest := helm.UpsertServiceRequest{
		HelmUpsertRequest: helm.UpsertRepositoryRequest{
			Name:       gofakeit.Word(),
			TimeoutSec: &timeoutSec,
			Source: helm.Source{
				HelmRepository: &helm.SourceHelmRepository{
					ChartName:    chartName,
					ChartVersion: chartVersion,
				},
			},
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
		DeploymentRestrictionsDiff:   deploymentrestriction.ServiceDeploymentRestrictionsDiff{},
	}

	invalidRequest := helm.UpsertServiceRequest{
		HelmUpsertRequest: helm.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedHelm := &helm.Helm{
		ID:                   uuid.MustParse(validHelmID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		Name:                 validRequest.HelmUpsertRequest.Name,
		IconUri:              "app://qovery-console/helm",
		TimeoutSec:           &timeoutSec,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		EnvironmentID string
		Request       helm.UpsertServiceRequest
		SetupMocks    func(*repoMocks.HelmRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_request",
			EnvironmentID: validEnvID,
			Request:       invalidRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_repository_create_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_variable_service_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_secret_service_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_variable_service_list_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_secret_service_list_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		{
			TestName:      "error_deployment_service_get_status_failure_in_refresh",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validHelmID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create helm",
		},
		// Note: Success case is skipped because refreshHelm calls deploymentRestrictionService
		// which is a concrete struct requiring a real API client. This is better tested
		// in integration tests.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockHelmRepo := &repoMocks.HelmRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewHelmService(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService, testHelmDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.EnvironmentID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			}

			mockHelmRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestHelmService_Get(t *testing.T) {
	t.Parallel()

	validHelmID := gofakeit.UUID()
	validEnvID := gofakeit.UUID()
	invalidHelmID := "invalid-uuid"
	emptyHelmID := ""
	advancedSettingsJson := `{"key": "value"}`

	timeoutSec := int32(600)

	expectedHelm := &helm.Helm{
		ID:                   uuid.MustParse(validHelmID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		Name:                 gofakeit.Word(),
		IconUri:              "app://qovery-console/helm",
		TimeoutSec:           &timeoutSec,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
		AdvancedSettingsJson: advancedSettingsJson,
	}

	testCases := []struct {
		TestName                      string
		HelmID                        string
		AdvancedSettingsJsonFromState string
		IsTriggeredFromImport         bool
		SetupMocks                    func(*repoMocks.HelmRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError                   bool
		ErrorContains                 string
	}{
		{
			TestName:                      "error_empty_helm_id",
			HelmID:                        emptyHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName:                      "error_invalid_helm_id",
			HelmID:                        invalidHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName:                      "error_repository_get_failure",
			HelmID:                        validHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Get(mock.Anything, validHelmID, advancedSettingsJson, false).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get helm",
		},
		{
			TestName:                      "error_variable_service_list_failure_in_refresh",
			HelmID:                        validHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Get(mock.Anything, validHelmID, advancedSettingsJson, false).
					Return(expectedHelm, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get helm",
		},
		{
			TestName:                      "error_secret_service_list_failure_in_refresh",
			HelmID:                        validHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Get(mock.Anything, validHelmID, advancedSettingsJson, false).
					Return(expectedHelm, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get helm",
		},
		{
			TestName:                      "error_deployment_service_get_status_failure_in_refresh",
			HelmID:                        validHelmID,
			AdvancedSettingsJsonFromState: advancedSettingsJson,
			IsTriggeredFromImport:         false,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Get(mock.Anything, validHelmID, advancedSettingsJson, false).
					Return(expectedHelm, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validHelmID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get helm",
		},
		// NOTE: Success tests are skipped because DeploymentRestrictionService is a concrete struct (not an interface)
		// and cannot be mocked. The refreshHelm() method calls deploymentRestrictionService.GetServiceDeploymentRestrictions()
		// which would hit the real API. Use acceptance tests for success paths.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockHelmRepo := &repoMocks.HelmRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewHelmService(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService, testHelmDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.HelmID, tc.AdvancedSettingsJsonFromState, tc.IsTriggeredFromImport)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedHelm.ID, result.ID)
				assert.Equal(t, expectedHelm.Name, result.Name)
			}

			mockHelmRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestHelmService_Update(t *testing.T) {
	t.Parallel()

	validHelmID := gofakeit.UUID()
	validEnvID := gofakeit.UUID()
	invalidHelmID := "invalid-uuid"
	emptyHelmID := ""

	chartName := gofakeit.Word()
	chartVersion := "2.0.0"
	timeoutSec := int32(900)

	validRequest := helm.UpsertServiceRequest{
		HelmUpsertRequest: helm.UpsertRepositoryRequest{
			Name:       gofakeit.Word(),
			TimeoutSec: &timeoutSec,
			Source: helm.Source{
				HelmRepository: &helm.SourceHelmRepository{
					ChartName:    chartName,
					ChartVersion: chartVersion,
				},
			},
		},
		EnvironmentVariables:         variable.DiffRequest{},
		EnvironmentVariableAliases:   variable.DiffRequest{},
		EnvironmentVariableOverrides: variable.DiffRequest{},
		Secrets:                      secret.DiffRequest{},
		SecretAliases:                secret.DiffRequest{},
		SecretOverrides:              secret.DiffRequest{},
		DeploymentRestrictionsDiff:   deploymentrestriction.ServiceDeploymentRestrictionsDiff{},
	}

	invalidRequest := helm.UpsertServiceRequest{
		HelmUpsertRequest: helm.UpsertRepositoryRequest{
			Name: "",
		},
	}

	expectedHelm := &helm.Helm{
		ID:                   uuid.MustParse(validHelmID),
		EnvironmentID:        uuid.MustParse(validEnvID),
		Name:                 validRequest.HelmUpsertRequest.Name,
		IconUri:              "app://qovery-console/helm",
		TimeoutSec:           &timeoutSec,
		EnvironmentVariables: variable.Variables{},
		Secrets:              secret.Secrets{},
		State:                status.StateDeployed,
	}

	testCases := []struct {
		TestName      string
		HelmID        string
		Request       helm.UpsertServiceRequest
		SetupMocks    func(*repoMocks.HelmRepository, *mocks_test.DeploymentService, *mocks_test.VariableService, *mocks_test.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName: "error_empty_helm_id",
			HelmID:   emptyHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName: "error_invalid_helm_id",
			HelmID:   invalidHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName: "error_invalid_request",
			HelmID:   validHelmID,
			Request:  invalidRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_repository_update_failure",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_variable_service_update_failure",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_secret_service_update_failure",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_variable_service_list_failure_in_refresh",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_secret_service_list_failure_in_refresh",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		{
			TestName: "error_deployment_service_get_status_failure_in_refresh",
			HelmID:   validHelmID,
			Request:  validRequest,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService, vs *mocks_test.VariableService, ss *mocks_test.SecretService) {
				hr.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.HelmUpsertRequest).
					Return(expectedHelm, nil)
				vs.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validHelmID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validHelmID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validHelmID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validHelmID).
					Return(nil, errors.New("deployment status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update helm",
		},
		// NOTE: Success test is skipped because DeploymentRestrictionService is a concrete struct (not an interface)
		// and cannot be mocked. The refreshHelm() method calls deploymentRestrictionService.GetServiceDeploymentRestrictions()
		// which would hit the real API. Use acceptance tests for success paths.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockHelmRepo := &repoMocks.HelmRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService)

			service, err := NewHelmService(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService, testHelmDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.HelmID, tc.Request)

			assert.Error(t, err)
			assert.Nil(t, result)
			if tc.ErrorContains != "" {
				assert.Contains(t, err.Error(), tc.ErrorContains)
			}

			mockHelmRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestHelmService_Delete(t *testing.T) {
	t.Parallel()

	validHelmID := gofakeit.UUID()
	invalidHelmID := "invalid-uuid"
	emptyHelmID := ""

	testCases := []struct {
		TestName      string
		HelmID        string
		SetupMocks    func(*repoMocks.HelmRepository, *mocks_test.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName: "error_empty_helm_id",
			HelmID:   emptyHelmID,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName: "error_invalid_helm_id",
			HelmID:   invalidHelmID,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid helm id param",
		},
		{
			TestName: "error_repository_delete_failure",
			HelmID:   validHelmID,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService) {
				hr.EXPECT().
					Delete(mock.Anything, validHelmID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete helm",
		},
		{
			TestName: "success",
			HelmID:   validHelmID,
			SetupMocks: func(hr *repoMocks.HelmRepository, ds *mocks_test.DeploymentService) {
				hr.EXPECT().
					Delete(mock.Anything, validHelmID).
					Return(nil)
				notFoundErr := apierrors.NewNotFoundAPIError(apierrors.APIResourceHelm, validHelmID)
				ds.EXPECT().
					GetStatus(mock.Anything, validHelmID).
					Return(nil, notFoundErr)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockHelmRepo := &repoMocks.HelmRepository{}
			mockDeploymentService := &mocks_test.DeploymentService{}
			mockVariableService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}

			tc.SetupMocks(mockHelmRepo, mockDeploymentService)

			service, err := NewHelmService(mockHelmRepo, mockDeploymentService, mockVariableService, mockSecretService, testHelmDeploymentRestrictionService())
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.HelmID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockHelmRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}
