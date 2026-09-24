//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func testNullConsolidation() types.Object {
	return types.ObjectNull(karpenterConsolidationAttrTypes())
}

func testNullLimits() types.Object {
	return types.ObjectNull(karpenterLimitsAttrTypes())
}

func testStableOverrideObject(spotEnabled types.Bool, consolidation, limits attr.Value) types.Object {
	return types.ObjectValueMust(karpenterStableOverrideAttrTypes(), map[string]attr.Value{
		"spot_enabled":  spotEnabled,
		"consolidation": consolidation,
		"limits":        limits,
	})
}

// testStableSpotOnly is a stable_override block that carries nothing but spot_enabled.
func testStableSpotOnly(spotEnabled types.Bool) types.Object {
	return testStableOverrideObject(spotEnabled, testNullConsolidation(), testNullLimits())
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
func testKarpenterObject(overrides map[string]attr.Value) types.Object {
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
		"gpu_override":     types.ObjectNull(karpenterGpuOverrideAttrTypes()),
	}
	for name, value := range overrides {
		nodePools[name] = value
	}

	return types.ObjectValueMust(createKarpenterFeatureAttrTypes(), map[string]attr.Value{
		"disk_size_in_gib":             types.Int64Value(50),
		"default_service_architecture": types.StringValue("AMD64"),
		"qovery_node_pools":            types.ObjectValueMust(karpenterNodePoolsAttrTypes(), nodePools),
	})
}

// testNoPlan stands for a read without a plan to stay consistent with: import or data source.
func testNoPlan() types.Object {
	return types.ObjectNull(createKarpenterFeatureAttrTypes())
}

// testFeaturesObject wraps a karpenter object in a complete cluster features object.
func testFeaturesObject(karpenter types.Object) types.Object {
	return types.ObjectValueMust(createFeaturesAttrTypes(), map[string]attr.Value{
		featureKeyVpcSubnet:      types.StringNull(),
		featureKeyStaticIP:       types.BoolNull(),
		featureKeyNatGateways:    types.ObjectNull(createNatGatewaysFeatureAttrTypes()),
		featureKeyExistingVpc:    types.ObjectNull(createExistingVpcFeatureAttrTypes()),
		featureKeyGcpExistingVpc: types.ObjectNull(createGcpExistingVpcFeatureAttrTypes()),
		featureKeyKarpenter:      karpenter,
		featureKeyGkeKmsKey:      types.StringNull(),
	})
}

func testApiLimits() *qovery.KarpenterNodePoolLimits {
	return qovery.NewKarpenterNodePoolLimits(true, 10, 20, 0)
}

func testApiStable(spotEnabled *bool) *qovery.KarpenterStableNodePoolOverride {
	o := &qovery.KarpenterStableNodePoolOverride{}
	if spotEnabled != nil {
		SetStableNodePoolSpotEnabled(o, *spotEnabled)
	}
	return o
}

func testApiDefault(spotEnabled *bool) *qovery.KarpenterDefaultNodePoolOverride {
	o := &qovery.KarpenterDefaultNodePoolOverride{}
	if spotEnabled != nil {
		SetDefaultNodePoolSpotEnabled(o, *spotEnabled)
	}
	return o
}

func testApiCronjob(spotEnabled *bool) *qovery.KarpenterCronjobNodePoolOverride {
	o := &qovery.KarpenterCronjobNodePoolOverride{}
	if spotEnabled != nil {
		SetCronjobNodePoolSpotEnabled(o, *spotEnabled)
	}
	return o
}

// testApiKarpenterParameters builds an API Karpenter payload the way the API returns it: the
// global flag it derived, and node pool overrides that omit any spot_enabled equal to it.
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

// stateSpot returns the spot_enabled stored for a node pool override, failing when the block is
// not in state.
func stateSpot(t *testing.T, attrVals map[string]attr.Value, name string) attr.Value {
	t.Helper()

	override := stateOverride(t, attrVals, name)
	require.False(t, override.IsNull(), "%s must be in state", name)
	return override.Attributes()["spot_enabled"]
}

// karpenterRequest builds the API request for a karpenter object and returns its Karpenter
// parameters.
func karpenterRequest(t *testing.T, karpenter types.Object) *qovery.ClusterFeatureKarpenterParameters {
	t.Helper()

	requestFeatures, err := toQoveryClusterFeatures(testFeaturesObject(karpenter), "MANAGED", "AWS")
	require.NoError(t, err)

	for _, f := range requestFeatures {
		value := f.GetValue()
		if parameters := value.ClusterFeatureKarpenterParameters; parameters != nil {
			return parameters
		}
	}

	t.Fatal("the request carries no karpenter feature")
	return nil
}

// --- TF -> API ---------------------------------------------------------------------------

