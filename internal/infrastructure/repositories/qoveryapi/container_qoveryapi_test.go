//go:build unit && !integration

package qoveryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

// alwaysOKRoundTripper answers every call with the same payload, so a repository Create
// can run all the way to its final conversion without any API failure.
type alwaysOKRoundTripper struct {
	payload any
}

func (rt alwaysOKRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := json.Marshal(rt.payload)
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

// newAPIClientWithTransport routes the generated client through rt. It also points the
// configured server URL at a local server that fails every request: the advanced settings
// service (internal/domain/advanced_settings) ignores cfg.HTTPClient and issues its calls
// with a bare http.Client against that URL, so without this a test whose first failing
// follow-up call is the advanced settings update would reach the real Qovery API.
func newAPIClientWithTransport(t *testing.T, rt http.RoundTripper) *qovery.APIClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.HTTPClient = &http.Client{Transport: rt}
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
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

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
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

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
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

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
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

// A response that converts cleanly and then hits a failing follow-up call must still come
// back as the converted container: the identity-only fallback must not hijack the normal
// path. (The harness fails every post-create call, so the failure here is the advanced
// settings update.)
func TestContainerQoveryAPI_Create_UsesRegularConversionWhenResponseIsValid(t *testing.T) {
	t.Parallel()

	containerID := uuid.New().String()
	environmentID := uuid.New().String()
	registryID := uuid.New().String()

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
		createPathFragment: "/container",
		createdPayload:     newQoveryContainerResponse(containerID, environmentID, registryID),
	})

	repository, err := newContainerQoveryAPI(client)
	require.NoError(t, err)

	// No deployment stage and no custom domains, so the first failing call is the advanced
	// settings update; the partial returned there must be the converted container.
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

// The environment reference in the response is not always usable. When it is broken the
// fallback falls back again, onto the environment the create was aimed at, rather than
// giving up on the entity entirely.
func TestContainerQoveryAPI_Create_FallsBackToRequestedEnvironmentWhenResponseReferenceIsBroken(t *testing.T) {
	t.Parallel()

	containerID := uuid.New().String()
	environmentID := uuid.New().String()
	registryID := uuid.New().String()

	// Zero MaxRunningInstances forces the identity-only path, and the mangled environment
	// reference forces that path to lean on the requested environment ID.
	payload := newQoveryContainerResponse(containerID, environmentID, registryID)
	payload.MaxRunningInstances = 0
	payload.Environment = qovery.ReferenceObject{Id: "not-a-uuid"}

	client := newAPIClientWithTransport(t, createThenFailRoundTripper{
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
	require.NotNil(t, cont, "a broken environment reference must not cost us the whole entity")
	assert.Equal(t, containerID, cont.ID.String())
	assert.Equal(t, environmentID, cont.EnvironmentID.String(), "expected the environment the create was aimed at")
}
