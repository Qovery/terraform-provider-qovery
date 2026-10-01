//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

const (
	testAccOwnershipClusterAddress     = "qovery_cluster.test"
	testAccOwnershipLabelsGroupAddress = "qovery_labels_group.test"
	testAccOwnershipDataSourceAddress  = "data.qovery_cluster.test"

	testAccDeclaredRouteHCL       = `[{ description = "vpn", destination = "172.30.0.0/16", target = "vgw-0123456789abcdef0" }]`
	testAccDeclaredLabelsGroupHCL = `[qovery_labels_group.test.id]`
)

// Karpenter makes the API return unstable node counts, so import cannot match them.
var testAccOwnershipImportIgnore = []string{"advanced_settings_json", "min_running_nodes", "max_running_nodes"}

var testAccDeclaredRoute = qovery.ClusterRoutingTableResultsInner{Description: "vpn", Destination: "172.30.0.0/16", Target: "vgw-0123456789abcdef0"}

// TestAcc_ClusterRoutingTableOwnership covers routing_table as desired state: a route added,
// changed or deleted outside Terraform shows in the plan and the next apply reverts it, and
// removing the attribute from the configuration deletes every remote route. Every step checks
// the routes the API holds, not only the state.
func TestAcc_ClusterRoutingTableOwnership(t *testing.T) {
	t.Parallel()

	testName := "cluster-routing-table-ownership"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccOwnershipClusterAddress),
		Steps: []resource.TestStep{
			// 1. routing_table omitted: stored as null, no route on the cluster.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccOwnershipClusterAddress),
					testAccCaptureResourceID(testAccOwnershipClusterAddress, &clusterID),
					resource.TestCheckNoResourceAttr(testAccOwnershipClusterAddress, "routing_table.#"),
					testAccCheckClusterRoutesInAPI(&clusterID),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. A route added outside Terraform shows in the plan while the attribute is omitted.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "", false),
				Check: testAccSetClusterRoutesOutOfBand(&clusterID, qovery.ClusterRoutingTableResultsInner{
					Description: "peering", Destination: "172.31.0.0/16", Target: "pcx-0123456789abcdef0",
				}),
				ConfigPlanChecks:   testAccExpectRoutingTablePlanned(knownvalue.Null()),
				ExpectNonEmptyPlan: true,
			},
			// 3. The corrective apply deletes it.
			{
				Config:           testAccClusterOwnershipConfig(testName, "", "", false),
				Check:            testAccCheckClusterRoutesInAPI(&clusterID),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Declare one route.
			{
				Config: testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipClusterAddress, "routing_table.#", "1"),
					testAccCheckClusterRoutesInAPI(&clusterID, testAccDeclaredRoute),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 5. The route's target changed outside Terraform shows in the plan.
			{
				Config: testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", false),
				Check: testAccSetClusterRoutesOutOfBand(&clusterID, qovery.ClusterRoutingTableResultsInner{
					Description: "vpn", Destination: "172.30.0.0/16", Target: "vgw-0fedcba9876543210",
				}),
				ConfigPlanChecks:   testAccExpectRoutingTablePlanned(knownvalue.SetSizeExact(1)),
				ExpectNonEmptyPlan: true,
			},
			// 6. The corrective apply restores the declared target.
			{
				Config:           testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", false),
				Check:            testAccCheckClusterRoutesInAPI(&clusterID, testAccDeclaredRoute),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 7. The route deleted outside Terraform shows in the plan.
			{
				Config:             testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", false),
				Check:              testAccSetClusterRoutesOutOfBand(&clusterID),
				ConfigPlanChecks:   testAccExpectRoutingTablePlanned(knownvalue.SetSizeExact(1)),
				ExpectNonEmptyPlan: true,
			},
			// 8. The corrective apply recreates it.
			{
				Config:           testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", false),
				Check:            testAccCheckClusterRoutesInAPI(&clusterID, testAccDeclaredRoute),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 9. Import records the remote route.
			{
				ResourceName:            testAccOwnershipClusterAddress,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: testAccOwnershipImportIgnore,
			},
			// 10. The data source reports the route.
			{
				Config: testAccClusterOwnershipConfig(testName, testAccDeclaredRouteHCL, "", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipDataSourceAddress, "routing_table.#", "1"),
					resource.TestCheckResourceAttr(testAccOwnershipDataSourceAddress, "routing_table.0.target", testAccDeclaredRoute.Target),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 11. Removing the attribute plans the route's deletion and the apply deletes it.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccOwnershipClusterAddress, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(testAccOwnershipClusterAddress, tfjsonpath.New("routing_table"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccOwnershipClusterAddress, "routing_table.#"),
					testAccCheckClusterRoutesInAPI(&clusterID),
				),
			},
			// 12. An explicit empty routing table keeps the cluster without routes and plans clean.
			{
				Config: testAccClusterOwnershipConfig(testName, "[]", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipClusterAddress, "routing_table.#", "0"),
					testAccCheckClusterRoutesInAPI(&clusterID),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterLabelsGroupIdsOwnership covers labels_group_ids as desired state: a labels group
// attached or detached outside Terraform shows in the plan and the next apply reverts it, and
// removing the attribute from the configuration detaches every labels group.
func TestAcc_ClusterLabelsGroupIdsOwnership(t *testing.T) {
	t.Parallel()

	testName := "cluster-labels-group-ownership"
	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccOwnershipClusterAddress),
		Steps: []resource.TestStep{
			// 1. labels_group_ids omitted: stored as null, nothing attached.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccOwnershipClusterAddress),
					testAccCaptureResourceID(testAccOwnershipClusterAddress, &clusterID),
					resource.TestCheckNoResourceAttr(testAccOwnershipClusterAddress, "labels_group_ids.#"),
					testAccCheckClusterLabelsGroupsInAPI(&clusterID, false),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. A labels group attached outside Terraform shows in the plan while the attribute is omitted.
			{
				Config:             testAccClusterOwnershipConfig(testName, "", "", false),
				Check:              testAccSetClusterLabelsGroupsOutOfBand(&clusterID, true),
				ConfigPlanChecks:   testAccExpectLabelsGroupIdsPlanned(knownvalue.Null()),
				ExpectNonEmptyPlan: true,
			},
			// 3. The corrective apply detaches it.
			{
				Config:           testAccClusterOwnershipConfig(testName, "", "", false),
				Check:            testAccCheckClusterLabelsGroupsInAPI(&clusterID, false),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 4. Declare the labels group.
			{
				Config: testAccClusterOwnershipConfig(testName, "", testAccDeclaredLabelsGroupHCL, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipClusterAddress, "labels_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(testAccOwnershipClusterAddress, "labels_group_ids.*", testAccOwnershipLabelsGroupAddress, "id"),
					testAccCheckClusterLabelsGroupsInAPI(&clusterID, true),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 5. The labels group detached outside Terraform shows in the plan.
			{
				Config:             testAccClusterOwnershipConfig(testName, "", testAccDeclaredLabelsGroupHCL, false),
				Check:              testAccSetClusterLabelsGroupsOutOfBand(&clusterID, false),
				ConfigPlanChecks:   testAccExpectLabelsGroupIdsPlanned(knownvalue.SetSizeExact(1)),
				ExpectNonEmptyPlan: true,
			},
			// 6. The corrective apply attaches it again.
			{
				Config:           testAccClusterOwnershipConfig(testName, "", testAccDeclaredLabelsGroupHCL, false),
				Check:            testAccCheckClusterLabelsGroupsInAPI(&clusterID, true),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 7. Import records the attached labels group.
			{
				ResourceName:            testAccOwnershipClusterAddress,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: testAccOwnershipImportIgnore,
			},
			// 8. The data source reports the labels group.
			{
				Config: testAccClusterOwnershipConfig(testName, "", testAccDeclaredLabelsGroupHCL, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipDataSourceAddress, "labels_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(testAccOwnershipDataSourceAddress, "labels_group_ids.*", testAccOwnershipLabelsGroupAddress, "id"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 9. Removing the attribute plans the detach and the apply detaches it.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccOwnershipClusterAddress, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(testAccOwnershipClusterAddress, tfjsonpath.New("labels_group_ids"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccOwnershipClusterAddress, "labels_group_ids.#"),
					testAccCheckClusterLabelsGroupsInAPI(&clusterID, false),
				),
			},
			// 10. An explicit empty list keeps the cluster without labels groups and plans clean.
			{
				Config: testAccClusterOwnershipConfig(testName, "", "[]", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccOwnershipClusterAddress, "labels_group_ids.#", "0"),
					testAccCheckClusterLabelsGroupsInAPI(&clusterID, false),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterRoutesLabelsUpgradeFrom0x upgrades from the last 0.x release a cluster that
// declares neither routing_table nor labels_group_ids. 0.x stored routing_table as [] for such a
// cluster; the state upgrade turns it into null, so the first 1.0 plan is empty.
func TestAcc_ClusterRoutesLabelsUpgradeFrom0x(t *testing.T) {
	t.Parallel()

	testName := "cluster-routes-labels-upgrade"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccQoveryClusterDestroy(testAccOwnershipClusterAddress),
		Steps: []resource.TestStep{
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            testAccClusterAWSReadyConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists(testAccOwnershipClusterAddress),
					resource.TestCheckResourceAttr(testAccOwnershipClusterAddress, "routing_table.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccClusterAWSReadyConfig(testName),
				PlanOnly:                 true,
			},
		},
	})
}

// --- configuration -------------------------------------------------------------------------------

// testAccClusterOwnershipConfig renders an AWS cluster in READY state (no cloud infrastructure is
// provisioned) next to a labels group. routingTable and labelsGroupIDs are HCL values for the
// matching attributes; an empty string leaves the attribute out.
func testAccClusterOwnershipConfig(testName, routingTable, labelsGroupIDs string, withDataSource bool) string {
	var attributes strings.Builder
	if routingTable != "" {
		attributes.WriteString("\n  routing_table = " + routingTable)
	}
	if labelsGroupIDs != "" {
		attributes.WriteString("\n  labels_group_ids = " + labelsGroupIDs)
	}

	dataSource := ""
	if withDataSource {
		dataSource = fmt.Sprintf(`
data "qovery_cluster" "test" {
  id              = qovery_cluster.test.id
  organization_id = "%s"
}
`, getTestOrganizationID())
	}

	return fmt.Sprintf(`
resource "qovery_labels_group" "test" {
  organization_id = "%s"
  name            = "%s-lg"
  labels = [{ key = "team", value = "platform", propagate_to_cloud_provider = true }]
}

resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"%s

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          { key = "InstanceSize",   operator = "In", values = ["small", "medium", "large"] },
          { key = "InstanceFamily", operator = "In", values = ["t3a"] },
          { key = "Arch",           operator = "In", values = ["AMD64"] },
        ]
      }
    }
  }
}
%s`, getTestOrganizationID(), generateTestName(testName),
		getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), attributes.String(),
		dataSource)
}

// --- plan checks ---------------------------------------------------------------------------------

// testAccExpectRoutingTablePlanned asserts that the refreshed plan updates the cluster back to the
// configured routing_table after an out-of-band change.
func testAccExpectRoutingTablePlanned(planned knownvalue.Check) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(testAccOwnershipClusterAddress, plancheck.ResourceActionUpdate),
			plancheck.ExpectKnownValue(testAccOwnershipClusterAddress, tfjsonpath.New("routing_table"), planned),
		},
	}
}

// testAccExpectLabelsGroupIdsPlanned asserts that the refreshed plan updates the cluster back to
// the configured labels_group_ids after an out-of-band change.
func testAccExpectLabelsGroupIdsPlanned(planned knownvalue.Check) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(testAccOwnershipClusterAddress, plancheck.ResourceActionUpdate),
			plancheck.ExpectKnownValue(testAccOwnershipClusterAddress, tfjsonpath.New("labels_group_ids"), planned),
		},
	}
}

