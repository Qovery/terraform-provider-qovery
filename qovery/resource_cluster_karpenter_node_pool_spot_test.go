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
	t.Skip("QOV-2155: per node pool spot_enabled is not deployed on the API yet — unskip once it ships")
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
	t.Skip("QOV-2155: per node pool spot_enabled is not deployed on the API yet — unskip once it ships")
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
