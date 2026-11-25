//go:build unit

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/application/services"
	servicemocks "github.com/qovery/terraform-provider-qovery/internal/application/services/mocks_test"
	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentrestriction"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
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
		deploymentrestriction.DeploymentRestrictionService{},
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
		deploymentrestriction.DeploymentRestrictionService{},
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
		deploymentrestriction.DeploymentRestrictionService{},
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
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	cont, err := service.Create(context.Background(), environmentID.String(), newValidContainerUpsertServiceRequest())

	assert.ErrorContains(t, err, refreshErr.Error())
	require.NotNil(t, cont, "variables and secrets were already written, the container must still reach the state")
	assert.Equal(t, created.ID, cont.ID)
}

// callLog records the calls a test sees, in the order they happen.
type callLog struct {
	mu    sync.Mutex
	calls []string
}

func (l *callLog) add(call string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calls = append(l.calls, call)
}

func (l *callLog) list() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.calls...)
}

// newRestrictionAPI serves the service deployment restriction API from rules, answers writeStatus to
// every write, and records each call in log.
func newRestrictionAPI(t *testing.T, log *callLog, rules string, writeStatus int) deploymentrestriction.DeploymentRestrictionService {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.add(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = fmt.Fprintf(w, `{"results":%s}`, rules)
		case http.MethodPost:
			w.WriteHeader(writeStatus)
			_, _ = fmt.Fprintf(w, `{"id":%q,"created_at":"2026-11-25T10:00:00Z","mode":"MATCH","type":"PATH","value":"src/"}`, uuid.NewString())
		default:
			w.WriteHeader(writeStatus)
		}
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	service, err := deploymentrestriction.NewDeploymentRestrictionService(*qovery.NewAPIClient(cfg))
	require.NoError(t, err)
	return service
}

// newRefreshableContainerService returns a container service whose variables, secrets and status
// calls succeed, so a create, get or update reaches the deployment restrictions.
func newRefreshableContainerService(t *testing.T, containerRepository *repomocks.ContainerRepository, containerID string, restrictionService deploymentrestriction.DeploymentRestrictionService) container.Service {
	t.Helper()
	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, containerID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	variableService.On("List", mock.Anything, containerID).Return(variable.Variables{}, nil)

	secretService := &servicemocks.SecretService{}
	secretService.On("Update", mock.Anything, containerID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	secretService.On("List", mock.Anything, containerID).Return(secret.Secrets{}, nil)

	deploymentService := &servicemocks.DeploymentService{}
	deploymentService.On("GetStatus", mock.Anything, containerID).
		Return(&status.Status{ID: uuid.MustParse(containerID), State: status.StateDeployed}, nil)

	service, err := services.NewContainerService(
		containerRepository,
		deploymentService,
		variableService,
		secretService,
		restrictionService,
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)
	return service
}

func newRuleToCreate() []deploymentrestriction.ServiceDeploymentRestriction {
	return []deploymentrestriction.ServiceDeploymentRestriction{
		{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "src/"},
	}
}

func TestContainerService_Create_WritesDeploymentRestrictionsAfterTheContainer(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedContainer(environmentID)
	containerID := created.ID.String()
	log := &callLog{}

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Run(func(mock.Arguments) { log.add("create container") }).
		Return(created, nil)

	service := newRefreshableContainerService(t, containerRepository, containerID, newRestrictionAPI(t, log, `[]`, http.StatusCreated))

	request := newValidContainerUpsertServiceRequest()
	request.DeploymentRestrictionsDiff = deploymentrestriction.ServiceDeploymentRestrictionsDiff{Create: newRuleToCreate()}
	cont, err := service.Create(context.Background(), environmentID.String(), request)

	require.NoError(t, err)
	require.NotNil(t, cont)
	assert.Equal(t, []string{
		"create container",
		"POST /service/" + containerID + "/deploymentRestriction",
		"GET /service/" + containerID + "/deploymentRestriction",
	}, log.list())
}

// The rules are written once the container exists in Qovery, so a failed rule write must still
// hand the container back for the state.
func TestContainerService_Create_ReturnsCreatedContainerWhenRestrictionsFail(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedContainer(environmentID)
	log := &callLog{}

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	service := newRefreshableContainerService(t, containerRepository, created.ID.String(), newRestrictionAPI(t, log, `[]`, http.StatusBadRequest))

	request := newValidContainerUpsertServiceRequest()
	request.DeploymentRestrictionsDiff = deploymentrestriction.ServiceDeploymentRestrictionsDiff{Create: newRuleToCreate()}
	cont, err := service.Create(context.Background(), environmentID.String(), request)

	assert.ErrorContains(t, err, container.ErrFailedToCreateContainer.Error())
	require.NotNil(t, cont, "the container was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, cont.ID)
}

func TestContainerService_Update_WritesDeploymentRestrictionsAfterTheContainer(t *testing.T) {
	t.Parallel()

	updated := newCreatedContainer(uuid.New())
	containerID := updated.ID.String()
	removedRuleID := uuid.NewString()
	keptRuleID := uuid.NewString()
	log := &callLog{}

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Update", mock.Anything, containerID, mock.Anything).
		Run(func(mock.Arguments) { log.add("update container") }).
		Return(updated, nil)

	rules := fmt.Sprintf(`[{"id":%q,"created_at":"2026-11-25T10:00:00Z","mode":"MATCH","type":"PATH","value":"src/"}]`, keptRuleID)
	service := newRefreshableContainerService(t, containerRepository, containerID, newRestrictionAPI(t, log, rules, http.StatusNoContent))

	request := newValidContainerUpsertServiceRequest()
	request.DeploymentRestrictionsDiff = deploymentrestriction.ServiceDeploymentRestrictionsDiff{
		Create: newRuleToCreate(),
		Delete: []string{removedRuleID},
	}
	cont, err := service.Update(context.Background(), containerID, request)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"update container",
		"DELETE /service/" + containerID + "/deploymentRestriction/" + removedRuleID,
		"POST /service/" + containerID + "/deploymentRestriction",
		"GET /service/" + containerID + "/deploymentRestriction",
	}, log.list())
	require.Len(t, cont.DeploymentRestrictions, 1)
	assert.Equal(t, keptRuleID, *cont.DeploymentRestrictions[0].Id)
}

// Rules stay on a container switched from a Git repository to a registry image, so the read
// fetches them whatever the source.
func TestContainerService_Get_ReadsDeploymentRestrictionsOfAnImageContainer(t *testing.T) {
	t.Parallel()

	imageContainer := newCreatedContainer(uuid.New())
	containerID := imageContainer.ID.String()
	ruleID := uuid.NewString()
	log := &callLog{}

	containerRepository := &repomocks.ContainerRepository{}
	containerRepository.On("Get", mock.Anything, containerID, mock.Anything, false).Return(imageContainer, nil)

	rules := fmt.Sprintf(`[{"id":%q,"created_at":"2026-11-25T10:00:00Z","mode":"EXCLUDE","type":"PATH","value":"docs/"}]`, ruleID)
	service := newRefreshableContainerService(t, containerRepository, containerID, newRestrictionAPI(t, log, rules, http.StatusNoContent))

	cont, err := service.Get(context.Background(), containerID, "", false)

	require.NoError(t, err)
	assert.Nil(t, cont.GitSource)
	require.Len(t, cont.DeploymentRestrictions, 1)
	assert.Equal(t, ruleID, *cont.DeploymentRestrictions[0].Id)
	assert.Equal(t, qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE, cont.DeploymentRestrictions[0].Mode)
	assert.Equal(t, "docs/", cont.DeploymentRestrictions[0].Value)
}