func TestExtractStableNodePoolOverrideFromTypesObject_SpotEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		StableOverride attr.Value
		ExpectSpot     bool
		ExpectLimits   bool
	}{
		{
			// The stable node pool always exists. Leaving it out of the request would let the API
			// hand it the global flag, so an absent block is sent as on-demand.
			TestName:       "absent_block_is_sent_as_on_demand",
			StableOverride: types.ObjectNull(karpenterStableOverrideAttrTypes()),
			ExpectSpot:     false,
		},
		{
			TestName:       "spot_enabled_true_is_sent",
			StableOverride: testStableSpotOnly(types.BoolValue(true)),
			ExpectSpot:     true,
		},
		{
			TestName:       "spot_enabled_false_is_sent",
			StableOverride: testStableSpotOnly(types.BoolValue(false)),
			ExpectSpot:     false,
		},
		{
			TestName:       "null_spot_enabled_is_sent_as_on_demand",
			StableOverride: testStableOverrideObject(types.BoolNull(), testNullConsolidation(), testKarpenterLimitsObject()),
			ExpectSpot:     false,
			ExpectLimits:   true,
		},
		{
			TestName:       "unknown_spot_enabled_is_sent_as_on_demand",
			StableOverride: testStableOverrideObject(types.BoolUnknown(), testNullConsolidation(), testKarpenterLimitsObject()),
			ExpectSpot:     false,
			ExpectLimits:   true,
		},
		{
			TestName:       "spot_enabled_alongside_limits",
			StableOverride: testStableOverrideObject(types.BoolValue(true), testNullConsolidation(), testKarpenterLimitsObject()),
			ExpectSpot:     true,
			ExpectLimits:   true,
		},
		{
			// With spot_enabled defaulting to false an empty block is simply the on-demand pool.
			TestName:       "empty_block_is_sent_as_on_demand",
			StableOverride: testStableSpotOnly(types.BoolNull()),
			ExpectSpot:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			override, err := extractStableNodePoolOverrideFromTypesObject(testKarpenterObject(map[string]attr.Value{"stable_override": tc.StableOverride}))
			require.NoError(t, err)
			require.NotNil(t, override, "the stable node pool must always be sent")

			spotEnabled := GetStableNodePoolSpotEnabled(override)
			require.NotNil(t, spotEnabled, "spot_enabled must always be explicit")
			assert.Equal(t, tc.ExpectSpot, *spotEnabled)
			assert.Equal(t, tc.ExpectLimits, override.Limits != nil)
		})
	}
}

func TestExtractStableNodePoolOverrideFromTypesObject_ConsolidationStillWorks(t *testing.T) {
	t.Parallel()

	karpenter := testKarpenterObject(map[string]attr.Value{
		"stable_override": testStableOverrideObject(types.BoolNull(), testKarpenterConsolidationObject(), testNullLimits()),
	})

	override, err := extractStableNodePoolOverrideFromTypesObject(karpenter)
	require.NoError(t, err)
	require.NotNil(t, override)
	require.NotNil(t, override.Consolidation)
	assert.Equal(t, "PT02:00", override.Consolidation.StartTime)
	assert.Equal(t, boolPtr(false), GetStableNodePoolSpotEnabled(override))
}

func TestExtractDefaultNodePoolOverrideFromTypesObject_SpotEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName        string
		DefaultOverride attr.Value
		ExpectSpot      bool
		ExpectLimits    bool
	}{
		{
			TestName:        "absent_block_is_sent_as_on_demand",
			DefaultOverride: types.ObjectNull(karpenterDefaultOverrideAttrTypes()),
			ExpectSpot:      false,
		},
		{
			TestName:        "spot_enabled_true_is_sent",
			DefaultOverride: testDefaultOverrideObject(types.BoolValue(true), testNullLimits()),
			ExpectSpot:      true,
		},
		{
			TestName:        "spot_enabled_false_is_sent",
			DefaultOverride: testDefaultOverrideObject(types.BoolValue(false), testNullLimits()),
			ExpectSpot:      false,
		},
		{
			TestName:        "null_spot_enabled_is_sent_as_on_demand",
			DefaultOverride: testDefaultOverrideObject(types.BoolNull(), testKarpenterLimitsObject()),
			ExpectSpot:      false,
			ExpectLimits:    true,
		},
		{
			TestName:        "unknown_spot_enabled_is_sent_as_on_demand",
			DefaultOverride: testDefaultOverrideObject(types.BoolUnknown(), testKarpenterLimitsObject()),
			ExpectSpot:      false,
			ExpectLimits:    true,
		},
		{
			TestName:        "spot_enabled_alongside_limits",
			DefaultOverride: testDefaultOverrideObject(types.BoolValue(true), testKarpenterLimitsObject()),
			ExpectSpot:      true,
			ExpectLimits:    true,
		},
		{
			TestName:        "empty_block_is_sent_as_on_demand",
			DefaultOverride: testDefaultOverrideObject(types.BoolNull(), testNullLimits()),
			ExpectSpot:      false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			override, err := extractDefaultNodePoolOverrideFromTypesObject(testKarpenterObject(map[string]attr.Value{"default_override": tc.DefaultOverride}))
			require.NoError(t, err)
			require.NotNil(t, override, "the default node pool must always be sent")

			spotEnabled := GetDefaultNodePoolSpotEnabled(override)
			require.NotNil(t, spotEnabled, "spot_enabled must always be explicit")
			assert.Equal(t, tc.ExpectSpot, *spotEnabled)
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
		ExpectSpot      bool
	}{
		{
			// The presence of the block is what enables the dedicated cronjob node pool, so it
			// must never be synthesized for a configuration that does not declare it.
			TestName:        "absent_block_is_never_synthesized",
			CronjobOverride: types.ObjectNull(karpenterCronjobOverrideAttrTypes()),
			ExpectNil:       true,
		},
		{
			TestName:        "empty_block_is_sent_as_on_demand",
			CronjobOverride: testCronjobOverrideObject(types.BoolNull()),
			ExpectSpot:      false,
		},
		{
			TestName:        "unknown_spot_enabled_is_sent_as_on_demand",
			CronjobOverride: testCronjobOverrideObject(types.BoolUnknown()),
			ExpectSpot:      false,
		},
		{
			TestName:        "spot_enabled_true",
			CronjobOverride: testCronjobOverrideObject(types.BoolValue(true)),
			ExpectSpot:      true,
		},
		{
			TestName:        "spot_enabled_false",
			CronjobOverride: testCronjobOverrideObject(types.BoolValue(false)),
			ExpectSpot:      false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			override, err := extractCronjobNodePoolOverrideFromTypesObject(testKarpenterObject(map[string]attr.Value{"cronjob_override": tc.CronjobOverride}))
			require.NoError(t, err)

			if tc.ExpectNil {
				assert.Nil(t, override)
				return
			}
			require.NotNil(t, override)
			spotEnabled := GetCronjobNodePoolSpotEnabled(override)
			require.NotNil(t, spotEnabled, "spot_enabled must always be explicit")
			assert.Equal(t, tc.ExpectSpot, *spotEnabled)
		})
	}
}

