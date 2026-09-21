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
	"github.com/qovery/terraform-provider-qovery/internal/domain/terraformservice"
	repomocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newValidTerraformServiceUpsertServiceRequest() terraformservice.UpsertServiceRequest {
	return terraformservice.UpsertServiceRequest{
		TerraformServiceUpsertRequest: terraformservice.UpsertRepositoryRequest{
			Name: "tf-template-for-imhotep-staging",
			GitRepository: terraformservice.GitRepository{
				URL:      "https://github.com/org/repo",
				Branch:   "main",
				RootPath: "/",
			},
			Backend:       terraformservice.Backend{Kubernetes: &terraformservice.KubernetesBackend{}},
			Engine:        terraformservice.EngineTerraform,
			EngineVersion: terraformservice.EngineVersion{ExplicitVersion: "1.5.7"},
			JobResources:  terraformservice.JobResources{CPUMilli: 1000, RAMMiB: 1024, StorageGiB: 20},
		},
	}
}

func newCreatedTerraformService(environmentID uuid.UUID) *terraformservice.TerraformService {
	return &terraformservice.TerraformService{
		ID:            uuid.New(),
		EnvironmentID: environmentID,
		Name:          "tf-template-for-imhotep-staging",
	}
}

// Same partial-create hazard as containers, jobs and helm services: here the follow-up
// call is the external secrets read that runs right after the create.
func TestTerraformServiceService_Create_ReturnsCreatedServiceOnPartialFailure(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedTerraformService(environmentID)
	externalSecretErr := errors.New("500 Internal Server Error")

	terraformServiceRepository := &repomocks.TerraformServiceRepository{}
	terraformServiceRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	service, err := services.NewTerraformServiceService(
		terraformServiceRepository,
		stubExternalSecretRepository{listErr: externalSecretErr},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newService, err := service.Create(context.Background(), environmentID.String(), newValidTerraformServiceUpsertServiceRequest())

	assert.ErrorContains(t, err, externalSecretErr.Error())
	require.NotNil(t, newService, "the terraform service was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, newService.ID)
}

func TestTerraformServiceService_Create_ReturnsNilWhenCreateFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	createErr := errors.New("403 Forbidden")

	terraformServiceRepository := &repomocks.TerraformServiceRepository{}
	terraformServiceRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(nil, createErr)

	service, err := services.NewTerraformServiceService(
		terraformServiceRepository,
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newService, err := service.Create(context.Background(), environmentID.String(), newValidTerraformServiceUpsertServiceRequest())

	assert.ErrorContains(t, err, createErr.Error())
	assert.Nil(t, newService)
}

// The repository create is itself a create-then-configure sequence (deployment stage
// attach, advanced settings), so the terraform service may already exist in Qovery when
// the repository reports an error.
func TestTerraformServiceService_Create_PropagatesPartialServiceFromRepository(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedTerraformService(environmentID)
	repositoryErr := errors.New("500 Internal Server Error on deployment stage")

	terraformServiceRepository := &repomocks.TerraformServiceRepository{}
	terraformServiceRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, repositoryErr)

	service, err := services.NewTerraformServiceService(
		terraformServiceRepository,
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newService, err := service.Create(context.Background(), environmentID.String(), newValidTerraformServiceUpsertServiceRequest())

	assert.ErrorContains(t, err, repositoryErr.Error())
	require.NotNil(t, newService, "the repository reported the service as created, the service must pass it through")
	assert.Equal(t, created.ID, newService.ID)
}
