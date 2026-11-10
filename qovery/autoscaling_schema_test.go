//go:build unit && !integration

package qovery

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/autoscaling"
)

func i32(v int32) *int32  { return &v }
func sp(v string) *string { return &v }

func TestAutoscaling_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers: []autoscaling.Scaler{
			{
				ScalerType: "prometheus",
				Enabled:    true,
				Role:       autoscaling.RolePrimary,
				Config:     autoscaling.Config{ConfigJSON: `{"query":"up","threshold":"1"}`},
				TriggerAuth: &autoscaling.TriggerAuth{
					Name:       "auth",
					ConfigYAML: sp("foo: bar"),
				},
			},
			{
				ScalerType: "cron",
				Enabled:    false,
				Role:       autoscaling.RoleSafety,
				Config:     autoscaling.Config{ConfigYAML: "start: 0 0 * * *"},
			},
		},
	}

	obj := fromAutoscaling(original, types.ObjectNull(autoscalingAttrTypes()))
	require.False(t, obj.IsNull())

	got := toQoveryAutoscaling(obj)
	require.NotNil(t, got)

	require.NotNil(t, got.PollingIntervalSeconds)
	assert.Equal(t, int32(30), *got.PollingIntervalSeconds)
	require.NotNil(t, got.CooldownPeriodSeconds)
	assert.Equal(t, int32(300), *got.CooldownPeriodSeconds)
	require.Len(t, got.Scalers, 2)

	byType := map[string]autoscaling.Scaler{}
	for _, s := range got.Scalers {
		byType[s.ScalerType] = s
	}

	prom := byType["prometheus"]
	assert.True(t, prom.Enabled)
	assert.Equal(t, autoscaling.RolePrimary, prom.Role)
	assert.JSONEq(t, `{"query":"up","threshold":"1"}`, prom.Config.ConfigJSON)
	assert.Empty(t, prom.Config.ConfigYAML)
	require.NotNil(t, prom.TriggerAuth)
	assert.Equal(t, "auth", prom.TriggerAuth.Name)
	require.NotNil(t, prom.TriggerAuth.ConfigYAML)
	assert.Equal(t, "foo: bar", *prom.TriggerAuth.ConfigYAML)

	cron := byType["cron"]
	assert.False(t, cron.Enabled)
	assert.Equal(t, autoscaling.RoleSafety, cron.Role)
	assert.Equal(t, "start: 0 0 * * *", cron.Config.ConfigYAML)
	assert.Empty(t, cron.Config.ConfigJSON)
	assert.Nil(t, cron.TriggerAuth)
}

func TestAutoscaling_NilRoundTrip(t *testing.T) {
	t.Parallel()

	obj := fromAutoscaling(nil, types.ObjectNull(autoscalingAttrTypes()))
	assert.True(t, obj.IsNull())
	assert.Nil(t, toQoveryAutoscaling(obj))
}

// promScaler returns a prometheus PRIMARY scaler on query, with a trigger
// authentication named name (none when name is "").
func promScaler(query, name string) autoscaling.Scaler {
	s := autoscaling.Scaler{
		ScalerType: "prometheus",
		Enabled:    true,
		Role:       autoscaling.RolePrimary,
		Config:     autoscaling.Config{ConfigJSON: `{"query":"` + query + `"}`},
	}
	if name != "" {
		s.TriggerAuth = &autoscaling.TriggerAuth{Name: name, ConfigYAML: sp("secretTargetRef: []")}
	}
	return s
}

