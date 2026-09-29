//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests below cover the Karpenter settings the Console sets and q-core rebuilds from each
// request: consolidate_after on every node pool, consolidation and limits on the cronjob node
// pool, and disk_iops and disk_throughput on the karpenter feature. Before 1.0 the provider left
// them out of every request, so any apply wiped a value set from the Console.

// --- helpers -----------------------------------------------------------------------------

// testOverrideWith returns an override object with the given attributes replaced.
func testOverrideWith(override types.Object, replaced map[string]attr.Value) types.Object {
	attrs := override.Attributes()
	for name, value := range replaced {
		attrs[name] = value
	}
	return types.ObjectValueMust(override.AttributeTypes(context.Background()), attrs)
}

// testKarpenterObjectWithDisk returns a karpenter object with disk_iops and disk_throughput set.
func testKarpenterObjectWithDisk(karpenter types.Object, diskIops, diskThroughput types.Int64) types.Object {
	attrs := karpenter.Attributes()
	attrs["disk_iops"] = diskIops
	attrs["disk_throughput"] = diskThroughput
	return types.ObjectValueMust(createKarpenterFeatureAttrTypes(), attrs)
}

// testEveryNodePoolSetting declares every node pool with every setting these tests cover.
func testEveryNodePoolSetting() types.Object {
	karpenter := testKarpenterObject(map[string]attr.Value{
		"stable_override": testOverrideWith(testStableOverrideObject(types.BoolValue(false), testKarpenterConsolidationObject(), testKarpenterLimitsObject()), map[string]attr.Value{
			"consolidate_after": types.StringValue("1h"),
		}),
		"default_override": testOverrideWith(testDefaultOverrideObject(types.BoolValue(false), testKarpenterLimitsObject()), map[string]attr.Value{
			"consolidate_after": types.StringValue("2m"),
		}),
		"cronjob_override": testOverrideWith(testCronjobOverrideObject(types.BoolValue(true)), map[string]attr.Value{
			"consolidation":     testKarpenterConsolidationObject(),
			"limits":            testKarpenterLimitsObject(),
			"consolidate_after": types.StringValue("30s"),
		}),
		"gpu_override": testGpuOverrideObject(testGpuOverrideEveryField),
	})
	return testKarpenterObjectWithDisk(karpenter, types.Int64Value(4000), types.Int64Value(250))
}

func testApiConsolidation() *qovery.KarpenterNodePoolConsolidation {
	return qovery.NewKarpenterNodePoolConsolidation(true, []qovery.WeekdayEnum{qovery.WEEKDAYENUM_MONDAY}, "PT02:00", "PT04H00M")
}

// --- TF -> API ---------------------------------------------------------------------------

func TestToQoveryClusterFeatures_KarpenterNodePoolSettings(t *testing.T) {
	t.Parallel()

	request := karpenterRequest(t, testEveryNodePoolSetting())

	assert.Equal(t, new(int32(4000)), request.DiskIops)
	assert.Equal(t, new(int32(250)), request.DiskThroughput)

	nodePools := request.QoveryNodePools
	require.NotNil(t, nodePools.StableOverride)
	assert.Equal(t, new("1h"), nodePools.StableOverride.ConsolidateAfter)
	require.NotNil(t, nodePools.DefaultOverride)
	assert.Equal(t, new("2m"), nodePools.DefaultOverride.ConsolidateAfter)
	require.NotNil(t, nodePools.GpuOverride)
	assert.Equal(t, new("10m"), nodePools.GpuOverride.ConsolidateAfter)

	require.NotNil(t, nodePools.CronjobOverride)
	assert.Equal(t, new("30s"), nodePools.CronjobOverride.ConsolidateAfter)
	assert.Equal(t, testApiConsolidation(), nodePools.CronjobOverride.Consolidation)
	assert.Equal(t, testApiLimits(), nodePools.CronjobOverride.Limits)
	assert.Equal(t, new(true), GetCronjobNodePoolSpotEnabled(nodePools.CronjobOverride))
}

