//go:build unit
// +build unit

package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	mock_service "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deployment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	mock_repository "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

type ContainerServiceTestSuite struct {
	suite.Suite

	repository        *mock_repository.ContainerRepository
	variableService   *mock_service.VariableService
	secretService     *mock_service.SecretService
	deploymentService *mock_service.DeploymentService
	service           container.Service
}

func (ts *ContainerServiceTestSuite) SetupTest() {
	t := ts.T()

	// Initialize deployment, variable & secret service
	deploymentService := mock_service.NewDeploymentService(t)
	variableService := mock_service.NewVariableService(t)
	secretService := mock_service.NewSecretService(t)

	// Initialize container repository
	containerRepository := mock_repository.NewContainerRepository(t)

	// Initialize container service
	containerService, err := services.NewContainerService(containerRepository, deploymentService, variableService, secretService)
	require.NoError(t, err)
	require.NotNil(t, containerService)

	ts.repository = containerRepository
	ts.service = containerService
	ts.deploymentService = deploymentService
	ts.variableService = variableService
	ts.secretService = secretService

}

func TestContainerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerServiceTestSuite))
}

func (ts *ContainerServiceTestSuite) TestNew_FailWithInvalidRepository() {
	t := ts.T()

	contService, err := services.NewContainerService(nil, ts.deploymentService, ts.variableService, ts.secretService)
	assert.Nil(t, contService)
	assert.ErrorContains(t, err, services.ErrInvalidRepository.Error())
}

func (ts *ContainerServiceTestSuite) TestNew_FailWithInvalidService() {
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
			contService, err := services.NewContainerService(ts.repository, tc.DeploymentService, tc.VariableService, tc.SecretService)
			assert.Nil(t, contService)
			assert.ErrorContains(t, err, services.ErrInvalidService.Error())
		})
	}

}

func (ts *ContainerServiceTestSuite) TestNew_Success() {
	t := ts.T()

	contService, err := services.NewContainerService(ts.repository, ts.deploymentService, ts.variableService, ts.secretService)
	assert.Nil(t, err)
	assert.NotNil(t, contService)
}

