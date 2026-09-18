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
	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	repomocks "github.com/qovery/terraform-provider-qovery/internal/infrastructure/repositories/mocks_test"
)

func newValidHelmUpsertServiceRequest() helm.UpsertServiceRequest {
	return helm.UpsertServiceRequest{
		HelmUpsertRequest: helm.UpsertRepositoryRequest{
			Name: "role-imhotep-staging",
		},
	}
}

func newCreatedHelm(environmentID uuid.UUID) *helm.Helm {
	return &helm.Helm{
		ID:            uuid.New(),
		EnvironmentID: environmentID,
		Name:          "role-imhotep-staging",
	}
}

// Same partial-create hazard as containers and jobs.
func TestHelmService_Create_ReturnsCreatedHelmOnPartialFailure(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	created := newCreatedHelm(environmentID)
	variableErr := errors.New("409 Conflict - Variable already exists: AWS_REGION")

	helmRepository := &repomocks.HelmRepository{}
	helmRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(created, nil)

	variableService := &servicemocks.VariableService{}
	variableService.On("Update", mock.Anything, created.ID.String(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, variableErr)

	service, err := services.NewHelmService(
		helmRepository,
		&servicemocks.DeploymentService{},
		variableService,
		&servicemocks.SecretService{},
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newHelm, err := service.Create(context.Background(), environmentID.String(), newValidHelmUpsertServiceRequest())

	assert.ErrorContains(t, err, variableErr.Error())
	require.NotNil(t, newHelm, "the helm service was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, created.ID, newHelm.ID)
}

func TestHelmService_Create_ReturnsNilWhenCreateFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()
	createErr := errors.New("403 Forbidden")

	helmRepository := &repomocks.HelmRepository{}
	helmRepository.On("Create", mock.Anything, environmentID.String(), mock.Anything).
		Return(nil, createErr)

	service, err := services.NewHelmService(
		helmRepository,
		&servicemocks.DeploymentService{},
		&servicemocks.VariableService{},
		&servicemocks.SecretService{},
		deploymentrestriction.DeploymentRestrictionService{},
		stubExternalSecretRepository{},
		stubExternalSecretFileRepository{},
	)
	require.NoError(t, err)

	newHelm, err := service.Create(context.Background(), environmentID.String(), newValidHelmUpsertServiceRequest())

	assert.ErrorContains(t, err, createErr.Error())
	assert.Nil(t, newHelm)
}
