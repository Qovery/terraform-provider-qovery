//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers -----------------------------------------------------------------------------

func testKarpenterLimitsObject() types.Object {
	return types.ObjectValueMust(karpenterLimitsAttrTypes(), map[string]attr.Value{
		"enabled":                 types.BoolValue(true),
		"max_cpu_in_vcpu":         types.Int64Value(10),
		"max_memory_in_gibibytes": types.Int64Value(20),
	})
}

func testKarpenterConsolidationObject() types.Object {
	return types.ObjectValueMust(karpenterConsolidationAttrTypes(), map[string]attr.Value{
		"enabled":    types.BoolValue(true),
		"days":       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("MONDAY")}),
		"start_time": types.StringValue("PT02:00"),
		"duration":   types.StringValue("PT04H00M"),
	})
}

func testStableOverrideObject(spotEnabled types.Bool, consolidation, limits attr.Value) types.Object {
	return types.ObjectValueMust(karpenterStableOverrideAttrTypes(), map[string]attr.Value{
		"spot_enabled":  spotEnabled,
		"consolidation": consolidation,
		"limits":        limits,
	})
}

func testDefaultOverrideObject(spotEnabled types.Bool, limits attr.Value) types.Object {
	return types.ObjectValueMust(karpenterDefaultOverrideAttrTypes(), map[string]attr.Value{
		"spot_enabled": spotEnabled,
		"limits":       limits,
	})
}

func testCronjobOverrideObject(spotEnabled types.Bool) types.Object {
	return types.ObjectValueMust(karpenterCronjobOverrideAttrTypes(), map[string]attr.Value{
		"spot_enabled": spotEnabled,
	})
}

// testKarpenterObject builds the Terraform karpenter feature object, with every node pool
// override absent unless overrides says otherwise.
func testKarpenterObject(spotEnabled types.Bool, overrides map[string]attr.Value) types.Object {
	nodePools := map[string]attr.Value{
		"requirements": types.ListValueMust(
			types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()},
			[]attr.Value{
				types.ObjectValueMust(karpenterRequirementAttrTypes(), map[string]attr.Value{
					"key":      types.StringValue("InstanceFamily"),
					"operator": types.StringValue("In"),
					"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("t3a")}),
				}),
				types.ObjectValueMust(karpenterRequirementAttrTypes(), map[string]attr.Value{
					"key":      types.StringValue("InstanceSize"),
					"operator": types.StringValue("In"),
					"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("medium")}),
				}),
				types.ObjectValueMust(karpenterRequirementAttrTypes(), map[string]attr.Value{
					"key":      types.StringValue("Arch"),
					"operator": types.StringValue("In"),
					"values":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("AMD64")}),
				}),
			},
		),
		"stable_override":  types.ObjectNull(karpenterStableOverrideAttrTypes()),
		"default_override": types.ObjectNull(karpenterDefaultOverrideAttrTypes()),
		"cronjob_override": types.ObjectNull(karpenterCronjobOverrideAttrTypes()),
	}
	for name, value := range overrides {
		nodePools[name] = value
	}

	return types.ObjectValueMust(createKarpenterFeatureAttrTypes(), map[string]attr.Value{
		"spot_enabled":                 spotEnabled,
		"disk_size_in_gib":             types.Int64Value(50),
		"default_service_architecture": types.StringValue("AMD64"),
		"qovery_node_pools":            types.ObjectValueMust(karpenterNodePoolsAttrTypes(), nodePools),
	})
}

func testApiLimits() *qovery.KarpenterNodePoolLimits {
	return qovery.NewKarpenterNodePoolLimits(true, 10, 20, 0)
}

// testApiKarpenterParameters builds an API Karpenter payload with the given global flag and
// node pool overrides.
func testApiKarpenterParameters(globalSpotEnabled bool, nodePools qovery.KarpenterNodePool) *qovery.ClusterFeatureKarpenterParameters {
	nodePools.Requirements = []qovery.KarpenterNodePoolRequirement{{
		Key:      qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH,
		Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN,
		Values:   []string{"AMD64"},
	}}

	return &qovery.ClusterFeatureKarpenterParameters{
		SpotEnabled:                globalSpotEnabled,
		DiskSizeInGib:              50,
		DefaultServiceArchitecture: qovery.CPUARCHITECTUREENUM_AMD64,
		QoveryNodePools:            nodePools,
	}
}

