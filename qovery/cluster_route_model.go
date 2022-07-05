package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/qovery/terraform-provider-qovery/client"
)

var clusterRouteAttrTypes = map[string]attr.Type{
	"description": types.StringType,
	"destination": types.StringType,
	"target":      types.StringType,
}

type ClusterRouteList []ClusterRoute

func (routes ClusterRouteList) toTerraformSet() types.Set {
	set := types.Set{
		ElemType: types.ObjectType{
			AttrTypes: clusterRouteAttrTypes,
		},
	}

	if routes == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(routes))
	for _, v := range routes {
		set.Elems = append(set.Elems, v.toTerraformObject())
	}
	return set
}

func (routes ClusterRouteList) toUpsertRequest() client.ClusterRoutingTable {
	list := make([]client.ClusterRoute, 0, len(routes))

	for _, r := range routes {
		list = append(list, r.toUpsertRequest())
	}

	return client.ClusterRoutingTable{
		Routes: list,
	}
}

type ClusterRoute struct {
	Description types.String `tfsdk:"description"`
	Destination types.String `tfsdk:"destination"`
	Target      types.String `tfsdk:"target"`
}

func (r ClusterRoute) toTerraformObject() types.Object {
	return types.Object{
		AttrTypes: clusterRouteAttrTypes,
		Attrs: map[string]attr.Value{
			"description": r.Description,
			"destination": r.Destination,
			"target":      r.Target,
		},
	}
}

func (r ClusterRoute) toUpsertRequest() client.ClusterRoute {
	return client.ClusterRoute{
		Description: toString(r.Description),
		Destination: toString(r.Destination),
		Target:      toString(r.Target),
	}
}

func fromClusterRoute(r client.ClusterRoute) ClusterRoute {
	return ClusterRoute{
		Description: fromString(r.Description),
		Destination: fromString(r.Destination),
		Target:      fromString(r.Target),
	}
}

func fromClusterRoutingTable(routingTable *client.ClusterRoutingTable) ClusterRouteList {
	if routingTable == nil {
		return nil
	}

	list := make([]ClusterRoute, 0, len(routingTable.Routes))
	for _, v := range routingTable.Routes {
		list = append(list, fromClusterRoute(v))
	}

	if len(list) == 0 {
		return nil
	}
	return list
}

func toClusterRoute(r types.Object) ClusterRoute {
	return ClusterRoute{
		Description: r.Attrs["description"].(types.String),
		Destination: r.Attrs["destination"].(types.String),
		Target:      r.Attrs["target"].(types.String),
	}
}

func toClusterRouteList(routes types.Set) ClusterRouteList {
	if routes.Null || routes.Unknown {
		return nil
	}

	clusterRoutes := make([]ClusterRoute, 0, len(routes.Elems))
	for _, elem := range routes.Elems {
		clusterRoutes = append(clusterRoutes, toClusterRoute(elem.(types.Object)))
	}

	return clusterRoutes
}