func TestToQoveryNodePools_SendsEveryNodePool(t *testing.T) {
	t.Parallel()

	t.Run("declared_overrides", func(t *testing.T) {
		t.Parallel()

		nodePools, err := toQoveryNodePools(testKarpenterObject(map[string]attr.Value{
			"stable_override":  testStableSpotOnly(types.BoolValue(false)),
			"default_override": testDefaultOverrideObject(types.BoolValue(true), testKarpenterLimitsObject()),
			"cronjob_override": testCronjobOverrideObject(types.BoolValue(true)),
		}))
		require.NoError(t, err)

		assert.Equal(t, boolPtr(false), GetStableNodePoolSpotEnabled(nodePools.StableOverride))
		assert.Equal(t, boolPtr(true), GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride))
		assert.Equal(t, boolPtr(true), GetCronjobNodePoolSpotEnabled(nodePools.CronjobOverride))
		assert.Nil(t, nodePools.GpuOverride, "an undeclared gpu_override would create the GPU node pool")
	})

	t.Run("no_override_declared", func(t *testing.T) {
		t.Parallel()

		nodePools, err := toQoveryNodePools(testKarpenterObject(nil))
		require.NoError(t, err)

		assert.Equal(t, boolPtr(false), GetStableNodePoolSpotEnabled(nodePools.StableOverride))
		assert.Equal(t, boolPtr(false), GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride))
		assert.Nil(t, nodePools.CronjobOverride, "an undeclared cronjob_override would enable the dedicated pool")
		assert.Nil(t, nodePools.GpuOverride, "an undeclared gpu_override would create the GPU node pool")
	})
}

func TestToQoveryClusterFeatures_GlobalSpotEnabledIsTheOrOfTheNodePools(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName     string
		Overrides    map[string]attr.Value
		ExpectGlobal bool
	}{
		{
			TestName:     "no_override_declared",
			ExpectGlobal: false,
		},
		{
			TestName: "every_node_pool_on_demand",
			Overrides: map[string]attr.Value{
				"stable_override":  testStableSpotOnly(types.BoolValue(false)),
				"default_override": testDefaultOverrideObject(types.BoolValue(false), testNullLimits()),
				"cronjob_override": testCronjobOverrideObject(types.BoolValue(false)),
			},
			ExpectGlobal: false,
		},
		{
			TestName:     "stable_on_spot",
			Overrides:    map[string]attr.Value{"stable_override": testStableSpotOnly(types.BoolValue(true))},
			ExpectGlobal: true,
		},
		{
			TestName:     "default_on_spot",
			Overrides:    map[string]attr.Value{"default_override": testDefaultOverrideObject(types.BoolValue(true), testNullLimits())},
			ExpectGlobal: true,
		},
		{
			TestName:     "cronjob_on_spot",
			Overrides:    map[string]attr.Value{"cronjob_override": testCronjobOverrideObject(types.BoolValue(true))},
			ExpectGlobal: true,
		},
		{
			// q-core leaves the GPU node pool out of the global flag it recomputes.
			TestName: "gpu_on_spot_is_not_counted",
			Overrides: map[string]attr.Value{"gpu_override": testGpuOverrideObject(func(attrs map[string]attr.Value) {
				attrs["spot_enabled"] = types.BoolValue(true)
			})},
			ExpectGlobal: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			//nolint:staticcheck // SA1019: asserting the value sent for the deprecated global flag
			assert.Equal(t, tc.ExpectGlobal, karpenterRequest(t, testKarpenterObject(tc.Overrides)).SpotEnabled)
		})
	}
}

// TestCluster_toUpsertClusterRequest_UnsetNodePoolStaysOnDemandOnUpdate is the regression test for
// the 0.x write path. It sent the global flag from state on update, and the API hands that flag to
// every node pool the request leaves without a value. With only default_override pinned to spot,
// the stable node pool was created on-demand and then moved to spot by the next unrelated apply,
// because the global in state was the OR, true. The request must now be the same on create and on
// update, and never leave a pool without a value.
func TestCluster_toUpsertClusterRequest_UnsetNodePoolStaysOnDemandOnUpdate(t *testing.T) {
	t.Parallel()

	newCluster := func(description string) Cluster {
		return Cluster{
			Id:              types.StringValue("cluster-123"),
			OrganizationId:  types.StringValue("org-123"),
			CredentialsId:   types.StringValue("cred-123"),
			Name:            types.StringValue("test-cluster"),
			Description:     types.StringValue(description),
			CloudProvider:   types.StringValue("AWS"),
			Region:          types.StringValue("us-east-1"),
			KubernetesMode:  types.StringValue("MANAGED"),
			State:           types.StringValue("DEPLOYED"),
			InstanceType:    types.StringUnknown(),
			MinRunningNodes: types.Int64Unknown(),
			MaxRunningNodes: types.Int64Unknown(),
			DiskSize:        types.Int64Unknown(),
			Features: testFeaturesObject(testKarpenterObject(map[string]attr.Value{
				"default_override": testDefaultOverrideObject(types.BoolValue(true), testNullLimits()),
			})),
		}
	}

	requestNodePools := func(t *testing.T, plan Cluster, state *Cluster) qovery.KarpenterNodePool {
		t.Helper()
		request, err := plan.toUpsertClusterRequest(state)
		require.NoError(t, err)
		for _, f := range request.ClusterRequest.Features {
			value := f.GetValue()
			if parameters := value.ClusterFeatureKarpenterParameters; parameters != nil {
				return parameters.QoveryNodePools
			}
		}
		t.Fatal("the request carries no karpenter feature")
		return qovery.KarpenterNodePool{}
	}

	state := newCluster("before")

	for name, nodePools := range map[string]qovery.KarpenterNodePool{
		"create":           requestNodePools(t, newCluster("before"), nil),
		"unrelated_update": requestNodePools(t, newCluster("after"), &state),
	} {
		assert.Equal(t, boolPtr(false), GetStableNodePoolSpotEnabled(nodePools.StableOverride), "%s: the stable node pool must stay on-demand", name)
		assert.Equal(t, boolPtr(true), GetDefaultNodePoolSpotEnabled(nodePools.DefaultOverride), "%s", name)
	}
}

