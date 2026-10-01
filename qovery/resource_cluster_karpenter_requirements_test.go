//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testApiRequirement builds a requirement of an API response.
func testApiRequirement(key qovery.KarpenterNodePoolRequirementKey, values ...string) qovery.KarpenterNodePoolRequirement {
	return qovery.KarpenterNodePoolRequirement{Key: key, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: values}
}

const (
	testFamily = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY
	testSize   = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE
	testArch   = qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH
)

// TestKarpenterFeatureAttrValue_RequirementsKeepThePlannedForm pins the read of the requirements:
// the Console instance filter lists them as InstanceSize, InstanceFamily and Arch, with their
// values in the order of its instance type catalog and without duplicates. A requirement matches
// any of its values, so a response that differs from the plan or the state only in these ways
// keeps the planned form. Any other difference follows the API, so it shows in the plan.
func TestKarpenterFeatureAttrValue_RequirementsKeepThePlannedForm(t *testing.T) {
	t.Parallel()

	planned := testGpuRequirements(
		testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
		testKarpenterRequirement("InstanceSize", "large", "medium"),
		testKarpenterRequirement("Arch", "AMD64", "ARM64"),
	)
	// What a Console save of the same instances sends.
	fromConsole := []qovery.KarpenterNodePoolRequirement{
		testApiRequirement(testSize, "medium", "large"),
		testApiRequirement(testFamily, "c5", "c5a", "m5", "m5a", "t3", "t3a"),
		testApiRequirement(testArch, "ARM64", "AMD64"),
	}
	inConsoleOrder := testGpuRequirements(
		testKarpenterRequirement("InstanceSize", "medium", "large"),
		testKarpenterRequirement("InstanceFamily", "c5", "c5a", "m5", "m5a", "t3", "t3a"),
		testKarpenterRequirement("Arch", "ARM64", "AMD64"),
	)

	testCases := []struct {
		TestName string
		Api      []qovery.KarpenterNodePoolRequirement
		Prior    types.List
		Mode     clusterReadMode
		Expect   types.List
	}{
		{
			TestName: "console_save_keeps_the_planned_form",
			Api:      fromConsole,
			Prior:    planned,
			Expect:   planned,
		},
		{
			// The Console drops a repeated value.
			TestName: "console_deduplication_keeps_the_planned_values",
			Api:      []qovery.KarpenterNodePoolRequirement{testApiRequirement(testFamily, "t3", "m5"), testApiRequirement(testSize, "large"), testApiRequirement(testArch, "AMD64")},
			Prior: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "m5", "t3", "m5"),
				testKarpenterRequirement("InstanceSize", "large"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "m5", "t3", "m5"),
				testKarpenterRequirement("InstanceSize", "large"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
		},
		{
			// The requirements keep the planned order, and only the changed values follow the API.
			TestName: "added_value_follows_the_api",
			Api: []qovery.KarpenterNodePoolRequirement{
				testApiRequirement(testSize, "medium", "large"),
				testApiRequirement(testFamily, "c5", "c5a", "c6i", "m5", "m5a", "t3", "t3a"),
				testApiRequirement(testArch, "ARM64", "AMD64"),
			},
			Prior: planned,
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "c5", "c5a", "c6i", "m5", "m5a", "t3", "t3a"),
				testKarpenterRequirement("InstanceSize", "large", "medium"),
				testKarpenterRequirement("Arch", "AMD64", "ARM64"),
			),
		},
		{
			TestName: "removed_value_follows_the_api",
			Api: []qovery.KarpenterNodePoolRequirement{
				testApiRequirement(testSize, "medium", "large"),
				testApiRequirement(testFamily, "c5", "c5a", "m5", "m5a", "t3", "t3a"),
				testApiRequirement(testArch, "AMD64"),
			},
			Prior: planned,
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
				testKarpenterRequirement("InstanceSize", "large", "medium"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
		},
		{
			// A requirement the plan does not hold keeps the API order of every requirement.
			TestName: "other_requirements_follow_the_api_order",
			Api:      fromConsole,
			Prior: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
				testKarpenterRequirement("Arch", "AMD64", "ARM64"),
			),
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceSize", "medium", "large"),
				testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
				testKarpenterRequirement("Arch", "AMD64", "ARM64"),
			),
		},
		{
			// The write path accepts two requirements with the same key, which cannot be matched to
			// the API ones: the values follow the API, here an echo of the request.
			TestName: "repeated_key_follows_the_api",
			Api: []qovery.KarpenterNodePoolRequirement{
				testApiRequirement(testFamily, "t3", "m5"),
				testApiRequirement(testFamily, "m5", "t3"),
				testApiRequirement(testSize, "large"),
				testApiRequirement(testArch, "AMD64"),
			},
			Prior: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "t3", "m5"),
				testKarpenterRequirement("InstanceFamily", "m5", "t3"),
				testKarpenterRequirement("InstanceSize", "large"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceFamily", "t3", "m5"),
				testKarpenterRequirement("InstanceFamily", "m5", "t3"),
				testKarpenterRequirement("InstanceSize", "large"),
				testKarpenterRequirement("Arch", "AMD64"),
			),
		},
		{
			// Values unknown at plan time may resolve to anything: they follow the API.
			TestName: "unknown_planned_values_follow_the_api",
			Api:      fromConsole,
			Prior: testGpuRequirements(
				types.ObjectValueMust(karpenterRequirementAttrTypes(), map[string]attr.Value{
					"key":      types.StringValue("InstanceSize"),
					"operator": types.StringValue("In"),
					"values":   types.ListUnknown(types.StringType),
				}),
				testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
				testKarpenterRequirement("Arch", "AMD64", "ARM64"),
			),
			Expect: testGpuRequirements(
				testKarpenterRequirement("InstanceSize", "medium", "large"),
				testKarpenterRequirement("InstanceFamily", "t3", "t3a", "m5", "m5a", "c5", "c5a"),
				testKarpenterRequirement("Arch", "AMD64", "ARM64"),
			),
		},
		{
			TestName: "import_reports_the_api_form",
			Api:      fromConsole,
			Prior:    types.ListNull(types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()}),
			Expect:   inConsoleOrder,
		},
		{
			TestName: "data_source_reports_the_api_form",
			Api:      fromConsole,
			Prior:    types.ListNull(types.ObjectType{AttrTypes: karpenterRequirementAttrTypes()}),
			Mode:     clusterReadModeDataSource,
			Expect:   inConsoleOrder,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			plan := testNoPlan()
			if !tc.Prior.IsNull() {
				plan = testKarpenterObject(map[string]attr.Value{"requirements": tc.Prior})
			}
			parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{})
			parameters.QoveryNodePools.Requirements = tc.Api

			attrVals := karpenterFeatureAttrValue(parameters, plan, tc.Mode)
			require.NotNil(t, attrVals)

			nodePools, ok := attrVals["qovery_node_pools"].(types.Object)
			require.True(t, ok, "qovery_node_pools missing from the converted state")
			got := nodePools.Attributes()["requirements"]
			assert.True(t, tc.Expect.Equal(got), "want %s, got %s", tc.Expect, got)
		})
	}
}