// TestToQoveryClusterFeatures_KarpenterNodePoolSettingsUnset covers the removal of every setting:
// q-core rebuilds the Karpenter parameters from each request, so leaving a value out of the
// request is what clears it. The request depends on the configuration alone, so a value the
// state still holds is not sent either.
func TestToQoveryClusterFeatures_KarpenterNodePoolSettingsUnset(t *testing.T) {
	t.Parallel()

	karpenter := testKarpenterObject(map[string]attr.Value{
		"stable_override":  testStableSpotOnly(types.BoolValue(false)),
		"default_override": testDefaultOverrideObject(types.BoolValue(false), testNullLimits()),
		"cronjob_override": testCronjobOverrideObject(types.BoolValue(false)),
		"gpu_override":     testGpuOverrideObject(nil),
	})
	request := karpenterRequest(t, karpenter)

	assert.Nil(t, request.DiskIops)
	assert.Nil(t, request.DiskThroughput)

	nodePools := request.QoveryNodePools
	assert.Nil(t, nodePools.StableOverride.ConsolidateAfter)
	assert.Nil(t, nodePools.DefaultOverride.ConsolidateAfter)
	assert.Nil(t, nodePools.GpuOverride.ConsolidateAfter)
	require.NotNil(t, nodePools.CronjobOverride)
	assert.Nil(t, nodePools.CronjobOverride.ConsolidateAfter)
	assert.Nil(t, nodePools.CronjobOverride.Consolidation)
	assert.Nil(t, nodePools.CronjobOverride.Limits)
}

func TestCluster_hasFeaturesDiff_KarpenterNodePoolSettings(t *testing.T) {
	t.Parallel()

	state := testEveryNodePoolSetting()
	stableOverride := state.Attributes()["qovery_node_pools"].(types.Object).Attributes()["stable_override"].(types.Object)

	testCases := []struct {
		TestName   string
		Plan       types.Object
		ExpectDiff bool
	}{
		{TestName: "unchanged", Plan: state},
		{
			TestName:   "disk_iops_changed",
			Plan:       testKarpenterObjectWithDisk(state, types.Int64Value(5000), types.Int64Value(250)),
			ExpectDiff: true,
		},
		{
			TestName:   "disk_throughput_removed",
			Plan:       testKarpenterObjectWithDisk(state, types.Int64Value(4000), types.Int64Null()),
			ExpectDiff: true,
		},
		{
			TestName: "consolidate_after_changed",
			Plan: func() types.Object {
				attrs := state.Attributes()
				nodePools := attrs["qovery_node_pools"].(types.Object).Attributes()
				nodePools["stable_override"] = testOverrideWith(stableOverride, map[string]attr.Value{"consolidate_after": types.StringValue("2h")})
				attrs["qovery_node_pools"] = types.ObjectValueMust(karpenterNodePoolsAttrTypes(), nodePools)
				return types.ObjectValueMust(createKarpenterFeatureAttrTypes(), attrs)
			}(),
			ExpectDiff: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			plan := Cluster{KubernetesMode: types.StringValue("MANAGED"), CloudProvider: types.StringValue("AWS"), Features: testFeaturesObject(tc.Plan)}
			prior := Cluster{KubernetesMode: types.StringValue("MANAGED"), CloudProvider: types.StringValue("AWS"), Features: testFeaturesObject(state)}
			assert.Equal(t, tc.ExpectDiff, plan.hasFeaturesDiff(&prior))
		})
	}
}

// --- API -> TF ---------------------------------------------------------------------------

