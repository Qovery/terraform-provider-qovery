//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// TestAcc_Cluster is the main AWS EKS lifecycle test using Karpenter and READY state
// (no cloud infra is provisioned). Covers create, update, labels groups, and import.
func TestAcc_Cluster(t *testing.T) {
	t.SkipNow() // AWS Karpenter updates trigger Karpenter IAM validation that fails with test credentials
	t.Parallel()
	testName := "cluster"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccClusterKarpenterConfig(testName, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "region", "eu-west-3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "labels_group_ids"),
				),
			},
			// Add description
			{
				Config: testAccClusterKarpenterConfigWithDescription(testName, "my cluster"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", "my cluster"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Remove description
			{
				Config: testAccClusterKarpenterConfig(testName, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
				),
			},
			// Attach labels group
			{
				Config: testAccClusterKarpenterConfig(testName, "", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "labels_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"qovery_cluster.test", "labels_group_ids.0",
						"qovery_labels_group.test", "id",
					),
				),
			},
			// Detach labels group
			{
				Config: testAccClusterKarpenterConfig(testName, "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "labels_group_ids"),
				),
			},
			// Set advanced_settings_json
			{
				Config: testAccClusterKarpenterConfigWithAdvancedSettings(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttrSet("qovery_cluster.test", "advanced_settings_json"),
				),
			},
			// Import
			{
				ResourceName:            "qovery_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
		},
	})
}

// TestAcc_ClusterWithStaticIP verifies that static_ip feature config is persisted correctly.
// Uses Karpenter (required for new AWS MANAGED) + READY state (no cloud infra provisioned).
func TestAcc_ClusterWithStaticIP(t *testing.T) {
	t.Parallel()
	testName := "cluster-with-static-ip"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterKarpenterConfigWithStaticIP(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			{
				ResourceName:            "qovery_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
		},
	})
}

func TestAcc_ClusterWithKeda(t *testing.T) {
	t.Parallel()
	testName := "cluster-with-keda"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create with KEDA enabled
			{
				Config: testAccClusterKarpenterConfigWithKeda(testName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "keda.enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Plan stability — no diff on re-apply of the same config
			{
				Config:   testAccClusterKarpenterConfigWithKeda(testName, true),
				PlanOnly: true,
			},
			// Update KEDA to disabled
			{
				Config: testAccClusterKarpenterConfigWithKeda(testName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "keda.enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Import
			{
				ResourceName:            "qovery_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
		},
	})
}

// TestAcc_ClusterWithVpcPeering is kept skipped — it requires pre-existing VPC infrastructure.
func TestAcc_ClusterWithVpcPeering(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "cluster-with-vpc-peering"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDefaultConfigWithVpcPeering(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
					"10.42.0.0/16",
					map[string]string{
						"172.30.0.0/16": "target",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.42.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.description", "route-0"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.destination", "172.30.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.target", "target"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
		},
	})
}

// TestAcc_ClusterWithReadyState verifies that clusters can be created in READY state
// (config only, no cloud infrastructure provisioned) across all supported providers.
func TestAcc_ClusterWithReadyState(t *testing.T) {
	testCases := []struct {
		name                    string
		config                  func(string) string
		checks                  []resource.TestCheckFunc
		importStateVerifyIgnore []string
	}{
		{
			name:   "aws_eks",
			config: testAccClusterAWSReadyConfig,
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
				resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
			},
			importStateVerifyIgnore: []string{"advanced_settings_json"},
		},
		{
			name:   "scw",
			config: testAccClusterSCWReadyConfig,
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "SCW"),
				resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
			},
			importStateVerifyIgnore: []string{"advanced_settings_json"},
		},
		{
			name:   "azure",
			config: testAccClusterAzureReadyConfig,
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AZURE"),
				resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
			},
			importStateVerifyIgnore: []string{"advanced_settings_json"},
		},
		{
			name:   "gcp",
			config: testAccClusterGCPReadyConfig,
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "GCP"),
				resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
			},
			// GCP AUTO_PILOT returns sentinel INT_MAX for min/max_running_nodes — ignore on import.
			importStateVerifyIgnore: []string{"advanced_settings_json", "min_running_nodes", "max_running_nodes"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testName := fmt.Sprintf("cluster-ready-%s", tc.name)
			checks := append(
				[]resource.TestCheckFunc{
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
				},
				tc.checks...,
			)
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
				Steps: []resource.TestStep{
					{
						Config: tc.config(testName),
						Check:  resource.ComposeAggregateTestCheckFunc(checks...),
					},
					{
						ResourceName:            "qovery_cluster.test",
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
						ImportStateVerifyIgnore: tc.importStateVerifyIgnore,
					},
				},
			})
		})
	}
}

