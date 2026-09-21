//go:build unit && !integration

package qoveryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/container"
)

// createThenFailRoundTripper answers the service creation call with the given payload
// and fails every later call with a 500. It reproduces the real hazard: the Qovery API
// has created the service, and one of the follow-up calls the repository makes on it
// (custom domains, deployment stage, advanced settings) goes wrong.
type createThenFailRoundTripper struct {
	createPathFragment string
	createdPayload     any
}

func (rt createThenFailRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Match the creation endpoint exactly. A "contains" check would also catch the
	// follow-up calls the repository makes on the new service (POST
	// /container/{id}/customDomain, /container/{id}/advancedSettings) and answer them with
	// a 200, which would quietly defeat the point of this harness.
	if req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, rt.createPathFragment) {
		body, err := json.Marshal(rt.createdPayload)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    req,
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"message":"boom"}`)),
		Request:    req,
	}, nil
}

func newAPIClientWithTransport(rt http.RoundTripper) *qovery.APIClient {
	cfg := qovery.NewConfiguration()
	cfg.HTTPClient = &http.Client{Transport: rt}
	return qovery.NewAPIClient(cfg)
}

func newQoveryContainerResponse(containerID, environmentID, registryID string) *qovery.ContainerResponse {
	return qovery.NewContainerResponse(
		containerID,
		time.Now(),
		"xfunctional/imhotep/backend",
		"latest",
		qovery.ContainerRegistryProviderDetailsResponse{
			Id:   registryID,
			Name: "xfunctional-imhotep",
			Url:  "https://reg.eu1.mrva.tech",
			Kind: qovery.CONTAINERREGISTRYKINDENUM_DOCKER_HUB,
		},
		qovery.ReferenceObject{Id: environmentID},
		1999560,
		2047027,
		0,
		"worker",
		500,
		1024,
		0,
		1,
		1,
		qovery.Healthcheck{},
		false,
		"app://qovery-console/container",
		qovery.SERVICETYPEENUM_CONTAINER,
	)
}

// The container exists in Qovery as soon as CreateContainer returns. If a later call in
// the repository fails, returning nil hides it from the Terraform state and the next
// apply dies on "a container named X already exists". The repository must hand the
// created container back alongside the error.
func TestContainerQoveryAPI_Create_ReturnsCreatedContainerWhenFollowUpCallFails(t *testing.T) {
	t.Parallel()

	containerID := uuid.New().String()
	environmentID := uuid.New().String()
	registryID := uuid.New().String()

	client := newAPIClientWithTransport(createThenFailRoundTripper{
		createPathFragment: "/container",
		createdPayload:     newQoveryContainerResponse(containerID, environmentID, registryID),
	})

	repository, err := newContainerQoveryAPI(client)
	require.NoError(t, err)

	cont, err := repository.Create(context.Background(), environmentID, container.UpsertRepositoryRequest{
		RegistryID:           registryID,
		Name:                 "worker",
		ImageName:            "xfunctional/imhotep/backend",
		Tag:                  "latest",
		DeploymentStageID:    uuid.New().String(),
		AdvancedSettingsJson: "{}",
	})

	assert.Error(t, err)
	require.NotNil(t, cont, "the container was created in Qovery, it must be returned so its ID reaches the state")
	assert.Equal(t, containerID, cont.ID.String())
}

// Nothing was created when the creation call itself fails, so there is nothing to put in
// the state.
func TestContainerQoveryAPI_Create_ReturnsNilWhenCreationCallFails(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New().String()

	client := newAPIClientWithTransport(createThenFailRoundTripper{
		createPathFragment: "/never-matches",
	})

	repository, err := newContainerQoveryAPI(client)
	require.NoError(t, err)

	cont, err := repository.Create(context.Background(), environmentID, container.UpsertRepositoryRequest{
		RegistryID: uuid.New().String(),
		Name:       "worker",
		ImageName:  "xfunctional/imhotep/backend",
		Tag:        "latest",
	})

	assert.Error(t, err)
	assert.Nil(t, cont)
}

// A response the API accepted can still be impossible to convert into a domain container
// (unknown enum, a field the provider cannot parse, a validation the domain applies but
// the API does not). The container exists all the same, so the repository must fall back
// to its identifiers rather than return nil and orphan it.
func TestContainerQoveryAPI_Create_ReturnsIdentityOnlyContainerWhenConversionFails(t *testing.T) {
	t.Parallel()

	containerID := uuid.New().String()
	environmentID := uuid.New().String()
	registryID := uuid.New().String()

	// MaxRunningInstances is `validate:"required"` on the domain entity, so zero makes the
	// conversion fail while the generated client still accepts the payload.
	payload := newQoveryContainerResponse(containerID, environmentID, registryID)
	payload.MaxRunningInstances = 0

	client := newAPIClientWithTransport(createThenFailRoundTripper{
		createPathFragment: "/container",
		createdPayload:     payload,
	})

	repository, err := newContainerQoveryAPI(client)
	require.NoError(t, err)

	cont, err := repository.Create(context.Background(), environmentID, container.UpsertRepositoryRequest{
		RegistryID:           registryID,
		Name:                 "worker",
		ImageName:            "xfunctional/imhotep/backend",
		Tag:                  "latest",
		DeploymentStageID:    uuid.New().String(),
		AdvancedSettingsJson: "{}",
	})

	assert.Error(t, err)
	require.NotNil(t, cont, "the container exists even though its response could not be converted")
	assert.Equal(t, containerID, cont.ID.String())
	// Pin that this really came through identityOnlyContainer rather than the regular
	// conversion: the fallback carries identifiers and the name only, so the registry the
	// payload advertises must be absent.
	assert.Equal(t, uuid.Nil, cont.RegistryID, "expected the identity-only fallback, not a converted container")
	assert.Equal(t, environmentID, cont.EnvironmentID.String())
}

// A response that converts cleanly but whose follow-up calls all succeed still goes
// through the regular conversion — the fallback must not hijack the happy path.
func TestContainerQoveryAPI_Create_UsesRegularConversionWhenResponseIsValid(t *testing.T) {
	t.Parallel()

	containerID := uuid.New().String()
	environmentID := uuid.New().String()
	registryID := uuid.New().String()

	client := newAPIClientWithTransport(createThenFailRoundTripper{
		createPathFragment: "/container",
		createdPayload:     newQoveryContainerResponse(containerID, environmentID, registryID),
	})

	repository, err := newContainerQoveryAPI(client)
	require.NoError(t, err)

	// No deployment stage and no custom domains, so the first failing call is the advanced
	// settings update; the partial returned there is the converted container.
	cont, err := repository.Create(context.Background(), environmentID, container.UpsertRepositoryRequest{
		RegistryID:           registryID,
		Name:                 "worker",
		ImageName:            "xfunctional/imhotep/backend",
		Tag:                  "latest",
		AdvancedSettingsJson: "{}",
	})

	assert.Error(t, err)
	require.NotNil(t, cont)
	assert.Equal(t, registryID, cont.RegistryID.String(), "expected the converted container, not the identity-only fallback")
}
