//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"
)

const testAccContractClusterAddress = "qovery_cluster.test"

// testAccClusterNodeSizing is the node sizing a cluster holds in the API.
type testAccClusterNodeSizing struct {
	InstanceType    string
	DiskSize        int32
	MinRunningNodes int32
	MaxRunningNodes int32
}

// The Qovery defaults of a Scaleway cluster.
var testAccScalewayNodeSizingDefaults = testAccClusterNodeSizing{InstanceType: "DEV1-L", DiskSize: 40, MinRunningNodes: 3, MaxRunningNodes: 10}

// TestAcc_ClusterNodeSizingDefaults covers instance_type, disk_size, min_running_nodes and
// max_running_nodes on a cluster whose node group Qovery sizes (Scaleway): removing them plans the
// Qovery default, a value changed outside Terraform shows up in the plan and the next apply
// reverts it, and import records the remote values. Every step checks what the API holds.
func TestAcc_ClusterNodeSizingDefaults(t *testing.T) {
	t.Parallel()

	testName := "cluster-node-sizing-defaults"
	custom := testAccClusterNodeSizing{InstanceType: "DEV1-XL", DiskSize: 50, MinRunningNodes: 4, MaxRunningNodes: 6}
	customHCL := fmt.Sprintf("instance_type = %q\n  disk_size = %d\n  min_running_nodes = %d\n  max_running_nodes = %d",
		custom.InstanceType, custom.DiskSize, custom.MinRunningNodes, custom.MaxRunningNodes)

	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccContractClusterAddress),
		Steps: []resource.TestStep{
			// 1. Create with explicit non-default values.
			{
				Config: testAccClusterScalewayContractConfig(testName, customHCL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureResourceID(testAccContractClusterAddress, &clusterID),
					resource.TestCheckResourceAttr(testAccContractClusterAddress, "instance_type", custom.InstanceType),
					resource.TestCheckResourceAttr(testAccContractClusterAddress, "disk_size", "50"),
					testAccCheckClusterNodeSizingInAPI(&clusterID, custom),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Removing the four attributes plans the Qovery defaults, and the apply sets them.
			{
				Config: testAccClusterScalewayContractConfig(testName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: append([]plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccContractClusterAddress, plancheck.ResourceActionUpdate),
					}, testAccExpectNodeSizingPlanned(testAccScalewayNodeSizingDefaults)...),
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckClusterNodeSizingInAPI(&clusterID, testAccScalewayNodeSizingDefaults),
			},
			// 3. Values changed outside Terraform show up in the plan while the attributes are omitted.
			{
				Config: testAccClusterScalewayContractConfig(testName, ""),
				Check: func(_ *terraform.State) error {
					return testAccEditClusterOutOfBand(clusterID, nil, func(request *qovery.ClusterRequest) {
						request.DiskSize = new(int32(60))
						request.MinRunningNodes = new(int32(5))
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: append([]plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccContractClusterAddress, plancheck.ResourceActionUpdate),
					}, testAccExpectNodeSizingPlanned(testAccScalewayNodeSizingDefaults)...),
				},
				ExpectNonEmptyPlan: true,
			},
			// 4. Import records the remote values.
			{
				ResourceName:        testAccContractClusterAddress,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateCheck:    testAccImportStateCheckAttributes(map[string]string{"disk_size": "60", "min_running_nodes": "5"}),
			},
			// 5. The corrective apply reverts them.
			{
				Config:           testAccClusterScalewayContractConfig(testName, ""),
				Check:            testAccCheckClusterNodeSizingInAPI(&clusterID, testAccScalewayNodeSizingDefaults),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 6. Import matches the applied state.
			{
				ResourceName:            testAccContractClusterAddress,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
		},
	})
}

