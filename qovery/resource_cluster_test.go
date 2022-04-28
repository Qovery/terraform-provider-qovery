package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_Cluster(t *testing.T) {
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
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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

func TestAcc_ClusterWithVpcSubnet(t *testing.T) {
	t.Parallel()
	testName := "cluster-with-vpc-subnet"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterDefaultConfigWithVpcSubnet(
					testName,
					"AWS",
					"eu-west-3",
					"T3A_MEDIUM",
					"10.42.0.0/16",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "instance_type", "T3A_MEDIUM"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "features.vpc_subnet", "10.42.0.0/16"),
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

func testAccClusterDefaultConfigWithVpcSubnet(testName string, cloudProvider string, region string, instanceType string, vpcSubnet string) string {
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
}
`, getTestAWSCredentialsID(), getTestOrganizationID(), generateTestName(testName), cloudProvider, region, instanceType, vpcSubnet,
	)
}