func (ts *ContainerServiceTestSuite) TestCreate_FailWithInvalidEnvironmentID() {
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
			cont, err := ts.service.Create(context.Background(), tc.EnvironmentID, assertNewContainerUpsertServiceRequest(t))
			assert.Nil(t, cont)
			assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidEnvironmentIDParam.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestCreate_FailWithInvalidCreateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		CreateRequest container.UpsertServiceRequest
	}{
		{
			TestName: "invalid_container_upsert_repository_request",
			CreateRequest: container.UpsertServiceRequest{
				ContainerUpsertRequest: container.UpsertRepositoryRequest{},
				EnvironmentVariables:   assertNewVariableDiffRequest(t),
			},
		},
		{
			TestName: "invalid_container_environment_variable_request",
			CreateRequest: container.UpsertServiceRequest{
				ContainerUpsertRequest: assertNewContainerUpsertRepositoryRequest(t),
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
			cont, err := ts.service.Create(context.Background(), gofakeit.UUID(), tc.CreateRequest)
			assert.Nil(t, cont)
			assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestCreate_FailedToCreateContainer() {
	t := ts.T()

	environmentID := gofakeit.UUID()
	createRequest := assertNewContainerUpsertServiceRequest(t)

	ts.repository.EXPECT().
		Create(mock.Anything, environmentID, createRequest.ContainerUpsertRequest).
		Return(nil, container.ErrInvalidContainer)

	cont, err := ts.service.Create(context.Background(), environmentID, createRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
	assert.ErrorContains(t, err, container.ErrInvalidContainer.Error())
}

func (ts *ContainerServiceTestSuite) TestCreate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	createRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedContainer.EnvironmentID.String(), createRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.EnvironmentVariables).
		Return(nil, variable.ErrInvalidVariable)

	cont, err := ts.service.Create(context.Background(), expectedContainer.EnvironmentID.String(), createRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
	assert.ErrorContains(t, err, variable.ErrInvalidVariable.Error())
}

func (ts *ContainerServiceTestSuite) TestCreate_FailedToUpdateSecrets() {
	t := ts.T()

	createRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedContainer.EnvironmentID.String(), createRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)
	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.Secrets).
		Return(nil, secret.ErrInvalidSecret)

	cont, err := ts.service.Create(context.Background(), expectedContainer.EnvironmentID.String(), createRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
	assert.ErrorContains(t, err, secret.ErrInvalidSecret.Error())
}

func (ts *ContainerServiceTestSuite) TestCreate_FailedToDeploy() {
	t := ts.T()

	createRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedContainer.EnvironmentID.String(), createRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)
	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.Secrets).
		Return(assertCreateSecrets(t), nil)
	ts.deploymentService.EXPECT().
		Deploy(mock.Anything, expectedContainer.ID.String(), expectedContainer.Tag).
		Return(nil, deployment.ErrFailedToDeploy)

	cont, err := ts.service.Create(context.Background(), expectedContainer.EnvironmentID.String(), createRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
	assert.ErrorContains(t, err, deployment.ErrFailedToDeploy.Error())
}

func (ts *ContainerServiceTestSuite) TestCreate_Success() {
	t := ts.T()

	newContainer := assertCreateContainer(t)
	createRequest := assertNewContainerUpsertServiceRequest(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)
	expectedStatus := assertCreateStatus(t)

	expectedContainer := *newContainer
	err := expectedContainer.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetState(expectedStatus.State)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Create(mock.Anything, expectedContainer.EnvironmentID.String(), createRequest.ContainerUpsertRequest).
		Return(newContainer, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.EnvironmentVariables).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), createRequest.Secrets).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedSecrets, nil)

	ts.deploymentService.EXPECT().
		Deploy(mock.Anything, expectedContainer.ID.String(), expectedContainer.Tag).
		Return(expectedStatus, nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, expectedContainer.ID.String()).
		Return(expectedStatus, nil)

	cont, err := ts.service.Create(context.Background(), expectedContainer.EnvironmentID.String(), createRequest)
	assert.Nil(t, err)
	assertEqualContainer(t, &expectedContainer, cont)
}

