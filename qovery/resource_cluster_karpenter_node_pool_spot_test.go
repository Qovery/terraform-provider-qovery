//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAcc_ClusterKarpenterNodePoolSpot covers the per node pool spot_enabled flags: the
// deprecated global flag is left out of the configuration, each node pool carries its own
// value, and declaring cronjob_override enables the dedicated cronjob node pool.
//
// The PlanOnly steps are the point of the test: the API returns stable_override and
// default_override for every Karpenter cluster once the backfill runs, and it recomputes the
// global spot_enabled as the OR of the per node pool values, so both are prime candidates for
// a perpetual diff.
func TestAcc_ClusterKarpenterNodePoolSpot(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-node-pool-spot"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create with a spot value on every node pool and no global flag.
			{
				Config: testAccClusterKarpenterNodePoolSpotConfig(testName, false, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.cronjob_override.spot_enabled", "true"),
					// The deprecated global flag is derived by the API as the OR of the values above.
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Re-plan the same config after refresh: must be empty.
			{
				Config:             testAccClusterKarpenterNodePoolSpotConfig(testName, false, true, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Flip the default node pool to on-demand.
			{
				Config: testAccClusterKarpenterNodePoolSpotConfig(testName, false, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Drop the cronjob node pool: removing the block disables the dedicated pool.
			{
				Config: testAccClusterKarpenterNodePoolSpotConfigWithoutCronjob(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.cronjob_override.spot_enabled"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			{
				Config:             testAccClusterKarpenterNodePoolSpotConfigWithoutCronjob(testName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAcc_ClusterKarpenterLegacyGlobalSpot covers the configurations written before per node
// pool support: only the deprecated global flag is set and no node pool declares one. The
// backend backfill makes the API return per node pool values anyway, which must not show up in
// state as a diff.
func TestAcc_ClusterKarpenterLegacyGlobalSpot(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-legacy-spot"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterKarpenterLegacyGlobalSpotConfig(testName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "true"),
					// The backfilled node pool overrides must not land in state.
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			{
				Config:             testAccClusterKarpenterLegacyGlobalSpotConfig(testName, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccClusterKarpenterNodePoolSpotConfig(testName string, stableSpot, defaultSpot, cronjobSpot bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
        stable_override = {
          spot_enabled = %t
        }
        default_override = {
          spot_enabled = %t
        }
        cronjob_override = {
          spot_enabled = %t
        }
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), stableSpot, defaultSpot, cronjobSpot)
}

func testAccClusterKarpenterNodePoolSpotConfigWithoutCronjob(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
        stable_override = {
          spot_enabled = false
        }
        default_override = {
          spot_enabled = false
        }
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterKarpenterLegacyGlobalSpotConfig(testName string, spotEnabled bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      spot_enabled                 = %t
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), spotEnabled)
}

// TestAcc_ClusterKarpenterSpotMigration walks the migration path the deprecation documentation
// tells users to take, one phase at a time. Every phase is a place where the plan modifier, the
// plan-unknown handling and the write-path state fallback meet Terraform's proposed-new-state
// machinery — interactions that conversion-level unit tests cannot see, which is exactly how the
// import-path bug got past them before.
//
// What each phase pins:
//   - steps 1-2: a legacy configuration is quiet, and the node pool overrides the API backfills
//     do not leak into its state.
//   - steps 3-4: per node pool values land while the global is still configured, and the API
//     derives the global as the OR of them.
//   - steps 5-6: dropping the deprecated global leaves no perpetual diff — the plan modifier must
//     plan it unknown so the derived value lands instead of a frozen state value.
//   - step 7: THE GHOST DETECTOR. With every pool flipped to false the derived global must become
//     false. A stale `true` preserved from step 3's state would surface here as `true` and fail
//     this assertion — and in production it is what silently re-enables spot on every node pool
//     once the override blocks are removed.
//   - step 8: the end state is quiet too.
func TestAcc_ClusterKarpenterSpotMigration(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-spot-migration"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// 1. Legacy configuration: the global flag only.
			{
				Config: testAccClusterKarpenterLegacyGlobalSpotConfig(testName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "true"),
					// The backfilled node pool overrides must stay out of state.
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// 2. The legacy configuration is quiet.
			{
				Config:             testAccClusterKarpenterLegacyGlobalSpotConfig(testName, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// 3. Add per node pool values while keeping the global.
			{
				Config: testAccClusterKarpenterSpotMigrationConfigWithGlobal(testName, true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					// Derived as the OR of the per node pool values.
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "true"),
				),
			},
			// 4. Quiet with both levels configured.
			{
				Config:             testAccClusterKarpenterSpotMigrationConfigWithGlobal(testName, true, false, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// 5. Drop the deprecated global, keep the per node pool values.
			{
				Config: testAccClusterKarpenterSpotMigrationConfig(testName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					// Still the API's derived OR, now read back rather than configured.
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "true"),
				),
			},
			// 6. No perpetual diff after dropping the deprecated field.
			{
				Config:             testAccClusterKarpenterSpotMigrationConfig(testName, false, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// 7. Ghost detector: every pool off spot, so the derived global must be false.
			{
				Config: testAccClusterKarpenterSpotMigrationConfig(testName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "false"),
					// A stale true from step 3 would show up right here.
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.spot_enabled", "false"),
				),
			},
			// 8. The end state is quiet.
			{
				Config:             testAccClusterKarpenterSpotMigrationConfig(testName, false, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAcc_ClusterKarpenterDivergedImport pins import behaviour for a cluster whose node pools
// diverge from each other, covering both branches of the documented trade-off.
//
// Case A (spot-only overrides) is the branch that costs something: import takes the no-plan path,
// where an override carrying nothing but spot_enabled is dropped on purpose, so the imported state
// lacks blocks the applied state has. That asymmetry is deliberate — injecting such blocks on the
// no-plan path would put a permanent block-removal diff in front of every legacy user, and it is
// what broke ImportStateVerify on the generic cluster tests before (see the unit tests
// TestCreateKarpenterFeatureAttrValue_WithoutPlanSpotOnlyOverrideIsDropped and
// ..._ImportAgreesWithApply). The ignores below are that accepted cost, scoped to the two blocks.
//
// Case B (overrides carrying limits) is the stronger half and takes no ignores: it proves the
// divergence itself survives import whenever the block holds anything besides spot_enabled, so
// the trade-off costs users only the spot-only shape.
func TestAcc_ClusterKarpenterDivergedImport(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-diverged-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Case A — diverging spot values and nothing else in the blocks.
			{
				Config: testAccClusterKarpenterSpotMigrationConfig(testName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
				),
			},
			{
				ResourceName:        "qovery_cluster.test",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				// ImportStateVerifyIgnore entries are prefix matches, so naming each block covers
				// every attribute under it — the object's own "%" count marker included.
				//
				// The two node pool entries are this feature's accepted trade-off. min/max_running_nodes
				// are unrelated: Karpenter manages node scaling and the API returns sentinel values that
				// the resource preserves from plan while import reads them back raw, so every Karpenter
				// import step in this package ignores them (see TestAcc_ClusterWithKeda). Nothing here
				// asserts node counts.
				ImportStateVerifyIgnore: []string{
					"advanced_settings_json",
					"min_running_nodes",
					"max_running_nodes",
					"features.karpenter.qovery_node_pools.stable_override",
					"features.karpenter.qovery_node_pools.default_override",
				},
			},
			// Case B — the same divergence, in blocks that also carry limits.
			{
				Config: testAccClusterKarpenterDivergedImportConfigWithLimits(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.spot_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.limits.max_cpu_in_vcpu", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.spot_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.limits.max_cpu_in_vcpu", "20"),
				),
			},
			{
				// No node pool ignores — that is the point of this case: the blocks carry limits, so
				// import keeps them and their diverging spot values must round-trip. The entries below
				// are the same unrelated Karpenter/import quirks as in case A.
				ResourceName:            "qovery_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json", "min_running_nodes", "max_running_nodes"},
			},
		},
	})
}

func testAccClusterKarpenterSpotMigrationConfig(testName string, stableSpot, defaultSpot bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
        stable_override = {
          spot_enabled = %t
        }
        default_override = {
          spot_enabled = %t
        }
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), stableSpot, defaultSpot)
}

func testAccClusterKarpenterSpotMigrationConfigWithGlobal(testName string, globalSpot, stableSpot, defaultSpot bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      spot_enabled                 = %t
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
        stable_override = {
          spot_enabled = %t
        }
        default_override = {
          spot_enabled = %t
        }
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), globalSpot, stableSpot, defaultSpot)
}

func testAccClusterKarpenterDivergedImportConfigWithLimits(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large", "xlarge", "2xlarge"] },
          { key = "InstanceFamily", operator = "In", values = ["t3", "t3a", "m5", "m5a", "c5", "c5a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
        stable_override = {
          spot_enabled = false
          limits = {
            enabled                 = true
            max_cpu_in_vcpu         = 10
            max_memory_in_gibibytes = 20
          }
        }
        default_override = {
          spot_enabled = true
          limits = {
            enabled                 = true
            max_cpu_in_vcpu         = 20
            max_memory_in_gibibytes = 40
          }
        }
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}