func TestCluster_hasFeaturesDiff_SpotChanges(t *testing.T) {
	t.Parallel()

	// hasFeaturesDiff drives ForceUpdate, i.e. whether the cluster gets redeployed.
	newCluster := func(overrides map[string]attr.Value) Cluster {
		return Cluster{
			CloudProvider:  types.StringValue("AWS"),
			KubernetesMode: types.StringValue("MANAGED"),
			Features:       testFeaturesObject(testKarpenterObject(overrides)),
		}
	}
	stableSpot := map[string]attr.Value{"stable_override": testStableSpotOnly(types.BoolValue(true))}
	stableOnDemand := map[string]attr.Value{"stable_override": testStableSpotOnly(types.BoolValue(false))}

	state := newCluster(stableSpot)

	assert.False(t, newCluster(stableSpot).hasFeaturesDiff(&state), "an unchanged configuration must not force a redeploy")
	assert.True(t, newCluster(stableOnDemand).hasFeaturesDiff(&state), "moving a node pool off spot must force a redeploy")

	// The read path stores an undeclared stable node pool that runs on spot, so removing the block
	// from the configuration is a real move to on-demand and must be deployed like one.
	assert.True(t, newCluster(nil).hasFeaturesDiff(&state), "removing a spot node pool's block must force a redeploy")

	// An undeclared on-demand pool and a declared on-demand pool send the same request.
	onDemandState := newCluster(stableOnDemand)
	assert.False(t, newCluster(nil).hasFeaturesDiff(&onDemandState), "removing an on-demand block changes nothing")
}

// --- API -> TF ---------------------------------------------------------------------------

func TestKarpenterFeatureAttrValue_UndeclaredOnDemandNodePoolsStayOutOfState(t *testing.T) {
	t.Parallel()

	// The API returns stable_override and default_override for every Karpenter cluster. Storing
	// them for a configuration that never declared them would be permanent plan noise, so an
	// on-demand pool with nothing else in its block stays out of state.
	responses := map[string]qovery.KarpenterNodePool{
		"values_omitted":  {StableOverride: testApiStable(nil), DefaultOverride: testApiDefault(nil)},
		"values_explicit": {StableOverride: testApiStable(boolPtr(false)), DefaultOverride: testApiDefault(boolPtr(false))},
		"blocks_absent":   {},
	}

	for name, nodePools := range responses {
		name, nodePools := name, nodePools
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attrVals := karpenterFeatureAttrValue(testApiKarpenterParameters(false, nodePools), testKarpenterObject(nil), clusterReadModeResource)
			require.NotNil(t, attrVals)

			assert.True(t, stateOverride(t, attrVals, "stable_override").IsNull())
			assert.True(t, stateOverride(t, attrVals, "default_override").IsNull())
			assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull())
		})
	}
}

func TestKarpenterFeatureAttrValue_UndeclaredSpotNodePoolsAreStored(t *testing.T) {
	t.Parallel()

	// A 0.x configuration that relied on the global flag, refreshed by 1.0: the API runs both pools
	// on spot and omits their values because they equal the global. Storing the blocks is what
	// makes the plan show their removal, i.e. the move to on-demand, instead of an apply doing it
	// silently.
	t.Run("every_pool_on_spot", func(t *testing.T) {
		t.Parallel()

		parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
			StableOverride:  testApiStable(nil),
			DefaultOverride: testApiDefault(nil),
		})

		attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
		require.NotNil(t, attrVals)

		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "stable_override"))
		assert.True(t, stateOverride(t, attrVals, "stable_override").Attributes()["limits"].IsNull())
		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "default_override"))
		assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull(), "the cronjob pool is not enabled on the API")
	})

	t.Run("only_the_spot_pool_is_stored", func(t *testing.T) {
		t.Parallel()

		// stable deviates from the global, so the API sends it; default equals it and is omitted.
		parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
			StableOverride:  testApiStable(boolPtr(false)),
			DefaultOverride: testApiDefault(nil),
		})

		attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
		require.NotNil(t, attrVals)

		assert.True(t, stateOverride(t, attrVals, "stable_override").IsNull(), "an on-demand pool matches the default and stays out of state")
		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "default_override"))
	})
}

