//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	mock_service "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	mock_repository "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type EnvironmentServiceTestSuite struct {
	suite.Suite

	repository         *mock_repository.EnvironmentRepository
	deploymentService  *mock_service.DeploymentService
	variableService    *mock_service.VariableService
	secretService      *mock_service.SecretService
	environmentService environment.Service
}

func (ts *EnvironmentServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize variable & secret service
	deploymentService := mock_service.NewDeploymentService(t)
	variableService := mock_service.NewVariableService(t)
	secretService := mock_service.NewSecretService(t)

	// Initialize environment repository
	envRepository := mock_repository.NewEnvironmentRepository(t)

	// Initialize environment service
	environmentService, err := services.NewEnvironmentService(envRepository, deploymentService, variableService, secretService)
	require.NoError(t, err)
	require.NotNil(t, environmentService)

	ts.repository = envRepository
	ts.deploymentService = deploymentService
	ts.environmentService = environmentService
	ts.variableService = variableService
	ts.secretService = secretService
}

func (ts *EnvironmentServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	envService, err := services.NewEnvironmentService(nil, ts.deploymentService, ts.variableService, ts.secretService)
	assert.Nil(t, envService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *EnvironmentServiceTestSuite) TestNew_FailWithInvalidService() {
	t := ts.T()

	testCases := []struct {
		TestName          string
		DeploymentService deployment.Service
		VariableService   variable.Service
		SecretService     secret.Service
	}{
		{
			TestName:        "invalid_deployment_service",
			VariableService: ts.variableService,
			SecretService:   ts.secretService,
		},
		{
			TestName:          "invalid_variable_service",
			DeploymentService: ts.deploymentService,
			SecretService:     ts.secretService,
		},
		{
			TestName:          "invalid_secret_service",
			DeploymentService: ts.deploymentService,
			VariableService:   ts.variableService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			envService, err := services.NewEnvironmentService(ts.repository, tc.DeploymentService, tc.VariableService, tc.SecretService)
			assert.Nil(t, envService)
			assert.ErrorContains(t, err, services.ErrInvalidService.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestNew_Success() {
	t := ts.T()

	envService, err := services.NewEnvironmentService(ts.repository, ts.deploymentService, ts.variableService, ts.secretService)
	assert.Nil(t, err)
	assert.NotNil(t, envService)
}

func (ts *EnvironmentServiceTestSuite) TestCreate_FailWithInvalidProjectID() {
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
			env, err := ts.environmentService.Create(context.Background(), tc.ProjectID, assertNewEnvironmentCreateServiceRequest(t))
			assert.Nil(t, env)
			assert.ErrorContains(t, err, environment.ErrFailedToCreateEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidProjectIDParam.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest environment.CreateServiceRequest
	}{
		{
			TestName: "invalid_environment_upsert_repository_request",
			CreateRequest: environment.CreateServiceRequest{
				EnvironmentCreateRequest: environment.CreateRepositoryRequest{},
				EnvironmentVariables:     assertNewVariableDiffRequest(t),
			},
		},
		{
			TestName: "invalid_environment_environment_variable_request",
			CreateRequest: environment.CreateServiceRequest{
				EnvironmentCreateRequest: assertNewEnvironmentCreateRepositoryRequest(t),
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
			env, err := ts.environmentService.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, env)
			assert.ErrorContains(t, err, environment.ErrFailedToCreateEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidCreateRequest.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestCreate_FailedToCreateEnvironment() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	createRequest := assertNewEnvironmentCreateServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, createRequest.EnvironmentCreateRequest).
		Return(nil, environment.ErrInvalidEnvironment)

	env, err := ts.environmentService.Create(context.Background(), organizationID, createRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToCreateEnvironment.Error())
	assert.ErrorContains(t, err, environment.ErrInvalidEnvironment.Error())
}

func (ts *EnvironmentServiceTestSuite) TestCreate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	expectedEnvironment := assertCreateEnvironment(t)
	createRequest := assertNewEnvironmentCreateServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, createRequest.EnvironmentCreateRequest).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), createRequest.EnvironmentVariables).
		Return(nil, variable.ErrFailedToUpdateVariables)

	env, err := ts.environmentService.Create(context.Background(), organizationID, createRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToCreateEnvironment.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
}

func (ts *EnvironmentServiceTestSuite) TestCreate_FailedToUpdateSecrets() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	expectedEnvironment := assertCreateEnvironment(t)
	createRequest := assertNewEnvironmentCreateServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, createRequest.EnvironmentCreateRequest).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), createRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(nil, secret.ErrFailedToUpdateSecrets)

	env, err := ts.environmentService.Create(context.Background(), organizationID, createRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToCreateEnvironment.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
}

func (ts *EnvironmentServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	organizationID := gofakeit.UUID()
	newEnvironment := assertCreateEnvironment(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedEnvironment := *newEnvironment
	err := expectedEnvironment.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedEnvironment.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Create(mock.Anything, organizationID, mock.Anything).
		Return(newEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedSecrets, nil)

	env, err := ts.environmentService.Create(context.Background(), organizationID, assertNewEnvironmentCreateServiceRequest(t))
	assert.Nil(t, err)
	assertEqualEnvironment(t, &expectedEnvironment, env)
}

func (ts *EnvironmentServiceTestSuite) TestGet_FailWithInvalidEnvironmentID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		EnvironmentID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			EnvironmentID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			env, err := ts.environmentService.Get(context.Background(), tc.EnvironmentID)
			assert.Nil(t, env)
			assert.ErrorContains(t, err, environment.ErrFailedToGetEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidEnvironmentIDParam.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestGet_FailEnvironmentNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID).
		Return(nil, environment.ErrInvalidEnvironment)

	env, err := ts.environmentService.Get(context.Background(), fakeID)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToGetEnvironment.Error())
	assert.ErrorContains(t, err, environment.ErrInvalidEnvironment.Error())
}

func (ts *EnvironmentServiceTestSuite) TestGet_FailedToListEnvironmentVariables() {
	t := ts.T()

	expectedEnvironment := assertCreateEnvironment(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(nil, variable.ErrFailedToListVariables)

	env, err := ts.environmentService.Get(context.Background(), expectedEnvironment.ID.String())
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToGetEnvironment.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToListVariables.Error())
}

func (ts *EnvironmentServiceTestSuite) TestGet_FailedToListSecret() {
	t := ts.T()

	expectedEnvironment := assertCreateEnvironment(t)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(nil, secret.ErrFailedToListSecrets)

	env, err := ts.environmentService.Get(context.Background(), expectedEnvironment.ID.String())
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToGetEnvironment.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToListSecrets.Error())
}