func stateOverride(t *testing.T, attrVals map[string]attr.Value, name string) types.Object {
	t.Helper()

	nodePools, ok := attrVals["qovery_node_pools"].(types.Object)
	require.True(t, ok, "qovery_node_pools missing from the converted state")

	override, ok := nodePools.Attributes()[name].(types.Object)
	require.True(t, ok, "%s missing from the converted state", name)

	return override
}

// --- TF -> API ---------------------------------------------------------------------------

func TestExtractStableNodePoolOverrideFromTypesObject_SpotEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		StableOverride attr.Value
		ExpectNil      bool
		ExpectSpot     *bool
		ExpectLimits   bool
		ErrorContains  string
	}{
		{
			TestName:       "absent_block_is_not_sent",
			StableOverride: types.ObjectNull(karpenterStableOverrideAttrTypes()),
			ExpectNil:      true,
		},
		{
			TestName:       "spot_enabled_only_is_a_valid_block",
			StableOverride: testStableOverrideObject(types.BoolValue(true), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
			ExpectSpot:     boolPtr(true),
		},
		{
			TestName:       "spot_enabled_false_is_sent_explicitly",
			StableOverride: testStableOverrideObject(types.BoolValue(false), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
			ExpectSpot:     boolPtr(false),
		},
		{
			TestName:       "null_spot_enabled_is_left_out_of_the_request",
			StableOverride: testStableOverrideObject(types.BoolNull(), types.ObjectNull(karpenterConsolidationAttrTypes()), testKarpenterLimitsObject()),
			ExpectSpot:     nil,
			ExpectLimits:   true,
		},
		{
			TestName:       "unknown_spot_enabled_is_left_out_of_the_request",
			StableOverride: testStableOverrideObject(types.BoolUnknown(), types.ObjectNull(karpenterConsolidationAttrTypes()), testKarpenterLimitsObject()),
			ExpectSpot:     nil,
			ExpectLimits:   true,
		},
		{
			TestName:       "spot_enabled_alongside_limits",
			StableOverride: testStableOverrideObject(types.BoolValue(true), types.ObjectNull(karpenterConsolidationAttrTypes()), testKarpenterLimitsObject()),
			ExpectSpot:     boolPtr(true),
			ExpectLimits:   true,
		},
		{
			TestName:       "error_empty_block",
			StableOverride: testStableOverrideObject(types.BoolNull(), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
			ErrorContains:  "you must define at least its `spot_enabled`, its `consolidation` or its `limits`",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			karpenter := testKarpenterObject(types.BoolValue(true), map[string]attr.Value{"stable_override": tc.StableOverride})

			override, err := extractStableNodePoolOverrideFromTypesObject(karpenter)
			if tc.ErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ErrorContains)
				return
			}
			require.NoError(t, err)

			if tc.ExpectNil {
				assert.Nil(t, override)
				return
			}
			require.NotNil(t, override)

			assert.Equal(t, tc.ExpectSpot, GetStableNodePoolSpotEnabled(override))
			assert.Equal(t, tc.ExpectLimits, override.Limits != nil)
		})
	}
}

func TestExtractStableNodePoolOverrideFromTypesObject_ConsolidationStillWorks(t *testing.T) {
	t.Parallel()

	karpenter := testKarpenterObject(types.BoolValue(true), map[string]attr.Value{
		"stable_override": testStableOverrideObject(types.BoolNull(), testKarpenterConsolidationObject(), types.ObjectNull(karpenterLimitsAttrTypes())),
	})

	override, err := extractStableNodePoolOverrideFromTypesObject(karpenter)
	require.NoError(t, err)
	require.NotNil(t, override)
	require.NotNil(t, override.Consolidation)
	assert.Equal(t, "PT02:00", override.Consolidation.StartTime)
	assert.Nil(t, GetStableNodePoolSpotEnabled(override))
}