// TestAcc_ClusterGcpNatGateways verifies the value-based semantics of
// features.nat_gateways against the real API, on a GCP cluster in READY state
// (no cloud infra provisioned). It pins the Terraform-visible invariants
// of the v3 design:
//  1. explicit {static_ips_enabled=true, static_ips_count=3} round-trips through
//     create/Read (both fields visible in state),
//  2. static_ips_count updates in place (3 → 5) on the existing cluster,
//  3. removing the block resets to the default {false,1} with a visible diff —
//     it does NOT silently keep the previous value; post-apply Read matches the
//     planned default (no "inconsistent result" error),
//  4. re-adding the block re-enables the feature (false → true transition on the
//     existing cluster, not just at create),
//  5. static_ip = false disables the feature while nat_gateways stays at the
//     default object {false,1} (never null) in state,
//  6. import verify (same ignores as other GCP tests).
func TestAcc_ClusterGcpNatGateways(t *testing.T) {
	t.Parallel()
	testName := "cluster-gcp-nat-gateways"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Step 1: create with static_ip=true and an explicit enabled=true, count=3.
			{
				Config: testAccClusterGCPNatGatewaysConfig(testName, true, "nat_gateways = { static_ips_enabled = true, static_ips_count = 3 }"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "GCP"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_count", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "READY"),
				),
			},
			// Step 2: update static_ips_count in place (3 → 5) while enabled.
			{
				Config: testAccClusterGCPNatGatewaysConfig(testName, true, "nat_gateways = { static_ips_enabled = true, static_ips_count = 5 }"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_count", "5"),
				),
			},
			// Step 3: remove the block — ObjectDefault resets to {false,1} (visible diff).
			{
				Config: testAccClusterGCPNatGatewaysConfig(testName, true, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_count", "1"),
				),
			},
			// Step 4: re-add the block — re-enable on the existing cluster (false → true).
			{
				Config: testAccClusterGCPNatGatewaysConfig(testName, true, "nat_gateways = { static_ips_enabled = true, static_ips_count = 2 }"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_enabled", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_count", "2"),
				),
			},
			// Step 5: disable static_ip (block removed) — nat_gateways resets to default {false,1}.
			// NOTE: this transition is only accepted because the cluster is in READY
			// state (never deployed). q-core rejects enabling/disabling static_ip on an
			// already DEPLOYED cluster (isStaticIpUpdateForbiddenOnDeployedCluster).
			{
				Config: testAccClusterGCPNatGatewaysConfig(testName, false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_enabled", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.nat_gateways.static_ips_count", "1"),
				),
			},
			// Step 6: import verify.
			{
				ResourceName:        "qovery_cluster.test",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
				// GCP AUTO_PILOT returns sentinel INT_MAX for min/max_running_nodes — ignore on import.
				ImportStateVerifyIgnore: []string{"advanced_settings_json", "min_running_nodes", "max_running_nodes"},
			},
		},
	})
}

