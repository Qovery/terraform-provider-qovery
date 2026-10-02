//go:build unit && !integration
// +build unit,!integration

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// TestPreserveUnmanagedClusterFields checks the edit request resends the current value of every
// field the provider does not manage. The KEDA profiles are checked on the JSON sent to the API:
// client-go's ClusterKeda only models `enabled`, and q-core resets an omitted profile to NORMAL.
func TestPreserveUnmanagedClusterFields(t *testing.T) {
	t.Parallel()

	currentKeda := func(t *testing.T, raw string) *qovery.ClusterKeda {
		var keda qovery.ClusterKeda
		require.NoError(t, json.Unmarshal([]byte(raw), &keda))
		return &keda
	}

	testCases := []struct {
		TestName     string
		RequestKeda  *qovery.ClusterKeda
		CurrentKeda  string
		ExpectedKeda string
	}{
		{
			TestName:     "keda_enabled_resends_current_profiles",
			RequestKeda:  qovery.NewClusterKeda(true),
			CurrentKeda:  `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
			ExpectedKeda: `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
		{
			TestName:     "keda_disabled_resends_current_profiles",
			RequestKeda:  qovery.NewClusterKeda(false),
			CurrentKeda:  `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
			ExpectedKeda: `{"enabled":false,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
		{
			TestName:     "current_keda_without_profiles_sends_enabled_only",
			RequestKeda:  qovery.NewClusterKeda(true),
			CurrentKeda:  `{"enabled":true}`,
			ExpectedKeda: `{"enabled":true}`,
		},
		{
			TestName:     "no_current_keda_sends_enabled_only",
			RequestKeda:  qovery.NewClusterKeda(true),
			ExpectedKeda: `{"enabled":true}`,
		},
		{
			TestName:    "no_request_keda_stays_omitted",
			CurrentKeda: `{"enabled":true,"availability_profile":"HIGH","resource_profile":"LOW"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			current := &qovery.Cluster{}
			if tc.CurrentKeda != "" {
				current.Keda = currentKeda(t, tc.CurrentKeda)
			}
			request := qovery.ClusterRequest{Keda: tc.RequestKeda}

			preserveUnmanagedClusterFields(&request, current)

			if tc.ExpectedKeda == "" {
				assert.Nil(t, request.Keda)
				return
			}
			body, err := json.Marshal(request.Keda)
			require.NoError(t, err)
			assert.JSONEq(t, tc.ExpectedKeda, string(body))
		})
	}

	t.Run("disk_size_and_metrics_parameters", func(t *testing.T) {
		t.Parallel()

		currentDiskSize := int32(50)
		currentMetrics := &qovery.MetricsParameters{Enabled: qovery.PtrBool(true)}
		current := &qovery.Cluster{DiskSize: &currentDiskSize, MetricsParameters: currentMetrics}

		omitted := qovery.ClusterRequest{}
		preserveUnmanagedClusterFields(&omitted, current)
		assert.Equal(t, &currentDiskSize, omitted.DiskSize)
		assert.Equal(t, currentMetrics, omitted.MetricsParameters)

		plannedDiskSize := int32(100)
		plannedMetrics := &qovery.MetricsParameters{Enabled: qovery.PtrBool(false)}
		planned := qovery.ClusterRequest{DiskSize: &plannedDiskSize, MetricsParameters: plannedMetrics}
		preserveUnmanagedClusterFields(&planned, current)
		assert.Equal(t, &plannedDiskSize, planned.DiskSize)
		assert.Equal(t, plannedMetrics, planned.MetricsParameters)
	})
}

// fakeClusterListAPI serves the organization cluster list. The first failures calls answer 500,
// the next ones list the given cluster IDs.
type fakeClusterListAPI struct {
	mu         sync.Mutex
	failures   int
	clusterIDs []string
	calls      int
}

func (f *fakeClusterListAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Method != http.MethodGet || r.URL.Path != "/organization/org-1/cluster" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.calls++
	if f.calls <= f.failures {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	results := make([]map[string]any, 0, len(f.clusterIDs))
	for _, id := range f.clusterIDs {
		results = append(results, map[string]any{
			"id":             id,
			"created_at":     "2026-10-01T00:00:00Z",
			"organization":   map[string]any{"id": "org-1"},
			"name":           id,
			"region":         "eu-west-3",
			"cloud_provider": "AWS",
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
}

// TestClient_getClusterByID pins the retry of the cluster list call: a transient 500 is retried
// (QOV-2356), a cluster missing from the list is a 404 that is not retried.
func TestClient_getClusterByID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		Failures        int
		ClusterIDs      []string
		ExpectCalls     int
		ExpectError     bool
		ExpectNotFound  bool
		ExpectClusterID string
	}{
		{
			TestName:        "success_first_attempt",
			ClusterIDs:      []string{"cluster-0", "cluster-1"},
			ExpectCalls:     1,
			ExpectClusterID: "cluster-1",
		},
		{
			TestName:        "success_after_one_500",
			Failures:        1,
			ClusterIDs:      []string{"cluster-1"},
			ExpectCalls:     2,
			ExpectClusterID: "cluster-1",
		},
		{
			TestName:    "error_500_on_every_attempt",
			Failures:    maxRetryAttempts,
			ClusterIDs:  []string{"cluster-1"},
			ExpectCalls: maxRetryAttempts,
			ExpectError: true,
		},
		{
			TestName:       "error_cluster_not_listed_is_not_retried",
			ClusterIDs:     []string{"cluster-0"},
			ExpectCalls:    1,
			ExpectError:    true,
			ExpectNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &fakeClusterListAPI{failures: tc.Failures, clusterIDs: tc.ClusterIDs}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			c := New("token", "test", server.URL)
			cluster, apiErr := c.getClusterByID(context.Background(), "org-1", "cluster-1")

			assert.Equal(t, tc.ExpectCalls, api.calls)
			if tc.ExpectError {
				require.NotNil(t, apiErr)
				assert.Nil(t, cluster)
				assert.Equal(t, tc.ExpectNotFound, apierrors.IsNotFound(apiErr))
				return
			}
			require.Nil(t, apiErr)
			require.NotNil(t, cluster)
			assert.Equal(t, tc.ExpectClusterID, cluster.Id)
		})
	}
}
