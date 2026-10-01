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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRoutingTableAPI serves the cluster routing table endpoints and records every edit.
type fakeRoutingTableAPI struct {
	mu     sync.Mutex
	routes []ClusterRoute
	edits  [][]ClusterRoute
}

type fakeRoutingTableRoute struct {
	Description string `json:"description"`
	Destination string `json:"destination"`
	Target      string `json:"target"`
}

func (f *fakeRoutingTableAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.URL.Path != "/organization/org-1/cluster/cluster-1/routingTable" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPut {
		var body struct {
			Routes []fakeRoutingTableRoute `json:"routes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		routes := make([]ClusterRoute, 0, len(body.Routes))
		for _, route := range body.Routes {
			routes = append(routes, ClusterRoute(route))
		}
		f.routes = routes
		f.edits = append(f.edits, routes)
	}

	results := make([]fakeRoutingTableRoute, 0, len(f.routes))
	for _, route := range f.routes {
		results = append(results, fakeRoutingTableRoute(route))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
}

func TestClient_syncClusterRoutingTable(t *testing.T) {
	t.Parallel()

	routeA := ClusterRoute{Description: "vpn", Destination: "10.1.0.0/16", Target: "vgw-1"}
	routeB := ClusterRoute{Description: "peering", Destination: "10.2.0.0/16", Target: "pcx-1"}
	routeARetargeted := ClusterRoute{Description: "vpn", Destination: "10.1.0.0/16", Target: "vgw-2"}

	testCases := []struct {
		TestName     string
		Remote       []ClusterRoute
		Desired      []ClusterRoute
		ExpectEdit   bool
		ExpectRoutes []ClusterRoute
	}{
		{
			TestName:     "no_routes_on_either_side_skips_the_edit",
			ExpectEdit:   false,
			ExpectRoutes: []ClusterRoute{},
		},
		{
			TestName:     "same_routes_in_another_order_skip_the_edit",
			Remote:       []ClusterRoute{routeA, routeB},
			Desired:      []ClusterRoute{routeB, routeA},
			ExpectEdit:   false,
			ExpectRoutes: []ClusterRoute{routeA, routeB},
		},
		{
			TestName:     "empty_desired_table_deletes_every_remote_route",
			Remote:       []ClusterRoute{routeA, routeB},
			ExpectEdit:   true,
			ExpectRoutes: []ClusterRoute{},
		},
		{
			TestName:     "route_added_in_config_is_written",
			Remote:       []ClusterRoute{routeA},
			Desired:      []ClusterRoute{routeA, routeB},
			ExpectEdit:   true,
			ExpectRoutes: []ClusterRoute{routeA, routeB},
		},
		{
			TestName:     "route_changed_outside_terraform_is_reverted",
			Remote:       []ClusterRoute{routeARetargeted},
			Desired:      []ClusterRoute{routeA},
			ExpectEdit:   true,
			ExpectRoutes: []ClusterRoute{routeA},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			api := &fakeRoutingTableAPI{routes: tc.Remote}
			server := httptest.NewServer(api)
			t.Cleanup(server.Close)

			c := New("token", "test", server.URL)
			got, apiErr := c.syncClusterRoutingTable(context.Background(), "org-1", "cluster-1", ClusterRoutingTable{Routes: tc.Desired})
			require.Nil(t, apiErr)
			require.NotNil(t, got)

			assert.ElementsMatch(t, tc.ExpectRoutes, got.Routes)
			assert.ElementsMatch(t, tc.ExpectRoutes, api.routes)
			if tc.ExpectEdit {
				require.Len(t, api.edits, 1)
				assert.ElementsMatch(t, tc.ExpectRoutes, api.edits[0])
			} else {
				assert.Empty(t, api.edits)
			}
		})
	}
}