func TestTriggerAuthNames(t *testing.T) {
	t.Parallel()

	reformatted := promScaler("up", "configured")
	reformatted.Config.ConfigJSON = "{ \"query\" : \"up\" }"

	otherAuthConfig := promScaler("up", "configured")
	otherAuthConfig.TriggerAuth.ConfigYAML = sp("secretTargetRef: [{parameter: apiKey}]")

	testCases := []struct {
		TestName string
		Actual   []autoscaling.Scaler
		Prior    []autoscaling.Scaler
		Expected []string
	}{
		{
			TestName: "no_prior_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    nil,
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "matching_scaler_keeps_configured_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    []autoscaling.Scaler{promScaler("up", "configured")},
			Expected: []string{"configured"},
		},
		{
			TestName: "config_json_compared_as_json",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    []autoscaling.Scaler{reformatted},
			Expected: []string{"configured"},
		},
		{
			TestName: "each_scaler_keeps_its_own_name",
			Actual: []autoscaling.Scaler{
				promScaler("up", "a676e575-1-auth"),
				promScaler("down", "a676e575-2-auth"),
			},
			Prior: []autoscaling.Scaler{
				promScaler("down", "configured-down"),
				promScaler("up", "configured-up"),
			},
			Expected: []string{"configured-up", "configured-down"},
		},
		{
			TestName: "scalers_equal_but_name_each_take_one_name",
			Actual: []autoscaling.Scaler{
				promScaler("up", "a676e575-1-auth"),
				promScaler("up", "a676e575-2-auth"),
			},
			Prior: []autoscaling.Scaler{
				promScaler("up", "first"),
				promScaler("up", "second"),
			},
			Expected: []string{"first", "second"},
		},
		{
			TestName: "scaler_changed_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("down", "a676e575-1-auth")},
			Prior:    []autoscaling.Scaler{promScaler("up", "configured")},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "trigger_auth_config_changed_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    []autoscaling.Scaler{otherAuthConfig},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "trigger_auth_added_outside_terraform_keeps_api_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "a676e575-1-auth")},
			Prior:    []autoscaling.Scaler{promScaler("up", "")},
			Expected: []string{"a676e575-1-auth"},
		},
		{
			TestName: "scaler_without_trigger_auth_has_no_name",
			Actual:   []autoscaling.Scaler{promScaler("up", "")},
			Prior:    []autoscaling.Scaler{promScaler("up", "configured")},
			Expected: []string{""},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.Expected, triggerAuthNames(tc.Actual, tc.Prior))
		})
	}
}

// The state keeps the configured name of an inline trigger authentication,
// which the API replaces with a generated one.
func TestFromAutoscaling_KeepsConfiguredTriggerAuthName(t *testing.T) {
	t.Parallel()

	prior := fromAutoscaling(&autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers:                []autoscaling.Scaler{promScaler("up", "vocca-livekit-eu-staging-datadog-calls")},
	}, types.ObjectNull(autoscalingAttrTypes()))

	actual := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers:                []autoscaling.Scaler{promScaler("up", "a676e575-1793957741362-auth")},
	}

	assert.True(t, fromAutoscaling(actual, prior).Equal(prior))
	assert.Equal(t, "a676e575-1793957741362-auth", actual.Scalers[0].TriggerAuth.Name, "the API policy must not be modified")
}

func withID(s autoscaling.Scaler, id string) autoscaling.Scaler {
	s.ID = id
	return s
}

func scalerQueries(scalers []autoscaling.Scaler) []string {
	queries := make([]string, len(scalers))
	for i, s := range scalers {
		queries[i] = s.Config.ConfigJSON + "#" + s.ID
	}
	return queries
}

