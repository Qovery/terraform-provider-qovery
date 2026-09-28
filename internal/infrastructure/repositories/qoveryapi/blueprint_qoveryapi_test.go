//go:build unit && !integration

package qoveryapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

type erroringRoundTripper struct{ err error }

func (f erroringRoundTripper) RoundTrip(*http.Request) (*http.Response, error) { return nil, f.err }

type recordedRequest struct {
	method string
	path   string
}

// newBlueprintQoveryAPIAnswering serves every request with status and records what was called.
func newBlueprintQoveryAPIAnswering(t *testing.T, status int) (blueprint.Repository, func() []recordedRequest) {
	t.Helper()
	var (
		mu       sync.Mutex
		requests []recordedRequest
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, recordedRequest{method: r.Method, path: r.URL.Path})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status >= http.StatusBadRequest {
			_, _ = w.Write([]byte(`{"message":"stub error"}`))
		}
	}))
	t.Cleanup(server.Close)

	cfg := qovery.NewConfiguration()
	cfg.Servers = qovery.ServerConfigurations{{URL: server.URL}}
	repo, err := newBlueprintQoveryAPI(qovery.NewAPIClient(cfg))
	require.NoError(t, err)
	return repo, func() []recordedRequest {
		mu.Lock()
		defer mu.Unlock()
		return append([]recordedRequest(nil), requests...)
	}
}

func TestBlueprintQoveryAPIDeleteService(t *testing.T) {
	t.Parallel()
	serviceID := uuid.NewString()

	testCases := []struct {
		name        string
		serviceType blueprint.ServiceType
		status      int
		wantPath    string
		wantExisted bool
		wantAPIErr  bool
	}{
		{name: "terraform service deleted", serviceType: blueprint.ServiceTypeTerraform, status: http.StatusNoContent, wantPath: "/terraform/" + serviceID, wantExisted: true},
		{name: "helm service deleted", serviceType: blueprint.ServiceTypeHelm, status: http.StatusNoContent, wantPath: "/helm/" + serviceID, wantExisted: true},
		{name: "service already gone", serviceType: blueprint.ServiceTypeHelm, status: http.StatusNotFound, wantPath: "/helm/" + serviceID, wantExisted: false},
		{name: "forbidden is a real error, not a deleted service", serviceType: blueprint.ServiceTypeTerraform, status: http.StatusForbidden, wantPath: "/terraform/" + serviceID, wantAPIErr: true},
		{name: "server error", serviceType: blueprint.ServiceTypeHelm, status: http.StatusInternalServerError, wantPath: "/helm/" + serviceID, wantAPIErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, requests := newBlueprintQoveryAPIAnswering(t, tc.status)

			existed, err := repo.DeleteService(context.Background(), tc.serviceType, serviceID)

			assert.Equal(t, []recordedRequest{{method: http.MethodDelete, path: tc.wantPath}}, requests())
			if tc.wantAPIErr {
				var apiErr *apierrors.APIError
				require.ErrorAs(t, err, &apiErr)
				assert.False(t, existed)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantExisted, existed)
		})
	}

	t.Run("transport failure with no response is an API error", func(t *testing.T) {
		t.Parallel()
		transportErr := errors.New("connection refused")
		cfg := qovery.NewConfiguration()
		cfg.Servers = qovery.ServerConfigurations{{URL: "http://qovery.invalid"}}
		cfg.HTTPClient = &http.Client{Transport: erroringRoundTripper{err: transportErr}}
		repo, err := newBlueprintQoveryAPI(qovery.NewAPIClient(cfg))
		require.NoError(t, err)

		existed, err := repo.DeleteService(context.Background(), blueprint.ServiceTypeHelm, serviceID)

		var apiErr *apierrors.APIError
		require.ErrorAs(t, err, &apiErr)
		// APIError keeps the cause's text but has no Unwrap, so errors.Is cannot see it
		assert.ErrorContains(t, err, transportErr.Error())
		assert.False(t, existed)
	})

	t.Run("unknown service type calls nothing", func(t *testing.T) {
		t.Parallel()
		repo, requests := newBlueprintQoveryAPIAnswering(t, http.StatusNoContent)

		existed, err := repo.DeleteService(context.Background(), blueprint.ServiceType("JOB"), serviceID)

		assert.ErrorIs(t, err, blueprint.ErrUnknownServiceType)
		assert.False(t, existed)
		assert.Empty(t, requests())
	})
}
