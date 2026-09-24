//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

// testAccLastProvider0x is the last 0.x release, the version users upgrade to 1.0 from.
const testAccLastProvider0x = "0.89.0"

const testAccSpotClusterAddress = "qovery_cluster.test"

// testAccLastProvider0xFromRegistry runs a step with the released 0.x provider instead of the one
// under test.
var testAccLastProvider0xFromRegistry = map[string]resource.ExternalProvider{
	"qovery": {Source: "qovery/qovery", VersionConstraint: testAccLastProvider0x},
}

// Node pool override blocks used by the configurations below.
const (
	testAccStableSpot       = "stable_override = {\n  spot_enabled = true\n}"
	testAccStableOnDemand   = "stable_override = {\n  spot_enabled = false\n}"
	testAccDefaultSpot      = "default_override = {\n  spot_enabled = true\n}"
	testAccDefaultEmpty     = "default_override = {}"
	testAccCronjobSpot      = "cronjob_override = {\n  spot_enabled = true\n}"
	testAccGlobalSpot0x     = "spot_enabled = true"
	testAccNoKarpenterExtra = ""
)

// TestAcc_ClusterKarpenterNodePoolSpot covers the per node pool spot_enabled flags: every node pool
// carries its own value, a value or block left out means on-demand, and declaring
// cronjob_override enables the dedicated cronjob node pool. Where each pool runs is checked
// against the API, not only against the state the provider writes.
func TestAcc_ClusterKarpenterNodePoolSpot(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-node-pool-spot"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. A value on every node pool.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand, testAccDefaultSpot, testAccCronjobSpot),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.cronjob_override.spot_enabled", "true"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": true, "cronjob": true}),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "state", "READY"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Removing the value from a declared block moves the pool to on-demand, and the plan
			// says so.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand, testAccDefaultEmpty, testAccCronjobSpot),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(testAccSpotClusterAddress, testAccNodePoolPath("default_override").AtMapKey("spot_enabled"), knownvalue.Bool(false)),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.spot_enabled", "false"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false, "cronjob": true}),
				),
			},
			// 3. Back on spot.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand, testAccDefaultSpot, testAccCronjobSpot),
				Check:            testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": true, "cronjob": true}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Removing the whole block moves the pool to on-demand too, and the plan shows the
			// block going away.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand, testAccCronjobSpot),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolsLeaveSpot(testAccSpotClusterAddress, "default_override"),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.spot_enabled"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false, "cronjob": true}),
				),
			},
			// 5. Dropping cronjob_override disables the dedicated cronjob node pool.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.cronjob_override.spot_enabled"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "state", "READY"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterPartialSpotConfig is the regression test for the 0.x write path, which
// resent the global flag it kept in state on every update. The API hands that flag to any node
// pool the request leaves without a value, so with only default_override pinned to spot, the
// stable node pool was created on-demand and then moved to spot by the next unrelated apply.
func TestAcc_ClusterKarpenterPartialSpotConfig(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-partial-spot"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "before", testAccNoKarpenterExtra, testAccDefaultSpot),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.spot_enabled"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": true}),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// An update that has nothing to do with spot instances.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "after", testAccNoKarpenterExtra, testAccDefaultSpot),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "description", "after"),
					resource.TestCheckNoResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.spot_enabled"),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": true}),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterSpotDriftOnUndeclaredNodePool covers a node pool moved to spot outside