func TestKarpenterFeatureAttrValue_DeclaredOverridesResolveOmittedValuesToTheGlobal(t *testing.T) {
	t.Parallel()

	// stable deviates from the derived global and is sent; default and cronjob equal it and are
	// omitted, which means they run on spot.
	parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
		StableOverride:  testApiStable(boolPtr(false)),
		DefaultOverride: &qovery.KarpenterDefaultNodePoolOverride{Limits: testApiLimits()},
		CronjobOverride: testApiCronjob(nil),
	})

	plan := testKarpenterObject(map[string]attr.Value{
		"stable_override":  testStableSpotOnly(types.BoolValue(false)),
		"default_override": testDefaultOverrideObject(types.BoolValue(true), testKarpenterLimitsObject()),
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(true)),
	})

	attrVals := karpenterFeatureAttrValue(parameters, plan, clusterReadModeResource)
	require.NotNil(t, attrVals)

	assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "stable_override"))
	assert.True(t, stateOverride(t, attrVals, "stable_override").Attributes()["limits"].IsNull())
	assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "default_override"))
	assert.False(t, stateOverride(t, attrVals, "default_override").Attributes()["limits"].IsNull())
	assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "cronjob_override"))
}

func TestKarpenterFeatureAttrValue_ReportsOutOfBandChangesOnDeclaredNodePools(t *testing.T) {
	t.Parallel()

	// Both pools were declared on spot, then moved to on-demand from the console. Every value now
	// equals the global false and is omitted. Falling back to the planned value, as 0.x did, kept
	// `true` in state and hid the change; resolving to the global reports it as drift.
	parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{
		StableOverride:  testApiStable(nil),
		DefaultOverride: testApiDefault(nil),
	})

	prior := testKarpenterObject(map[string]attr.Value{
		"stable_override":  testStableSpotOnly(types.BoolValue(true)),
		"default_override": testDefaultOverrideObject(types.BoolValue(true), testNullLimits()),
	})

	attrVals := karpenterFeatureAttrValue(parameters, prior, clusterReadModeResource)
	require.NotNil(t, attrVals)

	assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "stable_override"))
	assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "default_override"))
}

func TestKarpenterFeatureAttrValue_ContentOverrideIsStoredWithWhereItRuns(t *testing.T) {
	t.Parallel()

	// An override with limits is stored even when the configuration does not declare it, as it
	// always was. Its spot_enabled is now where the pool actually runs rather than null.
	for _, globalSpotEnabled := range []bool{true, false} {
		globalSpotEnabled := globalSpotEnabled
		t.Run(map[bool]string{true: "spot", false: "on_demand"}[globalSpotEnabled], func(t *testing.T) {
			t.Parallel()

			parameters := testApiKarpenterParameters(globalSpotEnabled, qovery.KarpenterNodePool{
				StableOverride:  &qovery.KarpenterStableNodePoolOverride{Limits: testApiLimits()},
				DefaultOverride: &qovery.KarpenterDefaultNodePoolOverride{Limits: testApiLimits()},
			})

			attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
			require.NotNil(t, attrVals)

			assert.Equal(t, types.BoolValue(globalSpotEnabled), stateSpot(t, attrVals, "stable_override"))
			assert.False(t, stateOverride(t, attrVals, "stable_override").Attributes()["limits"].IsNull())
			assert.Equal(t, types.BoolValue(globalSpotEnabled), stateSpot(t, attrVals, "default_override"))
		})
	}
}

// TestKarpenterFeatureAttrValue_CronjobPoolFollowsTheAPI pins the cronjob_override read rule: the
// block is stored exactly when the API returns it, because its presence is what enables the pool
// and q-core returns it only while the pool is enabled. A pool enabled or disabled outside
// Terraform must reach the state, so the plan shows the change instead of the next apply
// reverting it silently.
func TestKarpenterFeatureAttrValue_CronjobPoolFollowsTheAPI(t *testing.T) {
	t.Parallel()

	declaredCronjob := testKarpenterObject(map[string]attr.Value{"cronjob_override": testCronjobOverrideObject(types.BoolValue(true))})

	testCases := []struct {
		TestName    string
		ApiGlobal   bool
		ApiCronjob  *qovery.KarpenterCronjobNodePoolOverride
		Plan        types.Object
		ExpectBlock bool
		ExpectSpot  bool
	}{
		{
			// Enabled from the Console while the configuration does not declare it.
			TestName:    "undeclared_pool_enabled_on_the_api_is_stored",
			ApiGlobal:   false,
			ApiCronjob:  testApiCronjob(nil),
			Plan:        testKarpenterObject(nil),
			ExpectBlock: true,
			ExpectSpot:  false,
		},
		{
			// Its spot_enabled equals the global true, so the API omits it.
			TestName:    "undeclared_pool_on_spot_resolves_to_the_global",
			ApiGlobal:   true,
			ApiCronjob:  testApiCronjob(nil),
			Plan:        testKarpenterObject(nil),
			ExpectBlock: true,
			ExpectSpot:  true,
		},
		{
			// Disabled from the Console while the configuration declares it.
			TestName:    "declared_pool_disabled_on_the_api_is_dropped",
			ApiGlobal:   true,
			ApiCronjob:  nil,
			Plan:        declaredCronjob,
			ExpectBlock: false,
		},
		{
			TestName:    "declared_pool_enabled_on_the_api_is_stored",
			ApiGlobal:   false,
			ApiCronjob:  testApiCronjob(boolPtr(true)),
			Plan:        declaredCronjob,
			ExpectBlock: true,
			ExpectSpot:  true,
		},
		{
			TestName:    "import_stores_an_enabled_pool",
			ApiGlobal:   false,
			ApiCronjob:  testApiCronjob(nil),
			Plan:        testNoPlan(),
			ExpectBlock: true,
			ExpectSpot:  false,
		},
		{
			TestName:    "no_pool_anywhere",
			ApiGlobal:   false,
			ApiCronjob:  nil,
			Plan:        testKarpenterObject(nil),
			ExpectBlock: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			parameters := testApiKarpenterParameters(tc.ApiGlobal, qovery.KarpenterNodePool{CronjobOverride: tc.ApiCronjob})

			attrVals := karpenterFeatureAttrValue(parameters, tc.Plan, clusterReadModeResource)
			require.NotNil(t, attrVals)

			if !tc.ExpectBlock {
				assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull())
				return
			}
			assert.Equal(t, types.BoolValue(tc.ExpectSpot), stateSpot(t, attrVals, "cronjob_override"))
		})
	}
}

