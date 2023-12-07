package qovery

import (
	"context"

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

func (routes ClusterRouteList) toTerraformSet(ctx context.Context, initialPlanClusterRouteSet types.Set) types.Set {
	var clusterRouteObjectType = types.ObjectType{
		AttrTypes: clusterRouteAttrTypes,
	}

	if len(initialPlanClusterRouteSet.Elements()) == 0 {
		return types.SetValueMust(clusterRouteObjectType, []attr.Value{})
	}

	if routes == nil {
		return types.SetNull(clusterRouteObjectType)
	}

	var elements = make([]attr.Value, 0, len(routes))
	for _, v := range routes {
		elements = append(elements, v.toTerraformObject())
	}
	set, diagnostics := types.SetValueFrom(ctx, clusterRouteObjectType, elements)
	if diagnostics.HasError() {
		panic("TODO")
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
	var attributes = map[string]attr.Value{
		"description": r.Description,
		"destination": r.Destination,
		"target":      r.Target,
	}
	terraformObjectValue, diagnostics := types.ObjectValue(clusterRouteAttrTypes, attributes)
	if diagnostics.HasError() {
		// TODO (framework-migration) Add new error checks
		panic("TODO")
	}
	return terraformObjectValue
}

func (r ClusterRoute) toUpsertRequest() client.ClusterRoute {
	return client.ClusterRoute{
		Description: ToString(r.Description),
		Destination: ToString(r.Destination),
		Target:      ToString(r.Target),
	}
}

func fromClusterRoute(r client.ClusterRoute) ClusterRoute {
	return ClusterRoute{
		Description: FromString(r.Description),
		Destination: FromString(r.Destination),
		Target:      FromString(r.Target),
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
		Description: r.Attributes()["description"].(types.String),
		Destination: r.Attributes()["destination"].(types.String),
		Target:      r.Attributes()["target"].(types.String),
	}
}

func toClusterRouteList(routes types.Set) ClusterRouteList {
	if routes.IsNull() || routes.IsUnknown() {
		return nil
	}

	clusterRoutes := make([]ClusterRoute, 0, len(routes.Elements()))
	for _, elem := range routes.Elements() {
		clusterRoutes = append(clusterRoutes, toClusterRoute(elem.(types.Object)))
	}

	return clusterRoutes
}