// Terraform, e.g. from the Qovery Console, while the configuration does not declare it: the plan
// must show the move back to on-demand rather than the next apply performing it silently.
func TestAcc_ClusterKarpenterSpotDriftOnUndeclaredNodePool(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-spot-drift"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCaptureResourceID(testAccSpotClusterAddress, &clusterID),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Move the default node pool to spot out of band. The harness plans with refresh after
			// this step: the plan must show default_override leaving spot.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						if parameters.QoveryNodePools.DefaultOverride == nil {
							parameters.QoveryNodePools.DefaultOverride = &qovery.KarpenterDefaultNodePoolOverride{}
						}
						parameters.QoveryNodePools.DefaultOverride.SetSpotEnabled(true)
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolsLeaveSpot(testAccSpotClusterAddress, "default_override"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// The corrective apply moves the pool back to on-demand.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check:            testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterCronjobPoolDrift covers the dedicated cronjob node pool enabled or
// disabled outside Terraform, e.g. from the Qovery Console. The presence of cronjob_override is
// what enables the pool, so either change must show in the plan as the block being removed or
// added back, instead of the next apply reverting it silently.
func TestAcc_ClusterKarpenterCronjobPoolDrift(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-cronjob-drift"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. No cronjob pool.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCaptureResourceID(testAccSpotClusterAddress, &clusterID),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Enable the pool out of band: the plan shows cronjob_override being removed.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						cronjob := &qovery.KarpenterCronjobNodePoolOverride{}
						cronjob.SetSpotEnabled(false)
						parameters.QoveryNodePools.CronjobOverride = cronjob
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "cronjob_override", true, false),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 3. The corrective apply disables the pool again.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				Check:            testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Declare the pool.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccCronjobSpot),
				Check:            testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false, "cronjob": true}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 5. Disable the pool out of band: the plan shows cronjob_override being added back.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccCronjobSpot),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						parameters.QoveryNodePools.CronjobOverride = nil
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "cronjob_override", false, true),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 6. The corrective apply enables the pool again.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccCronjobSpot),
				Check:            testAccCheckClusterKarpenterSpot(map[string]bool{"stable": false, "default": false, "cronjob": true}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterSpotUpgradeWithoutPinning upgrades from the last 0.x release a cluster
// whose node pools run on spot through the removed global flag, skipping the pinning step of the
// upgrade guide. The first 1.0 plan must show both pools leaving spot; pinning them then plans
// clean and keeps them on spot through an unrelated update.
func TestAcc_ClusterKarpenterSpotUpgradeWithoutPinning(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-spot-upgrade"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. Last 0.x release, global flag only.
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            testAccClusterKarpenterSpotConfig(testName, "", testAccGlobalSpot0x),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
				),
			},
			// 2. 1.0 with the global flag removed and nothing pinned: the plan shows both pools
			// leaving spot.
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolsLeaveSpot(testAccSpotClusterAddress, "stable_override", "default_override"),
					},
				},
			},
			// 3. Pinning both pools on spot changes nothing on the cluster.
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableSpot, testAccDefaultSpot),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
			},
			// 4. An unrelated update keeps both pools on spot.
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterKarpenterSpotConfig(testName, "after upgrade", testAccNoKarpenterExtra, testAccStableSpot, testAccDefaultSpot),
				Check:                    testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
				ConfigPlanChecks:         testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterSpotUpgradePinnedOn0x follows the upgrade guide: pin every pool's
// effective spot value on the last 0.x release, then remove the global flag and upgrade. The
// first 1.0 plan must be empty, and the pools keep their spot placement through an unrelated
// update.
func TestAcc_ClusterKarpenterSpotUpgradePinnedOn0x(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-spot-pinned"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. Last 0.x release, global flag only.
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            testAccClusterKarpenterSpotConfig(testName, "", testAccGlobalSpot0x),
				Check:             testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
			},
			// 2. Guide step 1, still on 0.x: pin each pool to the value it inherits.
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            testAccClusterKarpenterSpotConfig(testName, "", testAccGlobalSpot0x, testAccStableSpot, testAccDefaultSpot),
				Check:             testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
			},
			// 3. Guide step 2: remove the global flag and upgrade. The plan is empty.
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableSpot, testAccDefaultSpot),
				PlanOnly:                 true,
			},
			// 4. An unrelated update keeps both pools on spot.
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterKarpenterSpotConfig(testName, "after upgrade", testAccNoKarpenterExtra, testAccStableSpot, testAccDefaultSpot),
				Check:                    testAccCheckClusterKarpenterSpot(map[string]bool{"stable": true, "default": true}),
				ConfigPlanChecks:         testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterDivergedImport pins what an import of a cluster with mixed spot values