func TestExtractDefaultNodePoolOverrideFromTypesObject_SpotEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		DefaultOverride attr.Value
		ExpectNil       bool
		ExpectSpot      *bool
		ExpectLimits    bool
		ErrorContains   string
	}{
		{
			TestName:        "absent_block_is_not_sent",
			DefaultOverride: types.ObjectNull(karpenterDefaultOverrideAttrTypes()),
			ExpectNil:       true,
		},
		{
			TestName:        "spot_enabled_only_is_a_valid_block",
			DefaultOverride: testDefaultOverrideObject(types.BoolValue(true), types.ObjectNull(karpenterLimitsAttrTypes())),
			ExpectSpot:      boolPtr(true),
		},
		{
			TestName:        "limits_only_keeps_working",
			DefaultOverride: testDefaultOverrideObject(types.BoolNull(), testKarpenterLimitsObject()),
			ExpectLimits:    true,
		},
		{
			TestName:        "unknown_spot_enabled_is_left_out_of_the_request",
			DefaultOverride: testDefaultOverrideObject(types.BoolUnknown(), testKarpenterLimitsObject()),
			ExpectLimits:    true,
		},
		{
			TestName:        "spot_enabled_alongside_limits",
			DefaultOverride: testDefaultOverrideObject(types.BoolValue(false), testKarpenterLimitsObject()),
			ExpectSpot:      boolPtr(false),
			ExpectLimits:    true,
		},
		{
			TestName:        "error_empty_block",
			DefaultOverride: testDefaultOverrideObject(types.BoolNull(), types.ObjectNull(karpenterLimitsAttrTypes())),
			ErrorContains:   "you must define at least its `spot_enabled` or its `limits`",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			karpenter := testKarpenterObject(types.BoolValue(true), map[string]attr.Value{"default_override": tc.DefaultOverride})

			override, err := extractDefaultNodePoolOverrideFromTypesObject(karpenter)
			if tc.ErrorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ErrorContains)
				return
			}
			require.NoError(t, err)

			if tc.ExpectNil {
				assert.Nil(t, override)
				return
			}
			require.NotNil(t, override)

			assert.Equal(t, tc.ExpectSpot, GetDefaultNodePoolSpotEnabled(override))
			assert.Equal(t, tc.ExpectLimits, override.Limits != nil)
		})
	}
}

func TestExtractCronjobNodePoolOverrideFromTypesObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		CronjobOverride attr.Value
		ExpectNil       bool
		ExpectSpot      *bool
	}{
		{
			// The presence of the block is what enables the dedicated cronjob node pool, so it
			// must never be synthesized for a configuration that does not declare it.
			TestName:        "absent_block_is_never_synthesized",
			CronjobOverride: types.ObjectNull(karpenterCronjobOverrideAttrTypes()),
			ExpectNil:       true,
		},
		{
			TestName:        "empty_block_is_sent_and_enables_the_pool",
			CronjobOverride: testCronjobOverrideObject(types.BoolNull()),
			ExpectSpot:      nil,
		},
		{
			TestName:        "unknown_spot_enabled_still_sends_the_block",
			CronjobOverride: testCronjobOverrideObject(types.BoolUnknown()),
			ExpectSpot:      nil,
		},
		{
			TestName:        "spot_enabled_true",
			CronjobOverride: testCronjobOverrideObject(types.BoolValue(true)),
			ExpectSpot:      boolPtr(true),
		},
		{
			TestName:        "spot_enabled_false",
			CronjobOverride: testCronjobOverrideObject(types.BoolValue(false)),
			ExpectSpot:      boolPtr(false),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			karpenter := testKarpenterObject(types.BoolValue(true), map[string]attr.Value{"cronjob_override": tc.CronjobOverride})

			override, err := extractCronjobNodePoolOverrideFromTypesObject(karpenter)
			require.NoError(t, err)

			if tc.ExpectNil {
				assert.Nil(t, override)
				return
			}
			require.NotNil(t, override)
			assert.Equal(t, tc.ExpectSpot, GetCronjobNodePoolSpotEnabled(override))
		})
	}
}

