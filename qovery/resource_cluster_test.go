package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_Cluster(t *testing.T) {
	t.Parallel()
	clusterNameSuffix := uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryClusterDestroy("qovery_cluster.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterConfigRequired(
					getTestAWSCredentialsID(),
					getTestOrganizationID(),
					generateClusterName(clusterNameSuffix),
					"AWS",
					"eu-west-3",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateClusterName(clusterNameSuffix)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cpu", "2000"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "memory", "4096"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
				),
			},
			// Add description
			{
				Config: testAccClusterConfigWithDescription(
					getTestAWSCredentialsID(),
					getTestOrganizationID(),
					generateClusterName(clusterNameSuffix),
					"AWS",
					"eu-west-3",
					"cluster description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateClusterName(clusterNameSuffix)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", "cluster description"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cpu", "2000"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "memory", "4096"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
				),
			},
			// Update State -> STOPPED
			{
				Config: testAccClusterConfigWithState(
					getTestAWSCredentialsID(),
					getTestOrganizationID(),
					generateClusterName(clusterNameSuffix),
					"AWS",
					"eu-west-3",
					"STOPPED",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateClusterName(clusterNameSuffix)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cpu", "2000"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "memory", "4096"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "3"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "10"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "STOPPED"),
				),
			},
			// Update Resources
			{
				Config: testAccClusterConfigWithResources(
					getTestAWSCredentialsID(),
					getTestOrganizationID(),
					generateClusterName(clusterNameSuffix),
					"AWS",
					"eu-west-3",
					"3000",
					"8192",
					"4",
					"11",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryClusterExists("qovery_cluster.test"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "credentials_id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_cluster.test", "name", generateClusterName(clusterNameSuffix)),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_cluster.test", "cpu", "3000"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "memory", "8192"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "min_running_nodes", "4"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "max_running_nodes", "11"),
					resource.TestCheckResourceAttr("qovery_cluster.test", "state", "RUNNING"),
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

func testAccClusterConfigRequired(credentialsID string, organizationID string, name string, cloudProvider string, region string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
}
`, credentialsID, organizationID, name, cloudProvider, region,
	)
}

func testAccClusterConfigWithDescription(credentialsID string, organizationID string, name string, cloudProvider string, region string, description string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  description = "%s"
}
`, credentialsID, organizationID, name, cloudProvider, region, description,
	)
}

func testAccClusterConfigWithState(credentialsID string, organizationID string, name string, cloudProvider string, region string, state string) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  state = "%s"
}
`, credentialsID, organizationID, name, cloudProvider, region, state,
	)
}

func testAccClusterConfigWithResources(
	credentialsID string, organizationID string, name string, cloudProvider string, region string,
	cpu string, memory string, minRunningNodes string, maxRunningNodes string,
) string {
	return fmt.Sprintf(`
resource "qovery_cluster" "test" {
  credentials_id = "%s"
  organization_id = "%s"
  name = "%s"
  cloud_provider = "%s"
  region = "%s"
  cpu = "%s"
  memory = "%s"
  min_running_nodes = "%s"
  max_running_nodes = "%s"
}
`, credentialsID, organizationID, name, cloudProvider, region, cpu, memory, minRunningNodes, maxRunningNodes,
	)
}

func generateClusterName(suffix string) string {
	return fmt.Sprintf("%s-cluster-%s", testResourcePrefix, suffix)
}