// recovers. The API omits every per node pool value equal to the global flag it derived, and the
// read path resolves an omitted value to that global, so where each pool runs always round-trips.
// What import cannot recover is whether an on-demand block with nothing else in it was declared:
// such a block says nothing the schema default would not, so it stays out of the imported state.
func TestAcc_ClusterKarpenterDivergedImport(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-diverged-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// Case A: spot values and nothing else in the blocks.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra, testAccStableOnDemand, testAccDefaultSpot),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
				),
			},
			{
				ResourceName:        testAccSpotClusterAddress,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				// ImportStateVerifyIgnore entries are prefix matches, so naming a block covers every
				// attribute under it, the object's own "%" count marker included.
				//
				// stable_override holds nothing but spot_enabled = false, the schema default, so
				// import has no way to tell it was declared and leaves it out (unit test
				// TestKarpenterFeatureAttrValue_WithoutPlan). default_override runs on spot, which
				// differs from the default, so it round-trips and needs no ignore.
				//
				// min/max_running_nodes are unrelated to this feature: Karpenter manages node scaling and
				// the API returns sentinel values that the resource preserves from plan while import reads
				// them back raw, so every Karpenter import step in this package ignores them (see
				// TestAcc_ClusterWithKeda).
				ImportStateVerifyIgnore: []string{
					"advanced_settings_json",
					"min_running_nodes",
					"max_running_nodes",
					"features.karpenter.qovery_node_pools.stable_override",
				},
			},
			// Case B: the same values, in blocks that also carry limits. Both blocks, their limits and
			// both spot values round-trip with no ignore of their own.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "", testAccNoKarpenterExtra,
					"stable_override = {\n  spot_enabled = false\n  limits = {\n    enabled = true\n    max_cpu_in_vcpu = 10\n    max_memory_in_gibibytes = 20\n  }\n}",
					"default_override = {\n  spot_enabled = true\n  limits = {\n    enabled = true\n    max_cpu_in_vcpu = 20\n    max_memory_in_gibibytes = 40\n  }\n}",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.limits.max_cpu_in_vcpu", "10"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.default_override.limits.max_cpu_in_vcpu", "20"),
				),
			},
			{
				ResourceName:        testAccSpotClusterAddress,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{
					"advanced_settings_json",
					"min_running_nodes",
					"max_running_nodes",
				},
			},
		},
	})
}

// --- configuration ------------------------------------------------------------------------------

// testAccClusterKarpenterSpotConfig renders an AWS Karpenter cluster in READY state, i.e. never
// deployed. karpenterExtra goes into the karpenter block (the 0.x global flag), and every
// nodePoolBlocks entry into qovery_node_pools after the requirements.
func testAccClusterKarpenterSpotConfig(testName, description, karpenterExtra string, nodePoolBlocks ...string) string {
	descriptionLine := ""
	if description != "" {
		descriptionLine = fmt.Sprintf("\n  description     = %q", description)
	}

	blocks := ""
	for _, block := range nodePoolBlocks {
		blocks += "\n" + block
	}

	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"%s
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      %s
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]%s
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), descriptionLine, karpenterExtra, blocks)
}

// --- checks -------------------------------------------------------------------------------------

func testAccNodePoolPath(override string) tfjsonpath.Path {
	return tfjsonpath.New("features").AtMapKey("karpenter").AtMapKey("qovery_node_pools").AtMapKey(override)
}