func TestOrderScalersLikePrior(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Actual   []autoscaling.Scaler
		Prior    []autoscaling.Scaler
		Expected []autoscaling.Scaler
	}{
		{
			TestName: "no_prior_keeps_api_order",
			Actual:   []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2")},
			Expected: []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2")},
		},
		{
			TestName: "prior_order_followed_by_id",
			Actual:   []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2")},
			Prior:    []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("a", ""), "1")},
			Expected: []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("a", ""), "1")},
		},
		{
			TestName: "id_matches_even_when_attributes_changed",
			Actual:   []autoscaling.Scaler{withID(promScaler("a2", ""), "1"), withID(promScaler("b2", ""), "2")},
			Prior:    []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("a", ""), "1")},
			Expected: []autoscaling.Scaler{withID(promScaler("b2", ""), "2"), withID(promScaler("a2", ""), "1")},
		},
		{
			TestName: "scalers_without_id_matched_by_attributes",
			Actual:   []autoscaling.Scaler{withID(promScaler("a", "gen-1"), "1"), withID(promScaler("b", ""), "2")},
			Prior:    []autoscaling.Scaler{promScaler("b", ""), promScaler("a", "configured")},
			Expected: []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("a", "gen-1"), "1")},
		},
		{
			TestName: "id_match_wins_over_attribute_match",
			Actual:   []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("a", ""), "2")},
			Prior:    []autoscaling.Scaler{promScaler("a", ""), withID(promScaler("a", ""), "1")},
			Expected: []autoscaling.Scaler{withID(promScaler("a", ""), "2"), withID(promScaler("a", ""), "1")},
		},
		{
			TestName: "added_scaler_placed_like_the_plan",
			Actual:   []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2"), withID(promScaler("new", ""), "3")},
			Prior:    []autoscaling.Scaler{withID(promScaler("b", ""), "2"), promScaler("new", ""), withID(promScaler("a", ""), "1")},
			Expected: []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("new", ""), "3"), withID(promScaler("a", ""), "1")},
		},
		{
			TestName: "scaler_unknown_to_prior_goes_last",
			Actual:   []autoscaling.Scaler{withID(promScaler("outside", ""), "3"), withID(promScaler("a", ""), "1")},
			Prior:    []autoscaling.Scaler{withID(promScaler("a", ""), "1")},
			Expected: []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("outside", ""), "3")},
		},
		{
			TestName: "scaler_deleted_outside_terraform_is_dropped",
			Actual:   []autoscaling.Scaler{withID(promScaler("b", ""), "2")},
			Prior:    []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2")},
			Expected: []autoscaling.Scaler{withID(promScaler("b", ""), "2")},
		},
		{
			TestName: "scaler_recreated_outside_terraform_matched_by_attributes",
			Actual:   []autoscaling.Scaler{withID(promScaler("b", ""), "2"), withID(promScaler("a", ""), "9")},
			Prior:    []autoscaling.Scaler{withID(promScaler("a", ""), "1"), withID(promScaler("b", ""), "2")},
			Expected: []autoscaling.Scaler{withID(promScaler("a", ""), "9"), withID(promScaler("b", ""), "2")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, scalerQueries(tc.Expected), scalerQueries(orderScalersLikePrior(tc.Actual, tc.Prior)))
		})
	}
}

// The API returns scalers in creation order; the state keeps the order of the
// configuration and the id of each scaler, so a refresh plans no change.
func TestFromAutoscaling_KeepsPriorOrderAndIDs(t *testing.T) {
	t.Parallel()

	prior := fromAutoscaling(&autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers: []autoscaling.Scaler{
			withID(promScaler("b", "configured-b"), "2"),
			withID(promScaler("a", ""), "1"),
		},
	}, types.ObjectNull(autoscalingAttrTypes()))

	actual := &autoscaling.AutoscalingPolicy{
		PollingIntervalSeconds: i32(30),
		CooldownPeriodSeconds:  i32(300),
		Scalers: []autoscaling.Scaler{
			withID(promScaler("a", ""), "1"),
			withID(promScaler("b", "gen-auth"), "2"),
		},
	}

	got := fromAutoscaling(actual, prior)
	assert.True(t, got.Equal(prior), "got %s, want %s", got, prior)
}

func TestToQoveryAutoscaling_ScalerID(t *testing.T) {
	t.Parallel()

	obj := fromAutoscaling(&autoscaling.AutoscalingPolicy{
		Scalers: []autoscaling.Scaler{withID(promScaler("a", ""), "1"), promScaler("b", "")},
	}, types.ObjectNull(autoscalingAttrTypes()))

	scalers := obj.Attributes()["scalers"].(types.List).Elements()
	assert.True(t, scalers[1].(types.Object).Attributes()["id"].IsNull(), "a scaler without id has a null id")

	got := toQoveryAutoscaling(obj)
	require.Len(t, got.Scalers, 2)
	assert.Equal(t, "1", got.Scalers[0].ID)
	assert.Empty(t, got.Scalers[1].ID)

	unknownAttrs := scalers[0].(types.Object).Attributes()
	unknownAttrs["id"] = types.StringUnknown()
	planned := types.ObjectValueMust(autoscalingAttrTypes(), map[string]attr.Value{
		"polling_interval_seconds": types.Int64Null(),
		"cooldown_period_seconds":  types.Int64Null(),
		"scalers": types.ListValueMust(types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}, []attr.Value{
			types.ObjectValueMust(autoscalingScalerAttrTypes(), unknownAttrs),
		}),
	})
	assert.Empty(t, toQoveryAutoscaling(planned).Scalers[0].ID, "an unknown id is not sent")
}