func TestToQoveryNodePools_SendsEveryConfiguredOverride(t *testing.T) {
	t.Parallel()

	karpenter := testKarpenterObject(types.BoolValue(false), map[string]attr.Value{
		"stable_override":  testStableOverrideObject(types.BoolValue(false), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
		"default_override": testDefaultOverrideObject(types.BoolValue(true), testKarpenterLimitsObject()),
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(true)),
	})

	nodePools, err := toQoveryNodePools(karpenter)
	require.NoError(t, err)

	require.NotNil(t, nodePools.StableOverride)
	require.NotNil(t, nodePools.DefaultOverride)
	require.NotNil(t, nodePools.CronjobOverride)
	assert.Equal(t, false, *GetStableNodePoolSpotEnabled(nodePools.StableOverride))
	assert.Equal(t, true, *GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride))
	assert.Equal(t, true, *GetCronjobNodePoolSpotEnabled(nodePools.CronjobOverride))
	assert.Nil(t, nodePools.GpuOverride, "gpu_override is out of scope and must stay untouched")
}

// --- API -> TF ---------------------------------------------------------------------------

func TestCreateKarpenterFeatureAttrValue_BackfilledOverridesDoNotPolluteState(t *testing.T) {
	t.Parallel()

	// After the backend backfill every Karpenter cluster gets stable_override and
	// default_override with spot_enabled in the response, even when the configuration never
	// declared them. Storing those blocks would put a permanent diff in front of every user
	// who never wrote them.
	stable := qovery.KarpenterStableNodePoolOverride{}
	SetStableNodePoolSpotEnabled(&stable, true)
	defaultOverride := qovery.KarpenterDefaultNodePoolOverride{}
	SetDefaultNodePoolSpotEnabled(&defaultOverride, true)
	cronjob := qovery.KarpenterCronjobNodePoolOverride{}
	SetCronjobNodePoolSpotEnabled(&cronjob, true)

	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  &stable,
		DefaultOverride: &defaultOverride,
		CronjobOverride: &cronjob,
	})

	plan := testKarpenterObject(types.BoolValue(true), nil)

	attrVals := createKarpenterFeatureAttrValue(parameters, plan)
	require.NotNil(t, attrVals)

	assert.True(t, stateOverride(t, attrVals, "stable_override").IsNull())
	assert.True(t, stateOverride(t, attrVals, "default_override").IsNull())
	assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull())
}

func TestCreateKarpenterFeatureAttrValue_InjectsDeclaredOverrides(t *testing.T) {
	t.Parallel()

	stable := qovery.KarpenterStableNodePoolOverride{}
	SetStableNodePoolSpotEnabled(&stable, false)
	defaultOverride := qovery.KarpenterDefaultNodePoolOverride{Limits: testApiLimits()}
	SetDefaultNodePoolSpotEnabled(&defaultOverride, true)
	cronjob := qovery.KarpenterCronjobNodePoolOverride{}
	SetCronjobNodePoolSpotEnabled(&cronjob, true)

	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  &stable,
		DefaultOverride: &defaultOverride,
		CronjobOverride: &cronjob,
	})

	plan := testKarpenterObject(types.BoolNull(), map[string]attr.Value{
		"stable_override":  testStableOverrideObject(types.BoolValue(false), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
		"default_override": testDefaultOverrideObject(types.BoolValue(true), testKarpenterLimitsObject()),
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(true)),
	})

	attrVals := createKarpenterFeatureAttrValue(parameters, plan)
	require.NotNil(t, attrVals)

	stableState := stateOverride(t, attrVals, "stable_override")
	require.False(t, stableState.IsNull())
	assert.Equal(t, types.BoolValue(false), stableState.Attributes()["spot_enabled"])
	assert.True(t, stableState.Attributes()["limits"].IsNull())

	defaultState := stateOverride(t, attrVals, "default_override")
	require.False(t, defaultState.IsNull())
	assert.Equal(t, types.BoolValue(true), defaultState.Attributes()["spot_enabled"])
	assert.False(t, defaultState.Attributes()["limits"].IsNull())

	cronjobState := stateOverride(t, attrVals, "cronjob_override")
	require.False(t, cronjobState.IsNull())
	assert.Equal(t, types.BoolValue(true), cronjobState.Attributes()["spot_enabled"])
}