// TestAcc_ClusterFeaturesAndKedaDefaults covers features and keda as desired state on a Scaleway
// cluster: removing the blocks plans their defaults, a static IP or KEDA enabled outside
// Terraform shows up in the plan and the next apply disables it, and import records the remote
// values. It also creates a Scaleway cluster with a features block, which the Qovery API rejected
// while the provider sent it the VPC subnet feature.
func TestAcc_ClusterFeaturesAndKedaDefaults(t *testing.T) {
	t.Parallel()

	testName := "cluster-features-keda-defaults"
	declaredHCL := "features = { static_ip = true }\n  keda = { enabled = true }"

	var clusterID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccContractClusterAddress),
		Steps: []resource.TestStep{
			// 1. Create with a static IP and KEDA.
			{
				Config: testAccClusterScalewayContractConfig(testName, declaredHCL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureResourceID(testAccContractClusterAddress, &clusterID),
					resource.TestCheckResourceAttr(testAccContractClusterAddress, "features.static_ip", "true"),
					resource.TestCheckResourceAttr(testAccContractClusterAddress, "keda.enabled", "true"),
					testAccCheckClusterStaticIPAndKedaInAPI(&clusterID, true, true),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// 2. Removing both blocks plans their defaults, and the apply disables both.
			{
				Config: testAccClusterScalewayContractConfig(testName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: append([]plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccContractClusterAddress, plancheck.ResourceActionUpdate),
					}, testAccExpectStaticIPAndKedaPlanned(false, false)...),
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckClusterStaticIPAndKedaInAPI(&clusterID, false, false),
			},
			// 3. A static IP and KEDA enabled outside Terraform show up in the plan.
			{
				Config: testAccClusterScalewayContractConfig(testName, ""),
				Check: func(_ *terraform.State) error {
					return testAccEditClusterOutOfBand(clusterID, nil, func(request *qovery.ClusterRequest) {
						request.Keda = qovery.NewClusterKeda(true)
						request.Features = testAccWithStaticIPFeature(request.Features, true)
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: append([]plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccContractClusterAddress, plancheck.ResourceActionUpdate),
					}, testAccExpectStaticIPAndKedaPlanned(false, false)...),
				},
				ExpectNonEmptyPlan: true,
			},
			// 4. Import records them.
			{
				ResourceName:        testAccContractClusterAddress,
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateCheck:    testAccImportStateCheckAttributes(map[string]string{"features.static_ip": "true", "keda.enabled": "true"}),
			},
			// 5. The corrective apply disables them.
			{
				Config:           testAccClusterScalewayContractConfig(testName, ""),
				Check:            testAccCheckClusterStaticIPAndKedaInAPI(&clusterID, false, false),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
		},
	})
}

// TestAcc_ClusterKarpenterCannotBeDisabled covers the plan error on a configuration that drops
// the Karpenter feature, which the Qovery API cannot disable, and the data source reporting the
// sizing values the API derives for a Karpenter cluster.
func TestAcc_ClusterKarpenterCannotBeDisabled(t *testing.T) {
	t.Parallel()

	testName := "cluster-karpenter-cannot-be-disabled"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccContractClusterAddress),
		Steps: []resource.TestStep{
			{
				Config:           testAccClusterAWSReadyConfig(testName),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// The data source reports the API values, sentinels included.
			{
				Config: testAccClusterAWSReadyConfig(testName) + testAccContractClusterDataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "instance_type", "KARPENTER"),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "max_running_nodes", "2147483647"),
					resource.TestCheckResourceAttrPair("data.qovery_cluster.test", "disk_size", testAccContractClusterAddress, "disk_size"),
				),
			},
			// Omitting features plans no Karpenter, which fails the plan instead of the apply.
			{
				Config:      testAccClusterAWSWithoutFeaturesConfig(testName),
				ExpectError: regexp.MustCompile(`Cannot\s+disable\s+Karpenter`),
			},
		},
	})
}

// TestAcc_ClusterGkeKmsKeyCannotChange covers the plan error on a GKE KMS key added after the
// cluster was created: the Qovery API only takes the key on create.
func TestAcc_ClusterGkeKmsKeyCannotChange(t *testing.T) {
	t.Parallel()

	testName := "cluster-gke-kms-key-cannot-change"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy(testAccContractClusterAddress),
		Steps: []resource.TestStep{
			{
				Config:           testAccClusterGCPReadyConfig(testName),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			{
				Config: testAccClusterFeaturesConfig(testName, getTestGCPCredentialsID(), "GCP", "europe-west9", "AUTO_PILOT",
					`gke_kms_key = "projects/qovery/locations/europe-west9/keyRings/ring/cryptoKeys/key"`),
				ExpectError: regexp.MustCompile(`Cannot\s+change\s+features.gke_kms_key\s+after\s+creation`),
			},
		},
	})
}

// TestAcc_ClusterContractUpgradeFrom0x upgrades from the last 0.x release clusters that omit
// every node sizing attribute, features and keda: the defaults 1.0 plans are the values the
// Qovery API stored for them, so the first 1.0 plan is empty.
func TestAcc_ClusterContractUpgradeFrom0x(t *testing.T) {
	testCases := []struct {
		name   string
		config func(string) string
	}{
		{name: "scw", config: func(testName string) string { return testAccClusterScalewayContractConfig(testName, "") }},
		{name: "azure", config: testAccClusterAzureMinimalConfig},
		{name: "gcp", config: testAccClusterGCPReadyConfig},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testName := "cluster-contract-upgrade-" + tc.name
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				CheckDestroy: testAccQoveryClusterDestroy(testAccContractClusterAddress),
				Steps: []resource.TestStep{
					{
						ExternalProviders: testAccLastProvider0xFromRegistry,
						Config:            tc.config(testName),
						Check:             testAccQoveryClusterExists(testAccContractClusterAddress),
					},
					{
						ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
						Config:                   tc.config(testName),
						PlanOnly:                 true,
					},
				},
			})
		})
	}
}