// --- API checks and out-of-band changes ----------------------------------------------------------

// testAccCheckClusterRoutesInAPI checks that the cluster's routing table in the API holds exactly
// the expected routes.
func testAccCheckClusterRoutesInAPI(clusterID *string, expected ...qovery.ClusterRoutingTableResultsInner) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		table, res, err := qoveryAPIClient.ClustersAPI.GetRoutingTable(context.TODO(), getTestOrganizationID(), *clusterID).Execute()
		if err != nil || (res != nil && res.StatusCode >= 400) {
			return fmt.Errorf("failed to read routing table of cluster %s: %v", *clusterID, err)
		}
		got := routeKeys(table.GetResults())
		want := routeKeys(expected)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("cluster %s routes in API: got %v, want %v", *clusterID, got, want)
		}
		return nil
	}
}

func routeKeys(routes []qovery.ClusterRoutingTableResultsInner) []string {
	keys := make([]string, 0, len(routes))
	for _, route := range routes {
		keys = append(keys, route.Description+","+route.Destination+","+route.Target)
	}
	sort.Strings(keys)
	return keys
}

// testAccSetClusterRoutesOutOfBand replaces the cluster's routing table through the API, bypassing
// Terraform, the way the Qovery Console does. No route deletes them all.
func testAccSetClusterRoutesOutOfBand(clusterID *string, routes ...qovery.ClusterRoutingTableResultsInner) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if routes == nil {
			routes = []qovery.ClusterRoutingTableResultsInner{}
		}
		_, res, err := qoveryAPIClient.ClustersAPI.EditRoutingTable(context.TODO(), getTestOrganizationID(), *clusterID).
			ClusterRoutingTableRequest(qovery.ClusterRoutingTableRequest{Routes: routes}).
			Execute()
		if err != nil || (res != nil && res.StatusCode >= 400) {
			return fmt.Errorf("failed to edit routing table of cluster %s out of band: %v", *clusterID, err)
		}
		return nil
	}
}

