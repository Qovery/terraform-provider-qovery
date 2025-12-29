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
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/docker"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/git_repository"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	repoMocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

// testJobDeploymentRestrictionService creates a test-only deployment restriction service
// that returns empty deployment restrictions to avoid nil pointer panics in unit tests
func testJobDeploymentRestrictionService() deploymentrestriction.DeploymentRestrictionService {
	cfg := qovery.NewConfiguration()
	apiClient := qovery.NewAPIClient(cfg)
	service, _ := deploymentrestriction.NewDeploymentRestrictionService(*apiClient)
	return service
}

func TestNewJobService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName                     string
		JobRepository                job.Repository
		JobDeploymentService         deployment.Service
		VariableService              variable.Service
		SecretService                secret.Service
		DeploymentRestrictionService deploymentrestriction.DeploymentRestrictionService
		ExpectError                  bool
		ExpectedError                error
	}{
		{
			TestName:                     "success_with_valid_dependencies",
			JobRepository:                &repoMocks.JobRepository{},
			JobDeploymentService:         &mocks_test.DeploymentService{},
			VariableService:              &mocks_test.VariableService{},
			SecretService:                &mocks_test.SecretService{},
			DeploymentRestrictionService: testJobDeploymentRestrictionService(),
			ExpectError:                  false,
		},
		{
			TestName:                     "error_with_nil_repository",
			JobRepository:                nil,
			JobDeploymentService:         &mocks_test.DeploymentService{},
			VariableService:              &mocks_test.VariableService{},
			SecretService:                &mocks_test.SecretService{},
			DeploymentRestrictionService: testJobDeploymentRestrictionService(),
			ExpectError:                  true,
			ExpectedError:                ErrInvalidRepository,
		},
		{
			TestName:                     "error_with_nil_deployment_service",
			JobRepository:                &repoMocks.JobRepository{},
			JobDeploymentService:         nil,
			VariableService:              &mocks_test.VariableService{},
			SecretService:                &mocks_test.SecretService{},
			DeploymentRestrictionService: testJobDeploymentRestrictionService(),
			ExpectError:                  true,
			ExpectedError:                ErrInvalidService,
		},
		{
			TestName:                     "error_with_nil_variable_service",
			JobRepository:                &repoMocks.JobRepository{},
			JobDeploymentService:         &mocks_test.DeploymentService{},
			VariableService:              nil,
			SecretService:                &mocks_test.SecretService{},
			DeploymentRestrictionService: testJobDeploymentRestrictionService(),
			ExpectError:                  true,
			ExpectedError:                ErrInvalidService,
		},
		{
			TestName:                     "error_with_nil_secret_service",
			JobRepository:                &repoMocks.JobRepository{},
			JobDeploymentService:         &mocks_test.DeploymentService{},
			VariableService:              &mocks_test.VariableService{},
			SecretService:                nil,
			DeploymentRestrictionService: testJobDeploymentRestrictionService(),
			ExpectError:                  true,
			ExpectedError:                ErrInvalidService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewJobService(
				tc.JobRepository,
				tc.JobDeploymentService,
				tc.VariableService,
				tc.SecretService,
				tc.DeploymentRestrictionService,
			)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				assert.Equal(t, tc.ExpectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestJobService_Create(t *testing.T) {
	t.Parallel()

	validEnvID := gofakeit.UUID()
	invalidEnvID := "invalid-uuid"
	emptyEnvID := ""

	validJobID := uuid.New()
	cpu := int32(500)
	memory := int32(512)

	dockerGitURL := gofakeit.URL()
	dockerBranch := "main"
	dockerRootPath := "/"
	entrypoint := gofakeit.Word()

	onStartCmd := execution_command.ExecutionCommand{
		Entrypoint: &entrypoint,
		Arguments:  []string{gofakeit.Word()},
	}

	validRequest := job.UpsertServiceRequest{
		JobUpsertRequest: job.UpsertRepositoryRequest{
			Name:   gofakeit.Word(),
			CPU:    &cpu,
			Memory: &memory,
			Source: job.Source{
				Docker: &docker.Docker{
					GitRepository: git_repository.GitRepository{
						Url:      dockerGitURL,
						Branch:   &dockerBranch,
						RootPath: &dockerRootPath,
					},
				},
			},
			Schedule: job.JobSchedule{
				OnStart: &onStartCmd,
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

	invalidRequest := job.UpsertServiceRequest{
		JobUpsertRequest: job.UpsertRepositoryRequest{
			Name: "",
		},
	}

	createdJob := &job.Job{
		ID:            validJobID,
		EnvironmentID: uuid.MustParse(validEnvID),
		Name:          validRequest.JobUpsertRequest.Name,
		CPU:           *validRequest.JobUpsertRequest.CPU,
		Memory:        *validRequest.JobUpsertRequest.Memory,
		Source:        validRequest.JobUpsertRequest.Source,
		Schedule:      validRequest.JobUpsertRequest.Schedule,
		State:         status.StateStopped,
	}

	testCases := []struct {
		TestName          string
		EnvironmentID     string
		Request           job.UpsertServiceRequest
		SetupMocks        func(*repoMocks.JobRepository, *mocks_test.VariableService, *mocks_test.SecretService, *mocks_test.DeploymentService)
		ExpectError       bool
		ErrorContains     string
	}{
		{
			TestName:      "error_empty_environment_id",
			EnvironmentID: emptyEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_environment_id",
			EnvironmentID: invalidEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid environment id param",
		},
		{
			TestName:      "error_invalid_request",
			EnvironmentID: validEnvID,
			Request:       invalidRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_repository_create_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_variable_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(createdJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_secret_update_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(createdJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_refresh_variable_list_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(createdJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID.String()).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_refresh_secret_list_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(createdJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID.String()).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID.String()).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		{
			TestName:      "error_refresh_status_failure",
			EnvironmentID: validEnvID,
			Request:       validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Create(mock.Anything, validEnvID, validRequest.JobUpsertRequest).
					Return(createdJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID.String(), validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID.String()).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID.String()).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validJobID.String()).
					Return(nil, errors.New("status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create job",
		},
		// Note: Success case is skipped because refreshJob calls deploymentRestrictionService
		// which is a concrete struct requiring a real API client. This is better tested
		// in integration tests.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockJobRepo := &repoMocks.JobRepository{}
			mockVarService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}
			mockDeploymentService := &mocks_test.DeploymentService{}

			tc.SetupMocks(mockJobRepo, mockVarService, mockSecretService, mockDeploymentService)

			service, err := NewJobService(mockJobRepo, mockDeploymentService, mockVarService, mockSecretService, testJobDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.EnvironmentID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			}

			mockJobRepo.AssertExpectations(t)
			mockVarService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}

func TestJobService_Get(t *testing.T) {
	t.Parallel()

	validJobID := gofakeit.UUID()
	invalidJobID := "invalid-uuid"
	emptyJobID := ""
	advancedSettings := ""
	isTriggeredFromImport := false

	dockerGitURL := gofakeit.URL()
	dockerBranch := "main"
	dockerRootPath := "/"
	entrypoint := gofakeit.Word()

	onStartCmd := execution_command.ExecutionCommand{
		Entrypoint: &entrypoint,
		Arguments:  []string{gofakeit.Word()},
	}

	fetchedJob := &job.Job{
		ID:            uuid.MustParse(validJobID),
		EnvironmentID: uuid.New(),
		Name:          gofakeit.Word(),
		CPU:           500,
		Memory:        512,
		Source: job.Source{
			Docker: &docker.Docker{
				GitRepository: git_repository.GitRepository{
					Url:      dockerGitURL,
					Branch:   &dockerBranch,
					RootPath: &dockerRootPath,
				},
			},
		},
		Schedule: job.JobSchedule{
			OnStart: &onStartCmd,
		},
		State: status.StateStopped,
	}

	testCases := []struct {
		TestName      string
		JobID         string
		SetupMocks    func(*repoMocks.JobRepository, *mocks_test.VariableService, *mocks_test.SecretService, *mocks_test.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName: "error_empty_job_id",
			JobID:    emptyJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_invalid_job_id",
			JobID:    invalidJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_repository_get_failure",
			JobID:    validJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Get(mock.Anything, validJobID, advancedSettings, isTriggeredFromImport).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get job",
		},
		{
			TestName: "error_refresh_variable_list_failure",
			JobID:    validJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Get(mock.Anything, validJobID, advancedSettings, isTriggeredFromImport).
					Return(fetchedJob, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get job",
		},
		{
			TestName: "error_refresh_secret_list_failure",
			JobID:    validJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Get(mock.Anything, validJobID, advancedSettings, isTriggeredFromImport).
					Return(fetchedJob, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get job",
		},
		{
			TestName: "error_refresh_status_failure",
			JobID:    validJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Get(mock.Anything, validJobID, advancedSettings, isTriggeredFromImport).
					Return(fetchedJob, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validJobID).
					Return(nil, errors.New("status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get job",
		},
		// Note: Success case is skipped because refreshJob calls deploymentRestrictionService
		// which is a concrete struct requiring a real API client. This is better tested
		// in integration tests.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockJobRepo := &repoMocks.JobRepository{}
			mockVarService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}
			mockDeploymentService := &mocks_test.DeploymentService{}

			tc.SetupMocks(mockJobRepo, mockVarService, mockSecretService, mockDeploymentService)

			service, err := NewJobService(mockJobRepo, mockDeploymentService, mockVarService, mockSecretService, testJobDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.JobID, advancedSettings, isTriggeredFromImport)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			}

			mockJobRepo.AssertExpectations(t)
			mockVarService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}

func TestJobService_Update(t *testing.T) {
	t.Parallel()

	validJobID := gofakeit.UUID()
	invalidJobID := "invalid-uuid"
	emptyJobID := ""

	cpu := int32(500)
	memory := int32(512)

	dockerGitURL := gofakeit.URL()
	dockerBranch := "main"
	dockerRootPath := "/"
	entrypoint := gofakeit.Word()

	onStartCmd := execution_command.ExecutionCommand{
		Entrypoint: &entrypoint,
		Arguments:  []string{gofakeit.Word()},
	}

	validRequest := job.UpsertServiceRequest{
		JobUpsertRequest: job.UpsertRepositoryRequest{
			Name:   gofakeit.Word(),
			CPU:    &cpu,
			Memory: &memory,
			Source: job.Source{
				Docker: &docker.Docker{
					GitRepository: git_repository.GitRepository{
						Url:      dockerGitURL,
						Branch:   &dockerBranch,
						RootPath: &dockerRootPath,
					},
				},
			},
			Schedule: job.JobSchedule{
				OnStart: &onStartCmd,
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

	invalidRequest := job.UpsertServiceRequest{
		JobUpsertRequest: job.UpsertRepositoryRequest{
			Name: "",
		},
	}

	updatedJob := &job.Job{
		ID:            uuid.MustParse(validJobID),
		EnvironmentID: uuid.New(),
		Name:          validRequest.JobUpsertRequest.Name,
		CPU:           *validRequest.JobUpsertRequest.CPU,
		Memory:        *validRequest.JobUpsertRequest.Memory,
		Source:        validRequest.JobUpsertRequest.Source,
		Schedule:      validRequest.JobUpsertRequest.Schedule,
		State:         status.StateStopped,
	}

	testCases := []struct {
		TestName      string
		JobID         string
		Request       job.UpsertServiceRequest
		SetupMocks    func(*repoMocks.JobRepository, *mocks_test.VariableService, *mocks_test.SecretService, *mocks_test.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName: "error_empty_job_id",
			JobID:    emptyJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_invalid_job_id",
			JobID:    invalidJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_invalid_request",
			JobID:    validJobID,
			Request:  invalidRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_repository_update_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_variable_update_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(updatedJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_secret_update_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(updatedJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_refresh_variable_list_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(updatedJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_refresh_secret_list_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(updatedJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		{
			TestName: "error_refresh_status_failure",
			JobID:    validJobID,
			Request:  validRequest,
			SetupMocks: func(jr *repoMocks.JobRepository, vs *mocks_test.VariableService, ss *mocks_test.SecretService, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Update(mock.Anything, validJobID, validRequest.JobUpsertRequest).
					Return(updatedJob, nil)
				vs.EXPECT().
					Update(mock.Anything, validJobID, validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, validRequest.EnvironmentVariableOverrides, mock.Anything).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					Update(mock.Anything, validJobID, validRequest.Secrets, validRequest.SecretAliases, validRequest.SecretOverrides, mock.Anything).
					Return(secret.Secrets{}, nil)
				vs.EXPECT().
					List(mock.Anything, validJobID).
					Return(variable.Variables{}, nil)
				ss.EXPECT().
					List(mock.Anything, validJobID).
					Return(secret.Secrets{}, nil)
				ds.EXPECT().
					GetStatus(mock.Anything, validJobID).
					Return(nil, errors.New("status error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update job",
		},
		// Note: Success case is skipped because refreshJob calls deploymentRestrictionService
		// which is a concrete struct requiring a real API client. This is better tested
		// in integration tests.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockJobRepo := &repoMocks.JobRepository{}
			mockVarService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}
			mockDeploymentService := &mocks_test.DeploymentService{}

			tc.SetupMocks(mockJobRepo, mockVarService, mockSecretService, mockDeploymentService)

			service, err := NewJobService(mockJobRepo, mockDeploymentService, mockVarService, mockSecretService, testJobDeploymentRestrictionService())
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.JobID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			}

			mockJobRepo.AssertExpectations(t)
			mockVarService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}

func TestJobService_Delete(t *testing.T) {
	t.Parallel()

	validJobID := gofakeit.UUID()
	invalidJobID := "invalid-uuid"
	emptyJobID := ""

	testCases := []struct {
		TestName      string
		JobID         string
		SetupMocks    func(*repoMocks.JobRepository, *mocks_test.DeploymentService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName: "error_empty_job_id",
			JobID:    emptyJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_invalid_job_id",
			JobID:    invalidJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, ds *mocks_test.DeploymentService) {
			},
			ExpectError:   true,
			ErrorContains: "invalid job id param",
		},
		{
			TestName: "error_repository_delete_failure",
			JobID:    validJobID,
			SetupMocks: func(jr *repoMocks.JobRepository, ds *mocks_test.DeploymentService) {
				jr.EXPECT().
					Delete(mock.Anything, validJobID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete job",
		},
		// Note: Success case for Delete is skipped because it requires the wait function
		// which polls GetStatus expecting a proper 404 APIError. This is better tested
		// in integration tests where the full API behavior can be validated.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockJobRepo := &repoMocks.JobRepository{}
			mockVarService := &mocks_test.VariableService{}
			mockSecretService := &mocks_test.SecretService{}
			mockDeploymentService := &mocks_test.DeploymentService{}

			tc.SetupMocks(mockJobRepo, mockDeploymentService)

			service, err := NewJobService(mockJobRepo, mockDeploymentService, mockVarService, mockSecretService, testJobDeploymentRestrictionService())
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.JobID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockJobRepo.AssertExpectations(t)
			mockDeploymentService.AssertExpectations(t)
		})
	}
}
