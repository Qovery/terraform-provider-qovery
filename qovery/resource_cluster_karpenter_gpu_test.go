//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func testKarpenterRequirement(key string, values ...string) attr.Value {
	valueList := make([]attr.Value, len(values))
	for i, value := range values {
		valueList[i] = types.StringValue(value)
	}
	return types.ObjectValueMust(karpenterRequirementAttrTypes(), map[string]attr.Value{
		"key":      types.StringValue(key),
		"operator": types.StringValue("In"),
		"values":   types.ListValueMust(types.StringType, valueList),
	})
}

func testGpuRequirements(requirements ...attr.Value) types.List {
	return types.ListValueMust(types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()}, requirements)
}

func testGpuLimitsObject(maxGpu int64) types.Object {
	return types.ObjectValueMust(karpenterGpuLimitsAttrTypes(), map[string]attr.Value{
		"enabled":                 types.BoolValue(true),
		"max_cpu_in_vcpu":         types.Int64Value(10),
		"max_memory_in_gibibytes": types.Int64Value(20),
		"max_gpu":                 types.Int64Value(maxGpu),
	})
}

// testGpuOverrideObject builds the smallest gpu_override block q-core accepts: g4dn, xlarge and
// AMD64 requirements with a 50 GiB disk. mutate changes attributes before the block is built.
func testGpuOverrideObject(mutate func(attrs map[string]attr.Value)) types.Object {
	attrs := map[string]attr.Value{
		"requirements": testGpuRequirements(
			testKarpenterRequirement("InstanceFamily", "g4dn"),
			testKarpenterRequirement("InstanceSize", "xlarge"),
			testKarpenterRequirement("Arch", "AMD64"),
		),
		"disk_size_in_gib": types.Int64Value(50),
		"disk_iops":        types.Int64Null(),
		"disk_throughput":  types.Int64Null(),
		"spot_enabled":     types.BoolValue(false),
		"consolidation":    testNullConsolidation(),
		"limits":           types.ObjectNull(karpenterGpuLimitsAttrTypes()),
	}
	if mutate != nil {
		mutate(attrs)
	}
	return types.ObjectValueMust(karpenterGpuOverrideAttrTypes(), attrs)
}

// testGpuOverrideEveryField sets every optional attribute of the block.
func testGpuOverrideEveryField(attrs map[string]attr.Value) {
	attrs["disk_size_in_gib"] = types.Int64Value(100)
	attrs["disk_iops"] = types.Int64Value(3000)
	attrs["disk_throughput"] = types.Int64Value(125)
	attrs["spot_enabled"] = types.BoolValue(true)
	attrs["consolidation"] = testKarpenterConsolidationObject()
	attrs["limits"] = testGpuLimitsObject(4)
}

func testGpuKarpenterObject(gpuOverride attr.Value) types.Object {
	return testKarpenterObject(map[string]attr.Value{"gpu_override": gpuOverride})
}

// testApiGpu is a GPU node pool the way the API returns it: every field explicit.
func testApiGpu() *qovery.KarpenterGpuNodePoolOverride {
	return &qovery.KarpenterGpuNodePoolOverride{
		Requirements: []qovery.KarpenterNodePoolRequirement{
			{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"g4dn"}},
			{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"xlarge"}},
			{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"AMD64"}},
		},
		DiskSizeInGib: new(int32(50)),
		SpotEnabled:   boolPtr(false),
	}
}

// --- TF -> API ---------------------------------------------------------------------------

