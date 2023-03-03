//go:build integration && !unit
// +build integration,!unit

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
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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
					resource.TestCheckNoResourceAttr("qovery_cluster.test", "features.vpc_subnet"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.42.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.description", "route-0"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.destination", "172.30.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "routing_table.0.targer", "target"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.static_ip", "true"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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

		_, apiErr := apiClient.GetCluster(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
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

		_, apiErr := apiClient.GetCluster(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
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
  advanced_settings = { loki.log_retention_in_week = 1 }
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

func testAccClusterDefaultK3SConfig(testName string, cloudProvider string, region string, instanceType string) string {
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
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType,
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
