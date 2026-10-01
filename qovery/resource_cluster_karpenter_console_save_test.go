//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/qovery/qovery-client-go"
)

// testAccGpuUnsortedRequirements is a GPU node pool whose requirements are in another order than
// the Console writes them, with values in another order too.
const testAccGpuUnsortedRequirements = `gpu_override = {
  requirements = [
    { key = "InstanceFamily", operator = "In", values = ["g5", "g4dn"] },
    { key = "InstanceSize",   operator = "In", values = ["xlarge", "2xlarge"] },
    { key = "Arch",           operator = "In", values = ["AMD64"] },
  ]
  disk_size_in_gib = 50
}`

// testAccConsoleRequirements rewrites requirements the way the Console instance filter does: it
// lists them as InstanceSize, InstanceFamily and Arch, with their values in the order of its
// instance type catalog. The catalog is sorted by instance name, so the values are sorted here:
// that is the InstanceFamily order the Console sent in the QOV-2348 test plan.
func testAccConsoleRequirements(requirements []qovery.KarpenterNodePoolRequirement) []qovery.KarpenterNodePoolRequirement {
	rewritten := make([]qovery.KarpenterNodePoolRequirement, 0, len(requirements))
	for _, key := range []qovery.KarpenterNodePoolRequirementKey{
		qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE,
		qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY,
		qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH,
	} {
		for _, requirement := range requirements {
			if requirement.Key != key {
				continue
			}
			values := slices.Clone(requirement.Values)
			slices.Sort(values)
			rewritten = append(rewritten, qovery.KarpenterNodePoolRequirement{Key: key, Operator: requirement.Operator, Values: values})
		}
	}
	return rewritten
}

// testAccCheckClusterKarpenterRequirements asserts the requirements of the default and GPU node
// pools according to the API, order included.
func testAccCheckClusterKarpenterRequirements(clusterID *string, nodePools, gpu []qovery.KarpenterNodePoolRequirement) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		parameters, err := testAccClusterKarpenterParameters(*clusterID)
		if err != nil {
			return err
		}
		strip := func(requirements []qovery.KarpenterNodePoolRequirement) []qovery.KarpenterNodePoolRequirement {
			stripped := make([]qovery.KarpenterNodePoolRequirement, len(requirements))
			for i, requirement := range requirements {
				stripped[i] = qovery.KarpenterNodePoolRequirement{Key: requirement.Key, Operator: requirement.Operator, Values: requirement.Values}
			}
			return stripped
		}
		if got := strip(parameters.QoveryNodePools.Requirements); !reflect.DeepEqual(got, nodePools) {
			return fmt.Errorf("qovery_node_pools requirements on the API are %+v, want %+v", got, nodePools)
		}
		if parameters.QoveryNodePools.GpuOverride == nil {
			return fmt.Errorf("the cluster has no GPU node pool on the API")
		}
		if got := strip(parameters.QoveryNodePools.GpuOverride.Requirements); !reflect.DeepEqual(got, gpu) {
			return fmt.Errorf("gpu_override requirements on the API are %+v, want %+v", got, gpu)
		}
		return nil
	}
}

// TestAcc_ClusterKarpenterRequirementsFromConsole covers a Console save of the Karpenter
// instances: it rewrites the requirements in its own order, which 1.0.0-rc.1 showed as a plan
// difference that changed nothing. The plan must stay empty, and a value removed from the Console
// must still show.
func TestAcc_ClusterKarpenterRequirementsFromConsole(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-requirements-console"
	config := testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuUnsortedRequirements)
	var clusterID string

	// What the configuration sends, and what a Console save of the same instances writes.
	configured := []qovery.KarpenterNodePoolRequirement{
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"small", "medium", "large", "xlarge", "2xlarge"}},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"t3", "t3a", "m5", "m5a", "c5", "c5a"}},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"AMD64"}},
	}
	configuredGpu := []qovery.KarpenterNodePoolRequirement{
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"g5", "g4dn"}},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"xlarge", "2xlarge"}},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"AMD64"}},
	}
	fromConsole, fromConsoleGpu := testAccConsoleRequirements(configured), testAccConsoleRequirements(configuredGpu)
	requirementsPath := testAccNodePoolPath("requirements")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. The requirements reach the API in the configured order.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCaptureResourceID(testAccSpotClusterAddress, &clusterID),
					testAccCheckClusterKarpenterRequirements(&clusterID, configured, configuredGpu),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Save the instances from the Console: the API holds them in the Console order, and
			// there is nothing to plan.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(_ *terraform.State) error {
						return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
							parameters.QoveryNodePools.Requirements = fromConsole
							parameters.QoveryNodePools.GpuOverride.Requirements = fromConsoleGpu
						})
					},
					testAccCheckClusterKarpenterRequirements(&clusterID, fromConsole, fromConsoleGpu),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 3. The refresh keeps the configured order.
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.requirements.1.key", "InstanceFamily"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.requirements.1.values.0", "t3"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.requirements.0.key", "InstanceFamily"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.requirements.0.values.0", "g5"),
				),
			},
			// 4. A family removed from the Console shows in the plan, being added back.
			{
				Config: config,
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						for i, requirement := range parameters.QoveryNodePools.Requirements {
							if requirement.Key == qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY {
								parameters.QoveryNodePools.Requirements[i].Values = []string{"c5", "m5", "m5a", "t3", "t3a"}
							}
						}
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectPlannedChange(requirementsPath.AtSliceIndex(1).AtMapKey("values"),
							[]any{"c5", "m5", "m5a", "t3", "t3a"}, []any{"t3", "t3a", "m5", "m5a", "c5", "c5a"}),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 5. The corrective apply restores the configured requirements.
			{
				Config:           config,
				Check:            testAccCheckClusterKarpenterRequirements(&clusterID, configured, configuredGpu),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}