func TestKarpenterFeatureAttrValue_DeclaredEmptyCronjobResolvesToAKnownValue(t *testing.T) {
	t.Parallel()

	// An empty cronjob_override is legal: its presence alone enables the dedicated node pool. The
	// API omits its spot_enabled because it equals the global; state must hold a known value.
	parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{CronjobOverride: testApiCronjob(nil)})

	attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(map[string]attr.Value{
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(false)),
	}), clusterReadModeResource)
	require.NotNil(t, attrVals)

	assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "cronjob_override"))
}

func TestKarpenterFeatureAttrValue_WithoutPlan(t *testing.T) {
	t.Parallel()

	t.Run("empty_overrides_are_dropped", func(t *testing.T) {
		t.Parallel()

		// Regression: q-core returns a present-but-EMPTY stable_override for Karpenter clusters.
		// Injecting it on presence alone made import store an object where the apply path stores
		// null, and TestAcc_ClusterWithStaticIP, TestAcc_ClusterWithKeda and
		// TestAcc_ClusterWithReadyState/aws_eks all failed ImportStateVerify.
		parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: testApiStable(nil)})

		attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeResource)
		require.NotNil(t, attrVals)

		assert.True(t, stateOverride(t, attrVals, "stable_override").IsNull())
		assert.True(t, stateOverride(t, attrVals, "default_override").IsNull())
		assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull())
	})

	t.Run("spot_pools_and_content_are_stored", func(t *testing.T) {
		t.Parallel()

		// stable carries limits and deviates from the global; default equals the global true and is
		// omitted, so it runs on spot; cronjob is enabled, which import sees from its presence.
		parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
			StableOverride:  &qovery.KarpenterStableNodePoolOverride{Limits: testApiLimits(), SpotEnabled: *qovery.NewNullableBool(boolPtr(false))},
			DefaultOverride: testApiDefault(nil),
			CronjobOverride: testApiCronjob(nil),
		})

		attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeResource)
		require.NotNil(t, attrVals)

		assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "stable_override"))
		assert.False(t, stateOverride(t, attrVals, "stable_override").Attributes()["limits"].IsNull())
		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "default_override"))
		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "cronjob_override"))
	})

	t.Run("a_declared_on_demand_block_is_not_recoverable", func(t *testing.T) {
		t.Parallel()

		// The documented import trade-off: a block holding only spot_enabled = false says nothing
		// the default would not, so import cannot tell it was declared. The first plan after import
		// proposes adding it back, which changes nothing on the cluster.
		parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{
			StableOverride:  testApiStable(nil),
			DefaultOverride: testApiDefault(nil),
		})

		attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeResource)
		require.NotNil(t, attrVals)

		assert.True(t, stateOverride(t, attrVals, "stable_override").IsNull())
		assert.True(t, stateOverride(t, attrVals, "default_override").IsNull())
	})
}

// TestKarpenterFeatureAttrValue_ImportAgreesWithApply pins the invariant ImportStateVerify actually
// checks: for one API response, the state produced with a plan (apply/refresh) and the state
// produced without one (import) must be identical for a configuration that declares no node pool
// override. Asserting the two paths agree catches a divergence in either of them, which asserting
// each one against a literal does not.
func TestKarpenterFeatureAttrValue_ImportAgreesWithApply(t *testing.T) {
	t.Parallel()

	apiResponses := map[string]*qovery.ClusterFeatureKarpenterParameters{
		// What q-core returns for an on-demand Karpenter cluster: a present-but-empty override.
		"empty_stable_override": testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: testApiStable(nil)}),
		// Both pools on spot, values omitted because they equal the global.
		"spot_pools": testApiKarpenterParameters(true, qovery.KarpenterNodePool{StableOverride: testApiStable(nil), DefaultOverride: testApiDefault(nil)}),
		// One pool on spot, the other deviating to on-demand.
		"mixed_pools": testApiKarpenterParameters(true, qovery.KarpenterNodePool{StableOverride: testApiStable(boolPtr(false)), DefaultOverride: testApiDefault(nil)}),
		// An override with real content must survive both paths identically too.
		"stable_override_with_limits": testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: &qovery.KarpenterStableNodePoolOverride{Limits: testApiLimits()}}),
		// A cronjob pool enabled outside Terraform.
		"cronjob_pool_enabled": testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: testApiStable(nil), CronjobOverride: testApiCronjob(nil)}),
		// A GPU pool created outside Terraform.
		"gpu_pool_created": testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: testApiStable(nil), GpuOverride: testApiGpu()}),
	}

	for name, parameters := range apiResponses {
		name, parameters := name, parameters
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			applyState := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
			importState := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeResource)
			require.NotNil(t, applyState)
			require.NotNil(t, importState)

			for _, override := range karpenterNodePoolOverrideNames {
				assert.True(t,
					stateOverride(t, applyState, override).Equal(stateOverride(t, importState, override)),
					"%s differs between the apply and import paths: ImportStateVerify would fail", override)
			}
		})
	}
}

