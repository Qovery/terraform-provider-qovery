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

	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	repoMocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
	serviceMocks "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
)

func TestNewProjectService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		ProjectRepo     project.Repository
		VariableService variable.Service
		SecretService   secret.Service
		ExpectError     bool
		ErrorContains   string
	}{
		{
			TestName:        "success_with_valid_dependencies",
			ProjectRepo:     &repoMocks.ProjectRepository{},
			VariableService: &serviceMocks.VariableService{},
			SecretService:   &serviceMocks.SecretService{},
			ExpectError:     false,
		},
		{
			TestName:        "error_with_nil_repository",
			ProjectRepo:     nil,
			VariableService: &serviceMocks.VariableService{},
			SecretService:   &serviceMocks.SecretService{},
			ExpectError:     true,
			ErrorContains:   "invalid repository",
		},
		{
			TestName:        "error_with_nil_variable_service",
			ProjectRepo:     &repoMocks.ProjectRepository{},
			VariableService: nil,
			SecretService:   &serviceMocks.SecretService{},
			ExpectError:     true,
			ErrorContains:   "invalid service",
		},
		{
			TestName:        "error_with_nil_secret_service",
			ProjectRepo:     &repoMocks.ProjectRepository{},
			VariableService: &serviceMocks.VariableService{},
			SecretService:   nil,
			ExpectError:     true,
			ErrorContains:   "invalid service",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			service, err := NewProjectService(tc.ProjectRepo, tc.VariableService, tc.SecretService)
			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestProjectService_Create(t *testing.T) {
	t.Parallel()

	validOrgID := gofakeit.UUID()
	invalidOrgID := "invalid-uuid"
	emptyOrgID := ""

	validRequest := project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        gofakeit.Word(),
			Description: nil,
		},
		EnvironmentVariables:       variable.DiffRequest{},
		EnvironmentVariableAliases: variable.DiffRequest{},
		Secrets:                    secret.DiffRequest{},
		SecretAliases:              secret.DiffRequest{},
	}

	invalidRequest := project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        "", // Invalid: empty name
			Description: nil,
		},
		EnvironmentVariables:       variable.DiffRequest{},
		EnvironmentVariableAliases: variable.DiffRequest{},
		Secrets:                    secret.DiffRequest{},
		SecretAliases:              secret.DiffRequest{},
	}

	createdProject := &project.Project{
		ID:             uuid.MustParse(gofakeit.UUID()),
		OrganizationID: uuid.MustParse(validOrgID),
		Name:           validRequest.ProjectUpsertRequest.Name,
		Description:    validRequest.ProjectUpsertRequest.Description,
	}

	testCases := []struct {
		TestName         string
		OrganizationID   string
		Request          project.UpsertServiceRequest
		SetupMocks       func(*repoMocks.ProjectRepository, *serviceMocks.VariableService, *serviceMocks.SecretService)
		ExpectError      bool
		ErrorContains    string
	}{
		{
			TestName:       "error_empty_organization_id",
			OrganizationID: emptyOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_organization_id",
			OrganizationID: invalidOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:    true,
			ErrorContains:  "invalid organization id param",
		},
		{
			TestName:       "error_invalid_request",
			OrganizationID: validOrgID,
			Request:        invalidRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:    true,
			ErrorContains:  "failed to create project",
		},
		{
			TestName:       "error_repository_create_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create project",
		},
		{
			TestName:       "error_variable_service_update_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(createdProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create project",
		},
		{
			TestName:       "error_secret_service_update_failure",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(createdProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create project",
		},
		{
			TestName:       "error_variable_service_list_failure_in_refresh",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(createdProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, createdProject.ID.String()).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create project",
		},
		{
			TestName:       "error_secret_service_list_failure_in_refresh",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(createdProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, createdProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, createdProject.ID.String()).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to create project",
		},
		{
			TestName:       "success",
			OrganizationID: validOrgID,
			Request:        validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Create(mock.Anything, validOrgID, validRequest.ProjectUpsertRequest).
					Return(createdProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, createdProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, createdProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, createdProject.ID.String()).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockProjectRepo := &repoMocks.ProjectRepository{}
			mockVariableService := &serviceMocks.VariableService{}
			mockSecretService := &serviceMocks.SecretService{}
			tc.SetupMocks(mockProjectRepo, mockVariableService, mockSecretService)

			service, err := NewProjectService(mockProjectRepo, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Create(context.Background(), tc.OrganizationID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, createdProject.ID, result.ID)
				assert.Equal(t, createdProject.Name, result.Name)
			}

			mockProjectRepo.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestProjectService_Get(t *testing.T) {
	t.Parallel()

	validProjectID := gofakeit.UUID()
	invalidProjectID := "invalid-uuid"
	emptyProjectID := ""

	expectedProject := &project.Project{
		ID:             uuid.MustParse(validProjectID),
		OrganizationID: uuid.MustParse(gofakeit.UUID()),
		Name:           gofakeit.Word(),
		Description:    nil,
	}

	testCases := []struct {
		TestName      string
		ProjectID     string
		SetupMocks    func(*repoMocks.ProjectRepository, *serviceMocks.VariableService, *serviceMocks.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:  "error_empty_project_id",
			ProjectID: emptyProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:  "error_invalid_project_id",
			ProjectID: invalidProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:  "error_repository_get_failure",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Get(mock.Anything, validProjectID).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get project",
		},
		{
			TestName:  "error_variable_service_list_failure",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Get(mock.Anything, validProjectID).
					Return(expectedProject, nil)

				vs.EXPECT().
					List(mock.Anything, expectedProject.ID.String()).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get project",
		},
		{
			TestName:  "error_secret_service_list_failure",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Get(mock.Anything, validProjectID).
					Return(expectedProject, nil)

				vs.EXPECT().
					List(mock.Anything, expectedProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, expectedProject.ID.String()).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to get project",
		},
		{
			TestName:  "success",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Get(mock.Anything, validProjectID).
					Return(expectedProject, nil)

				vs.EXPECT().
					List(mock.Anything, expectedProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, expectedProject.ID.String()).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockProjectRepo := &repoMocks.ProjectRepository{}
			mockVariableService := &serviceMocks.VariableService{}
			mockSecretService := &serviceMocks.SecretService{}
			tc.SetupMocks(mockProjectRepo, mockVariableService, mockSecretService)

			service, err := NewProjectService(mockProjectRepo, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Get(context.Background(), tc.ProjectID)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, expectedProject.ID, result.ID)
				assert.Equal(t, expectedProject.Name, result.Name)
			}

			mockProjectRepo.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestProjectService_Update(t *testing.T) {
	t.Parallel()

	validProjectID := gofakeit.UUID()
	invalidProjectID := "invalid-uuid"
	emptyProjectID := ""

	validRequest := project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        gofakeit.Word(),
			Description: nil,
		},
		EnvironmentVariables:       variable.DiffRequest{},
		EnvironmentVariableAliases: variable.DiffRequest{},
		Secrets:                    secret.DiffRequest{},
		SecretAliases:              secret.DiffRequest{},
	}

	invalidRequest := project.UpsertServiceRequest{
		ProjectUpsertRequest: project.UpsertRepositoryRequest{
			Name:        "", // Invalid: empty name
			Description: nil,
		},
		EnvironmentVariables:       variable.DiffRequest{},
		EnvironmentVariableAliases: variable.DiffRequest{},
		Secrets:                    secret.DiffRequest{},
		SecretAliases:              secret.DiffRequest{},
	}

	updatedProject := &project.Project{
		ID:             uuid.MustParse(validProjectID),
		OrganizationID: uuid.MustParse(gofakeit.UUID()),
		Name:           validRequest.ProjectUpsertRequest.Name,
		Description:    validRequest.ProjectUpsertRequest.Description,
	}

	testCases := []struct {
		TestName      string
		ProjectID     string
		Request       project.UpsertServiceRequest
		SetupMocks    func(*repoMocks.ProjectRepository, *serviceMocks.VariableService, *serviceMocks.SecretService)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:   "error_empty_project_id",
			ProjectID:  emptyProjectID,
			Request:    validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:   "error_invalid_project_id",
			ProjectID:  invalidProjectID,
			Request:    validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:   "error_invalid_request",
			ProjectID:  validProjectID,
			Request:    invalidRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "error_repository_update_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(nil, errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "error_variable_service_update_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(updatedProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, errors.New("variable service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "error_secret_service_update_failure",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(updatedProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, errors.New("secret service error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "error_variable_service_list_failure_in_refresh",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(updatedProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, updatedProject.ID.String()).
					Return(nil, errors.New("variable list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "error_secret_service_list_failure_in_refresh",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(updatedProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, updatedProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, updatedProject.ID.String()).
					Return(nil, errors.New("secret list error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to update project",
		},
		{
			TestName:  "success",
			ProjectID: validProjectID,
			Request:   validRequest,
			SetupMocks: func(pr *repoMocks.ProjectRepository, vs *serviceMocks.VariableService, ss *serviceMocks.SecretService) {
				pr.EXPECT().
					Update(mock.Anything, validProjectID, validRequest.ProjectUpsertRequest).
					Return(updatedProject, nil)

				emptyDiffRequest := variable.DiffRequest{
					Create: []variable.DiffCreateRequest{},
					Update: []variable.DiffUpdateRequest{},
					Delete: []variable.DiffDeleteRequest{},
				}
				emptySecretDiffRequest := secret.DiffRequest{
					Create: []secret.DiffCreateRequest{},
					Update: []secret.DiffUpdateRequest{},
					Delete: []secret.DiffDeleteRequest{},
				}
				overridesAuthorizedScopes := make(map[variable.Scope]struct{})

				vs.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.EnvironmentVariables, validRequest.EnvironmentVariableAliases, emptyDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				ss.EXPECT().
					Update(mock.Anything, updatedProject.ID.String(), validRequest.Secrets, validRequest.SecretAliases, emptySecretDiffRequest, overridesAuthorizedScopes).
					Return(nil, nil)

				vs.EXPECT().
					List(mock.Anything, updatedProject.ID.String()).
					Return(variable.Variables{}, nil)

				ss.EXPECT().
					List(mock.Anything, updatedProject.ID.String()).
					Return(secret.Secrets{}, nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockProjectRepo := &repoMocks.ProjectRepository{}
			mockVariableService := &serviceMocks.VariableService{}
			mockSecretService := &serviceMocks.SecretService{}
			tc.SetupMocks(mockProjectRepo, mockVariableService, mockSecretService)

			service, err := NewProjectService(mockProjectRepo, mockVariableService, mockSecretService)
			require.NoError(t, err)

			result, err := service.Update(context.Background(), tc.ProjectID, tc.Request)

			if tc.ExpectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, updatedProject.ID, result.ID)
				assert.Equal(t, updatedProject.Name, result.Name)
			}

			mockProjectRepo.AssertExpectations(t)
			mockVariableService.AssertExpectations(t)
			mockSecretService.AssertExpectations(t)
		})
	}
}

func TestProjectService_Delete(t *testing.T) {
	t.Parallel()

	validProjectID := gofakeit.UUID()
	invalidProjectID := "invalid-uuid"
	emptyProjectID := ""

	testCases := []struct {
		TestName      string
		ProjectID     string
		SetupMocks    func(*repoMocks.ProjectRepository)
		ExpectError   bool
		ErrorContains string
	}{
		{
			TestName:      "error_empty_project_id",
			ProjectID:     emptyProjectID,
			SetupMocks:    func(pr *repoMocks.ProjectRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:      "error_invalid_project_id",
			ProjectID:     invalidProjectID,
			SetupMocks:    func(pr *repoMocks.ProjectRepository) {},
			ExpectError:   true,
			ErrorContains: "invalid project id param",
		},
		{
			TestName:  "error_repository_delete_failure",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository) {
				pr.EXPECT().
					Delete(mock.Anything, validProjectID).
					Return(errors.New("repository error"))
			},
			ExpectError:   true,
			ErrorContains: "failed to delete project",
		},
		{
			TestName:  "success",
			ProjectID: validProjectID,
			SetupMocks: func(pr *repoMocks.ProjectRepository) {
				pr.EXPECT().
					Delete(mock.Anything, validProjectID).
					Return(nil)
			},
			ExpectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockProjectRepo := &repoMocks.ProjectRepository{}
			mockVariableService := &serviceMocks.VariableService{}
			mockSecretService := &serviceMocks.SecretService{}
			tc.SetupMocks(mockProjectRepo)

			service, err := NewProjectService(mockProjectRepo, mockVariableService, mockSecretService)
			require.NoError(t, err)

			err = service.Delete(context.Background(), tc.ProjectID)

			if tc.ExpectError {
				assert.Error(t, err)
				if tc.ErrorContains != "" {
					assert.Contains(t, err.Error(), tc.ErrorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockProjectRepo.AssertExpectations(t)
		})
	}
}