// TestKarpenterFeatureAttrValue_NodePoolSettingsFollowTheAPI covers the refresh: every setting is
// read from the API, so a value set from the Console shows in the plan instead of being wiped by
// the next apply without notice.
func TestKarpenterFeatureAttrValue_NodePoolSettingsFollowTheAPI(t *testing.T) {
	t.Parallel()

	withConsolidateAfter := func(value string) func(*qovery.KarpenterNodePool) {
		return func(nodePools *qovery.KarpenterNodePool) {
			nodePools.StableOverride = testApiStable(nil)
			nodePools.StableOverride.ConsolidateAfter = new(value)
		}
	}

	testCases := []struct {
		TestName       string
		ApiNodePools   func(*qovery.KarpenterNodePool)
		Plan           types.Object
		Mode           clusterReadMode
		Override       string
		ExpectOverride types.Object
	}{
		{
			// A stable node pool whose only setting is consolidate_after has real content: it is
			// stored even though the configuration does not declare the block, so that the plan
			// shows its removal.
			TestName:     "undeclared_stable_pool_with_consolidate_after_is_stored",
			ApiNodePools: withConsolidateAfter("5m"),
			Plan:         testKarpenterObject(nil),
			Override:     "stable_override",
			ExpectOverride: testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{
				"consolidate_after": types.StringValue("5m"),
			}),
		},
		{
			TestName: "undeclared_default_pool_with_consolidate_after_is_stored",
			ApiNodePools: func(nodePools *qovery.KarpenterNodePool) {
				nodePools.DefaultOverride = testApiDefault(nil)
				nodePools.DefaultOverride.ConsolidateAfter = new("1h")
			},
			Plan:     testKarpenterObject(nil),
			Override: "default_override",
			ExpectOverride: testOverrideWith(testDefaultOverrideObject(types.BoolValue(false), testNullLimits()), map[string]attr.Value{
				"consolidate_after": types.StringValue("1h"),
			}),
		},
		{
			TestName:     "console_change_on_a_declared_pool_is_stored",
			ApiNodePools: withConsolidateAfter("2h"),
			Plan: testKarpenterObject(map[string]attr.Value{
				"stable_override": testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{
					"consolidate_after": types.StringValue("1h"),
				}),
			}),
			Override: "stable_override",
			ExpectOverride: testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{
				"consolidate_after": types.StringValue("2h"),
			}),
		},
		{
			TestName:     "console_removal_on_a_declared_pool_is_stored",
			ApiNodePools: func(nodePools *qovery.KarpenterNodePool) { nodePools.StableOverride = testApiStable(nil) },
			Plan: testKarpenterObject(map[string]attr.Value{
				"stable_override": testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{
					"consolidate_after": types.StringValue("1h"),
				}),
			}),
			Override:       "stable_override",
			ExpectOverride: testStableSpotOnly(types.BoolValue(false)),
		},
		{
			TestName:       "import_stores_the_pool",
			ApiNodePools:   withConsolidateAfter("5m"),
			Plan:           testNoPlan(),
			Override:       "stable_override",
			ExpectOverride: testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{"consolidate_after": types.StringValue("5m")}),
		},
		{
			TestName: "cronjob_pool_settings_are_stored",
			ApiNodePools: func(nodePools *qovery.KarpenterNodePool) {
				nodePools.CronjobOverride = testApiCronjob(nil)
				nodePools.CronjobOverride.Consolidation = testApiConsolidation()
				nodePools.CronjobOverride.Limits = testApiLimits()
				nodePools.CronjobOverride.ConsolidateAfter = new("30s")
			},
			Plan:     testKarpenterObject(map[string]attr.Value{"cronjob_override": testCronjobOverrideObject(types.BoolValue(false))}),
			Override: "cronjob_override",
			ExpectOverride: testOverrideWith(testCronjobOverrideObject(types.BoolValue(false)), map[string]attr.Value{
				"consolidation":     testKarpenterConsolidationObject(),
				"limits":            testKarpenterLimitsObject(),
				"consolidate_after": types.StringValue("30s"),
			}),
		},
		{
			TestName: "gpu_pool_consolidate_after_is_stored",
			ApiNodePools: func(nodePools *qovery.KarpenterNodePool) {
				nodePools.GpuOverride = testApiGpu()
				nodePools.GpuOverride.ConsolidateAfter = new("10m")
			},
			Plan:     testKarpenterObject(nil),
			Override: "gpu_override",
			ExpectOverride: testGpuOverrideObject(func(attrs map[string]attr.Value) {
				attrs["consolidate_after"] = types.StringValue("10m")
			}),
		},
		{
			TestName:     "data_source_reports_the_pool",
			ApiNodePools: withConsolidateAfter("5m"),
			Plan:         testNoPlan(),
			Mode:         clusterReadModeDataSource,
			Override:     "stable_override",
			ExpectOverride: testOverrideWith(testStableSpotOnly(types.BoolValue(false)), map[string]attr.Value{
				"consolidate_after": types.StringValue("5m"),
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			nodePools := qovery.KarpenterNodePool{}
			tc.ApiNodePools(&nodePools)
			parameters := testApiKarpenterParameters(false, nodePools)

			attrVals := karpenterFeatureAttrValue(parameters, tc.Plan, tc.Mode)
			require.NotNil(t, attrVals)

			got := stateOverride(t, attrVals, tc.Override)
			assert.True(t, tc.ExpectOverride.Equal(got), "want %s, got %s", tc.ExpectOverride, got)
		})
	}
}

func TestKarpenterFeatureAttrValue_Disk(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName             string
		ApiDiskIops          *int32
		ApiDiskThroughput    *int32
		ExpectDiskIops       types.Int64
		ExpectDiskThroughput types.Int64
	}{
		{
			TestName:             "set",
			ApiDiskIops:          new(int32(4000)),
			ApiDiskThroughput:    new(int32(250)),
			ExpectDiskIops:       types.Int64Value(4000),
			ExpectDiskThroughput: types.Int64Value(250),
		},
		{
			TestName:             "unset",
			ExpectDiskIops:       types.Int64Null(),
			ExpectDiskThroughput: types.Int64Null(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{})
			parameters.DiskIops = tc.ApiDiskIops
			parameters.DiskThroughput = tc.ApiDiskThroughput

			attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
			require.NotNil(t, attrVals)
			assert.Equal(t, tc.ExpectDiskIops, attrVals["disk_iops"])
			assert.Equal(t, tc.ExpectDiskThroughput, attrVals["disk_throughput"])
		})
	}
}