func TestCreateKarpenterFeatureAttrValue_LimitsOnlyOverrideKeepsBeingInjected(t *testing.T) {
	t.Parallel()

	// Legacy behavior: an override the API returns with actual content is stored in state even
	// when the plan does not declare it, and its spot_enabled reads as null while the API does
	// not send one.
	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  &qovery.KarpenterStableNodePoolOverride{Limits: testApiLimits()},
		DefaultOverride: &qovery.KarpenterDefaultNodePoolOverride{Limits: testApiLimits()},
	})

	attrVals := createKarpenterFeatureAttrValue(parameters, testKarpenterObject(types.BoolValue(true), nil))
	require.NotNil(t, attrVals)

	stableState := stateOverride(t, attrVals, "stable_override")
	require.False(t, stableState.IsNull())
	assert.True(t, stableState.Attributes()["spot_enabled"].IsNull())
	assert.False(t, stableState.Attributes()["limits"].IsNull())

	defaultState := stateOverride(t, attrVals, "default_override")
	require.False(t, defaultState.IsNull())
	assert.True(t, defaultState.Attributes()["spot_enabled"].IsNull())
}

func TestCreateKarpenterFeatureAttrValue_KeepsPlannedSpotWhileApiDoesNotEchoIt(t *testing.T) {
	t.Parallel()

	// The backend rollout (QOV-2155) is not deployed everywhere yet: until it is, the API drops
	// the per node pool flag. Falling back to the planned value keeps the apply consistent.
	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  &qovery.KarpenterStableNodePoolOverride{},
		CronjobOverride: &qovery.KarpenterCronjobNodePoolOverride{},
	})

	plan := testKarpenterObject(types.BoolValue(true), map[string]attr.Value{
		"stable_override":  testStableOverrideObject(types.BoolValue(false), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(true)),
	})

	attrVals := createKarpenterFeatureAttrValue(parameters, plan)
	require.NotNil(t, attrVals)

	assert.Equal(t, types.BoolValue(false), stateOverride(t, attrVals, "stable_override").Attributes()["spot_enabled"])
	assert.Equal(t, types.BoolValue(true), stateOverride(t, attrVals, "cronjob_override").Attributes()["spot_enabled"])
}

func TestCreateKarpenterFeatureAttrValue_EmptyDeclaredBlockStaysUnset(t *testing.T) {
	t.Parallel()

	// An empty cronjob_override is legal — its presence alone enables the dedicated node pool —
	// and on create its spot_enabled is unknown. It must resolve to a known value (null here,
	// since the API sends none) rather than stay unknown, which Terraform would reject.
	parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{
		CronjobOverride: &qovery.KarpenterCronjobNodePoolOverride{},
	})

	plan := testKarpenterObject(types.BoolUnknown(), map[string]attr.Value{
		"cronjob_override": testCronjobOverrideObject(types.BoolUnknown()),
	})

	attrVals := createKarpenterFeatureAttrValue(parameters, plan)
	require.NotNil(t, attrVals)

	cronjobState := stateOverride(t, attrVals, "cronjob_override")
	require.False(t, cronjobState.IsNull(), "a declared cronjob_override must stay in state")
	spotEnabled := cronjobState.Attributes()["spot_enabled"]
	assert.True(t, spotEnabled.IsNull())
	assert.False(t, spotEnabled.IsUnknown())
}

