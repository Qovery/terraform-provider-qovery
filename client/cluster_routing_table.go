package client

import (
	"context"

	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

type ClusterRoutingTable struct {
	Routes []ClusterRoute
}

func (rt ClusterRoutingTable) toQoveryRequest() qovery.ClusterRoutingTableRequest {
	routes := make([]qovery.ClusterRoutingTableResultsInner, 0, len(rt.Routes))
	for _, route := range rt.Routes {
		routes = append(routes, route.toQoveryRequest())
	}

	return qovery.ClusterRoutingTableRequest{
		Routes: routes,
	}
}

func newClusterRoutingTableFromQoveryResponse(resp *qovery.ClusterRoutingTable) ClusterRoutingTable {
	routes := make([]ClusterRoute, 0, len(resp.GetResults()))
	for _, route := range resp.Results {
		routes = append(routes, newClusterRouteFromQoveryResponse(route))
	}

	return ClusterRoutingTable{
		Routes: routes,
	}
}

type ClusterRoute struct {
	Description string
	Destination string
	Target      string
}

func (cr ClusterRoute) toQoveryRequest() qovery.ClusterRoutingTableResultsInner {
	return qovery.ClusterRoutingTableResultsInner{
		Description: cr.Description,
		Destination: cr.Destination,
		Target:      cr.Target,
	}
}

func newClusterRouteFromQoveryResponse(resp qovery.ClusterRoutingTableResultsInner) ClusterRoute {
	return ClusterRoute{
		Description: resp.Description,
		Destination: resp.Destination,
		Target:      resp.Target,
	}
}

func (c *Client) getClusterRoutingTable(ctx context.Context, organizationID string, clusterID string) (*ClusterRoutingTable, *apierrors.APIError) {
	routingTable, res, err := c.api.ClustersAPI.
		GetRoutingTable(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterRoutingTable, clusterID, res, err)
	}

	resp := newClusterRoutingTableFromQoveryResponse(routingTable)
	return &resp, nil
}

func (c *Client) editClusterRoutingTable(ctx context.Context, organizationID string, clusterID string, request ClusterRoutingTable) (*ClusterRoutingTable, *apierrors.APIError) {
	routingTable, res, err := c.api.ClustersAPI.
		EditRoutingTable(ctx, organizationID, clusterID).
		ClusterRoutingTableRequest(request.toQoveryRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterRoutingTable, clusterID, res, err)
	}

	resp := newClusterRoutingTableFromQoveryResponse(routingTable)
	return &resp, nil
}

// syncClusterRoutingTable makes the remote routing table hold exactly the desired routes (an
// empty list means no route) and returns the resulting table. The edit endpoint is only called
// when the tables differ, so a cluster that never had routes never reaches it.
func (c *Client) syncClusterRoutingTable(ctx context.Context, organizationID string, clusterID string, desired ClusterRoutingTable) (*ClusterRoutingTable, *apierrors.APIError) {
	current, apiErr := c.getClusterRoutingTable(ctx, organizationID, clusterID)
	if apiErr != nil {
		return nil, apiErr
	}
	if current.hasSameRoutes(desired) {
		return current, nil
	}
	return c.editClusterRoutingTable(ctx, organizationID, clusterID, desired)
}

// hasSameRoutes reports whether both tables hold the same routes, in any order.
func (rt ClusterRoutingTable) hasSameRoutes(other ClusterRoutingTable) bool {
	if len(rt.Routes) != len(other.Routes) {
		return false
	}
	remaining := make(map[ClusterRoute]int, len(rt.Routes))
	for _, route := range rt.Routes {
		remaining[route]++
	}
	for _, route := range other.Routes {
		if remaining[route] == 0 {
			return false
		}
		remaining[route]--
	}
	return true
}