// TestKarpenterNodePoolSettings_RoundTrip sends every setting, feeds the request back as the API
// response and checks the refresh stores the karpenter object unchanged. A difference would fail
// the apply with "Provider produced inconsistent result after apply", or leave a permanent diff.
func TestKarpenterNodePoolSettings_RoundTrip(t *testing.T) {
	t.Parallel()

	karpenter := testEveryNodePoolSetting()
	request := karpenterRequest(t, karpenter)

	attrVals := karpenterFeatureAttrValue(request, karpenter, clusterReadModeResource)
	require.NotNil(t, attrVals)

	got := types.ObjectValueMust(createKarpenterFeatureAttrTypes(), attrVals)
	assert.True(t, karpenter.Equal(got), "want %s, got %s", karpenter, got)
}

// --- schema --------------------------------------------------------------------------------

// TestClusterSchema_KarpenterNodePoolSettings pins the schema of the settings: every node pool
// override has an Optional consolidate_after that only accepts the form q-core returns, the
// cronjob node pool has consolidation and limits, and the data source reports them all.
func TestClusterSchema_KarpenterNodePoolSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var resourceSchema resource.SchemaResponse
	clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &resourceSchema)
	require.False(t, resourceSchema.Diagnostics.HasError())

	karpenter := resourceSchema.Schema.Attributes["features"].(schema.SingleNestedAttribute).Attributes["karpenter"].(schema.SingleNestedAttribute)
	for _, name := range []string{"disk_iops", "disk_throughput"} {
		attribute := karpenter.Attributes[name].(schema.Int64Attribute)
		assert.True(t, attribute.Optional, "karpenter.%s must be Optional", name)
		assert.False(t, attribute.Computed, "karpenter.%s must not be Computed: removing it clears it", name)
	}

	nodePools := karpenter.Attributes["qovery_node_pools"].(schema.SingleNestedAttribute)
	for _, name := range karpenterNodePoolOverrideNames {
		consolidateAfter := nodePools.Attributes[name].(schema.SingleNestedAttribute).Attributes["consolidate_after"].(schema.StringAttribute)
		assert.True(t, consolidateAfter.Optional, "%s.consolidate_after must be Optional", name)
		assert.False(t, consolidateAfter.Computed, "%s.consolidate_after must not be Computed: removing it clears it", name)

		// "60m" comes back from q-core as "1h": the validator must reject it at plan time.
		require.Len(t, consolidateAfter.Validators, 1, "%s.consolidate_after must be validated", name)
		resp := &validator.StringResponse{}
		consolidateAfter.Validators[0].ValidateString(ctx, validator.StringRequest{Path: path.Root(name), ConfigValue: types.StringValue("60m")}, resp)
		assert.True(t, resp.Diagnostics.HasError(), "%s.consolidate_after must reject a value q-core would return differently", name)
	}

	cronjob := nodePools.Attributes["cronjob_override"].(schema.SingleNestedAttribute)
	assert.Contains(t, cronjob.Attributes, "consolidation")
	assert.Contains(t, cronjob.Attributes, "limits")

	var dataSourceSchema datasource.SchemaResponse
	clusterDataSource{}.Schema(ctx, datasource.SchemaRequest{}, &dataSourceSchema)
	require.False(t, dataSourceSchema.Diagnostics.HasError())

	dataSourceKarpenter := dataSourceSchema.Schema.Attributes["features"].(datasourceschema.SingleNestedAttribute).Attributes["karpenter"].(datasourceschema.SingleNestedAttribute)
	for _, name := range []string{"disk_iops", "disk_throughput"} {
		assert.True(t, dataSourceKarpenter.Attributes[name].(datasourceschema.Int64Attribute).Computed, "data source karpenter.%s must be Computed", name)
	}
	dataSourceNodePools := dataSourceKarpenter.Attributes["qovery_node_pools"].(datasourceschema.SingleNestedAttribute)
	for _, name := range karpenterNodePoolOverrideNames {
		consolidateAfter := dataSourceNodePools.Attributes[name].(datasourceschema.SingleNestedAttribute).Attributes["consolidate_after"].(datasourceschema.StringAttribute)
		assert.True(t, consolidateAfter.Computed, "data source %s.consolidate_after must be Computed", name)
	}
}
