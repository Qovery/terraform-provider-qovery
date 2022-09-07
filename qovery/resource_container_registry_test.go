//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

const (
	awsECRRegion = "eu-west-3"
	awsECRURL    = "https://default.com"
)

func TestAcc_ContainerRegistry(t *testing.T) {
	t.Parallel()
	testName := "container-registry"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerRegistryDestroy("qovery_container_registry.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerRegistryDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "kind", "ECR"),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "url", awsECRURL),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "description", ""),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.region", awsECRRegion),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
			// Add description
			{
				Config: testAccContainerRegistryDefaultConfigWithDescription(
					testName,
					"this is a description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "kind", "ECR"),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "url", awsECRURL),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "description", "this is a description"),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.region", awsECRRegion),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.access_key_id", getTestAWSCredentialsAccessKeyID()),
					resource.TestCheckResourceAttr("qovery_container_registry.test", "config.secret_access_key", getTestAWSCredentialsSecretAccessKey()),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_container_registry.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     fmt.Sprintf("%s,", getTestOrganizationID()),
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

func testAccQoveryContainerRegistryExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("container registry not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("container_registry.id not found")
		}

		_, err := qoveryServices.ContainerRegistry.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryContainerRegistryDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("container registry not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("container_registry.id not found")
		}

		_, err := qoveryServices.ContainerRegistry.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found container registry but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted container registry: %s", err.Error())
		}
		return nil
	}
}

func testAccContainerRegistryDefaultConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_container_registry" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "ECR"
  url = "%s"
  config = {
    region = "%s"
    access_key_id = "%s"
    secret_access_key = "%s"
  }
}
`, getTestOrganizationID(), generateTestName(testName), awsECRURL, awsECRRegion, getTestAWSCredentialsAccessKeyID(), getTestAWSCredentialsSecretAccessKey(),
	)
}

func testAccContainerRegistryDefaultConfigWithDescription(testName string, description string) string {
	return fmt.Sprintf(`
resource "qovery_container_registry" "test" {
  organization_id = "%s"
  name = "%s"
  kind = "ECR"
  url = "%s"
  config = {
    region = "%s"
    access_key_id = "%s"
    secret_access_key = "%s"
  }
  description = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), awsECRURL, awsECRRegion, getTestAWSCredentialsAccessKeyID(), getTestAWSCredentialsSecretAccessKey(), description,
	)
}
