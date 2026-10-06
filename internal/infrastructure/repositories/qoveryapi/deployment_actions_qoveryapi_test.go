//go:build unit && !integration
// +build unit,!integration

package qoveryapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/newdeployment"
)

type deploymentActionCall struct {
	method string
	path   string
	body   string
}

// newDeploymentActionAPIClient answers every call with status and body and records each call.
func newDeploymentActionAPIClient(t *testing.T, status int, body string) (*qovery.APIClient, func() []deploymentActionCall) {
	t.Helper()
	var (
		mu    sync.Mutex
		calls []deploymentActionCall
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		mu.Lock()
		calls = append(calls, deploymentActionCall{method: r.Method, path: r.URL.Path, body: string(requestBody)})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	return qovery.NewAPIClient(cfg), func() []deploymentActionCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]deploymentActionCall(nil), calls...)
	}
}

// TestContainerDeploymentQoveryAPIDeploy_SendsImageTag guards that the image tag is always sent,
// as when the client required it.
func TestContainerDeploymentQoveryAPIDeploy_SendsImageTag(t *testing.T) {
	t.Parallel()

	containerID := uuid.NewString()
	containerStatus := fmt.Sprintf(`{"id":%q,"state":"DEPLOYMENT_QUEUED","service_deployment_status":"NEVER_DEPLOYED","is_part_last_deployment":true,`+
		`"status_details":{"action":"DEPLOY","status":"QUEUED","sub_action":"NONE"},"deployment_request_id":null,"deployment_requests_count":1}`, containerID)

	testCases := []struct {
		TestName     string
		ImageTag     string
		ExpectedBody string
	}{
		{
			TestName:     "with_image_tag",
			ImageTag:     "1.2.3",
			ExpectedBody: `{"image_tag":"1.2.3"}`,
		},
		{
			TestName:     "with_empty_image_tag",
			ImageTag:     "",
			ExpectedBody: `{"image_tag":""}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			client, calls := newDeploymentActionAPIClient(t, http.StatusOK, containerStatus)
			repo, err := newContainerDeploymentQoveryAPI(client)
			require.NoError(t, err)

			st, err := repo.Deploy(context.Background(), containerID, tc.ImageTag)

			require.NoError(t, err)
			assert.Equal(t, "DEPLOYMENT_QUEUED", st.State.String())
			recorded := calls()
			require.Len(t, recorded, 1)
			assert.Equal(t, http.MethodPost, recorded[0].method)
			assert.Equal(t, "/container/"+containerID+"/deploy", recorded[0].path)
			assert.JSONEq(t, tc.ExpectedBody, recorded[0].body)
		})
	}
}

// TestNewDeploymentQoveryAPIRestart guards the restart of an environment: the redeploy call
// answers without a body, so only its status code decides the outcome.
func TestNewDeploymentQoveryAPIRestart(t *testing.T) {
	t.Parallel()

	environmentID := uuid.New()

	testCases := []struct {
		TestName    string
		Status      int
		Body        string
		ExpectError bool
	}{
		{
			TestName: "success_without_body",
			Status:   http.StatusOK,
		},
		{
			TestName: "success_accepted_without_body",
			Status:   http.StatusAccepted,
		},
		{
			TestName:    "error_api_failure",
			Status:      http.StatusInternalServerError,
			Body:        `{"message":"stub error"}`,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			client, calls := newDeploymentActionAPIClient(t, tc.Status, tc.Body)
			repo, err := newDeploymentEnvironmentQoveryAPI(client)
			require.NoError(t, err)
			deployment := newdeployment.Deployment{EnvironmentID: &environmentID}

			got, err := repo.Restart(context.Background(), deployment)

			recorded := calls()
			require.Len(t, recorded, 1)
			assert.Equal(t, http.MethodPost, recorded[0].method)
			assert.Equal(t, "/environment/"+environmentID.String()+"/redeploy", recorded[0].path)
			if tc.ExpectError {
				var apiErr *apierrors.APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, &deployment, got)
		})
	}
}
