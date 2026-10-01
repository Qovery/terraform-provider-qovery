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
)

// fakeClusterStatusAPI serves the cluster status and deploy endpoints and counts the deploys.
// A deploy moves the cluster to DEPLOYED.
type fakeClusterStatusAPI struct {
	mu      sync.Mutex
	status  qovery.ClusterStateEnum
	deploys int
}

func (f *fakeClusterStatusAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/organization/org-1/cluster/cluster-1/status":
	case r.Method == http.MethodPost && r.URL.Path == "/organization/org-1/cluster/cluster-1/deploy":
		f.deploys++
		f.status = qovery.CLUSTERSTATEENUM_DEPLOYED
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"cluster_id":  "cluster-1",
		"status":      f.status,
		"is_deployed": f.status == qovery.CLUSTERSTATEENUM_DEPLOYED,
		"reason":      qovery.DEPLOYMENTINFRAREASON_UNSPECIFIED,
	})
}

// TestClient_updateClusterStatus pins when a cluster gets deployed. A create never forces the
// deploy (QOV-2355): a new managed cluster is READY and gets deployed anyway, while a new
// self-managed cluster is already DEPLOYED and q-core refuses to deploy it.
func TestClient_updateClusterStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Status        qovery.ClusterStateEnum
		ForceUpdate   bool
		ExpectDeploys int
	}{
		{
			TestName:      "ready_cluster_is_deployed_without_force",
			Status:        qovery.CLUSTERSTATEENUM_READY,
			ExpectDeploys: 1,
		},
		{
			TestName:      "deployed_cluster_is_not_deployed_without_force",
			Status:        qovery.CLUSTERSTATEENUM_DEPLOYED,
			ExpectDeploys: 0,
		},
		{
			TestName:      "deployed_cluster_is_deployed_again_with_force",
			Status:        qovery.CLUSTERSTATEENUM_DEPLOYED,
			ForceUpdate:   true,
			ExpectDeploys: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &fakeClusterStatusAPI{status: tc.Status}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			c := New("token", "test", server.URL)
			got, apiErr := c.updateClusterStatus(context.Background(), "org-1", &qovery.Cluster{Id: "cluster-1"}, qovery.CLUSTERSTATEENUM_DEPLOYED, tc.ForceUpdate)
			require.Nil(t, apiErr)
			require.NotNil(t, got)

			assert.Equal(t, qovery.CLUSTERSTATEENUM_DEPLOYED, *got)
			assert.Equal(t, tc.ExpectDeploys, api.deploys)
		})
	}
}