func TestExtractGpuNodePoolOverrideFromTypesObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName    string
		GpuOverride attr.Value
		Expect      *qovery.KarpenterGpuNodePoolOverride
	}{
		{
			// q-core rebuilds the node pools from each request, so a synthesized block would create
			// a GPU node pool the configuration never asked for.
			TestName:    "absent_block_is_never_synthesized",
			GpuOverride: types.ObjectNull(karpenterGpuOverrideAttrTypes()),
		},
		{
			TestName:    "unknown_block_is_not_sent",
			GpuOverride: types.ObjectUnknown(karpenterGpuOverrideAttrTypes()),
		},
		{
			TestName:    "minimal_block",
			GpuOverride: testGpuOverrideObject(nil),
			Expect:      testApiGpu(),
		},
		{
			// The schema defaults spot_enabled to false, so null or unknown only come from a plan
			// that is not final yet; the request still carries an explicit value.
			TestName: "unset_spot_enabled_is_sent_as_on_demand",
			GpuOverride: testGpuOverrideObject(func(attrs map[string]attr.Value) {
				attrs["spot_enabled"] = types.BoolUnknown()
			}),
			Expect: testApiGpu(),
		},
		{
			TestName:    "every_field",
			GpuOverride: testGpuOverrideObject(testGpuOverrideEveryField),
			Expect: func() *qovery.KarpenterGpuNodePoolOverride {
				o := testApiGpu()
				o.DiskSizeInGib = new(int32(100))
				o.DiskIops = new(int32(3000))
				o.DiskThroughput = new(int32(125))
				o.SpotEnabled = boolPtr(true)
				o.Consolidation = qovery.NewKarpenterNodePoolConsolidation(true, []qovery.WeekdayEnum{qovery.WEEKDAYENUM_MONDAY}, "PT02:00", "PT04H00M")
				o.Limits = qovery.NewKarpenterNodePoolLimits(true, 10, 20, 4)
				return o
			}(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			override, err := extractGpuNodePoolOverrideFromTypesObject(testGpuKarpenterObject(tc.GpuOverride))
			require.NoError(t, err)
			assert.Equal(t, tc.Expect, override)
		})
	}
}

func TestExtractGpuNodePoolOverrideFromTypesObject_RequirementsAreValidated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Requirements  types.List
		ErrorContains string
	}{
		{
			TestName:      "no_requirement",
			Requirements:  testGpuRequirements(),
			ErrorContains: "requirements are mandatory",
		},
		{
			TestName: "missing_arch",
			Requirements: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "g4dn"),
				testKarpenterRequirement("InstanceSize", "xlarge"),
			),
			ErrorContains: "missing some karpenter nodepool requirement",
		},
		{
			TestName: "empty_values",
			Requirements: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily"),
				testKarpenterRequirement("InstanceSize", "xlarge"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
			ErrorContains: "values must not be empty",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			_, err := extractGpuNodePoolOverrideFromTypesObject(testGpuKarpenterObject(testGpuOverrideObject(func(attrs map[string]attr.Value) {
				attrs["requirements"] = tc.Requirements
			})))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "gpu_override")
			assert.Contains(t, err.Error(), tc.ErrorContains)
		})
	}
}

func TestToQoveryNodePools_GpuNodePool(t *testing.T) {
	t.Parallel()

	nodePools, err := toQoveryNodePools(testKarpenterObject(map[string]attr.Value{
		"stable_override": testStableOverrideObject(types.BoolValue(false), testNullConsolidation(), testKarpenterLimitsObject()),
		"gpu_override":    testGpuOverrideObject(testGpuOverrideEveryField),
	}))
	require.NoError(t, err)

	require.NotNil(t, nodePools.GpuOverride)
	assert.Equal(t, int32(4), nodePools.GpuOverride.Limits.MaxGpu)
	assert.Equal(t, int32(0), nodePools.StableOverride.Limits.MaxGpu, "only the GPU node pool limits carry max_gpu")
}

