//go:build unit && !integration
// +build unit,!integration

package deploymentrestriction

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var restrictionObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":    types.StringType,
		"mode":  types.StringType,
		"type":  types.StringType,
		"value": types.StringType,
	},
}

func restrictionObject(t *testing.T, id types.String, value string) attr.Value {
	t.Helper()
	obj, diags := types.ObjectValue(restrictionObjectType.AttrTypes, map[string]attr.Value{
		"id":    id,
		"mode":  types.StringValue("EXCLUDE"),
		"type":  types.StringValue("PATH"),
		"value": types.StringValue(value),
	})
	require.False(t, diags.HasError(), diags)
	return obj
}

func restrictionSet(t *testing.T, elements ...attr.Value) types.Set {
	t.Helper()
	set, diags := types.SetValue(restrictionObjectType, elements)
	require.False(t, diags.HasError(), diags)
	return set
}

func TestToDeploymentRestrictionDiff(t *testing.T) {
	t.Parallel()

	state := restrictionSet(t,
		restrictionObject(t, types.StringValue("rule-1"), "docs/"),
		restrictionObject(t, types.StringValue("rule-2"), "tests/"),
	)

	testCases := []struct {
		name         string
		plan         types.Set
		state        *types.Set
		expectCreate []string
		expectUpdate []string
		expectDelete []string
	}{
		{
			name:         "null_plan_deletes_every_rule_in_state",
			plan:         types.SetNull(restrictionObjectType),
			state:        &state,
			expectDelete: []string{"rule-1", "rule-2"},
		},
		{
			name:  "unknown_plan_gives_empty_diff",
			plan:  types.SetUnknown(restrictionObjectType),
			state: &state,
		},
		{
			name:         "empty_plan_deletes_every_rule_in_state",
			plan:         restrictionSet(t),
			state:        &state,
			expectDelete: []string{"rule-1", "rule-2"},
		},
		{
			name:  "null_plan_without_state_gives_empty_diff",
			plan:  types.SetNull(restrictionObjectType),
			state: nil,
		},
		{
			name: "explicit_plan_updates_kept_rule_creates_new_one_and_deletes_the_rest",
			plan: restrictionSet(t,
				restrictionObject(t, types.StringValue("rule-1"), "docs/"),
				restrictionObject(t, types.StringUnknown(), "infra/"),
			),
			state:        &state,
			expectCreate: []string{"infra/"},
			expectUpdate: []string{"docs/"},
			expectDelete: []string{"rule-2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff, err := ToDeploymentRestrictionDiff(tc.plan, tc.state)

			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expectCreate, restrictionValues(diff.Create))
			assert.ElementsMatch(t, tc.expectUpdate, restrictionValues(diff.Update))
			assert.ElementsMatch(t, tc.expectDelete, diff.Delete)
		})
	}
}

func restrictionValues(restrictions []ServiceDeploymentRestriction) []string {
	values := make([]string, 0, len(restrictions))
	for _, restriction := range restrictions {
		values = append(values, restriction.Value)
	}
	return values
}