func (ts *ContainerServiceTestSuite) TestGet_FailWithInvalidContainerID() {
	t := ts.T()

	testCases := []struct {
		TestName    string
		ContainerID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:    "invalid_uuid",
			ContainerID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			cont, err := ts.service.Get(context.Background(), tc.ContainerID)
			assert.Nil(t, cont)
			assert.ErrorContains(t, err, container.ErrFailedToGetContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidContainerIDParam.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestGet_FailContainerNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Get(mock.Anything, fakeID).
		Return(nil, container.ErrInvalidContainer)

	cont, err := ts.service.Get(context.Background(), fakeID)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToGetContainer.Error())
	assert.ErrorContains(t, err, container.ErrInvalidContainer.Error())
}

func (ts *ContainerServiceTestSuite) TestGet_Success() {
	t := ts.T()

	expectedContainer := assertCreateContainer(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)
	expectedStatus := assertCreateStatus(t)

	err := expectedContainer.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetState(expectedStatus.State)
	require.NoError(ts.T(), err)

	ts.variableService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedVariables, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedSecrets, nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, expectedContainer.ID.String()).
		Return(expectedStatus, nil)
	ts.repository.EXPECT().
		Get(mock.Anything, expectedContainer.ID.String()).
		Return(expectedContainer, nil)

	cont, err := ts.service.Get(context.Background(), expectedContainer.ID.String())
	assert.Nil(t, err)
	assertEqualContainer(t, expectedContainer, cont)
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailWithInvalidContainerID() {
	t := ts.T()

	testCases := []struct {
		TestName    string
		ContainerID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:    "invalid_uuid",
			ContainerID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			cont, err := ts.service.Update(context.Background(), tc.ContainerID, assertNewContainerUpsertServiceRequest(t))
			assert.Nil(t, cont)
			assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidContainerIDParam.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailContainerNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Update(mock.Anything, fakeID, mock.Anything).
		Return(nil, container.ErrInvalidContainer)

	cont, err := ts.service.Update(context.Background(), fakeID, assertNewContainerUpsertServiceRequest(t))
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
	assert.ErrorContains(t, err, container.ErrInvalidContainer.Error())
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailWithInvalidUpdateRequest() {
	t := ts.T()

	testCases := []struct {
		TestName      string
		UpdateRequest container.UpsertServiceRequest
	}{
		{
			TestName: "invalid_container_upsert_repository_request",
			UpdateRequest: container.UpsertServiceRequest{
				ContainerUpsertRequest: container.UpsertRepositoryRequest{},
				EnvironmentVariables:   assertNewVariableDiffRequest(t),
			},
		},
		{
			TestName: "invalid_container_environment_variable_request",
			UpdateRequest: container.UpsertServiceRequest{
				ContainerUpsertRequest: assertNewContainerUpsertRepositoryRequest(t),
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
			cont, err := ts.service.Update(context.Background(), gofakeit.UUID(), tc.UpdateRequest)
			assert.Nil(t, cont)
			assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidUpsertRequest.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailedToUpdateEnvironmentVariables() {
	t := ts.T()

	updateRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.EnvironmentVariables).
		Return(nil, variable.ErrInvalidVariable)

	cont, err := ts.service.Update(context.Background(), expectedContainer.ID.String(), updateRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
	assert.ErrorContains(t, err, variable.ErrInvalidVariable.Error())
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailedToUpdateSecrets() {
	t := ts.T()

	updateRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)
	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.Secrets).
		Return(nil, secret.ErrInvalidSecret)

	cont, err := ts.service.Update(context.Background(), expectedContainer.ID.String(), updateRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
	assert.ErrorContains(t, err, secret.ErrInvalidSecret.Error())
}

func (ts *ContainerServiceTestSuite) TestUpdate_FailedToDeploy() {
	t := ts.T()

	updateRequest := assertNewContainerUpsertServiceRequest(t)
	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.ContainerUpsertRequest).
		Return(expectedContainer, nil)
	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.EnvironmentVariables).
		Return(assertCreateVariables(t), nil)
	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.Secrets).
		Return(assertCreateSecrets(t), nil)
	ts.deploymentService.EXPECT().
		Deploy(mock.Anything, expectedContainer.ID.String(), expectedContainer.Tag).
		Return(nil, deployment.ErrFailedToDeploy)

	cont, err := ts.service.Update(context.Background(), expectedContainer.ID.String(), updateRequest)
	assert.Nil(t, cont)
	assert.ErrorContains(t, err, container.ErrFailedToUpdateContainer.Error())
	assert.ErrorContains(t, err, deployment.ErrFailedToDeploy.Error())
}

func (ts *ContainerServiceTestSuite) TestUpdate_Success() {
	t := ts.T()

	updatedContainer := assertCreateContainer(t)
	updateRequest := assertNewContainerUpsertServiceRequest(t)
	expectedVariables := assertCreateVariables(t)
	expectedSecrets := assertCreateSecrets(t)
	expectedStatus := assertCreateStatus(t)

	expectedContainer := *updatedContainer
	err := expectedContainer.SetEnvironmentVariables(expectedVariables)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetSecrets(expectedSecrets)
	require.NoError(ts.T(), err)
	err = expectedContainer.SetState(expectedStatus.State)
	require.NoError(ts.T(), err)

	ts.repository.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.ContainerUpsertRequest).
		Return(updatedContainer, nil)

	ts.variableService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.EnvironmentVariables).
		Return(expectedVariables, nil)
	ts.variableService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedVariables, nil)

	ts.secretService.EXPECT().
		Update(mock.Anything, expectedContainer.ID.String(), updateRequest.Secrets).
		Return(expectedSecrets, nil)
	ts.secretService.EXPECT().
		List(mock.Anything, expectedContainer.ID.String()).
		Return(expectedSecrets, nil)

	ts.deploymentService.EXPECT().
		Deploy(mock.Anything, expectedContainer.ID.String(), expectedContainer.Tag).
		Return(expectedStatus, nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, expectedContainer.ID.String()).
		Return(expectedStatus, nil)

	cont, err := ts.service.Update(context.Background(), expectedContainer.ID.String(), updateRequest)
	assert.Nil(t, err)
	assertEqualContainer(t, &expectedContainer, cont)
}

func (ts *ContainerServiceTestSuite) TestDelete_FailWithInvalidContainerID() {
	t := ts.T()

	testCases := []struct {
		TestName    string
		ContainerID string
	}{
		{
			TestName: "empty_string",
		},
		{
			TestName:    "invalid_uuid",
			ContainerID: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			err := ts.service.Delete(context.Background(), tc.ContainerID)
			assert.ErrorContains(t, err, container.ErrFailedToDeleteContainer.Error())
			assert.ErrorContains(t, err, container.ErrInvalidContainerIDParam.Error())
		})
	}
}

