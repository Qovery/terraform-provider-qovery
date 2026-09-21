//go:build unit && !integration

package qoveryapi

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/deploymentstage"
)

func newQoveryDeploymentStageResponse(stageID, environmentID string) *qovery.DeploymentStageResponse {
	name := "TERRAFORM DEFAULT"
	description := "Rename me to avoid default ordering"

	response := qovery.NewDeploymentStageResponse(stageID, time.Now(), qovery.ReferenceObject{Id: environmentID})
	response.Name = &name
	response.Description = &description

	return response
}

// The stage is created before it is moved, so a failing move must not cost us the stage.
func TestDeploymentStageQoveryAPI_Create_ReturnsCreatedStageWhenMoveFails(t *testing.T) {
	t.Parallel()

	stageID := uuid.New().String()
	environmentID := uuid.New().String()
	isAfter := uuid.New().String()

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
		createPathFragment: "/deploymentStage",
		createdPayload:     newQoveryDeploymentStageResponse(stageID, environmentID),
	})

	repository, err := newDeploymentStageQoveryAPI(client)
	require.NoError(t, err)

	stage, err := repository.Create(context.Background(), environmentID, deploymentstage.UpsertRepositoryRequest{
		Name:    "TERRAFORM DEFAULT",
		IsAfter: &isAfter,
	})

	assert.Error(t, err)
	require.NotNil(t, stage, "the stage was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, stageID, stage.ID.String())
}

// is_after and is_before come from the user's configuration and are parsed as UUIDs when
// the domain entity is built. A malformed one must cost the ordering, not the stage.
func TestDeploymentStageQoveryAPI_Create_ReturnsCreatedStageWhenOrderingReferenceIsMalformed(t *testing.T) {
	t.Parallel()

	stageID := uuid.New().String()
	environmentID := uuid.New().String()
	malformed := "not-a-uuid"

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
		createPathFragment: "/deploymentStage",
		createdPayload:     newQoveryDeploymentStageResponse(stageID, environmentID),
	})

	repository, err := newDeploymentStageQoveryAPI(client)
	require.NoError(t, err)

	stage, err := repository.Create(context.Background(), environmentID, deploymentstage.UpsertRepositoryRequest{
		Name:    "TERRAFORM DEFAULT",
		IsAfter: &malformed,
	})

	assert.Error(t, err)
	require.NotNil(t, stage, "a malformed ordering reference must not cost us the created stage")
	assert.Equal(t, stageID, stage.ID.String())
	assert.Nil(t, stage.IsAfter, "the ordering is dropped, the identity is kept")
}

// The environment reference in the created-stage response is not always usable. When it is
// broken the fallback leans on the environment the create was aimed at, rather than
// dropping the stage and orphaning it.
func TestDeploymentStageQoveryAPI_Create_FallsBackToRequestedEnvironmentWhenResponseReferenceIsBroken(t *testing.T) {
	t.Parallel()

	stageID := uuid.New().String()
	environmentID := uuid.New().String()
	malformed := "not-a-uuid"

	// NewDeploymentStage parses the environment before the ordering, so this mangled
	// environment reference is what fails the first construction. The fallback would fail
	// on it too if it trusted the response alone instead of the requested environment.
	payload := newQoveryDeploymentStageResponse(stageID, "broken-env-ref")

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
		createPathFragment: "/deploymentStage",
		createdPayload:     payload,
	})

	repository, err := newDeploymentStageQoveryAPI(client)
	require.NoError(t, err)

	stage, err := repository.Create(context.Background(), environmentID, deploymentstage.UpsertRepositoryRequest{
		Name:    "TERRAFORM DEFAULT",
		IsAfter: &malformed,
	})

	assert.Error(t, err)
	require.NotNil(t, stage, "a broken environment reference must not cost us the created stage")
	assert.Equal(t, stageID, stage.ID.String())
	assert.Equal(t, environmentID, stage.EnvironmentID.String(), "expected the environment the create was aimed at")
}

// Every call can succeed and the final conversion still fail, because it reads the same
// unusable response the first construction choked on. That is the one path where the
// fallback is the only thing standing between a created stage and an orphan.
func TestDeploymentStageQoveryAPI_Create_ReturnsFallbackWhenFinalConversionFails(t *testing.T) {
	t.Parallel()

	stageID := uuid.New().String()
	environmentID := uuid.New().String()
	isAfter := uuid.New().String()

	// Nothing fails at the API level here; the broken environment reference is what makes
	// both the first and the final NewDeploymentStage call fail.
	client := newAPIClientWithTransport(t, alwaysOKRoundTripper{
		payload: newQoveryDeploymentStageResponse(stageID, "broken-env-ref"),
	})

	repository, err := newDeploymentStageQoveryAPI(client)
	require.NoError(t, err)

	stage, err := repository.Create(context.Background(), environmentID, deploymentstage.UpsertRepositoryRequest{
		Name:    "TERRAFORM DEFAULT",
		IsAfter: &isAfter,
	})

	assert.Error(t, err)
	require.NotNil(t, stage, "the stage was created and moved, it must not be lost on the final conversion")
	assert.Equal(t, stageID, stage.ID.String())
	assert.Equal(t, environmentID, stage.EnvironmentID.String())
}