// TestCluster_hasFeaturesDiff_GpuNodePool pins that every GPU node pool change forces a redeploy,
// which is what applies node pool changes to the cluster.
func TestCluster_hasFeaturesDiff_GpuNodePool(t *testing.T) {
	t.Parallel()

	newCluster := func(gpuOverride attr.Value) Cluster {
		return Cluster{
			CloudProvider:  types.StringValue("AWS"),
			KubernetesMode: types.StringValue("MANAGED"),
			Features:       testFeaturesObject(testGpuKarpenterObject(gpuOverride)),
		}
	}
	noGpu := types.ObjectNull(karpenterGpuOverrideAttrTypes())
	gpu := testGpuOverrideObject(nil)
	biggerDisk := testGpuOverrideObject(func(attrs map[string]attr.Value) {
		attrs["disk_size_in_gib"] = types.Int64Value(100)
	})

	gpuState := newCluster(gpu)
	noGpuState := newCluster(noGpu)

	assert.False(t, newCluster(gpu).hasFeaturesDiff(&gpuState), "an unchanged GPU node pool must not force a redeploy")
	assert.True(t, newCluster(gpu).hasFeaturesDiff(&noGpuState), "adding the GPU node pool must force a redeploy")
	assert.True(t, newCluster(biggerDisk).hasFeaturesDiff(&gpuState), "changing the GPU node pool must force a redeploy")
	assert.True(t, newCluster(noGpu).hasFeaturesDiff(&gpuState), "removing the GPU node pool must force a redeploy")
}

// TestCluster_toUpsertClusterRequest_StateGpuPoolTheProviderWouldNotSend is the regression test for a
// GPU node pool created through the API with requirements the provider refuses to send: q-core
// accepts any GPU requirements, the provider asks for all three keys. The refresh stores the pool
// as the API returns it, and an update, including the one that removes the pool, must still go
// through: the prior state must not have to pass the write validation for Karpenter to count as
// installed.
func TestCluster_toUpsertClusterRequest_StateGpuPoolTheProviderWouldNotSend(t *testing.T) {
	t.Parallel()

	newCluster := func(gpuOverride attr.Value) Cluster {
		return Cluster{
			Id:              types.StringValue("cluster-123"),
			OrganizationId:  types.StringValue("org-123"),
			CredentialsId:   types.StringValue("cred-123"),
			Name:            types.StringValue("test-cluster"),
			CloudProvider:   types.StringValue("AWS"),
			Region:          types.StringValue("us-east-1"),
			KubernetesMode:  types.StringValue("MANAGED"),
			State:           types.StringValue("DEPLOYED"),
			InstanceType:    types.StringUnknown(),
			MinRunningNodes: types.Int64Unknown(),
			MaxRunningNodes: types.Int64Unknown(),
			DiskSize:        types.Int64Unknown(),
			Features:        testFeaturesObject(testGpuKarpenterObject(gpuOverride)),
		}
	}
	state := newCluster(testGpuOverrideObject(func(attrs map[string]attr.Value) {
		attrs["requirements"] = testGpuRequirements()
	}))

	assert.True(t, IsKarpenterAlreadyInstalled(&state))

	for name, gpuOverride := range map[string]attr.Value{
		"pool_removed":  types.ObjectNull(karpenterGpuOverrideAttrTypes()),
		"pool_declared": testGpuOverrideObject(nil),
	} {
		_, err := newCluster(gpuOverride).toUpsertClusterRequest(&state)
		assert.NoError(t, err, "%s: the update must not be refused because of the prior state", name)
	}
}

// --- API -> TF ---------------------------------------------------------------------------

