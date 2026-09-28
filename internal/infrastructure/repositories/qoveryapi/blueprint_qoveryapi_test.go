//go:build unit && !integration

package qoveryapi

import (
	"context"
	"encoding/json"
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
	return newBlueprintQoveryAPIServing(t, status, "")
}

// newBlueprintQoveryAPIServing is newBlueprintQoveryAPIAnswering with body as the success response.
func newBlueprintQoveryAPIServing(t *testing.T, status int, body string) (blueprint.Repository, func() []recordedRequest) {
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
			return
		}
		_, _ = w.Write([]byte(body))
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

func TestBlueprintQoveryAPIGetServiceIconURI(t *testing.T) {
	t.Parallel()
	serviceID := uuid.NewString()
	icon := "app://qovery-console/redis"

	testCases := []struct {
		name        string
		serviceType blueprint.ServiceType
		status      int
		wantPath    string
		wantIcon    *string
		wantAPIErr  bool
	}{
		{name: "terraform service icon", serviceType: blueprint.ServiceTypeTerraform, status: http.StatusOK, wantPath: "/terraform/" + serviceID, wantIcon: &icon},
		{name: "helm service icon", serviceType: blueprint.ServiceTypeHelm, status: http.StatusOK, wantPath: "/helm/" + serviceID, wantIcon: &icon},
		{name: "service gone", serviceType: blueprint.ServiceTypeHelm, status: http.StatusNotFound, wantPath: "/helm/" + serviceID},
		{name: "server error", serviceType: blueprint.ServiceTypeTerraform, status: http.StatusInternalServerError, wantPath: "/terraform/" + serviceID, wantAPIErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := terraformResponseJSON(t, icon)
			if tc.serviceType == blueprint.ServiceTypeHelm {
				body = helmResponseJSON(t, icon)
			}
			repo, requests := newBlueprintQoveryAPIServing(t, tc.status, body)

			got, err := repo.GetServiceIconURI(context.Background(), tc.serviceType, serviceID)

			assert.Equal(t, []recordedRequest{{method: http.MethodGet, path: tc.wantPath}}, requests())
			if tc.wantAPIErr {
				var apiErr *apierrors.APIError
				require.ErrorAs(t, err, &apiErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantIcon, got)
		})
	}

	t.Run("unknown service type calls nothing", func(t *testing.T) {
		t.Parallel()
		repo, requests := newBlueprintQoveryAPIAnswering(t, http.StatusOK)

		_, err := repo.GetServiceIconURI(context.Background(), blueprint.ServiceType("JOB"), serviceID)

		assert.ErrorIs(t, err, blueprint.ErrUnknownServiceType)
		assert.Empty(t, requests())
	})
}

func helmResponseJSON(t *testing.T, iconURI string) string {
	t.Helper()
	source := qovery.HelmResponseAllOfSourceOneOf1AsHelmResponseAllOfSource(&qovery.HelmResponseAllOfSourceOneOf1{
		Repository: qovery.HelmSourceRepositoryResponse{
			ChartName:    "redis",
			ChartVersion: "1.0.0",
			Repository:   qovery.HelmSourceRepositoryResponseRepository{Id: uuid.NewString(), Name: "catalog", Url: "https://charts.example.com"},
		},
	})
	body, err := json.Marshal(qovery.HelmResponse{
		Id:          uuid.NewString(),
		Environment: qovery.ReferenceObject{Id: uuid.NewString()},
		Name:        "redis",
		Source:      source,
		Arguments:   []string{},
		IconUri:     iconURI,
		ServiceType: qovery.SERVICETYPEENUM_HELM,
	})
	require.NoError(t, err)
	return string(body)
}

func terraformResponseJSON(t *testing.T, iconURI string) string {
	t.Helper()
	service := minimalTerraformResponse()
	service.IconUri = iconURI
	service.ServiceType = qovery.SERVICETYPEENUM_TERRAFORM
	body, err := json.Marshal(service)
	require.NoError(t, err)
	return string(body)
}

func TestBlueprintQoveryAPIGetVariableDefaults(t *testing.T) {
	t.Parallel()
	organizationID := uuid.NewString()
	environmentID := uuid.NewString()
	version := blueprint.CatalogVersion{Provider: "HELM", ServiceFamily: "redis", ServiceVersion: "8"}
	manifest := `{"results":[
		{"kind":"variable","name":"memory_limit","type":{"type":"string"},"required":false,"is_secret":false,"default_value":"512Mi"},
		{"kind":"variable","name":"password","type":{"type":"string","min_length":10},"required":true,"is_secret":true,"default_value":null},
		{"kind":"contextVariable","name":"region","source":"cluster.region","value":"eu-west-3"}
	]}`

	t.Run("variables with a default", func(t *testing.T) {
		t.Parallel()
		repo, requests := newBlueprintQoveryAPIServing(t, http.StatusOK, manifest)

		defaults, err := repo.GetVariableDefaults(context.Background(), organizationID, environmentID, version)

		require.NoError(t, err)
		assert.Equal(t, map[string]string{"memory_limit": "512Mi"}, defaults)
		assert.Equal(t, []recordedRequest{{method: http.MethodGet, path: "/organization/" + organizationID + "/blueprint/catalog/HELM/redis/8/manifest"}}, requests())
	})

	t.Run("catalog error", func(t *testing.T) {
		t.Parallel()
		repo, _ := newBlueprintQoveryAPIServing(t, http.StatusNotFound, "")

		_, err := repo.GetVariableDefaults(context.Background(), organizationID, environmentID, version)

		var apiErr *apierrors.APIError
		require.ErrorAs(t, err, &apiErr)
	})
}