// TestServiceResource_UpgradeStateAutoscaling runs the upgraders of application
// and container on a state whose scalers are a set without id.
func TestServiceResource_UpgradeStateAutoscaling(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	scalerAttrTypesV1 := autoscalingScalerAttrTypes()
	delete(scalerAttrTypesV1, "id")
	autoscalingAttrTypesV1 := autoscalingAttrTypes()
	autoscalingAttrTypesV1["scalers"] = types.SetType{ElemType: types.ObjectType{AttrTypes: scalerAttrTypesV1}}

	triggerAuth := types.ObjectValueMust(autoscalingTriggerAuthAttrTypes(), map[string]attr.Value{
		"name":        types.StringValue("auth"),
		"config_yaml": types.StringValue("secretTargetRef: []"),
	})
	scalerV1 := func(query string) attr.Value {
		return types.ObjectValueMust(scalerAttrTypesV1, map[string]attr.Value{
			"scaler_type":            types.StringValue("prometheus"),
			"enabled":                types.BoolValue(true),
			"role":                   types.StringValue("PRIMARY"),
			"config_json":            jsontypes.NewNormalizedValue(`{"query":"` + query + `"}`),
			"config_yaml":            types.StringNull(),
			"trigger_authentication": triggerAuth,
		})
	}
	autoscalingV1 := types.ObjectValueMust(autoscalingAttrTypesV1, map[string]attr.Value{
		"polling_interval_seconds": types.Int64Value(30),
		"cooldown_period_seconds":  types.Int64Value(300),
		"scalers":                  types.SetValueMust(types.ObjectType{AttrTypes: scalerAttrTypesV1}, []attr.Value{scalerV1("a"), scalerV1("b")}),
	})

	resources := map[string]interface {
		resource.Resource
		resource.ResourceWithUpgradeState
	}{
		"application": applicationResource{},
		"container":   containerResource{},
	}

	for name, r := range resources {
		for _, version := range []int64{0, 1} {
			for _, prior := range []types.Object{autoscalingV1, types.ObjectNull(autoscalingAttrTypesV1)} {
				t.Run(fmt.Sprintf("%s/v%d/null_%t", name, version, prior.IsNull()), func(t *testing.T) {
					t.Parallel()

					var schemaResp resource.SchemaResponse
					r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

					upgrader, ok := r.UpgradeState(ctx)[version]
					require.True(t, ok)
					require.NotNil(t, upgrader.PriorSchema)

					priorState := tfsdk.State{
						Schema: *upgrader.PriorSchema,
						Raw:    tftypes.NewValue(upgrader.PriorSchema.Type().TerraformType(ctx), nil),
					}
					require.False(t, priorState.SetAttribute(ctx, path.Root("id"), "service-123").HasError())
					diags := priorState.SetAttribute(ctx, path.Root("autoscaling"), prior)
					require.False(t, diags.HasError(), "%v", diags)

					resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
					upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
					require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

					var got types.Object
					require.False(t, resp.State.GetAttribute(ctx, path.Root("autoscaling"), &got).HasError())
					if prior.IsNull() {
						assert.True(t, got.IsNull())
						return
					}

					policy := toQoveryAutoscaling(got)
					require.NotNil(t, policy)
					assert.Equal(t, int32(30), *policy.PollingIntervalSeconds)
					require.Len(t, policy.Scalers, 2)
					assert.ElementsMatch(t, []string{`{"query":"a"}`, `{"query":"b"}`},
						[]string{policy.Scalers[0].Config.ConfigJSON, policy.Scalers[1].Config.ConfigJSON})
					for _, s := range got.Attributes()["scalers"].(types.List).Elements() {
						attrs := s.(types.Object).Attributes()
						assert.True(t, attrs["id"].IsNull(), "the next refresh fills the id")
						assert.True(t, attrs["trigger_authentication"].Equal(triggerAuth))
					}
				})
			}
		}
	}
}

func cronScaler(start string) autoscaling.Scaler {
	return autoscaling.Scaler{
		ScalerType: "cron",
		Enabled:    true,
		Role:       autoscaling.RoleSafety,
		Config:     autoscaling.Config{ConfigYAML: "start: " + start},
	}
}

// scalersListValue returns the `scalers` list of scalers, with an unknown id for
// a scaler whose ID is "?".
func scalersListValue(scalers ...autoscaling.Scaler) types.List {
	list := fromAutoscaling(&autoscaling.AutoscalingPolicy{Scalers: scalers}, types.ObjectNull(autoscalingAttrTypes())).Attributes()["scalers"].(types.List)
	elems := list.Elements()
	for i, s := range scalers {
		if s.ID != "?" {
			continue
		}
		attrs := elems[i].(types.Object).Attributes()
		attrs["id"] = types.StringUnknown()
		elems[i] = types.ObjectValueMust(autoscalingScalerAttrTypes(), attrs)
	}
	return types.ListValueMust(list.ElementType(context.Background()), elems)
}

