//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterResource_UpgradeStateV0ToV1(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	routeType := types.ObjectType{AttrTypes: clusterRouteAttrTypes}
	declaredRoute := types.ObjectValueMust(clusterRouteAttrTypes, map[string]attr.Value{
		"description": types.StringValue("vpn"),
		"destination": types.StringValue("10.1.0.0/16"),
		"target":      types.StringValue("vgw-1"),
	})

	testCases := []struct {
		TestName        string
		PriorRoutes     types.Set
		ExpectNull      bool
		ExpectRouteSize int
	}{
		{TestName: "empty_routing_table_written_by_0x_becomes_null", PriorRoutes: types.SetValueMust(routeType, []attr.Value{}), ExpectNull: true},
		{TestName: "declared_routes_are_kept", PriorRoutes: types.SetValueMust(routeType, []attr.Value{declaredRoute}), ExpectRouteSize: 1},
		{TestName: "null_routing_table_stays_null", PriorRoutes: types.SetNull(routeType), ExpectNull: true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			upgraders := clusterResource{}.UpgradeState(ctx)
			upgrader, ok := upgraders[0]
			require.True(t, ok)
			require.NotNil(t, upgrader.PriorSchema)

			priorState := tfsdk.State{
				Schema: *upgrader.PriorSchema,
				Raw:    tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), nil),
			}
			require.False(t, priorState.SetAttribute(ctx, path.Root("id"), "cluster-123").HasError())
			require.False(t, priorState.SetAttribute(ctx, path.Root("routing_table"), tc.PriorRoutes).HasError())

			var schemaResp resource.SchemaResponse
			clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

			upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

			var id types.String
			require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
			assert.Equal(t, "cluster-123", id.ValueString())

			var routes types.Set
			require.False(t, resp.State.GetAttribute(ctx, path.Root("routing_table"), &routes).HasError())
			if tc.ExpectNull {
				assert.True(t, routes.IsNull())
				return
			}
			assert.Len(t, routes.Elements(), tc.ExpectRouteSize)
		})
	}
}