func (ts *ContainerServiceTestSuite) TestDelete_FailContainerNotFound() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID).
		Return(container.ErrInvalidContainer)

	err := ts.service.Delete(context.Background(), fakeID)
	assert.ErrorContains(t, err, container.ErrFailedToDeleteContainer.Error())
	assert.ErrorContains(t, err, container.ErrInvalidContainer.Error())
}

func (ts *ContainerServiceTestSuite) TestDelete_FailToGetContainerStatus() {
	t := ts.T()

	fakeID := gofakeit.UUID()

	ts.repository.EXPECT().
		Delete(mock.Anything, fakeID).
		Return(container.ErrInvalidContainer)

	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, fakeID).
		Return(nil, container.ErrInvalidContainer)

	err := ts.service.Delete(context.Background(), fakeID)
	assert.ErrorContains(t, err, container.ErrFailedToDeleteContainer.Error())
	assert.ErrorContains(t, err, container.ErrInvalidContainer.Error())
}

func (ts *ContainerServiceTestSuite) TestDelete_Success() {
	t := ts.T()

	expectedContainer := assertCreateContainer(t)

	ts.repository.EXPECT().
		Delete(mock.Anything, expectedContainer.ID.String()).
		Return(nil)
	ts.deploymentService.EXPECT().
		GetStatus(mock.Anything, expectedContainer.ID.String()).
		Return(nil, container.ErrInvalidContainer)

	err := ts.service.Delete(context.Background(), expectedContainer.ID.String())
	assert.Nil(t, err)
}

func assertCreateStatus(t *testing.T) *status.Status {
	stateIdx := gofakeit.IntRange(0, len(status.AllowedDesiredStateValues)-1)
	statusIdx := gofakeit.IntRange(0, len(status.AllowedServiceDeploymentStatusValues)-1)

	st, err := status.NewStatus(status.NewStatusParams{
		StatusID:                gofakeit.UUID(),
		State:                   status.AllowedDesiredStateValues[stateIdx].String(),
		ServiceDeploymentStatus: status.AllowedServiceDeploymentStatusValues[statusIdx].String(),
	})
	require.NoError(t, err)
	require.NotNil(t, st)
	require.NoError(t, st.Validate())

	return st
}

func assertCreateContainer(t *testing.T) *container.Container {
	cont, err := container.NewContainer(container.NewContainerParams{
		ContainerID:         gofakeit.UUID(),
		EnvironmentID:       gofakeit.UUID(),
		RegistryID:          gofakeit.UUID(),
		Name:                gofakeit.Name(),
		ImageName:           gofakeit.Name(),
		Tag:                 gofakeit.AppVersion(),
		CPU:                 int32(gofakeit.IntRange(container.MinCPU, container.DefaultCPU)),
		Memory:              int32(gofakeit.IntRange(container.MinMemory, container.DefaultMemory)),
		MinRunningInstances: int32(gofakeit.IntRange(container.MinMinRunningInstances, container.DefaultMaxRunningInstances)),
		MaxRunningInstances: int32(gofakeit.IntRange(container.DefaultMaxRunningInstances, container.DefaultMaxRunningInstances*8)),
		AutoPreview:         gofakeit.Bool(),
	})
	require.NoError(t, err)
	require.NotNil(t, cont)
	require.NoError(t, cont.Validate())

	return cont
}