// TestKarpenterFeatureAttrValue_GpuPoolFollowsTheAPI pins the gpu_override read rule: the block is
// stored exactly when the API returns it, whatever the plan says. q-core deletes the GPU node pool
// on any request without the block, so a pool created outside Terraform must reach the state and
// show in the plan instead of being deleted by the next apply.
func TestKarpenterFeatureAttrValue_GpuPoolFollowsTheAPI(t *testing.T) {
	t.Parallel()

	declaredGpu := testGpuKarpenterObject(testGpuOverrideObject(nil))
	apiBiggerDisk := testApiGpu()
	apiBiggerDisk.DiskSizeInGib = new(int32(200))

	testCases := []struct {
		TestName    string
		ApiGpu      *qovery.KarpenterGpuNodePoolOverride
		Plan        types.Object
		Mode        clusterReadMode
		ExpectBlock types.Object
	}{
		{
			// Created from the Console while the configuration does not declare it.
			TestName:    "undeclared_pool_on_the_api_is_stored",
			ApiGpu:      testApiGpu(),
			Plan:        testKarpenterObject(nil),
			ExpectBlock: testGpuOverrideObject(nil),
		},
		{
			// Deleted from the Console while the configuration declares it.
			TestName:    "declared_pool_deleted_on_the_api_is_dropped",
			ApiGpu:      nil,
			Plan:        declaredGpu,
			ExpectBlock: types.ObjectNull(karpenterGpuOverrideAttrTypes()),
		},
		{
			TestName: "out_of_band_change_on_a_declared_pool_is_stored",
			ApiGpu:   apiBiggerDisk,
			Plan:     declaredGpu,
			ExpectBlock: testGpuOverrideObject(func(attrs map[string]attr.Value) {
				attrs["disk_size_in_gib"] = types.Int64Value(200)
			}),
		},
		{
			TestName:    "import_stores_the_pool",
			ApiGpu:      testApiGpu(),
			Plan:        testNoPlan(),
			ExpectBlock: testGpuOverrideObject(nil),
		},
		{
			TestName:    "data_source_reports_the_pool",
			ApiGpu:      testApiGpu(),
			Plan:        testNoPlan(),
			Mode:        clusterReadModeDataSource,
			ExpectBlock: testGpuOverrideObject(nil),
		},
		{
			TestName:    "no_pool_anywhere",
			ApiGpu:      nil,
			Plan:        testKarpenterObject(nil),
			ExpectBlock: types.ObjectNull(karpenterGpuOverrideAttrTypes()),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			// The global flag is true to show it never leaks into the GPU node pool.
			parameters := testApiKarpenterParameters(true, qovery.KarpenterNodePool{GpuOverride: tc.ApiGpu})

			attrVals := karpenterFeatureAttrValue(parameters, tc.Plan, tc.Mode)
			require.NotNil(t, attrVals)

			got := stateOverride(t, attrVals, "gpu_override")
			assert.True(t, tc.ExpectBlock.Equal(got), "want %s, got %s", tc.ExpectBlock, got)
		})
	}
}

// TestKarpenterGpuOverride_RoundTrip sends a gpu_override block, feeds the request back as the API
// response and checks the refresh stores the block unchanged. A difference would fail the apply
// with "Provider produced inconsistent result after apply", or leave a permanent diff.
func TestKarpenterGpuOverride_RoundTrip(t *testing.T) {
	t.Parallel()

	for name, gpuOverride := range map[string]types.Object{
		"minimal_block": testGpuOverrideObject(nil),
		"every_field":   testGpuOverrideObject(testGpuOverrideEveryField),
		"limits_disabled": testGpuOverrideObject(func(attrs map[string]attr.Value) {
			attrs["limits"] = types.ObjectValueMust(karpenterGpuLimitsAttrTypes(), map[string]attr.Value{
				"enabled":                 types.BoolValue(false),
				"max_cpu_in_vcpu":         types.Int64Value(10),
				"max_memory_in_gibibytes": types.Int64Value(20),
				"max_gpu":                 types.Int64Value(0),
			})
		}),
	} {
		name, gpuOverride := name, gpuOverride
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karpenter := testGpuKarpenterObject(gpuOverride)
			request := karpenterRequest(t, karpenter)
			require.NotNil(t, request.QoveryNodePools.GpuOverride)

			attrVals := karpenterFeatureAttrValue(request, karpenter, clusterReadModeResource)
			require.NotNil(t, attrVals)

			got := stateOverride(t, attrVals, "gpu_override")
			assert.True(t, gpuOverride.Equal(got), "want %s, got %s", gpuOverride, got)
		})
	}
}

// --- plan-time warnings ---------------------------------------------------------------------------