// --- configuration -------------------------------------------------------------------------------

const testAccContractClusterDataSource = `
data "qovery_cluster" "test" {
  id              = qovery_cluster.test.id
  organization_id = qovery_cluster.test.organization_id
}
`

// testAccClusterScalewayContractConfig renders a Scaleway cluster in READY state (no cloud
// infrastructure is provisioned) with the given extra HCL attributes.
func testAccClusterScalewayContractConfig(testName, attributes string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "SCW"
  region          = "pl-waw-1"
  kubernetes_mode = "MANAGED"
  state           = "READY"
  %s
}
`, getTestScalewayCredentialsID(), getTestOrganizationID(), generateTestName(testName), attributes)
}

func testAccClusterAzureMinimalConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AZURE"
  region          = "francecentral"
  kubernetes_mode = "MANAGED"
  state           = "READY"
}
`, getTestAzureCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

// testAccClusterAWSWithoutFeaturesConfig is testAccClusterAWSReadyConfig without its features
// block, i.e. without Karpenter.
func testAccClusterAWSWithoutFeaturesConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

// --- plan checks ---------------------------------------------------------------------------------

func testAccExpectNodeSizingPlanned(sizing testAccClusterNodeSizing) []plancheck.PlanCheck {
	return []plancheck.PlanCheck{
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("instance_type"), knownvalue.StringExact(sizing.InstanceType)),
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("disk_size"), knownvalue.Int64Exact(int64(sizing.DiskSize))),
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("min_running_nodes"), knownvalue.Int64Exact(int64(sizing.MinRunningNodes))),
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("max_running_nodes"), knownvalue.Int64Exact(int64(sizing.MaxRunningNodes))),
	}
}

func testAccExpectStaticIPAndKedaPlanned(staticIP, keda bool) []plancheck.PlanCheck {
	return []plancheck.PlanCheck{
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("features").AtMapKey("static_ip"), knownvalue.Bool(staticIP)),
		plancheck.ExpectKnownValue(testAccContractClusterAddress, tfjsonpath.New("keda").AtMapKey("enabled"), knownvalue.Bool(keda)),
	}
}

// --- API checks and out-of-band changes ----------------------------------------------------------

func testAccCheckClusterNodeSizingInAPI(clusterID *string, expected testAccClusterNodeSizing) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		cluster, err := testAccReadClusterFromAPI(*clusterID)
		if err != nil {
			return err
		}
		got := testAccClusterNodeSizing{
			InstanceType:    cluster.GetInstanceType(),
			DiskSize:        cluster.GetDiskSize(),
			MinRunningNodes: cluster.GetMinRunningNodes(),
			MaxRunningNodes: cluster.GetMaxRunningNodes(),
		}
		if got != expected {
			return fmt.Errorf("cluster %s node sizing in API: got %+v, want %+v", *clusterID, got, expected)
		}
		return nil
	}
}

func testAccCheckClusterStaticIPAndKedaInAPI(clusterID *string, staticIP, keda bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		cluster, err := testAccReadClusterFromAPI(*clusterID)
		if err != nil {
			return err
		}
		gotStaticIP := false
		for _, f := range cluster.Features {
			if value := f.GetValueObject().ClusterFeatureBooleanResponse; f.GetId() == "STATIC_IP" && value != nil {
				gotStaticIP = value.Value
			}
		}
		gotKeda := cluster.Keda != nil && cluster.Keda.GetEnabled()
		if gotStaticIP != staticIP || gotKeda != keda {
			return fmt.Errorf("cluster %s in API: static_ip %t and keda %t, want %t and %t", *clusterID, gotStaticIP, gotKeda, staticIP, keda)
		}
		return nil
	}
}

// testAccWithStaticIPFeature returns the request features with STATIC_IP set to enabled.
func testAccWithStaticIPFeature(features []qovery.ClusterRequestFeaturesInner, enabled bool) []qovery.ClusterRequestFeaturesInner {
	updated := make([]qovery.ClusterRequestFeaturesInner, 0, len(features)+1)
	for _, f := range features {
		if f.GetId() != "STATIC_IP" {
			updated = append(updated, f)
		}
	}
	return append(updated, qovery.ClusterRequestFeaturesInner{
		Id:    new("STATIC_IP"),
		Value: *qovery.NewNullableClusterRequestFeaturesInnerValue(&qovery.ClusterRequestFeaturesInnerValue{Bool: &enabled}),
	})
}

// testAccImportStateCheckAttributes checks attributes of the imported state.
func testAccImportStateCheckAttributes(expected map[string]string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("imported %d states, want 1", len(states))
		}
		for name, want := range expected {
			if got := states[0].Attributes[name]; got != want {
				return fmt.Errorf("imported %s = %q, want %q", name, got, want)
			}
		}
		return nil
	}
}