func (ts *EnvironmentServiceTestSuite) TestGet_Success() {
	t := ts.T()

	newEnvironment := assertCreateEnvironment(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedEnvironment := *newEnvironment
	err := expectedEnvironment.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedEnvironment.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedSecrets, nil)

	ts.repository.EXPECT().
		Get(mock.Anything, expectedEnvironment.ID.String()).
		Return(newEnvironment, nil)

	env, err := ts.environmentService.Get(context.Background(), expectedEnvironment.ID.String())
	assert.Nil(t, err)
	assertEqualEnvironment(t, &expectedEnvironment, env)
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_FailWithInvalidEnvironmentID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		EnvironmentID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			EnvironmentID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			env, err := ts.environmentService.Update(context.Background(), tc.EnvironmentID, assertNewEnvironmentUpdateServiceRequest(t))
			assert.Nil(t, env)
			assert.ErrorContains(t, err, environment.ErrFailedToUpdateEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidEnvironmentIDParam.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_FailEnvironmentNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()
	updateRequest := assertNewEnvironmentUpdateServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, updateRequest.EnvironmentUpdateRequest).
		Return(nil, environment.ErrInvalidEnvironment)

	env, err := ts.environmentService.Update(context.Background(), fakeID, updateRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToUpdateEnvironment.Error())
	assert.ErrorContains(t, err, environment.ErrInvalidEnvironment.Error())
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	expectedEnvironment := assertCreateEnvironment(t)
	updateRequest := assertNewEnvironmentUpdateServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), updateRequest.EnvironmentUpdateRequest).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), updateRequest.EnvironmentVariables).
		Return(nil, variable.ErrFailedToUpdateVariables)

	env, err := ts.environmentService.Update(context.Background(), expectedEnvironment.ID.String(), updateRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToUpdateEnvironment.Error())
	assert.ErrorContains(t, err, variable.ErrFailedToUpdateVariables.Error())
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_FailedToUpdateSecrets() {
	t := ts.T()

	expectedEnvironment := assertCreateEnvironment(t)
	updateRequest := assertNewEnvironmentUpdateServiceRequest(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), updateRequest.EnvironmentUpdateRequest).
		Return(expectedEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), updateRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(nil, secret.ErrFailedToUpdateSecrets)

	env, err := ts.environmentService.Update(context.Background(), expectedEnvironment.ID.String(), updateRequest)
	assert.Nil(t, env)
	assert.ErrorContains(t, err, environment.ErrFailedToUpdateEnvironment.Error())
	assert.ErrorContains(t, err, secret.ErrFailedToUpdateSecrets.Error())
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest environment.UpdateServiceRequest
	}{
		{
			TestName: "invalid_environment_environment_variables",
			UpdateRequest: environment.UpdateServiceRequest{
				EnvironmentUpdateRequest: assertNewEnvironmentUpdateRepositoryRequest(t),
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
			env, err := ts.environmentService.Update(context.Background(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, env)
			assert.ErrorContains(t, err, environment.ErrFailedToUpdateEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidUpdateRequest.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	newEnvironment := assertCreateEnvironment(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)

	expectedEnvironment := *newEnvironment
	err := expectedEnvironment.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedEnvironment.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(newEnvironment, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedEnvironment.ID.String(), mock.Anything).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedEnvironment.ID.String()).
		Return(expectedSecrets, nil)

	env, err := ts.environmentService.Update(context.Background(), expectedEnvironment.ID.String(), assertNewEnvironmentUpdateServiceRequest(t))
	assert.Nil(t, err)
	assertEqualEnvironment(t, &expectedEnvironment, env)
}

func (ts *EnvironmentServiceTestSuite) TestDelete_FailWithInvalidEnvironmentID() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		EnvironmentID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:      "invalid_uuid",
			EnvironmentID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.environmentService.Delete(context.Background(), tc.EnvironmentID)
			assert.ErrorContains(t, err, environment.ErrFailedToDeleteEnvironment.Error())
			assert.ErrorContains(t, err, environment.ErrInvalidEnvironmentIDParam.Error())
		})
	}
}