func TestWarnKarpenterGpuNodePoolRemoval(t *testing.T) {
	t.Parallel()

	gpu := map[string]attr.Value{"gpu_override": testGpuOverrideObject(nil)}
	gpuBiggerDisk := map[string]attr.Value{"gpu_override": testGpuOverrideObject(func(attrs map[string]attr.Value) {
		attrs["disk_size_in_gib"] = types.Int64Value(100)
	})}
	gpuPath := path.Root("features").AtName("karpenter").AtName("qovery_node_pools").AtName("gpu_override")

	testCases := []struct {
		TestName       string
		StateOverrides map[string]attr.Value
		PlanOverrides  map[string]attr.Value
		// PlanKarpenter replaces the planned karpenter object built from PlanOverrides.
		PlanKarpenter *types.Object
		ExpectWarning bool
	}{
		{
			// A pool created from the Console, which the refresh stored: the configuration does not
			// declare it.
			TestName:       "removed_block_warns",
			StateOverrides: gpu,
			ExpectWarning:  true,
		},
		{
			TestName:       "kept_block_is_quiet",
			StateOverrides: gpu,
			PlanOverrides:  gpu,
		},
		{
			TestName:       "changed_block_is_quiet",
			StateOverrides: gpu,
			PlanOverrides:  gpuBiggerDisk,
		},
		{
			TestName:      "added_block_is_quiet",
			PlanOverrides: gpu,
		},
		{
			// An override block computed from a value unknown at plan time may resolve to a block.
			TestName:       "unknown_planned_override_is_quiet",
			StateOverrides: gpu,
			PlanOverrides:  map[string]attr.Value{"gpu_override": types.ObjectUnknown(karpenterGpuOverrideAttrTypes())},
		},
		{
			TestName:       "unknown_planned_node_pools_are_quiet",
			StateOverrides: gpu,
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
			warnKarpenterGpuNodePoolRemoval(context.Background(),
				testKarpenterState(t, testKarpenterObject(tc.StateOverrides)),
				testKarpenterPlan(t, planKarpenter),
				&diags)

			require.False(t, diags.HasError(), "%v", diags)
			if !tc.ExpectWarning {
				assert.Empty(t, diags)
				return
			}
			require.Len(t, diags.Warnings(), 1)
			withPath, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
			require.True(t, ok, "the warning must point at gpu_override")
			assert.Equal(t, gpuPath, withPath.Path())
		})
	}

	t.Run("create_is_quiet", func(t *testing.T) {
		t.Parallel()

		_, sch := testKarpenterRawFeatures(t, testKarpenterObject(nil))
		var diags diag.Diagnostics
		warnKarpenterGpuNodePoolRemoval(context.Background(),
			tfsdk.State{Raw: tftypes.NewValue(sch.Type().TerraformType(context.Background()), nil), Schema: sch},
			testKarpenterPlan(t, testKarpenterObject(gpu)),
			&diags)
		assert.Empty(t, diags)
	})
}

// --- schema wiring -------------------------------------------------------------------------

// TestClusterSchema_KarpenterGpuOverride pins the gpu_override contract: an Optional block that is
// not Computed, since its presence is what creates the GPU node pool, and a max_gpu limit that
// defaults to 0, the value the API holds when nothing is set.
func TestClusterSchema_KarpenterGpuOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var resourceSchema resource.SchemaResponse
	clusterResource{}.Schema(ctx, resource.SchemaRequest{}, &resourceSchema)
	require.False(t, resourceSchema.Diagnostics.HasError())

	karpenter := resourceSchema.Schema.Attributes["features"].(schema.SingleNestedAttribute).Attributes["karpenter"].(schema.SingleNestedAttribute)
	nodePools := karpenter.Attributes["qovery_node_pools"].(schema.SingleNestedAttribute)
	gpuOverride := nodePools.Attributes["gpu_override"].(schema.SingleNestedAttribute)

	assert.True(t, gpuOverride.IsOptional())
	assert.False(t, gpuOverride.IsComputed(), "a computed gpu_override would keep a removed block, and the GPU node pool with it")

	maxGpu := gpuOverride.Attributes["limits"].(schema.SingleNestedAttribute).Attributes["max_gpu"].(schema.Int64Attribute)
	require.NotNil(t, maxGpu.Default)
	var resp defaults.Int64Response
	maxGpu.Default.DefaultInt64(ctx, defaults.Int64Request{}, &resp)
	assert.Equal(t, types.Int64Value(0), resp.PlanValue)
}