// TestAcc_ClusterGcpNatGatewaysValidationErrors pins the user-facing error paths of
// features.nat_gateways through the real plugin wiring (ValidateConfig at plan time,
// toUpsertClusterRequest at apply time). Every step fails before any cluster is
// created, so the test is cheap and CheckDestroy is trivially satisfied.
//  1. Rule B (plan time): static_ips_enabled=true requires features.static_ip=true,
//  2. Rule A (plan time): nat_gateways with static_ips_enabled=true is GCP-only,
//  3. apply-time guard: a custom vpc_subnet is rejected for GCP clusters before
//     any API call is made.
func TestAcc_ClusterGcpNatGatewaysValidationErrors(t *testing.T) {
	t.Parallel()
	testName := "cluster-gcp-nat-gateways-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Rule B: enabled=true with static_ip=false on GCP → plan-time error.
			{
				Config: testAccClusterFeaturesConfig(
					testName, getTestGCPCredentialsID(), "GCP", "europe-west9", "AUTO_PILOT",
					`static_ip    = false
    nat_gateways = { static_ips_enabled = true }`,
				),
				ExpectError: regexp.MustCompile(`static_ips_enabled requires`),
			},
			// Rule A: nat_gateways enabled on a non-GCP cluster → plan-time error.
			{
				Config: testAccClusterFeaturesConfig(
					testName, getTestAWSCredentialsID(), "AWS", "eu-west-3", "T3A_MEDIUM",
					`static_ip    = true
    nat_gateways = { static_ips_enabled = true }`,
				),
				ExpectError: regexp.MustCompile(`only supported for GCP`),
			},
			// Apply-time guard: custom vpc_subnet on GCP is rejected before any API call.
			{
				Config: testAccClusterFeaturesConfig(
					testName, getTestGCPCredentialsID(), "GCP", "europe-west9", "AUTO_PILOT",
					`vpc_subnet = "10.42.0.0/16"`,
				),
				ExpectError: regexp.MustCompile(`vpc_subnet is not supported for GCP`),
			},
		},
	})
}

// --- Test check helpers ---

func testAccQoveryClusterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("cluster not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("cluster.id not found")
		}
		_, apiErr := apiClient.GetCluster(context.TODO(), getTestOrganizationID(), rs.Primary.ID, "{}", false)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryClusterDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("cluster not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("cluster.id not found")
		}
		// Retry to handle transient non-404 API responses that some cloud providers
		// (e.g. GCP) briefly return right after cluster deletion.
		var lastErr *apierrors.APIError
		for attempt := 0; attempt < 3; attempt++ {
			_, lastErr = apiClient.GetCluster(context.TODO(), getTestOrganizationID(), rs.Primary.ID, "{}", false)
			if lastErr == nil {
				return fmt.Errorf("found cluster but expected it to be deleted")
			}
			if apierrors.IsNotFound(lastErr) {
				return nil
			}
			time.Sleep(3 * time.Second)
		}
		return fmt.Errorf("unexpected error checking for deleted cluster: %s", lastErr.Summary())
	}
}

// --- Config helpers ---

// testAccClusterKarpenterConfig builds an AWS EKS+Karpenter cluster in READY state.
// Pass a non-empty description to set it. Pass attachLabels=true to attach a labels group.
func testAccClusterKarpenterConfig(testName, description string, attachLabels bool) string {
	descriptionLine := ""
	if description != "" {
		descriptionLine = fmt.Sprintf("\n  description = %q", description)
	}
	labelsGroupResource := ""
	labelsGroupIds := ""
	if attachLabels {
		labelsGroupResource = fmt.Sprintf(`
resource "qovery_labels_group" "test" {
  organization_id = "%s"
  name            = "%s-lg"
  labels = [{ key = "team", value = "platform", propagate_to_cloud_provider = true }]
}
`, getTestOrganizationID(), generateTestName(testName))
		labelsGroupIds = "\n  labels_group_ids = [qovery_labels_group.test.id]"
	}
	return fmt.Sprintf(`%s
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
      spot_enabled                 = true
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
  }%s
}
`, labelsGroupResource,
		getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName),
		descriptionLine, labelsGroupIds)
}

