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

func TestAcc_DeploymentStage(t *testing.T) {
	t.Parallel()
	testName := "deployment-stage"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDeploymentStageDestroy("qovery_deployment_stage.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDeploymentStageDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "description", ""),
				),
			},
			// Update name
			{
				Config: testAccDeploymentStageConfig(testName+"-updated", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName+"-updated")),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "description", ""),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_deployment_stage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_DeploymentStageWithDescription(t *testing.T) {
	t.Parallel()
	testName := "deployment-stage-desc"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDeploymentStageDestroy("qovery_deployment_stage.test"),
		Steps: []resource.TestStep{
			// Create with description
			{
				Config: testAccDeploymentStageConfig(testName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "description", "Initial description"),
				),
			},
			// Update description
			{
				Config: testAccDeploymentStageConfig(testName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "description", "Updated description"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_deployment_stage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_DeploymentStage_Import(t *testing.T) {
	t.Parallel()
	testName := "deployment-stage-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryDeploymentStageDestroy("qovery_deployment_stage.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDeploymentStageDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_deployment_stage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryDeploymentStageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("deployment stage not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("deployment_stage.id not found")
		}

		_, err := qoveryServices.DeploymentStage.Get(context.TODO(), rs.Primary.Attributes["environment_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryDeploymentStageDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("deployment stage not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("deployment_stage.id not found")
		}

		_, err := qoveryServices.DeploymentStage.Get(context.TODO(), rs.Primary.Attributes["environment_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found deployment stage but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted deployment stage: %s", err.Error())
		}
		return nil
	}
}

func testAccDeploymentStageDefaultConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_deployment_stage" "test" {
  environment_id = "%s"
  name           = "%s"
}
`, getTestEnvironmentID(), generateTestName(testName))
}

func testAccDeploymentStageConfig(testName string, description string) string {
	if description == "" {
		return testAccDeploymentStageDefaultConfig(testName)
	}
	return fmt.Sprintf(`
resource "qovery_deployment_stage" "test" {
  environment_id = "%s"
  name           = "%s"
  description    = "%s"
}
`, getTestEnvironmentID(), generateTestName(testName), description)
}
