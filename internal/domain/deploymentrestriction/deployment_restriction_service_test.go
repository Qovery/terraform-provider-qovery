//go:build unit && !integration
// +build unit,!integration

package deploymentrestriction

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

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type restrictionCall struct {
	method string
	path   string
	body   string
}

type restrictionResponse struct {
	status int
	body   string
}

// newRestrictionService answers each call with the response set for its HTTP method and records each call.
func newRestrictionService(t *testing.T, responses map[string]restrictionResponse) (DeploymentRestrictionService, func() []restrictionCall) {
	t.Helper()
	var (
		mu    sync.Mutex
		calls []restrictionCall
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		mu.Lock()
		calls = append(calls, restrictionCall{method: r.Method, path: r.URL.Path, body: string(requestBody)})
		mu.Unlock()
		response := responses[r.Method]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.status)
		_, _ = w.Write([]byte(response.body))
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	service, err := NewDeploymentRestrictionService(*qovery.NewAPIClient(cfg))
	require.NoError(t, err)

	return service, func() []restrictionCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]restrictionCall(nil), calls...)
	}
}

func TestDeploymentRestrictionService_GetServiceDeploymentRestrictions(t *testing.T) {
	t.Parallel()

	t.Run("lists_the_rules_from_the_service_api", func(t *testing.T) {
		t.Parallel()

		serviceID := uuid.NewString()
		restrictionID := uuid.NewString()
		service, recordedCalls := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodGet: {
				status: http.StatusOK,
				body:   fmt.Sprintf(`{"results":[{"id":%q,"created_at":"2026-11-25T10:00:00Z","mode":"EXCLUDE","type":"PATH","value":"docs/"}]}`, restrictionID),
			},
		})

		restrictions, apiErr := service.GetServiceDeploymentRestrictions(context.Background(), serviceID)

		require.Nil(t, apiErr)
		assert.Equal(t, []ServiceDeploymentRestriction{{
			Id:    &restrictionID,
			Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE,
			Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
			Value: "docs/",
		}}, restrictions)
		assert.Equal(t, []restrictionCall{
			{method: http.MethodGet, path: "/service/" + serviceID + "/deploymentRestriction"},
		}, recordedCalls())
	})

	t.Run("returns_an_empty_list_when_the_service_has_no_rules", func(t *testing.T) {
		t.Parallel()

		service, _ := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodGet: {status: http.StatusOK, body: `{}`},
		})

		restrictions, apiErr := service.GetServiceDeploymentRestrictions(context.Background(), uuid.NewString())

		require.Nil(t, apiErr)
		assert.NotNil(t, restrictions)
		assert.Empty(t, restrictions)
	})

	t.Run("returns_a_read_error_when_the_api_fails", func(t *testing.T) {
		t.Parallel()

		serviceID := uuid.NewString()
		service, _ := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodGet: {status: http.StatusInternalServerError, body: `{}`},
		})

		restrictions, apiErr := service.GetServiceDeploymentRestrictions(context.Background(), serviceID)

		assert.Nil(t, restrictions)
		require.NotNil(t, apiErr)
		assert.Equal(t, apierrors.NewReadError(apierrors.APIResourceServiceDeploymentRestriction, serviceID, nil, nil).Summary(), apiErr.Summary())
		assert.Contains(t, apiErr.Detail(), serviceID)
	})
}

func TestDeploymentRestrictionService_UpdateServiceDeploymentRestrictions(t *testing.T) {
	t.Parallel()

	t.Run("deletes_then_creates_through_the_service_api", func(t *testing.T) {
		t.Parallel()

		serviceID := uuid.NewString()
		removedRestrictionID := uuid.NewString()
		service, recordedCalls := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodDelete: {status: http.StatusNoContent},
			http.MethodPost: {
				status: http.StatusCreated,
				body:   fmt.Sprintf(`{"id":%q,"created_at":"2026-11-25T10:00:00Z","mode":"MATCH","type":"PATH","value":"src/"}`, uuid.NewString()),
			},
		})

		apiErr := service.UpdateServiceDeploymentRestrictions(context.Background(), serviceID, ServiceDeploymentRestrictionsDiff{
			Create: []ServiceDeploymentRestriction{{
				Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH,
				Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
				Value: "src/",
			}},
			Delete: []string{removedRestrictionID},
		})

		require.Nil(t, apiErr)
		assert.Equal(t, []restrictionCall{
			{method: http.MethodDelete, path: "/service/" + serviceID + "/deploymentRestriction/" + removedRestrictionID},
			{method: http.MethodPost, path: "/service/" + serviceID + "/deploymentRestriction", body: `{"mode":"MATCH","type":"PATH","value":"src/"}` + "\n"},
		}, recordedCalls())
	})

	t.Run("stops_with_a_delete_error_before_creating", func(t *testing.T) {
		t.Parallel()

		serviceID := uuid.NewString()
		removedRestrictionID := uuid.NewString()
		service, recordedCalls := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodDelete: {status: http.StatusInternalServerError, body: `{}`},
		})

		apiErr := service.UpdateServiceDeploymentRestrictions(context.Background(), serviceID, ServiceDeploymentRestrictionsDiff{
			Create: []ServiceDeploymentRestriction{{
				Mode:  qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH,
				Type:  qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH,
				Value: "src/",
			}},
			Delete: []string{removedRestrictionID},
		})

		require.NotNil(t, apiErr)
		assert.Equal(t, apierrors.NewDeleteError(apierrors.APIResourceServiceDeploymentRestriction, removedRestrictionID, nil, nil).Summary(), apiErr.Summary())
		assert.Contains(t, apiErr.Detail(), removedRestrictionID)
		assert.Len(t, recordedCalls(), 1)
	})

	t.Run("stops_with_a_create_error_at_the_first_failed_rule", func(t *testing.T) {
		t.Parallel()

		service, recordedCalls := newRestrictionService(t, map[string]restrictionResponse{
			http.MethodPost: {status: http.StatusBadRequest, body: `{}`},
		})

		apiErr := service.UpdateServiceDeploymentRestrictions(context.Background(), uuid.NewString(), ServiceDeploymentRestrictionsDiff{
			Create: []ServiceDeploymentRestriction{
				{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_MATCH, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "src/"},
				{Mode: qovery.DEPLOYMENTRESTRICTIONMODEENUM_EXCLUDE, Type: qovery.DEPLOYMENTRESTRICTIONTYPEENUM_PATH, Value: "docs/"},
			},
		})

		require.NotNil(t, apiErr)
		assert.Equal(t, apierrors.NewCreateError(apierrors.APIResourceServiceDeploymentRestriction, "no id", nil, nil).Summary(), apiErr.Summary())
		assert.Contains(t, apiErr.Detail(), "no id")
		assert.Len(t, recordedCalls(), 1)
	})
}
