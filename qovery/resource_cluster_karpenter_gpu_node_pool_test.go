//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/qovery/qovery-client-go"
)

// GPU node pool blocks used by the configurations below. A never-deployed cluster accepts g4dn,
// xlarge and AMD64 with a 50 GiB disk, which keeps these tests cheap.
const (
	testAccGpuMinimal = `gpu_override = {
  requirements = [
    { key = "InstanceFamily", operator = "In", values = ["g4dn"] },
    { key = "InstanceSize",   operator = "In", values = ["xlarge"] },
    { key = "Arch",           operator = "In", values = ["AMD64"] },
  ]
  disk_size_in_gib = 50
}`
	testAccGpuEveryField = `gpu_override = {
  requirements = [
    { key = "InstanceFamily", operator = "In", values = ["g4dn", "g5"] },
    { key = "InstanceSize",   operator = "In", values = ["xlarge", "2xlarge"] },
    { key = "Arch",           operator = "In", values = ["AMD64"] },
  ]
  disk_size_in_gib = 100
  disk_iops        = 3000
  disk_throughput  = 125
  spot_enabled     = true
  consolidation = {
    enabled    = true
    days       = ["MONDAY", "WEDNESDAY"]
    start_time = "PT02:00"
    duration   = "PT04H00M"
  }
  limits = {
    enabled                 = true
    max_cpu_in_vcpu         = 16
    max_memory_in_gibibytes = 64
    max_gpu                 = 4
  }
}`
	// testAccGpuClusterDataSource reads the test cluster through the data source.
	testAccGpuClusterDataSource = `
data "qovery_cluster" "test" {
  id              = qovery_cluster.test.id
  organization_id = qovery_cluster.test.organization_id
}
`
)

// testAccGpuRequirements builds the requirements of testAccGpuMinimal, or of testAccGpuEveryField
// when every is true, the way the API returns them.
func testAccGpuRequirements(every bool) []qovery.KarpenterNodePoolRequirement {
	families, sizes := []string{"g4dn"}, []string{"xlarge"}
	if every {
		families, sizes = []string{"g4dn", "g5"}, []string{"xlarge", "2xlarge"}
	}
	return []qovery.KarpenterNodePoolRequirement{
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_FAMILY, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: families},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_INSTANCE_SIZE, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: sizes},
		{Key: qovery.KARPENTERNODEPOOLREQUIREMENTKEY_ARCH, Operator: qovery.KARPENTERNODEPOOLREQUIREMENTOPERATOR_IN, Values: []string{"AMD64"}},
	}
}

// testAccGpuMinimalOnAPI is the GPU node pool of testAccGpuMinimal as the API holds it.
func testAccGpuMinimalOnAPI() *qovery.KarpenterGpuNodePoolOverride {
	return &qovery.KarpenterGpuNodePoolOverride{
		Requirements:  testAccGpuRequirements(false),
		DiskSizeInGib: new(int32(50)),
		SpotEnabled:   new(false),
	}
}

// testAccGpuEveryFieldOnAPI is the GPU node pool of testAccGpuEveryField as the API holds it.
func testAccGpuEveryFieldOnAPI() *qovery.KarpenterGpuNodePoolOverride {
	return &qovery.KarpenterGpuNodePoolOverride{
		Requirements:   testAccGpuRequirements(true),
		DiskSizeInGib:  new(int32(100)),
		DiskIops:       new(int32(3000)),
		DiskThroughput: new(int32(125)),
		SpotEnabled:    new(true),
		Consolidation:  qovery.NewKarpenterNodePoolConsolidation(true, []qovery.WeekdayEnum{qovery.WEEKDAYENUM_MONDAY, qovery.WEEKDAYENUM_WEDNESDAY}, "PT02:00", "PT04H00M"),
		Limits:         qovery.NewKarpenterNodePoolLimits(true, 16, 64, 4),
	}
}