func testAccClusterKarpenterConfigWithDescription(testName, description string) string {
	return testAccClusterKarpenterConfig(testName, description, false)
}

func testAccClusterKarpenterConfigWithAdvancedSettings(testName string) string {
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
      spot_enabled                 = true
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

  advanced_settings_json = jsonencode({
    "loki.log_retention_in_week" = 2
  })
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterKarpenterConfigWithStaticIP(testName string) string {
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
    static_ip  = true
    karpenter = {
      spot_enabled                 = true
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
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterKarpenterConfigWithKeda(testName string, kedaEnabled bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "eu-west-3"
  kubernetes_mode = "MANAGED"
  state           = "READY"

  keda = {
    enabled = %t
  }

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      spot_enabled                 = true
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
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), kedaEnabled)
}

func testAccClusterAWSReadyConfig(testName string) string {
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
      spot_enabled                 = true
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
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterSCWReadyConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id    = "%s"
  organization_id   = "%s"
  name              = "%s"
  cloud_provider    = "SCW"
  region            = "pl-waw-1"
  kubernetes_mode   = "MANAGED"
  instance_type     = "DEV1-L"
  min_running_nodes = 3
  max_running_nodes = 3
  state             = "READY"
}
`, getTestScalewayCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterAzureReadyConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id    = "%s"
  organization_id   = "%s"
  name              = "%s"
  cloud_provider    = "AZURE"
  region            = "francecentral"
  kubernetes_mode   = "MANAGED"
  instance_type     = "Standard_B2s_v2"
  min_running_nodes = 3
  max_running_nodes = 10
  state             = "READY"
}
`, getTestAzureCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

func testAccClusterGCPReadyConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "GCP"
  region          = "europe-west9"
  kubernetes_mode = "MANAGED"
  instance_type   = "AUTO_PILOT"
  state           = "READY"
}
`, getTestGCPCredentialsID(), getTestOrganizationID(), generateTestName(testName))
}

// testAccClusterGCPNatGatewaysConfig builds a GCP cluster in READY state with
// features.static_ip set to staticIP and an optional nat_gateways block passed
// verbatim (empty string to omit the block).
func testAccClusterGCPNatGatewaysConfig(testName string, staticIP bool, natGatewaysBlock string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "GCP"
  region          = "europe-west9"
  kubernetes_mode = "MANAGED"
  instance_type   = "AUTO_PILOT"
  state           = "READY"
  features = {
    static_ip = %t
    %s
  }
}
`, getTestGCPCredentialsID(), getTestOrganizationID(), generateTestName(testName), staticIP, natGatewaysBlock)
}

// testAccClusterFeaturesConfig builds a READY cluster on the given provider with a
// verbatim features body, for validation-error steps that never create the resource.
func testAccClusterFeaturesConfig(testName string, credentialsID string, cloudProvider string, region string, instanceType string, featuresBody string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "%s"
  region          = "%s"
  kubernetes_mode = "MANAGED"
  instance_type   = "%s"
  state           = "READY"
  features = {
    %s
  }
}
`, credentialsID, getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, featuresBody)
}

func testAccClusterDefaultConfigWithVpcPeering(testName string, cloudProvider string, region string, instanceType string, vpcSubnet string, routingTable map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "%s"
  region          = "%s"
  instance_type   = "%s"
  features = {
    vpc_subnet = "%s"
  }
  routing_table = %s
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, vpcSubnet, convertRoutingTableToString(routingTable))
}

func convertRoutingTableToString(routingTable map[string]string) string {
	routes := make([]string, 0, len(routingTable))
	idx := 0
	for destination, target := range routingTable {
		routes = append(routes, fmt.Sprintf(`{description: "%s", destination: "%s", target: "%s"}`, fmt.Sprintf("route-%d", idx), destination, target))
		idx++
	}
	return fmt.Sprintf("[%s]", strings.Join(routes, ","))
}