func TestCreateKarpenterFeatureAttrValue_GlobalSpotEnabled(t *testing.T) {
	t.Parallel()

	stableTrue := func() *qovery.KarpenterStableNodePoolOverride {
		o := &qovery.KarpenterStableNodePoolOverride{}
		SetStableNodePoolSpotEnabled(o, true)
		return o
	}

	testCases := []struct {
		TestName     string
		ApiGlobal    bool
		NodePools    qovery.KarpenterNodePool
		Plan         types.Object
		ExpectGlobal types.Bool
	}{
		{
			// No per node pool value: the API echoes the flag, so drift is reported as before.
			TestName:     "api_value_wins_without_per_pool_values",
			ApiGlobal:    false,
			Plan:         testKarpenterObject(types.BoolValue(true), nil),
			ExpectGlobal: types.BoolValue(false),
		},
		{
			// The API derives the flag as the OR of the per node pool values, so it no longer
			// matches what Terraform planned. Keeping the planned value avoids failing the apply
			// with "provider produced inconsistent result after apply".
			TestName:  "planned_value_wins_with_per_pool_values",
			ApiGlobal: true,
			NodePools: qovery.KarpenterNodePool{StableOverride: stableTrue()},
			Plan: testKarpenterObject(types.BoolValue(false), map[string]attr.Value{
				"stable_override": testStableOverrideObject(types.BoolValue(true), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
			}),
			ExpectGlobal: types.BoolValue(false),
		},
		{
			// On create the deprecated flag is left out of the configuration, so the plan is
			// unknown and the API value is what lands in state.
			TestName:  "api_value_wins_when_the_plan_is_unknown",
			ApiGlobal: true,
			NodePools: qovery.KarpenterNodePool{StableOverride: stableTrue()},
			Plan: testKarpenterObject(types.BoolUnknown(), map[string]attr.Value{
				"stable_override": testStableOverrideObject(types.BoolValue(true), types.ObjectNull(karpenterConsolidationAttrTypes()), types.ObjectNull(karpenterLimitsAttrTypes())),
			}),
			ExpectGlobal: types.BoolValue(true),
		},
		{
			TestName:     "api_value_wins_without_a_plan",
			ApiGlobal:    true,
			NodePools:    qovery.KarpenterNodePool{StableOverride: stableTrue()},
			Plan:         types.ObjectNull(createKarpenterFeatureAttrTypes()),
			ExpectGlobal: types.BoolValue(true),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			attrVals := createKarpenterFeatureAttrValue(testApiKarpenterParameters(tc.ApiGlobal, tc.NodePools), tc.Plan)
			require.NotNil(t, attrVals)
			assert.Equal(t, tc.ExpectGlobal, attrVals["spot_enabled"])
		})
	}
}

func TestCreateKarpenterFeatureAttrValue_WithoutPlanEverythingIsInjected(t *testing.T) {
	t.Parallel()

	// The data source has no plan to stay consistent with, so it reports what the API returns,
	// cronjob_override included.
	stable := qovery.KarpenterStableNodePoolOverride{}
	SetStableNodePoolSpotEnabled(&stable, false)
	cronjob := qovery.KarpenterCronjobNodePoolOverride{}
	SetCronjobNodePoolSpotEnabled(&cronjob, true)

	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  &stable,
		CronjobOverride: &cronjob,
	})

	attrVals := createKarpenterFeatureAttrValue(parameters, types.ObjectNull(createKarpenterFeatureAttrTypes()))
	require.NotNil(t, attrVals)

	assert.Equal(t, types.BoolValue(false), stateOverride(t, attrVals, "stable_override").Attributes()["spot_enabled"])
	assert.Equal(t, types.BoolValue(true), stateOverride(t, attrVals, "cronjob_override").Attributes()["spot_enabled"])
	assert.True(t, stateOverride(t, attrVals, "default_override").IsNull())
}

// --- schema wiring -------------------------------------------------------------------------

// TestClusterFeaturesSchemasMatchModel guards the rule that the resource and the data source
// share the Cluster model: the features object the model builds must be assignable to both
// schemas, otherwise the mismatch only shows up at runtime ("mismatch between struct and
// object"), never at build time.
func TestClusterFeaturesSchemasMatchModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	modelType := types.ObjectType{AttrTypes: createFeaturesAttrTypes()}

	var resourceSchema resource.SchemaResponse
	clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &resourceSchema)
	require.False(t, resourceSchema.Diagnostics.HasError())

	var dataSourceSchema datasource.SchemaResponse
	clusterDataSource{}.Schema(ctx, datasource.SchemaRequest{}, &dataSourceSchema)
	require.False(t, dataSourceSchema.Diagnostics.HasError())

	assert.True(t, modelType.Equal(resourceSchema.Schema.Attributes["features"].GetType()),
		"qovery_cluster resource features schema drifted from createFeaturesAttrTypes()")
	assert.True(t, modelType.Equal(dataSourceSchema.Schema.Attributes["features"].GetType()),
		"qovery_cluster data source features schema drifted from createFeaturesAttrTypes()")
}
