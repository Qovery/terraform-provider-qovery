//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

// Karpenter settings used by the configurations below: consolidate_after on every node pool,
// consolidation and limits on the cronjob node pool, and the node disk IOPS and throughput.
const (
	testAccKarpenterDisk   = "disk_iops       = 4000\n      disk_throughput = 250"
	testAccStableSettings  = "stable_override = {\n  consolidate_after = \"1h\"\n}"
	testAccDefaultSettings = "default_override = {\n  consolidate_after = \"2m\"\n}"
	testAccCronjobSettings = `cronjob_override = {
  consolidate_after = "30s"
  consolidation = {
    enabled    = true
    days       = ["MONDAY", "WEDNESDAY"]
    start_time = "PT02:00"
    duration   = "PT04H00M"
  }
  limits = {
    enabled                 = true
    max_cpu_in_vcpu         = 8
    max_memory_in_gibibytes = 16
  }
}`
	testAccGpuSettings = `gpu_override = {
  requirements = [
    { key = "InstanceFamily", operator = "In", values = ["g4dn"] },
    { key = "InstanceSize",   operator = "In", values = ["xlarge"] },
    { key = "Arch",           operator = "In", values = ["AMD64"] },
  ]
  disk_size_in_gib  = 50
  consolidate_after = "10m"
}`
	// testAccCronjobLimitsTooLow sets cronjob limits below the 6 vCPU q-core requires.
	testAccCronjobLimitsTooLow = `cronjob_override = {
  limits = {
    enabled                 = true
    max_cpu_in_vcpu         = 2
    max_memory_in_gibibytes = 16
  }
}`
	testAccStableConsolidateAfterNotCanonical = "stable_override = {\n  consolidate_after = \"60m\"\n}"
	testAccCronjobEmpty                       = "cronjob_override = {}"
)

// testAccKarpenterSettings is what the API holds for the settings under test.
type testAccKarpenterSettings struct {
	DiskIops                *int32
	DiskThroughput          *int32
	StableConsolidateAfter  *string
	DefaultConsolidateAfter *string
	CronjobConsolidateAfter *string
	CronjobConsolidation    *qovery.KarpenterNodePoolConsolidation
	CronjobLimits           *qovery.KarpenterNodePoolLimits
	GpuConsolidateAfter     *string
}

// testAccKarpenterSettingsOfConfig is the API view of the configuration with every setting.
func testAccKarpenterSettingsOfConfig() testAccKarpenterSettings {
	return testAccKarpenterSettings{
		DiskIops:                new(int32(4000)),
		DiskThroughput:          new(int32(250)),
		StableConsolidateAfter:  new("1h"),
		DefaultConsolidateAfter: new("2m"),
		CronjobConsolidateAfter: new("30s"),
		CronjobConsolidation:    qovery.NewKarpenterNodePoolConsolidation(true, []qovery.WeekdayEnum{qovery.WEEKDAYENUM_MONDAY, qovery.WEEKDAYENUM_WEDNESDAY}, "PT02:00", "PT04H00M"),
		CronjobLimits:           qovery.NewKarpenterNodePoolLimits(true, 8, 16, 0),
		GpuConsolidateAfter:     new("10m"),
	}
}

// testAccKarpenterSettingsFromConsole is what the out-of-band edit of step 4 sets: a different
// value for every setting.
func testAccKarpenterSettingsFromConsole() testAccKarpenterSettings {
	return testAccKarpenterSettings{
		DiskIops:                new(int32(5000)),
		DiskThroughput:          new(int32(300)),
		StableConsolidateAfter:  new("2h"),
		DefaultConsolidateAfter: new("5m"),
		CronjobConsolidateAfter: new("45s"),
		CronjobConsolidation:    qovery.NewKarpenterNodePoolConsolidation(true, []qovery.WeekdayEnum{qovery.WEEKDAYENUM_FRIDAY}, "PT03:00", "PT02H00M"),
		CronjobLimits:           qovery.NewKarpenterNodePoolLimits(true, 10, 20, 0),
		GpuConsolidateAfter:     new("20m"),
	}
}