// testAccReadClusterFromAPI reads a cluster straight from the API. It retries like
// testAccQoveryClusterExists, because the cluster list briefly fails right after a write.
func testAccReadClusterFromAPI(clusterID string) (*qovery.Cluster, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		clusters, res, err := qoveryAPIClient.ClustersAPI.ListOrganizationCluster(context.TODO(), getTestOrganizationID()).Execute()
		switch {
		case err != nil:
			lastErr = err
		case res != nil && res.StatusCode >= 400:
			lastErr = fmt.Errorf("HTTP %d", res.StatusCode)
		default:
			for _, cluster := range clusters.GetResults() {
				if cluster.Id == clusterID {
					return &cluster, nil
				}
			}
			lastErr = fmt.Errorf("cluster not found")
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("failed to read cluster %s: %w", clusterID, lastErr)
}

// testAccClusterKarpenterParameters reads the Karpenter parameters of a cluster from the API.
func testAccClusterKarpenterParameters(clusterID string) (*qovery.ClusterFeatureKarpenterParameters, error) {
	cluster, err := testAccReadClusterFromAPI(clusterID)
	if err != nil {
		return nil, err
	}
	for _, f := range cluster.Features {
		if f.GetId() != "KARPENTER" {
			continue
		}
		if value := f.GetValueObject().ClusterFeatureKarpenterParametersResponse; value != nil {
			parameters := value.Value
			return &parameters, nil
		}
	}
	return nil, fmt.Errorf("cluster %s has no karpenter feature", clusterID)
}

// testAccEffectiveSpotEnabled resolves where a node pool runs: the API omits a per node pool value
// equal to the global flag it derived.
func testAccEffectiveSpotEnabled(value *bool, isSet bool, globalSpotEnabled bool) bool {
	if isSet && value != nil {
		return *value
	}
	return globalSpotEnabled
}

// testAccCheckClusterKarpenterSpot asserts where each node pool of the test cluster runs according
// to the API, independently of what the provider stores in state. A "cronjob" entry also asserts
// the cronjob node pool exists; leaving it out asserts it does not.
func testAccCheckClusterKarpenterSpot(want map[string]bool) resource.TestCheckFunc {
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

		//nolint:staticcheck // SA1019: the deprecated global flag is how omitted per node pool values resolve
		global := parameters.SpotEnabled
		nodePools := parameters.QoveryNodePools
		stableValue, stableSet := nodePools.StableOverride.GetSpotEnabledOk()
		defaultValue, defaultSet := nodePools.DefaultOverride.GetSpotEnabledOk()
		got := map[string]bool{
			"stable":  testAccEffectiveSpotEnabled(stableValue, stableSet, global),
			"default": testAccEffectiveSpotEnabled(defaultValue, defaultSet, global),
		}
		if nodePools.CronjobOverride != nil {
			cronjobValue, cronjobSet := nodePools.CronjobOverride.GetSpotEnabledOk()
			got["cronjob"] = testAccEffectiveSpotEnabled(cronjobValue, cronjobSet, global)
		}

		if len(got) != len(want) {
			return fmt.Errorf("%s: node pools on the API are %v, want %v", resourceName, got, want)
		}
		for pool, spotEnabled := range want {
			if got[pool] != spotEnabled {
				return fmt.Errorf("%s: %s node pool spot_enabled is %t on the API, want %t (all pools: %v)", resourceName, pool, got[pool], spotEnabled, got)
			}
		}
		return nil
	}
}