func (ts *EnvironmentServiceTestSuite) TestDelete_FailEnvironmentNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID).
		Return(environment.ErrInvalidEnvironment)

	err := ts.environmentService.Delete(context.Background(), fakeID)
	assert.ErrorContains(t, err, environment.ErrFailedToDeleteEnvironment.Error())
	assert.ErrorContains(t, err, environment.ErrInvalidEnvironment.Error())
}

func (ts *EnvironmentServiceTestSuite) TestDelete_FailToGetContainerStatus() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID).
		Return(nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, fakeID).
		Return(nil, environment.ErrInvalidEnvironment)

	err := ts.environmentService.Delete(context.Background(), fakeID)
	assert.ErrorContains(t, err, environment.ErrFailedToDeleteEnvironment.Error())
	assert.ErrorContains(t, err, environment.ErrInvalidEnvironment.Error())
}

func (ts *EnvironmentServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	expectedEnvironment := assertCreateEnvironment(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, expectedEnvironment.ID.String()).
		Return(nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, expectedEnvironment.ID.String()).
		Return(nil, apierrors.NewNotFoundApiError(apierrors.ApiResourceEnvironment, expectedEnvironment.ID.String()))

	err := ts.environmentService.Delete(context.Background(), expectedEnvironment.ID.String())
	assert.Nil(t, err)
}

func TestEnvironmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentServiceTestSuite))
}

func assertCreateEnvironment(t *testing.T) *environment.Environment {
	env, err := environment.NewEnvironment(environment.NewEnvironmentParams{
		EnvironmentID: gofakeit.UUID(),
		ClusterID:     gofakeit.UUID(),
		ProjectID:     gofakeit.UUID(),
		Name:          gofakeit.Name(),
		Mode:          environment.ModeDevelopment.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, env)
	require.NoError(t, env.Validate())

	return env
}

func assertNewEnvironmentCreateServiceRequest(t *testing.T) environment.CreateServiceRequest {
	req := environment.CreateServiceRequest{
		EnvironmentCreateRequest: assertNewEnvironmentCreateRepositoryRequest(t),
		EnvironmentVariables:     assertNewVariableDiffRequest(t),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewEnvironmentUpdateServiceRequest(t *testing.T) environment.UpdateServiceRequest {
	req := environment.UpdateServiceRequest{
		EnvironmentUpdateRequest: assertNewEnvironmentUpdateRepositoryRequest(t),
		EnvironmentVariables:     assertNewVariableDiffRequest(t),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewEnvironmentCreateRepositoryRequest(t *testing.T) environment.CreateRepositoryRequest {
	req := environment.CreateRepositoryRequest{
		Name:      gofakeit.Name(),
		ClusterID: pointer.To(gofakeit.UUID()),
		Mode:      pointer.To(environment.ModeDevelopment),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewEnvironmentUpdateRepositoryRequest(t *testing.T) environment.UpdateRepositoryRequest {
	req := environment.UpdateRepositoryRequest{
		Name: pointer.ToString(gofakeit.Name()),
		Mode: pointer.To(environment.ModeDevelopment),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertEqualEnvironment(t *testing.T, expected *environment.Environment, actual *environment.Environment) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.ProjectID, actual.ProjectID)
	assert.Equal(t, expected.ClusterID, actual.ClusterID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Mode, actual.Mode)
	assertEqualVariables(t, expected.EnvironmentVariables, actual.EnvironmentVariables)
	assertEqualVariables(t, expected.BuiltInEnvironmentVariables, actual.BuiltInEnvironmentVariables)
	assertEqualSecrets(t, expected.Secrets, actual.Secrets)
}
