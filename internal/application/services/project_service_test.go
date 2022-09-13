//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	mock_service "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	mock_repository "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type ProjectServiceTestSuite struct {
	suite.Suite

	repository      *mock_repository.ProjectRepository
	variableService *mock_service.VariableService
	secretService   *mock_service.SecretService
	projectService  project.Service
}

func (ts *ProjectServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize variable & secret service
	variableService := mock_service.NewVariableService(t)
	secretService := mock_service.NewSecretService(t)

	// Initialize project repository
	projRepository := mock_repository.NewProjectRepository(t)

	// Initialize project service
	projectService, err := services.NewProjectService(projRepository, variableService, secretService)
	require.NoError(t, err)
	require.NotNil(t, projectService)

	ts.repository = projRepository
	ts.projectService = projectService
	ts.variableService = variableService
	ts.secretService = secretService
}

func (ts *ProjectServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	projService, err := services.NewProjectService(nil, ts.variableService, ts.secretService)
	assert.Nil(t, projService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *ProjectServiceTestSuite) TestNew_FailWithInvalidService() {
	t := ts.T()

	testCases := []struct {
		TestName        string
		VariableService variable.Service
		SecretService   secret.Service
	}{
		{
			TestName:        "invalid_secret_service",
			VariableService: ts.variableService,
		},
		{
			TestName:      "invalid_variable_service",
			SecretService: ts.secretService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			projService, err := services.NewProjectService(ts.repository, tc.VariableService, tc.SecretService)
			assert.Nil(t, projService)
			assert.ErrorContains(t, err, services.ErrInvalidService.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestNew_Success() {
	t := ts.T()

	projService, err := services.NewProjectService(ts.repository, ts.variableService, ts.secretService)
	assert.Nil(t, err)
	assert.NotNil(t, projService)
}

func (ts *ProjectServiceTestSuite) TestCreate_FailWithInvalidOrganizationID() {
	t := ts.T()

	testCases := []struct {
		TestName       string
		OrganizationID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:       "invalid_uuid",
			OrganizationID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := ts.projectService.Create(context.Background(), tc.OrganizationID, assertNewProjectUpsertServiceRequest(t))
			assert.Nil(t, proj)
			assert.ErrorContains(t, err, project.ErrFailedToCreateProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidOrganizationIDParam.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest project.UpsertServiceRequest
	}{
		{
			TestName: "invalid_project_upsert_repository_request",
			CreateRequest: project.UpsertServiceRequest{
				ProjectUpsertRequest: project.UpsertRepositoryRequest{},
				EnvironmentVariables: assertNewVariableDiffRequest(t),
			},
		},
		{
			TestName: "invalid_project_environment_variable_request",
			CreateRequest: project.UpsertServiceRequest{
				ProjectUpsertRequest: assertNewProjectUpsertRepositoryRequest(t),
				EnvironmentVariables: variable.DiffRequest{
					Delete: []variable.DiffDeleteRequest{
						{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := ts.projectService.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, proj)
			assert.ErrorContains(t, err, project.ErrFailedToCreateProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestCreate_FailedToCreateProject() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	createRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, createRequest.ProjectUpsertRequest).
		Return(nil, errors.New(""))

	proj, err := ts.projectService.Create(context.Background(), organizationID, createRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToCreateProject.Error())
}

func (ts *ProjectServiceTestSuite) TestCreate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	expectedProject := assertCreateProject(t)
	createRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedProject.OrganizationID.String(), createRequest.ProjectUpsertRequest).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), createRequest.EnvironmentVariables).
		Return(nil, variable.ErrFailedToUpdateVariables)

	proj, err := ts.projectService.Create(context.Background(), expectedProject.OrganizationID.String(), createRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToCreateProject.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
}

func (ts *ProjectServiceTestSuite) TestCreate_FailedToUpdateSecrets() {
	t := ts.T()

	expectedProject := assertCreateProject(t)
	createRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedProject.OrganizationID.String(), createRequest.ProjectUpsertRequest).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), createRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(nil, secret.ErrFailedToUpdateSecrets)

	proj, err := ts.projectService.Create(context.Background(), expectedProject.OrganizationID.String(), createRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToCreateProject.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
}

func (ts *ProjectServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	newProject := assertCreateProject(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedProject := *newProject
	err := expectedProject.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedProject.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Create(mock.Anything, newProject.OrganizationID.String(), mock.Anything).
		Return(newProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedSecrets, nil)

	proj, err := ts.projectService.Create(context.Background(), expectedProject.OrganizationID.String(), assertNewProjectUpsertServiceRequest(t))
	assert.Nil(t, err)
	assertEqualProject(t, &expectedProject, proj)
}

func (ts *ProjectServiceTestSuite) TestGet_FailWithInvalidProjectID() {
	t := ts.T()

	testCases := []struct {
		TestName  string
		ProjectID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:  "invalid_uuid",
			ProjectID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := ts.projectService.Get(context.Background(), tc.ProjectID)
			assert.Nil(t, proj)
			assert.ErrorContains(t, err, project.ErrFailedToGetProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidProjectIDParam.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestGet_FailProjectNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID).
		Return(nil, errors.New(""))

	proj, err := ts.projectService.Get(context.Background(), fakeID)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToGetProject.Error())
}

func (ts *ProjectServiceTestSuite) TestGet_FailedToListEnvironmentVariables() {
	t := ts.T()

	expectedProject := assertCreateProject(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedProject.ID.String()).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(nil, variable.ErrFailedToListVariables)

	proj, err := ts.projectService.Get(context.Background(), expectedProject.ID.String())
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToGetProject.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToListVariables.Error())
}

func (ts *ProjectServiceTestSuite) TestGet_FailedToListSecret() {
	t := ts.T()

	expectedProject := assertCreateProject(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedProject.ID.String()).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(nil, secret.ErrFailedToListSecrets)

	proj, err := ts.projectService.Get(context.Background(), expectedProject.ID.String())
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToGetProject.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToListSecrets.Error())
}

func (ts *ProjectServiceTestSuite) TestGet_Success() {
	t := ts.T()

	newProject := assertCreateProject(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedProject := *newProject
	err := expectedProject.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedProject.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedSecrets, nil)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedProject.ID.String()).
		Return(newProject, nil)

	proj, err := ts.projectService.Get(context.Background(), expectedProject.ID.String())
	assert.Nil(t, err)
	assertEqualProject(t, &expectedProject, proj)
}

func (ts *ProjectServiceTestSuite) TestUpdate_FailWithInvalidProjectID() {
	t := ts.T()

	testCases := []struct {
		TestName  string
		ProjectID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:  "invalid_uuid",
			ProjectID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := ts.projectService.Update(context.Background(), tc.ProjectID, assertNewProjectUpsertServiceRequest(t))
			assert.Nil(t, proj)
			assert.ErrorContains(t, err, project.ErrFailedToUpdateProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidProjectIDParam.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestUpdate_FailProjectNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, updateRequest.ProjectUpsertRequest).
		Return(nil, errors.New(""))

	proj, err := ts.projectService.Update(context.Background(), fakeID, updateRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToUpdateProject.Error())
}

func (ts *ProjectServiceTestSuite) TestUpdate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	expectedProject := assertCreateProject(t)
	updateRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), updateRequest.ProjectUpsertRequest).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), updateRequest.EnvironmentVariables).
		Return(nil, variable.ErrFailedToUpdateVariables)

	proj, err := ts.projectService.Update(context.Background(), expectedProject.ID.String(), updateRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToUpdateProject.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
}

func (ts *ProjectServiceTestSuite) TestUpdate_FailedToUpdateSecrets() {
	t := ts.T()

	expectedProject := assertCreateProject(t)
	updateRequest := assertNewProjectUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), updateRequest.ProjectUpsertRequest).
		Return(expectedProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), updateRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(nil, secret.ErrFailedToUpdateSecrets)

	proj, err := ts.projectService.Update(context.Background(), expectedProject.ID.String(), updateRequest)
	assert.Nil(t, proj)
	assert.ErrorContains(t, err, project.ErrFailedToUpdateProject.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
}

func (ts *ProjectServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest project.UpsertServiceRequest
	}{
		{
			TestName: "invalid_project_upsert_request",
			UpdateRequest: project.UpsertServiceRequest{
				EnvironmentVariables: assertNewVariableDiffRequest(t),
			},
		},
		{
			TestName: "invalid_project_environment_variables",
			UpdateRequest: project.UpsertServiceRequest{
				ProjectUpsertRequest: assertNewProjectUpsertRepositoryRequest(t),
				EnvironmentVariables: variable.DiffRequest{
					Create: []variable.DiffCreateRequest{
						{},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			proj, err := ts.projectService.Update(context.Background(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, proj)
			assert.ErrorContains(t, err, project.ErrFailedToUpdateProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	newProject := assertCreateProject(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedProject := *newProject
	err := expectedProject.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedProject.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(newProject, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedProject.ID.String(), mock.Anything).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedProject.ID.String()).
		Return(expectedSecrets, nil)

	proj, err := ts.projectService.Update(context.Background(), expectedProject.ID.String(), assertNewProjectUpsertServiceRequest(t))
	assert.Nil(t, err)
	assertEqualProject(t, &expectedProject, proj)
}

func (ts *ProjectServiceTestSuite) TestDelete_FailWithInvalidProjectID() {
	t := ts.T()

	testCases := []struct {
		TestName  string
		ProjectID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:  "invalid_uuid",
			ProjectID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.projectService.Delete(context.Background(), tc.ProjectID)
			assert.ErrorContains(t, err, project.ErrFailedToDeleteProject.Error())
			assert.ErrorContains(t, err, project.ErrInvalidProjectIDParam.Error())
		})
	}
}

func (ts *ProjectServiceTestSuite) TestDelete_FailProjectNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID).
		Return(errors.New(""))

	err := ts.projectService.Delete(context.Background(), fakeID)
	assert.ErrorContains(t, err, project.ErrFailedToDeleteProject.Error())
}

func (ts *ProjectServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	expectedProject := assertCreateProject(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, expectedProject.ID.String()).
		Return(nil)

	err := ts.projectService.Delete(context.Background(), expectedProject.ID.String())
	assert.Nil(t, err)
}

func TestProjectServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

func assertCreateProject(t *testing.T) *project.Project {
	proj, err := project.NewProject(project.NewProjectParams{
		ProjectID:      gofakeit.UUID(),
		OrganizationID: gofakeit.UUID(),
		Name:           gofakeit.Name(),
		Description:    pointer.ToString(gofakeit.Word()),
	})
	require.NoError(t, err)
	require.NotNil(t, proj)
	require.NoError(t, proj.Validate())

	return proj
}

func assertNewProjectUpsertServiceRequest(t *testing.T) project.UpsertServiceRequest {
	req := project.UpsertServiceRequest{
		ProjectUpsertRequest: assertNewProjectUpsertRepositoryRequest(t),
		EnvironmentVariables: assertNewVariableDiffRequest(t),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewProjectUpsertRepositoryRequest(t *testing.T) project.UpsertRepositoryRequest {
	req := project.UpsertRepositoryRequest{
		Name:        gofakeit.Name(),
		Description: pointer.ToString(gofakeit.Word()),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertEqualProject(t *testing.T, expected *project.Project, actual *project.Project) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.OrganizationID, actual.OrganizationID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assertEqualVariables(t, expected.EnvironmentVariables, actual.EnvironmentVariables)
	assertEqualVariables(t, expected.BuiltInEnvironmentVariables, actual.BuiltInEnvironmentVariables)
	assertEqualSecrets(t, expected.Secrets, actual.Secrets)
}