func assertNewContainerUpsertServiceRequest(t *testing.T) container.UpsertServiceRequest {
	stateIdx := gofakeit.IntRange(0, len(status.AllowedDesiredStateValues)-1)

	req := container.UpsertServiceRequest{
		ContainerUpsertRequest: assertNewContainerUpsertRepositoryRequest(t),
		EnvironmentVariables:   assertNewVariableDiffRequest(t),
		Secrets:                assertNewSecretDiffRequest(t),
		DesiredState:           status.AllowedDesiredStateValues[stateIdx],
	}
	require.NoError(t, req.Validate())

	return req
}

func assertNewContainerUpsertRepositoryRequest(t *testing.T) container.UpsertRepositoryRequest {
	req := container.UpsertRepositoryRequest{
		RegistryID: gofakeit.UUID(),
		Name:       gofakeit.Name(),
		ImageName:  gofakeit.Name(),
		Tag:        gofakeit.AppVersion(),
	}
	require.NoError(t, req.Validate())

	return req
}

func assertEqualContainer(t *testing.T, expected *container.Container, actual *container.Container) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.EnvironmentID, actual.EnvironmentID)
	assert.Equal(t, expected.RegistryID, actual.RegistryID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ImageName, actual.ImageName)
	assert.Equal(t, expected.Tag, actual.Tag)
	assert.Equal(t, expected.CPU, actual.CPU)
	assert.Equal(t, expected.Memory, actual.Memory)
	assert.Equal(t, expected.MinRunningInstances, actual.MinRunningInstances)
	assert.Equal(t, expected.MaxRunningInstances, actual.MaxRunningInstances)
	assert.Equal(t, expected.AutoPreview, actual.AutoPreview)
	assert.Equal(t, expected.Entrypoint, actual.Entrypoint)
	assert.Equal(t, expected.Arguments, actual.Arguments)
	assert.Equal(t, expected.Storages, actual.Storages)
	assert.Equal(t, expected.Ports, actual.Ports)
	assert.Equal(t, expected.State, actual.State)
	assertEqualVariables(t, expected.EnvironmentVariables, actual.EnvironmentVariables)
	assertEqualVariables(t, expected.BuiltInEnvironmentVariables, actual.BuiltInEnvironmentVariables)
	assertEqualSecrets(t, expected.Secrets, actual.Secrets)
	assertEqualPorts(t, expected.Ports, actual.Ports)
	assertEqualStorages(t, expected.Storages, actual.Storages)

}

func assertEqualPort(t *testing.T, expected *port.Port, actual *port.Port) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.InternalPort, actual.InternalPort)
	assert.Equal(t, expected.PubliclyAccessible, actual.PubliclyAccessible)
	assert.Equal(t, expected.Protocol, actual.Protocol)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ExternalPort, actual.ExternalPort)
}

func assertEqualPorts(t *testing.T, expected port.Ports, actual port.Ports) {
	require.Len(t, expected, len(actual))

	actualByID := map[string]port.Port{}
	for _, v := range actual {
		actualByID[v.ID.String()] = v
	}

	for _, v := range expected {
		found, ok := actualByID[v.ID.String()]
		require.True(t, ok)
		assertEqualPort(t, &v, &found)
	}
}

func assertEqualStorage(t *testing.T, expected *storage.Storage, actual *storage.Storage) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Size, actual.Size)
	assert.Equal(t, expected.MountPoint, actual.MountPoint)
}

func assertEqualStorages(t *testing.T, expected storage.Storages, actual storage.Storages) {
	require.Len(t, expected, len(actual))

	actualByID := map[string]storage.Storage{}
	for _, v := range actual {
		actualByID[v.ID.String()] = v
	}

	for _, v := range expected {
		found, ok := actualByID[v.ID.String()]
		require.True(t, ok)
		assertEqualStorage(t, &v, &found)
	}
}