// plannedScalerIDs returns the planned id of each scaler: "?" when unknown, ""
// when null.
func plannedScalerIDs(list types.List) []string {
	ids := make([]string, 0, len(list.Elements()))
	for _, elem := range list.Elements() {
		id := elem.(types.Object).Attributes()["id"].(types.String)
		switch {
		case id.IsUnknown():
			ids = append(ids, "?")
		case id.IsNull():
			ids = append(ids, "")
		default:
			ids = append(ids, id.ValueString())
		}
	}
	return ids
}

func TestScalerIDsFromStateModifier(t *testing.T) {
	t.Parallel()

	promA := promScaler("a", "auth")
	cron := cronScaler("0 0 * * *")
	state := scalersListValue(withID(promA, "P"), withID(cron, "C"))

	unknownConfig := scalersListValue(withID(promA, "?"))
	unknownAttrs := unknownConfig.Elements()[0].(types.Object).Attributes()
	unknownAttrs["config_json"] = jsontypes.NewNormalizedUnknown()
	unknownConfig = types.ListValueMust(unknownConfig.ElementType(context.Background()), []attr.Value{
		types.ObjectValueMust(autoscalingScalerAttrTypes(), unknownAttrs),
	})

	testCases := []struct {
		TestName string
		State    types.List
		Plan     types.List
		Expected []string
	}{
		{
			TestName: "create_plans_unknown_ids",
			State:    types.ListNull(types.ObjectType{AttrTypes: autoscalingScalerAttrTypes()}),
			Plan:     scalersListValue(withID(promA, "?"), withID(cron, "?")),
			Expected: []string{"?", "?"},
		},
		{
			TestName: "unchanged_keeps_ids",
			State:    state,
			Plan:     scalersListValue(withID(promA, "P"), withID(cron, "C")),
			Expected: []string{"P", "C"},
		},
		{
			TestName: "config_order_differs_from_state_order",
			State:    state,
			Plan:     scalersListValue(withID(cron, "?"), withID(promA, "?")),
			Expected: []string{"C", "P"},
		},
		{
			TestName: "scaler_inserted_in_the_middle_is_created",
			State:    state,
			Plan:     scalersListValue(withID(promA, "?"), withID(promScaler("b", ""), "?"), withID(cron, "?")),
			Expected: []string{"P", "?", "C"},
		},
		{
			TestName: "first_scaler_removed",
			State:    state,
			Plan:     scalersListValue(withID(cron, "?")),
			Expected: []string{"C"},
		},
		{
			TestName: "modified_scaler_keeps_its_id",
			State:    state,
			Plan:     scalersListValue(withID(promScaler("x", "auth"), "?"), withID(cron, "?")),
			Expected: []string{"P", "C"},
		},
		{
			TestName: "reordered_and_modified_scalers_keep_their_ids",
			State:    state,
			Plan:     scalersListValue(withID(cronScaler("0 1 * * *"), "?"), withID(promScaler("x", "auth"), "?")),
			Expected: []string{"C", "P"},
		},
		{
			TestName: "scaler_type_changed_in_place",
			State:    scalersListValue(withID(promA, "P")),
			Plan:     scalersListValue(withID(cron, "?")),
			Expected: []string{"P"},
		},
		{
			TestName: "state_without_ids_plans_unknown_ids",
			State:    scalersListValue(promA, cron),
			Plan:     scalersListValue(withID(cron, "?"), withID(promA, "?")),
			Expected: []string{"?", "?"},
		},
		{
			TestName: "unknown_attribute_plans_unknown_ids",
			State:    state,
			Plan:     unknownConfig,
			Expected: []string{"?"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.ListRequest{
				Path:       path.Root("autoscaling").AtName("scalers"),
				StateValue: tc.State,
				PlanValue:  tc.Plan,
			}
			resp := &planmodifier.ListResponse{PlanValue: tc.Plan}
			scalerIDsFromStateModifier{}.PlanModifyList(context.Background(), req, resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Equal(t, tc.Expected, plannedScalerIDs(resp.PlanValue))
		})
	}
}