// --- data source mode -------------------------------------------------------------------------

func TestKarpenterFeatureAttrValue_DataSourceReportsEveryNodePool(t *testing.T) {
	t.Parallel()

	t.Run("resolves_omitted_values", func(t *testing.T) {
		t.Parallel()

		// stable equals the global true and is omitted; default and cronjob deviate and are sent.
		parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{
			StableOverride:  testApiStable(nil),
			DefaultOverride: testApiDefault(boolPtr(false)),
			CronjobOverride: testApiCronjob(boolPtr(false)),
		})

		attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeDataSource)
		require.NotNil(t, attrVals)

		assert.Equal(t, types.BoolValue(true), stateSpot(t, attrVals, "stable_override"))
		assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "default_override"))
		assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "cronjob_override"))
	})

	t.Run("on_demand_pools_are_reported_too", func(t *testing.T) {
		t.Parallel()

		// The data source has no plan to keep quiet: an on-demand pool is still where the pool
		// runs, and without the global flag it is the only place a reader can find it.
		parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{StableOverride: testApiStable(nil)})

		attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeDataSource)
		require.NotNil(t, attrVals)

		assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "stable_override"))
		assert.Equal(t, types.BoolValue(false), stateSpot(t, attrVals, "default_override"))
		assert.True(t, stateOverride(t, attrVals, "cronjob_override").IsNull(), "the cronjob pool is reported only while it is enabled")
	})
}

func TestKarpenterFeatureAttrValue_DataSourceKeepsContentOverrides(t *testing.T) {
	t.Parallel()

	parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{
		StableOverride:  &qovery.KarpenterStableNodePoolOverride{Limits: testApiLimits()},
		DefaultOverride: &qovery.KarpenterDefaultNodePoolOverride{Limits: testApiLimits()},
	})

	attrVals := karpenterFeatureAttrValue(parameters, testNoPlan(), clusterReadModeDataSource)
	require.NotNil(t, attrVals)

	assert.False(t, stateOverride(t, attrVals, "stable_override").Attributes()["limits"].IsNull())
	assert.False(t, stateOverride(t, attrVals, "default_override").Attributes()["limits"].IsNull())
}

// --- plan-time warning --------------------------------------------------------------------------

// testKarpenterRawFeatures builds a raw object holding a cluster features object whose karpenter
// block is the one given, together with a schema restricted to the resource's features attribute.
func testKarpenterRawFeatures(t *testing.T, karpenter types.Object) (tftypes.Value, schema.Schema) {
	t.Helper()

	var schemaResp resource.SchemaResponse
	clusterResource{}.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	features := testFeaturesObject(karpenter)

	raw, err := features.ToTerraformValue(context.Background())
	require.NoError(t, err)

	object := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"features": features.Type(context.Background()).TerraformType(context.Background())}},
		map[string]tftypes.Value{"features": raw},
	)

	return object, schema.Schema{Attributes: map[string]schema.Attribute{
		"features": schemaResp.Schema.Attributes["features"],
	}}
}

func testKarpenterState(t *testing.T, karpenter types.Object) tfsdk.State {
	t.Helper()

	raw, sch := testKarpenterRawFeatures(t, karpenter)
	return tfsdk.State{Raw: raw, Schema: sch}
}

func testKarpenterPlan(t *testing.T, karpenter types.Object) tfsdk.Plan {
	t.Helper()

	raw, sch := testKarpenterRawFeatures(t, karpenter)
	return tfsdk.Plan{Raw: raw, Schema: sch}
}

