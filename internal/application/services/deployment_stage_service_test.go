//go:build unit

package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

// stubDeploymentStageRepository stands in for the deployment stage repository, which has
// no generated mock. Create returns whatever the test asks for; the other methods are not
// reached.
type stubDeploymentStageRepository struct {
	created *deploymentstage.DeploymentStage
	err     error
}

func (s stubDeploymentStageRepository) Create(_ context.Context, _ string, _ deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	return s.created, s.err
}

func (stubDeploymentStageRepository) Get(_ context.Context, _ string, _ string) (*deploymentstage.DeploymentStage, error) {
	return nil, nil
}

func (stubDeploymentStageRepository) GetAllByEnvironmentID(_ context.Context, _ string) (*[]deploymentstage.DeploymentStage, error) {
	return nil, nil
}

func (stubDeploymentStageRepository) Update(_ context.Context, _ string, _ deploymentstage.UpsertRepositoryRequest) (*deploymentstage.DeploymentStage, error) {
	return nil, nil
}

func (stubDeploymentStageRepository) Delete(_ context.Context, _ string) error { return nil }

func newValidDeploymentStageUpsertServiceRequest() deploymentstage.UpsertServiceRequest {
	return deploymentstage.UpsertServiceRequest{
		DeploymentStageUpsertRequest: deploymentstage.UpsertRepositoryRequest{
			Name: "TERRAFORM DEFAULT",
		},
	}
}

// Creating a deployment stage is not atomic either: the stage is created, then moved
// relative to another one. A failed move leaves the stage in Qovery, so the service must
// pass it through for the state instead of flattening it to nil — otherwise the next
// apply collides on the stage name.
func TestDeploymentStageService_Create_PropagatesPartialStageFromRepository(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := &deploymentstage.DeploymentStage{
		ID:            uuid.New(),
		EnvironmentID: environmentID,
		Name:          "TERRAFORM DEFAULT",
	}
	repositoryErr := errors.New("500 Internal Server Error on move")

	service, err := services.NewDeploymentStageService(stubDeploymentStageRepository{created: created, err: repositoryErr})
	require.NoError(t, err)

	stage, err := service.Create(context.Background(), environmentID.String(), newValidDeploymentStageUpsertServiceRequest())

	assert.ErrorContains(t, err, repositoryErr.Error())
	require.NotNil(t, stage, "the stage was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, stage.ID)
}

// Nothing was created when the creation call itself fails.
func TestDeploymentStageService_Create_ReturnsNilWhenCreateFails(t *testing.T) {
	t.Parallel()

	createErr := errors.New("403 Forbidden")

	service, err := services.NewDeploymentStageService(stubDeploymentStageRepository{err: createErr})
	require.NoError(t, err)

	stage, err := service.Create(context.Background(), uuid.New().String(), newValidDeploymentStageUpsertServiceRequest())

	assert.ErrorContains(t, err, createErr.Error())
	assert.Nil(t, stage)
}