// testAccCheckClusterLabelsGroupsInAPI checks whether the labels group of the configuration is the
// only one attached to the cluster in the API (attached) or none is (not attached).
func testAccCheckClusterLabelsGroupsInAPI(clusterID *string, attached bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testAccReadClusterFromAPI(*clusterID)
		if err != nil {
			return err
		}
		got := make([]string, 0, len(cluster.LabelsGroups))
		for _, lg := range cluster.LabelsGroups {
			got = append(got, lg.GetId())
		}

		want := []string{}
		if attached {
			labelsGroupID, err := testAccLabelsGroupIDFromState(s)
			if err != nil {
				return err
			}
			want = append(want, labelsGroupID)
		}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("cluster %s labels groups in API: got %v, want %v", *clusterID, got, want)
		}
		return nil
	}
}

// testAccSetClusterLabelsGroupsOutOfBand attaches the configuration's labels group to the cluster
// (attach) or detaches every labels group (detach) through the API, bypassing Terraform.
func testAccSetClusterLabelsGroupsOutOfBand(clusterID *string, attach bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		labelsGroups := []qovery.ClusterLabelsGroup{}
		if attach {
			labelsGroupID, err := testAccLabelsGroupIDFromState(s)
			if err != nil {
				return err
			}
			labelsGroups = append(labelsGroups, qovery.ClusterLabelsGroup{Id: &labelsGroupID})
		}
		return testAccEditClusterOutOfBand(*clusterID, nil, func(request *qovery.ClusterRequest) {
			request.LabelsGroups = labelsGroups
		})
	}
}

func testAccLabelsGroupIDFromState(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccOwnershipLabelsGroupAddress]
	if !ok || rs.Primary.ID == "" {
		return "", fmt.Errorf("%s: id not found in state", testAccOwnershipLabelsGroupAddress)
	}
	return rs.Primary.ID, nil
}
