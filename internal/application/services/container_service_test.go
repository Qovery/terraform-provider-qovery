//go:build unit

package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	servicemocks "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	repomocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newValidContainerUpsertServiceRequest() container.UpsertServiceRequest {
	return container.UpsertServiceRequest{
		ContainerUpsertRequest: container.UpsertRepositoryRequest{
			RegistryID: uuid.New().String(),
			Name:       "worker",
			ImageName:  "xfunctional/imhotep/backend",
			Tag:        "latest",
		},
	}
}

func newCreatedContainer(environmentID uuid.UUID) *container.Container {
	return &container.Container{
		ID:                  uuid.New(),
		EnvironmentID:       environmentID,
		RegistryID:          uuid.New(),
		Name:                "worker",
		ImageName:           "xfunctional/imhotep/backend",
		Tag:                 "latest",
		CPU:                 500,
		Memory:              512,
		MinRunningInstances: 1,
		MaxRunningInstances: 1,
	}
}

// A container create is not atomic: the Qovery API creates the container first, then
// the provider pushes variables, secrets and external secrets onto it. When one of
// those follow-up calls fails (for instance a 409 because the variable already exists
// at environment scope), the container exists in Qovery. The service must hand it back
// so the resource layer can write its ID to the Terraform state; returning nil orphans
// the container and every later apply fails with "a container named X already exists".
func TestContainerService_Create_ReturnsCreatedContainerOnPartialFailure(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedContainer(environmentID)
	variableErr := errors.New("409 Conflict - Variable already exists: AWS_REGION")

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, variableErr)

	service, err := services.NewContainerService(
		containerRepository,
		&servicemocks.DeploymentService{},
		variableService,
		&servicemocks.SecretService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	cont, err := service.Create(context.Background(), environmentID.String(), newValidContainerUpsertServiceRequest())

	assert.ErrorContains(t, err, variableErr.Error())
	require.NotNil(t, cont, "the container was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, cont.ID)
}

// Nothing was created when the create call itself fails, so there is nothing to put in
// the state and the service must keep returning nil.
func TestContainerService_Create_ReturnsNilWhenCreateFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	createErr := errors.New("403 Forbidden")

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(nil, createErr)

	service, err := services.NewContainerService(
		containerRepository,
		&servicemocks.DeploymentService{},
		&servicemocks.VariableService{},
		&servicemocks.SecretService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	cont, err := service.Create(context.Background(), environmentID.String(), newValidContainerUpsertServiceRequest())

	assert.ErrorContains(t, err, createErr.Error())
	assert.Nil(t, cont)
}

// The repository create is itself a create-then-configure sequence: the container may
// already exist in Qovery when the repository returns an error. The service must pass
// that container through rather than flattening it to nil.
func TestContainerService_Create_PropagatesPartialContainerFromRepository(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedContainer(environmentID)
	repositoryErr := errors.New("500 Internal Server Error on deployment stage")

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, repositoryErr)

	service, err := services.NewContainerService(
		containerRepository,
		&servicemocks.DeploymentService{},
		&servicemocks.VariableService{},
		&servicemocks.SecretService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	cont, err := service.Create(context.Background(), environmentID.String(), newValidContainerUpsertServiceRequest())

	assert.ErrorContains(t, err, repositoryErr.Error())
	require.NotNil(t, cont, "the repository reported the container as created, the service must pass it through")
	assert.Equal(t, created.ID, cont.ID)
}

// The refresh is the last post-create step. It runs after variables and secrets have
// been written, so a failure there must not drop the container either.
func TestContainerService_Create_ReturnsCreatedContainerWhenRefreshFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedContainer(environmentID)
	refreshErr := errors.New("503 Service Unavailable while listing variables")

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	variableService.On("List", mock.Anything, created.ID.String()).
		Return(nil, refreshErr)

	secretService := &servicemocks.SecretService{}
	secretService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	service, err := services.NewContainerService(
		containerRepository,
		&servicemocks.DeploymentService{},
		variableService,
		secretService,
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	cont, err := service.Create(context.Background(), environmentID.String(), newValidContainerUpsertServiceRequest())

	assert.ErrorContains(t, err, refreshErr.Error())
	require.NotNil(t, cont, "variables and secrets were already written, the container must still reach the state")
	assert.Equal(t, created.ID, cont.ID)
}
