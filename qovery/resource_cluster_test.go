//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

// TestAcc_AWSClusterConfigOnly creates an AWS cluster without deploying it (config only)
// This is faster and cheaper than deploying a full cluster
// Uses Karpenter which is required for new AWS EKS clusters
// Note: state=STOPPED prevents deployment, but API returns READY for never-deployed clusters
// causing expected drift on refresh (hence ExpectNonEmptyPlan)
func TestAcc_AWSClusterConfigOnly(t *testing.T) {
	t.Parallel()
	testName := "aws-cluster-config-only"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create cluster config only (no deployment triggered)
			{
				Config: testAccAWSClusterWithKarpenterConfigWithState(
					testName,
					"eu-west-3",
					"STOPPED",
				),
				ExpectNonEmptyPlan: true, // API returns READY for never-deployed clusters
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "region", "eu-west-3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
				),
			},
		},
	})
}

// TestAcc_GCPClusterConfigOnly creates a GCP cluster without deploying it (config only)
// GCP uses AUTO_PILOT mode where node counts are managed automatically by the cloud provider
// Note: state=STOPPED prevents deployment, but API returns READY for never-deployed clusters
func TestAcc_GCPClusterConfigOnly(t *testing.T) {
	t.Parallel()
	testName := "gcp-cluster-config-only"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create GCP cluster config only (no deployment triggered)
			{
				Config: testAccGCPClusterConfigWithState(
					testName,
					"europe-west9",
					"STOPPED",
				),
				ExpectNonEmptyPlan: true, // API returns READY for never-deployed clusters
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestGCPCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "GCP"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "region", "europe-west9"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "AUTO_PILOT"),
				),
			},
		},
	})
}

// TestAcc_AzureClusterConfigOnly creates an Azure AKS cluster without deploying it (config only)
// Azure credentials must be created via the Qovery console (provisioning requires server-side scripts)
// Note: state=STOPPED prevents deployment, but API returns READY for never-deployed clusters
func TestAcc_AzureClusterConfigOnly(t *testing.T) {
	t.Parallel()
	testName := "azure-cluster-config-only"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create Azure AKS cluster config only (no deployment triggered)
			{
				Config: testAccAzureClusterConfigWithState(
					testName,
					"francecentral",
					"STOPPED",
				),
				ExpectNonEmptyPlan: true, // API returns READY for never-deployed clusters
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAzureCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "region", "francecentral"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "Standard_B2s_v2"),
				),
			},
		},
	})
}

// FIXME: disabled until ttl advanced setting has been implemented for cleaning
func TestAcc_Cluster(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "cluster"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterDefaultConfig(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "advanced_settings", "{ loki.log_retention_in_week = 1 }"),
				),
			},
			// Add description
			{
				Config: testAccClusterDefaultConfigWithDescription(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
					"cluster description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", "cluster description"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
			// Update State -> STOPPED
			{
				Config: testAccClusterDefaultConfigWithState(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
					"STOPPED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "STOPPED"),
				),
			},
			// Update Resources
			{
				Config: testAccClusterDefaultConfigWithResources(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_LARGE",
					"4",
					"11",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_LARGE"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "4"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "11"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
			// Check Import
			// Since this takes too much time to create a cluster, the import test is done here.
			{
				ResourceName:        "qovery_cluster.test",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
			},
		},
	})
}

// FIXME: disabled until ttl advanced setting has been implemented for cleaning
func TestAcc_ClusterWithKubernetesMode(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "cluster-with-kubernetes-mode"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterDefaultK3SConfig(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_SMALL",
					true,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "K3S"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MT3A_SMALLEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "1"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "1"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "true"),
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.vpc_subnet"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
			// Check Import
			// Since this takes too much time to create a cluster, the import test is done here.
			// TODO: uncomment when ImportStateIdPrefix is fixed
			//{
			//	ResourceName:        "qovery_cluster.test",
			//	ImportState:         true,
			//	ImportStateVerify:   true,
			//	ImportStateIdPrefix: fmt.Sprintf("%s,", getTestOrganizationID()),
			//},
		},
	})
}

// FIXME: disabled until ttl advanced setting has been implemented for cleaning
func TestAcc_ClusterWithVpcPeering(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "cluster-with-vpc-peering"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.42.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.description", "route-0"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.destination", "172.30.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.targer", "target"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
		},
	})
}