// TestAcc_ClusterKarpenterNodePoolSettings covers the Karpenter settings the Console sets and
// q-core rebuilds from each request. Before 1.0 the provider left them out of every request, so
// any apply, even one that only changed the description, wiped a value set from the Console while
// the plan showed nothing. Each setting must now round-trip, show in the plan when changed from
// the Console, survive an unrelated apply and be removed when the configuration drops it. What the
// cluster holds is checked against the API, not only against the state the provider writes.
func TestAcc_ClusterKarpenterNodePoolSettings(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-node-pool-settings"
	var clusterID string
	everySetting := func(description string) string {
		return testAccClusterKarpenterSpotConfig(testName, description, testAccKarpenterDisk,
			testAccStableSettings, testAccDefaultSettings, testAccCronjobSettings, testAccGpuSettings)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccSpotClusterAddress),
		Steps: []resource.TestStep{
			// 1. Every setting reaches the API.
			{
				Config: everySetting(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccSpotClusterAddress),
					testAccCaptureResourceID(testAccSpotClusterAddress, &clusterID),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.disk_iops", "4000"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.stable_override.consolidate_after", "1h"),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "features.karpenter.qovery_node_pools.cronjob_override.limits.max_cpu_in_vcpu", "8"),
					testAccCheckClusterKarpenterSettings(testAccKarpenterSettingsOfConfig()),
					resource.TestCheckResourceAttr(testAccSpotClusterAddress, "state", "READY"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. An unrelated apply keeps every setting, and the data source reports them.
			{
				Config: everySetting("unrelated change") + testAccGpuClusterDataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterKarpenterSettings(testAccKarpenterSettingsOfConfig()),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.disk_iops", "4000"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.disk_throughput", "250"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.stable_override.consolidate_after", "1h"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.default_override.consolidate_after", "2m"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.cronjob_override.consolidate_after", "30s"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.cronjob_override.consolidation.start_time", "PT02:00"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.cronjob_override.limits.max_memory_in_gibibytes", "16"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "features.karpenter.qovery_node_pools.gpu_override.consolidate_after", "10m"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 3. Import recovers every setting.
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
			// 4. Change every setting from the Console: the plan shows each change being reverted.
			{
				Config: everySetting("unrelated change"),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						testAccSetKarpenterSettings(parameters, testAccKarpenterSettingsFromConsole())
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectPlannedChange(tfjsonpath.New("features").AtMapKey("karpenter").AtMapKey("disk_iops"), 5000, 4000),
						testAccExpectPlannedChange(tfjsonpath.New("features").AtMapKey("karpenter").AtMapKey("disk_throughput"), 300, 250),
						testAccExpectPlannedChange(testAccNodePoolPath("stable_override").AtMapKey("consolidate_after"), "2h", "1h"),
						testAccExpectPlannedChange(testAccNodePoolPath("default_override").AtMapKey("consolidate_after"), "5m", "2m"),
						testAccExpectPlannedChange(testAccNodePoolPath("cronjob_override").AtMapKey("consolidate_after"), "45s", "30s"),
						testAccExpectPlannedChange(testAccNodePoolPath("cronjob_override").AtMapKey("consolidation").AtMapKey("start_time"), "PT03:00", "PT02:00"),
						testAccExpectPlannedChange(testAccNodePoolPath("cronjob_override").AtMapKey("limits").AtMapKey("max_cpu_in_vcpu"), 10, 8),
						testAccExpectPlannedChange(testAccNodePoolPath("gpu_override").AtMapKey("consolidate_after"), "20m", "10m"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 5. The corrective apply restores the configured settings.
			{
				Config:           everySetting("unrelated change"),
				Check:            testAccCheckClusterKarpenterSettings(testAccKarpenterSettingsOfConfig()),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 6. q-core validates the cronjob limits, and its message reaches the user.
			{
				Config:      testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccCronjobLimitsTooLow),
				ExpectError: regexp.MustCompile(`Wrong cpu value\s+for\s+Cronjob\s+node\s+pool`),
			},
			// 7. A consolidate_after q-core would return in another unit fails at plan time.
			{
				Config:      testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccStableConsolidateAfterNotCanonical),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Write\s+"1h"\s+instead`),
			},
			// 8. Removing every setting plans its removal and clears it on the API.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccCronjobEmpty),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectPlannedChange(tfjsonpath.New("features").AtMapKey("karpenter").AtMapKey("disk_iops"), 4000, nil),
						testAccExpectPlannedChange(testAccNodePoolPath("cronjob_override").AtMapKey("consolidate_after"), "30s", nil),
						testAccExpectPlannedChange(testAccNodePoolPath("cronjob_override").AtMapKey("limits").AtMapKey("max_cpu_in_vcpu"), 8, nil),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "stable_override", true, false),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "default_override", true, false),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckClusterKarpenterSettings(testAccKarpenterSettings{}),
			},
			// 9. A consolidate_after and a disk IOPS set from the Console on a cluster whose
			// configuration declares neither: the stable_override block the pool now needs and
			// disk_iops show in the plan, being removed.
			{
				Config: testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccCronjobEmpty),
				Check: func(_ *terraform.State) error {
					return testAccSetClusterKarpenterOutOfBand(clusterID, func(parameters *qovery.ClusterFeatureKarpenterParameters) {
						parameters.DiskIops = new(int32(5000))
						if parameters.QoveryNodePools.StableOverride == nil {
							parameters.QoveryNodePools.StableOverride = &qovery.KarpenterStableNodePoolOverride{}
						}
						parameters.QoveryNodePools.StableOverride.ConsolidateAfter = new("2h")
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccSpotClusterAddress, plancheck.ResourceActionUpdate),
						testAccExpectNodePoolBlock(testAccSpotClusterAddress, "stable_override", true, false),
						testAccExpectPlannedChange(tfjsonpath.New("features").AtMapKey("karpenter").AtMapKey("disk_iops"), 5000, nil),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// 10. The corrective apply clears them.
			{
				Config:           testAccClusterKarpenterSpotConfig(testName, "unrelated change", testAccNoKarpenterExtra, testAccCronjobEmpty),
				Check:            testAccCheckClusterKarpenterSettings(testAccKarpenterSettings{}),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// testAccSetKarpenterSettings writes settings into Karpenter parameters read from the API, the way
// the Console edits them.
func testAccSetKarpenterSettings(parameters *qovery.ClusterFeatureKarpenterParameters, settings testAccKarpenterSettings) {
	nodePools := &parameters.QoveryNodePools
	parameters.DiskIops = settings.DiskIops
	parameters.DiskThroughput = settings.DiskThroughput
	if nodePools.StableOverride == nil {
		nodePools.StableOverride = &qovery.KarpenterStableNodePoolOverride{}
	}
	nodePools.StableOverride.ConsolidateAfter = settings.StableConsolidateAfter
	if nodePools.DefaultOverride == nil {
		nodePools.DefaultOverride = &qovery.KarpenterDefaultNodePoolOverride{}
	}
	nodePools.DefaultOverride.ConsolidateAfter = settings.DefaultConsolidateAfter
	nodePools.CronjobOverride.ConsolidateAfter = settings.CronjobConsolidateAfter
	nodePools.CronjobOverride.Consolidation = settings.CronjobConsolidation
	nodePools.CronjobOverride.Limits = settings.CronjobLimits
	nodePools.GpuOverride.ConsolidateAfter = settings.GpuConsolidateAfter
}

// testAccCheckClusterKarpenterSettings asserts the settings under test according to the API,
// independently of what the provider stores in state. A zero value for a setting asserts the API
// holds none. The GPU node pool is only checked while it exists.
func testAccCheckClusterKarpenterSettings(want testAccKarpenterSettings) resource.TestCheckFunc {
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

		nodePools := parameters.QoveryNodePools
		got := testAccKarpenterSettings{
			DiskIops:       parameters.DiskIops,
			DiskThroughput: parameters.DiskThroughput,
		}
		if nodePools.StableOverride != nil {
			got.StableConsolidateAfter = nodePools.StableOverride.ConsolidateAfter
		}
		if nodePools.DefaultOverride != nil {
			got.DefaultConsolidateAfter = nodePools.DefaultOverride.ConsolidateAfter
		}
		if nodePools.CronjobOverride != nil {
			got.CronjobConsolidateAfter = nodePools.CronjobOverride.ConsolidateAfter
			got.CronjobConsolidation = nodePools.CronjobOverride.Consolidation
			got.CronjobLimits = nodePools.CronjobOverride.Limits
			// AdditionalProperties holds whatever the client does not model; it is not compared.
			if got.CronjobConsolidation != nil {
				got.CronjobConsolidation.AdditionalProperties = nil
			}
			if got.CronjobLimits != nil {
				got.CronjobLimits.AdditionalProperties = nil
			}
		}
		if nodePools.GpuOverride != nil {
			got.GpuConsolidateAfter = nodePools.GpuOverride.ConsolidateAfter
		}

		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("%s: Karpenter settings on the API are %s, want %s", resourceName, testAccFormatKarpenterSettings(got), testAccFormatKarpenterSettings(want))
		}
		return nil
	}
}

func testAccFormatKarpenterSettings(s testAccKarpenterSettings) string {
	deref := func(v any) any {
		value := reflect.ValueOf(v)
		if value.IsNil() {
			return nil
		}
		return value.Elem().Interface()
	}
	return fmt.Sprintf("{disk_iops:%v disk_throughput:%v stable:%v default:%v cronjob:%v cronjob_consolidation:%+v cronjob_limits:%+v gpu:%v}",
		deref(s.DiskIops), deref(s.DiskThroughput), deref(s.StableConsolidateAfter), deref(s.DefaultConsolidateAfter),
		deref(s.CronjobConsolidateAfter), deref(s.CronjobConsolidation), deref(s.CronjobLimits), deref(s.GpuConsolidateAfter))
}

// testAccExpectPlannedChange is a plan check asserting the value of an attribute of the test
// cluster before and after the planned change. A nil value asserts the attribute is null or absent.
func testAccExpectPlannedChange(attributePath tfjsonpath.Path, before, after any) plancheck.PlanCheck {
	return plannedChangeCheck{address: testAccSpotClusterAddress, path: attributePath, before: before, after: after}
}

type plannedChangeCheck struct {
	address string
	path    tfjsonpath.Path
	before  any
	after   any
}

func (c plannedChangeCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address != c.address || change.Change == nil {
			continue
		}
		// The plan JSON decodes numbers as json.Number or float64: compare the printed values.
		value := func(object any) string {
			v, err := tfjsonpath.Traverse(object, c.path)
			if err != nil || v == nil {
				return "<nil>"
			}
			return fmt.Sprint(v)
		}
		want := func(v any) string {
			if v == nil {
				return "<nil>"
			}
			return fmt.Sprint(v)
		}
		if got := value(change.Change.Before); got != want(c.before) {
			resp.Error = fmt.Errorf("%s: %s is %s before the change, want %s", c.address, c.path, got, want(c.before))
			return
		}
		if got := value(change.Change.After); got != want(c.after) {
			resp.Error = fmt.Errorf("%s: %s is planned as %s, want %s", c.address, c.path, got, want(c.after))
		}
		return
	}
	resp.Error = fmt.Errorf("%s: no planned change", c.address)
}