// TestKarpenterFeatureAttrValue_GpuRequirementsKeepThePlannedForm is the same rule for the GPU
// node pool.
func TestKarpenterFeatureAttrValue_GpuRequirementsKeepThePlannedForm(t *testing.T) {
	t.Parallel()

	configured := testGpuOverrideObject(func(attrs map[string]attr.Value) {
		attrs["requirements"] = testGpuRequirements(
			testKarpenterRequirement("InstanceFamily", "g5", "g4dn"),
			testKarpenterRequirement("InstanceSize", "xlarge"),
			testKarpenterRequirement("Arch", "AMD64"),
		)
	})
	apiGpu := testApiGpu()
	apiGpu.Requirements = []qovery.KarpenterNodePoolRequirement{
		testApiRequirement(testSize, "xlarge"),
		testApiRequirement(testFamily, "g4dn", "g5"),
		testApiRequirement(testArch, "AMD64"),
	}
	parameters := testApiKarpenterParameters(false, qovery.KarpenterNodePool{GpuOverride: apiGpu})

	t.Run("console_save_keeps_the_planned_form", func(t *testing.T) {
		t.Parallel()

		attrVals := karpenterFeatureAttrValue(parameters, testGpuKarpenterObject(configured), clusterReadModeResource)
		require.NotNil(t, attrVals)
		got := stateOverride(t, attrVals, "gpu_override")
		assert.True(t, configured.Equal(got), "want %s, got %s", configured, got)
	})

	t.Run("undeclared_pool_reports_the_api_form", func(t *testing.T) {
		t.Parallel()

		attrVals := karpenterFeatureAttrValue(parameters, testKarpenterObject(nil), clusterReadModeResource)
		require.NotNil(t, attrVals)
		got := stateOverride(t, attrVals, "gpu_override").Attributes()["requirements"]
		want := testGpuRequirements(
			testKarpenterRequirement("InstanceSize", "xlarge"),
			testKarpenterRequirement("InstanceFamily", "g4dn", "g5"),
			testKarpenterRequirement("Arch", "AMD64"),
		)
		assert.True(t, want.Equal(got), "want %s, got %s", want, got)
	})
}
