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
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/execution_command"
	"github.com/qovery/terraform-provider-qovery/internal/domain/image"
	"github.com/qovery/terraform-provider-qovery/internal/domain/job"
	repomocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newValidJobUpsertServiceRequest() job.UpsertServiceRequest {
	entrypoint := "/app/.venv/bin/python"
	return job.UpsertServiceRequest{
		JobUpsertRequest: job.UpsertRepositoryRequest{
			Name: "migrations-and-baseline",
			Source: job.Source{
				Image: &image.Image{
					RegistryID: uuid.New().String(),
					Name:       "xfunctional/imhotep/backend",
					Tag:        "latest",
				},
			},
			Schedule: job.JobSchedule{
				OnStart: &execution_command.ExecutionCommand{
					Entrypoint: &entrypoint,
					Arguments:  []string{"-m", "cockpit.staging_init"},
				},
			},
		},
	}
}

func newCreatedJob(environmentID uuid.UUID) *job.Job {
	return &job.Job{
		ID:            uuid.New(),
		EnvironmentID: environmentID,
		Name:          "migrations-and-baseline",
	}
}

// Same partial-create hazard as containers: the job exists in Qovery as soon as the
// create call returns, so a failure while pushing its variables must still hand the job
// back for the state. Otherwise the next apply hits "a job named X already exists".
func TestJobService_Create_ReturnsCreatedJobOnPartialFailure(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedJob(environmentID)
	variableErr := errors.New("409 Conflict - Variable already exists: AWS_REGION")

	jobRepository := &repomocks.JobRepository{}
	jobRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, variableErr)

	service, err := services.NewJobService(
		jobRepository,
		&servicemocks.DeploymentService{},
		variableService,
		&servicemocks.SecretService{},
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newJob, err := service.Create(context.Background(), environmentID.String(), newValidJobUpsertServiceRequest())

	assert.ErrorContains(t, err, variableErr.Error())
	require.NotNil(t, newJob, "the job was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, newJob.ID)
}

func TestJobService_Create_ReturnsNilWhenCreateFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	createErr := errors.New("403 Forbidden")

	jobRepository := &repomocks.JobRepository{}
	jobRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(nil, createErr)

	service, err := services.NewJobService(
		jobRepository,
		&servicemocks.DeploymentService{},
		&servicemocks.VariableService{},
		&servicemocks.SecretService{},
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newJob, err := service.Create(context.Background(), environmentID.String(), newValidJobUpsertServiceRequest())

	assert.ErrorContains(t, err, createErr.Error())
	assert.Nil(t, newJob)
}

// The repository create is itself a create-then-configure sequence, so the job may
// already exist in Qovery when it reports an error.
func TestJobService_Create_PropagatesPartialJobFromRepository(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedJob(environmentID)
	repositoryErr := errors.New("500 Internal Server Error on deployment stage")

	jobRepository := &repomocks.JobRepository{}
	jobRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, repositoryErr)

	service, err := services.NewJobService(
		jobRepository,
		&servicemocks.DeploymentService{},
		&servicemocks.VariableService{},
		&servicemocks.SecretService{},
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newJob, err := service.Create(context.Background(), environmentID.String(), newValidJobUpsertServiceRequest())

	assert.ErrorContains(t, err, repositoryErr.Error())
	require.NotNil(t, newJob, "the repository reported the job as created, the service must pass it through")
	assert.Equal(t, created.ID, newJob.ID)
}

// The refresh is the last post-create step, after variables and secrets are written.
func TestJobService_Create_ReturnsCreatedJobWhenRefreshFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedJob(environmentID)
	refreshErr := errors.New("503 Service Unavailable while listing variables")

	jobRepository := &repomocks.JobRepository{}
	jobRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	variableService.On("List", mock.Anything, created.ID.String()).Return(nil, refreshErr)

	secretService := &servicemocks.SecretService{}
	secretService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	service, err := services.NewJobService(
		jobRepository,
		&servicemocks.DeploymentService{},
		variableService,
		secretService,
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newJob, err := service.Create(context.Background(), environmentID.String(), newValidJobUpsertServiceRequest())

	assert.ErrorContains(t, err, refreshErr.Error())
	require.NotNil(t, newJob, "variables and secrets were already written, the job must still reach the state")
	assert.Equal(t, created.ID, newJob.ID)
}
