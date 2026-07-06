//go:build unit && !integration
// +build unit,!integration

// Schema-wiring tests for the immutable existing-VPC blocks of qovery_cluster.
// Modifier semantics are covered in plan_modifiers_test.go
// (TestRejectExistingVpcChange_*); these tests only guard that both blocks
// carry the block-level modifier and that it fires through the wired schema.
package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clusterFeaturesNestedAttribute returns the named nested block under the
// cluster resource's features attribute.
func clusterFeaturesNestedAttribute(t *testing.T, name string) schema.SingleNestedAttribute {
	t.Helper()

	var resp resource.SchemaResponse
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	features, ok := resp.Schema.Attributes["features"].(schema.SingleNestedAttribute)
	require.True(t, ok, "features must be a SingleNestedAttribute")
	block, ok := features.Attributes[name].(schema.SingleNestedAttribute)
	require.True(t, ok, "features.%s must be a SingleNestedAttribute", name)
	return block
}

// TestClusterExistingVpcBlocks_RejectImmutableChanges guards both existing-VPC
// blocks: the API ignores changes to them after cluster creation, so Terraform
// must reject any change — removal of the block or a child value change — with
// a plan error instead of replacing the cluster. Because the guard is wired at
// the object level, every child attribute (including ones added later) is
// immutable by default; there is no per-attribute wiring to forget.
func TestClusterExistingVpcBlocks_RejectImmutableChanges(t *testing.T) {
	t.Parallel()

	for _, blockName := range []string{"existing_vpc", "gcp_existing_vpc"} {
		blockName := blockName
		t.Run(blockName, func(t *testing.T) {
			t.Parallel()

			block := clusterFeaturesNestedAttribute(t, blockName)

			wired := false
			for _, mod := range block.PlanModifiers {
				if _, ok := mod.(rejectExistingVpcChangeModifier); ok {
					wired = true
				}
			}
			require.True(t, wired,
				"features.%s must carry RejectExistingVpcChange at the object level", blockName)

			attrTypes := map[string]attr.Type{"id": types.StringType}
			stateValue := types.ObjectValueMust(attrTypes, map[string]attr.Value{"id": types.StringValue("vpc-a")})
			raw := types.StringValue("non-empty-resource")

			immutableChanges := []struct {
				changeName string
				planValue  types.Object
			}{
				{"block_removal", types.ObjectNull(attrTypes)},
				{"child_value_change", types.ObjectValueMust(attrTypes, map[string]attr.Value{"id": types.StringValue("vpc-b")})},
			}
			for _, change := range immutableChanges {
				requiresReplace := false
				hasError := false
				for _, mod := range block.PlanModifiers {
					resp := &planmodifier.ObjectResponse{PlanValue: change.planValue}
					mod.PlanModifyObject(context.Background(), planmodifier.ObjectRequest{
						State:      buildState(&raw),
						StateValue: stateValue,
						Plan:       buildPlan(&raw),
						PlanValue:  change.planValue,
					}, resp)
					requiresReplace = requiresReplace || resp.RequiresReplace
					hasError = hasError || resp.Diagnostics.HasError()
				}
				assert.False(t, requiresReplace,
					"features.%s %s must not force cluster replacement", blockName, change.changeName)
				assert.True(t, hasError,
					"features.%s %s must produce a plan error", blockName, change.changeName)
			}
		})
	}
}