// TestAcc_ClusterKarpenterGpuNodePool covers the lifecycle of a GPU node pool declared in
// Terraform: created with the block, updated field by field, kept by an unrelated apply, imported,
// and deleted with the block. What the pool looks like is checked against the API, not only
// against the state the provider writes.
func TestAcc_ClusterKarpenterGpuNodePool(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-gpu-node-pool"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. Declaring the block creates the GPU node pool.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuMinimal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.disk_size_in_gib", "50"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.spot_enabled", "false"),
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.disk_iops"),
					testAccCheckClusterGpuNodePool(testAccGpuMinimalOnAPI()),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "state", "READY"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Every field of the block reaches the API.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuEveryField),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.limits.max_gpu", "4"),
					testAccCheckClusterGpuNodePool(testAccGpuEveryFieldOnAPI()),
					// The GPU node pool has its own spot flag: the other pools stay on-demand.
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
				),
			},
			// 3. An unrelated apply keeps the declared GPU node pool, and the data source reports it.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccGpuEveryField) + testAccGpuClusterDataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterGpuNodePool(testAccGpuEveryFieldOnAPI()),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.gpu_override.disk_size_in_gib", "100"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.gpu_override.spot_enabled", "true"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.gpu_override.limits.max_gpu", "4"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Import recovers the whole block.
			{
				ResourceName:        testAccSpotClusterAddress,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				// min/max_running_nodes: see TestAcc_ClusterKarpenterDivergedImport.
				ImportStateVerifyIgnore: []string{
					"advanced_settings_json",
					"min_running_nodes",
					"max_running_nodes",
				},
			},
			// 5. Removing the block deletes the GPU node pool, and the plan shows it.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "gpu_override", true, false),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.gpu_override.disk_size_in_gib"),
					testAccCheckClusterGpuNodePool(nil),
				),
			},
		},
	})
}

// TestAcc_ClusterKarpenterGpuPoolDrift covers a GPU node pool created or deleted outside
// Terraform, e.g. from the Qovery Console. q-core rebuilds the node pools from each request, so
// before 1.0 an unrelated apply deleted a Console-created GPU pool while the plan showed nothing.
// Either change must now show in the plan, and declaring the block adopts a Console-created pool.
func TestAcc_ClusterKarpenterGpuPoolDrift(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-gpu-drift"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. No GPU node pool.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCaptureResourceID(testAccSpotClusterAddress, &clusterID),
					testAccCheckClusterGpuNodePool(nil),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Create the pool out of band: the plan shows gpu_override being removed.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						parameters.QoveryNodePools.GpuOverride = testAccGpuMinimalOnAPI()
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "gpu_override", true, false),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 3. Declaring the same block adopts the pool: nothing to apply, and the pool stays.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuMinimal),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckClusterGpuNodePool(testAccGpuMinimalOnAPI()),
			},
			// 4. Delete the pool out of band: the plan shows gpu_override being added back.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuMinimal),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						parameters.QoveryNodePools.GpuOverride = nil
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "gpu_override", false, true),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 5. The corrective apply creates the pool again.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccGpuMinimal),
				Check:            testAccCheckClusterGpuNodePool(testAccGpuMinimalOnAPI()),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// testAccCheckClusterGpuNodePool asserts the GPU node pool of the test cluster according to the
// API, independently of what the provider stores in state. A nil want asserts the cluster has no
// GPU node pool.
func testAccCheckClusterGpuNodePool(want *qovery.KarpenterGpuNodePoolOverride) resource.TestCheckFunc {
	const resourceName = testAccSpotClusterAddress
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("%s: id not found in state", resourceName)
		}

		parameters, err := testAccClusterKarpenterParameters(rs.Primary.ID)
		if err != nil {
			return err
		}

		got := parameters.QoveryNodePools.GpuOverride
		if want == nil || got == nil {
			if want != got {
				return fmt.Errorf("%s: GPU node pool on the API is %+v, want %+v", resourceName, got, want)
			}
			return nil
		}

		// AdditionalProperties holds whatever the client does not model; it is not compared.
		got.AdditionalProperties = nil
		if got.Consolidation != nil {
			got.Consolidation.AdditionalProperties = nil
		}
		if got.Limits != nil {
			got.Limits.AdditionalProperties = nil
		}
		for i := range got.Requirements {
			got.Requirements[i].AdditionalProperties = nil
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("%s: GPU node pool on the API is %s, want %s", resourceName, testAccFormatGpuNodePool(got), testAccFormatGpuNodePool(want))
		}
		return nil
	}
}

func testAccFormatGpuNodePool(o *qovery.KarpenterGpuNodePoolOverride) string {
	json, err := o.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("%+v", *o)
	}
	return string(json)
}
