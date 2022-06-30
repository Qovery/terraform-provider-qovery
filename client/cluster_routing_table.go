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
	routes := make([]qovery.ClusterRoutingTableRequestRoutesInner, 0, len(rt.Routes))
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

func (cr ClusterRoute) toQoveryRequest() qovery.ClusterRoutingTableRequestRoutesInner {
	return qovery.ClusterRoutingTableRequestRoutesInner{
		Description: cr.Description,
		Destination: cr.Destination,
		Target:      cr.Target,
	}
}

func newClusterRouteFromQoveryResponse(resp qovery.ClusterRoutingTableResultsInner) ClusterRoute {
	return ClusterRoute{
		Description: *resp.Description,
		Destination: *resp.Destination,
		Target:      *resp.Target,
	}
}

func (c *Client) getClusterRoutingTable(ctx context.Context, organizationID string, clusterID string) (*ClusterRoutingTable, *apierrors.APIError) {
	routingTable, res, err := c.api.ClustersApi.
		GetRoutingTable(ctx, organizationID, clusterID).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewReadError(apierrors.APIResourceClusterRoutingTable, clusterID, res, err)
	}

	resp := newClusterRoutingTableFromQoveryResponse(routingTable)
	return &resp, nil
}

func (c *Client) editClusterRoutingTable(ctx context.Context, organizationID string, clusterID string, request ClusterRoutingTable) (*ClusterRoutingTable, *apierrors.APIError) {
	routingTable, res, err := c.api.ClustersApi.
		EditRoutingTable(ctx, organizationID, clusterID).
		ClusterRoutingTableRequest(request.toQoveryRequest()).
		Execute()
	if err != nil || res.StatusCode >= 400 {
		return nil, apierrors.NewUpdateError(apierrors.APIResourceClusterRoutingTable, clusterID, res, err)
	}

	resp := newClusterRoutingTableFromQoveryResponse(routingTable)
	return &resp, nil
}