func TestAcc_ClusterWithStaticIP(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "cluster-with-static-ip"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterDefaultConfigWithStaticIP(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
					true,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "kubernetes_mode", "MANAGED"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "production", "false"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "DEPLOYED"),
				),
			},
		},
	})
}

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

		_, apiErr := apiClient.GetCluster(context.TODO(), getTestOrganizationID(), rs.Primary.ID, "{}", false)
		if apiErr == nil {
			return fmt.Errorf("found cluster but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted cluster: %s", apiErr.Summary())
		}
		return nil
	}
}

func testAccClusterDefaultConfig(testName string, cloudProvider string, region string, instanceType string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  advanced_settings = jsonencode({ loki.log_retention_in_week = 1 })
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType,
	)
}

func testAccClusterDefaultConfigWithDescription(testName string, cloudProvider string, region string, instanceType string, description string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  description = "%s"
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, description,
	)
}

func testAccClusterDefaultConfigWithState(testName string, cloudProvider string, region string, instanceType string, state string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s" 
  state = "%s"
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, state,
	)
}

func testAccClusterDefaultK3SConfig(testName string, cloudProvider string, region string, instanceType string, production bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  kubernetes_mode = "K3S"
  min_running_nodes = 1
  max_running_nodes = 1
  production = %t
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, production,
	)
}

func testAccClusterDefaultConfigWithResources(
	testName string, cloudProvider string, region string, instanceType string,
	minRunningNodes string, maxRunningNodes string,
) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  min_running_nodes = "%s"
  max_running_nodes = "%s"
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, minRunningNodes, maxRunningNodes,
	)
}

func testAccClusterDefaultConfigWithVpcPeering(testName string, cloudProvider string, region string, instanceType string, vpcSubnet string, routingTable map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  features = {
    vpc_subnet = "%s"
  }
  routing_table = %s
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, vpcSubnet, convertRoutingTableToString(routingTable),
	)
}

func testAccClusterDefaultConfigWithStaticIP(testName string, cloudProvider string, region string, instanceType string, staticIP bool) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  instance_type = "%s"
  features = {
    static_ip = %t
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, staticIP,
	)
}

func convertRoutingTableToString(routingTable map[string]string) string {
	routes := make([]string, 0, len(routingTable))
	idx := 0
	for destination, target := range routingTable {
		routes = append(routes, fmt.Sprintf(`{description: "%s",  destination: "%s", target: "%s"}`, fmt.Sprintf("route-%d", idx), destination, target))
		idx++
	}
	return fmt.Sprintf("[%s]", strings.Join(routes, ","))
}

func testAccGCPClusterConfigWithState(testName string, region string, state string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "GCP"
  region          = "%s"
  instance_type   = "AUTO_PILOT"
  state           = "%s"
}
`, getTestGCPCredentialsID(), getTestOrganizationID(), generateTestName(testName), region, state,
	)
}

func testAccAWSClusterWithKarpenterConfigWithState(testName string, region string, state string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AWS"
  region          = "%s"
  state           = "%s"

  features = {
    vpc_subnet = "10.0.0.0/16"
    karpenter = {
      spot_enabled                 = true
      disk_size_in_gib             = 50
      default_service_architecture = "AMD64"
      qovery_node_pools = {
        requirements = [
          {
            key      = "InstanceSize"
            operator = "In"
            values   = ["small", "medium", "large", "xlarge", "2xlarge"]
          },
          {
            key      = "InstanceFamily"
            operator = "In"
            values   = ["t3", "t3a", "m5", "m5a", "m6i", "c5", "c5a"]
          },
          {
            key      = "Arch"
            operator = "In"
            values   = ["AMD64"]
          }
        ]
      }
    }
  }
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), region, state,
	)
}

func testAccAzureClusterConfigWithState(testName string, region string, state string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id  = "%s"
  organization_id = "%s"
  name            = "%s"
  cloud_provider  = "AZURE"
  region          = "%s"
  instance_type   = "Standard_B2s_v2"
  state           = "%s"
}
`, getTestAzureCredentialsID(), getTestOrganizationID(), generateTestName(testName), region, state,
	)
}