// testAccSetClusterKarpenterOutOfBand edits the Karpenter parameters of a cluster directly through
// the API, bypassing Terraform, the way the Qovery Console does. The request resends the settings
// the provider manages for the configurations in this file.
func testAccSetClusterKarpenterOutOfBand(clusterID string, mutate func(*qovery.ClusterFeatureKarpenterParameters)) error {
	if clusterID == "" {
		return fmt.Errorf("cluster id was not captured from state")
	}

	cluster, err := testAccReadClusterFromAPI(clusterID)
	if err != nil {
		return err
	}

	request := qovery.ClusterRequest{
		Name:                           cluster.Name,
		Description:                    cluster.Description,
		Region:                         cluster.Region,
		CloudProvider:                  cluster.CloudProvider,
		MinRunningNodes:                cluster.MinRunningNodes,
		MaxRunningNodes:                cluster.MaxRunningNodes,
		DiskSize:                       cluster.DiskSize,
		InstanceType:                   cluster.InstanceType,
		Kubernetes:                     cluster.Kubernetes,
		Production:                     cluster.Production,
		MetricsParameters:              cluster.MetricsParameters,
		InfrastructureChartsParameters: cluster.InfrastructureChartsParameters,
		Keda:                           cluster.Keda,
		LabelsGroups:                   cluster.LabelsGroups,
	}

	for _, f := range cluster.Features {
		id := f.GetId()
		valueObject := f.GetValueObject()
		var value qovery.ClusterRequestFeaturesInnerValue
		switch {
		case valueObject.ClusterFeatureKarpenterParametersResponse != nil:
			parameters := valueObject.ClusterFeatureKarpenterParametersResponse.Value
			mutate(&parameters)
			value.ClusterFeatureKarpenterParameters = &parameters
		case valueObject.ClusterFeatureStringResponse != nil:
			value.String = &valueObject.ClusterFeatureStringResponse.Value
		case valueObject.ClusterFeatureBooleanResponse != nil:
			value.Bool = &valueObject.ClusterFeatureBooleanResponse.Value
		default:
			continue
		}
		request.Features = append(request.Features, qovery.ClusterRequestFeaturesInner{
			Id:    &id,
			Value: *qovery.NewNullableClusterRequestFeaturesInnerValue(&value),
		})
	}

	_, httpRes, err := qoveryAPIClient.ClustersAPI.EditCluster(context.TODO(), getTestOrganizationID(), clusterID).ClusterRequest(request).Execute()
	if err != nil {
		return fmt.Errorf("failed to edit cluster %s out of band: %w", clusterID, err)
	}
	if httpRes != nil && httpRes.StatusCode >= 400 {
		return fmt.Errorf("failed to edit cluster %s out of band: HTTP %d", clusterID, httpRes.StatusCode)
	}
	return nil
}

// testAccExpectNodePoolsLeaveSpot is a plan check asserting that each named node pool override
// runs on spot before the change and is planned away, i.e. that the plan moves the pool to
// on-demand instances by removing its block.
func testAccExpectNodePoolsLeaveSpot(address string, overrides ...string) plancheck.PlanCheck {
	return nodePoolsLeaveSpotCheck{address: address, overrides: overrides}
}

type nodePoolsLeaveSpotCheck struct {
	address   string
	overrides []string
}

func (c nodePoolsLeaveSpotCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address != c.address || change.Change == nil {
			continue
		}
		for _, override := range c.overrides {
			before := testAccNodePoolOverride(change.Change.Before, override)
			if before == nil || before["spot_enabled"] != true {
				resp.Error = fmt.Errorf("%s: %s must run on spot before the change, got %v", c.address, override, before)
				return
			}
			if after := testAccNodePoolOverride(change.Change.After, override); after != nil {
				resp.Error = fmt.Errorf("%s: %s must be planned away, got %v", c.address, override, after)
				return
			}
		}
		return
	}
	resp.Error = fmt.Errorf("%s: no planned change", c.address)
}

// testAccExpectNodePoolBlock is a plan check asserting whether a node pool override block is
// present before and after the planned change.
func testAccExpectNodePoolBlock(address, override string, presentBefore, presentAfter bool) plancheck.PlanCheck {
	return nodePoolBlockCheck{address: address, override: override, presentBefore: presentBefore, presentAfter: presentAfter}
}

type nodePoolBlockCheck struct {
	address       string
	override      string
	presentBefore bool
	presentAfter  bool
}

func (c nodePoolBlockCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address != c.address || change.Change == nil {
			continue
		}
		before := testAccNodePoolOverride(change.Change.Before, c.override) != nil
		after := testAccNodePoolOverride(change.Change.After, c.override) != nil
		if before != c.presentBefore || after != c.presentAfter {
			resp.Error = fmt.Errorf("%s: %s present before=%t after=%t, want before=%t after=%t", c.address, c.override, before, after, c.presentBefore, c.presentAfter)
		}
		return
	}
	resp.Error = fmt.Errorf("%s: no planned change", c.address)
}

// testAccNodePoolOverride digs a node pool override out of a planned resource value, or returns
// nil when it is absent.
func testAccNodePoolOverride(value any, override string) map[string]any {
	for _, key := range []string{"features", "karpenter", "qovery_node_pools", override} {
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = object[key]
	}
	object, _ := value.(map[string]any)
	return object
}