func TestWarnKarpenterSpotToOnDemand(t *testing.T) {
	t.Parallel()

	stableSpot := map[string]attr.Value{"stable_override": testStableSpotOnly(types.BoolValue(true))}
	defaultSpot := map[string]attr.Value{"default_override": testDefaultOverrideObject(types.BoolValue(true), testNullLimits())}
	defaultOnDemand := map[string]attr.Value{"default_override": testDefaultOverrideObject(types.BoolValue(false), testNullLimits())}
	cronjobSpot := map[string]attr.Value{"cronjob_override": testCronjobOverrideObject(types.BoolValue(true))}
	cronjobOnDemand := map[string]attr.Value{"cronjob_override": testCronjobOverrideObject(types.BoolValue(false))}
	gpuSpot := map[string]attr.Value{"gpu_override": testGpuOverrideObject(func(attrs map[string]attr.Value) {
		attrs["spot_enabled"] = types.BoolValue(true)
	})}
	gpuOnDemand := map[string]attr.Value{"gpu_override": testGpuOverrideObject(nil)}

	nodePoolsPath := path.Root("features").AtName("karpenter").AtName("qovery_node_pools")

	testCases := []struct {
		TestName       string
		StateOverrides map[string]attr.Value
		PlanOverrides  map[string]attr.Value
		// PlanKarpenter replaces the planned karpenter object built from PlanOverrides.
		PlanKarpenter  *types.Object
		ExpectWarnings []path.Path
	}{
		{
			// The upgrade case: the read path stored the undeclared spot pool, the configuration
			// does not declare it.
			TestName:       "removed_spot_block_warns",
			StateOverrides: stableSpot,
			ExpectWarnings: []path.Path{nodePoolsPath.AtName("stable_override")},
		},
		{
			TestName:       "spot_to_on_demand_warns",
			StateOverrides: defaultSpot,
			PlanOverrides:  defaultOnDemand,
			ExpectWarnings: []path.Path{nodePoolsPath.AtName("default_override")},
		},
		{
			TestName:       "spot_kept_is_quiet",
			StateOverrides: defaultSpot,
			PlanOverrides:  defaultSpot,
		},
		{
			TestName:       "on_demand_to_spot_is_quiet",
			StateOverrides: defaultOnDemand,
			PlanOverrides:  defaultSpot,
		},
		{
			TestName:       "cronjob_spot_to_on_demand_warns",
			StateOverrides: cronjobSpot,
			PlanOverrides:  cronjobOnDemand,
			ExpectWarnings: []path.Path{nodePoolsPath.AtName("cronjob_override")},
		},
		{
			// Removing cronjob_override removes the dedicated pool: nothing moves to on-demand.
			TestName:       "removed_cronjob_block_is_quiet",
			StateOverrides: cronjobSpot,
		},
		{
			TestName:       "gpu_spot_to_on_demand_warns",
			StateOverrides: gpuSpot,
			PlanOverrides:  gpuOnDemand,
			ExpectWarnings: []path.Path{nodePoolsPath.AtName("gpu_override")},
		},
		{
			// Removing gpu_override deletes the GPU pool, which warnKarpenterGpuNodePoolRemoval
			// reports: nothing moves to on-demand.
			TestName:       "removed_gpu_block_is_quiet",
			StateOverrides: gpuSpot,
		},
		{
			TestName:       "unknown_planned_value_is_quiet",
			StateOverrides: defaultSpot,
			PlanOverrides:  map[string]attr.Value{"default_override": testDefaultOverrideObject(types.BoolUnknown(), testNullLimits())},
		},
		{
			// An override block computed from a value unknown at plan time may resolve to spot.
			TestName:       "unknown_planned_override_is_quiet",
			StateOverrides: stableSpot,
			PlanOverrides:  map[string]attr.Value{"stable_override": types.ObjectUnknown(karpenterStableOverrideAttrTypes())},
		},
		{
			TestName:       "unknown_planned_node_pools_are_quiet",
			StateOverrides: stableSpot,
			PlanKarpenter: func() *types.Object {
				karpenter := types.ObjectValueMust(createKarpenterFeatureAttrTypes(), map[string]attr.Value{
					"disk_size_in_gib":             types.Int64Value(50),
					"default_service_architecture": types.StringValue("AMD64"),
					"qovery_node_pools":            types.ObjectUnknown(karpenterNodePoolsAttrTypes()),
				})
				return &karpenter
			}(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			planKarpenter := testKarpenterObject(tc.PlanOverrides)
			if tc.PlanKarpenter != nil {
				planKarpenter = *tc.PlanKarpenter
			}

			var diags diag.Diagnostics
			warnKarpenterSpotToOnDemand(context.Background(),
				testKarpenterState(t, testKarpenterObject(tc.StateOverrides)),
				testKarpenterPlan(t, planKarpenter),
				&diags)

			require.False(t, diags.HasError(), "%v", diags)
			var warnings []path.Path
			for _, d := range diags.Warnings() {
				withPath, ok := d.(diag.DiagnosticWithPath)
				require.True(t, ok, "the warning must point at the node pool override")
				warnings = append(warnings, withPath.Path())
			}
			assert.ElementsMatch(t, tc.ExpectWarnings, warnings)
		})
	}

	t.Run("create_is_quiet", func(t *testing.T) {
		t.Parallel()

		_, sch := testKarpenterRawFeatures(t, testKarpenterObject(nil))
		var diags diag.Diagnostics
		warnKarpenterSpotToOnDemand(context.Background(),
			tfsdk.State{Raw: tftypes.NewValue(sch.Type().TerraformType(context.Background()), nil), Schema: sch},
			testKarpenterPlan(t, testKarpenterObject(stableSpot)),
			&diags)
		assert.Empty(t, diags)
	})
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

// TestClusterSchema_KarpenterSpotEnabled pins the 1.0 contract: no global spot flag on the
// resource or the data source, and every per node pool spot_enabled on the resource defaults to
// false, i.e. on-demand instances.
func TestClusterSchema_KarpenterSpotEnabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var resourceSchema resource.SchemaResponse
	clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &resourceSchema)
	require.False(t, resourceSchema.Diagnostics.HasError())

	karpenter := resourceSchema.Schema.Attributes["features"].(schema.SingleNestedAttribute).Attributes["karpenter"].(schema.SingleNestedAttribute)
	assert.NotContains(t, karpenter.Attributes, "spot_enabled", "the global spot flag is removed in 1.0")

	nodePools := karpenter.Attributes["qovery_node_pools"].(schema.SingleNestedAttribute)
	for _, name := range karpenterNodePoolOverrideNames {
		spotEnabled := nodePools.Attributes[name].(schema.SingleNestedAttribute).Attributes["spot_enabled"].(schema.BoolAttribute)
		require.NotNil(t, spotEnabled.Default, "%s.spot_enabled must have a default", name)

		var resp defaults.BoolResponse
		spotEnabled.Default.DefaultBool(ctx, defaults.BoolRequest{}, &resp)
		assert.Equal(t, types.BoolValue(false), resp.PlanValue, "%s.spot_enabled must default to on-demand", name)
	}

	var dataSourceSchema datasource.SchemaResponse
	clusterDataSource{}.Schema(ctx, datasource.SchemaRequest{}, &dataSourceSchema)
	require.False(t, dataSourceSchema.Diagnostics.HasError())

	dataSourceKarpenter := dataSourceSchema.Schema.Attributes["features"].(datasourceschema.SingleNestedAttribute).Attributes["karpenter"].(datasourceschema.SingleNestedAttribute)
	assert.NotContains(t, dataSourceKarpenter.Attributes, "spot_enabled", "the global spot flag is removed from the data source in 1.0")
}
