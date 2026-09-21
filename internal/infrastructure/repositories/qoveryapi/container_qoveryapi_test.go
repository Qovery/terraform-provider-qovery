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
	if req.Method == http.MethodPost && strings.Contains(req.URL.Path, rt.createPathFragment) {
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
