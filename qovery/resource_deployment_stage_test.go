//go:build integration && !unit

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
			// Step 1: Create
			{
				Config: testAccDeploymentStageDefaultConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)+"-stage"),
					resource.TestCheckResourceAttrSet("qovery_deployment_stage.test", "environment_id"),
					resource.TestCheckResourceAttrSet("qovery_deployment_stage.test", "id"),
				),
			},
			// Step 2: Update name and description
			{
				Config: testAccDeploymentStageUpdatedConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.test"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "name", generateTestName(testName)+"-stage-updated"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.test", "description", "updated description"),
					resource.TestCheckResourceAttrSet("qovery_deployment_stage.test", "environment_id"),
					resource.TestCheckResourceAttrSet("qovery_deployment_stage.test", "id"),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "qovery_deployment_stage.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["qovery_deployment_stage.test"]
					if !ok {
						return "", fmt.Errorf("deployment stage not found: qovery_deployment_stage.test")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["environment_id"], rs.Primary.Attributes["name"]), nil
				},
				ImportStateVerifyIgnore: []string{"is_after", "is_before"},
			},
		},
	})
}

func TestAcc_DeploymentStageWithOrdering(t *testing.T) {
	t.Parallel()
	testName := "deployment-stage-ordering"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccQoveryDeploymentStageDestroy("qovery_deployment_stage.first"),
			testAccQoveryDeploymentStageDestroy("qovery_deployment_stage.second"),
		),
		Steps: []resource.TestStep{
			// Step 1: Create two stages with is_after ordering
			{
				Config: testAccDeploymentStageOrderingConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.first"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.second"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.first", "name", generateTestName(testName)+"-first"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.second", "name", generateTestName(testName)+"-second"),
					resource.TestCheckResourceAttrPair("qovery_deployment_stage.second", "is_after", "qovery_deployment_stage.first", "id"),
				),
			},
			// Step 2: Add a third stage after first (displacing second)
			{
				Config: testAccDeploymentStageReorderingConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.first"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.second"),
					testAccQoveryDeploymentStageExists("qovery_deployment_stage.third"),
					resource.TestCheckResourceAttr("qovery_deployment_stage.third", "name", generateTestName(testName)+"-third"),
					resource.TestCheckResourceAttrPair("qovery_deployment_stage.third", "is_after", "qovery_deployment_stage.first", "id"),
				),
			},
		},
	})
}

// Helper: check that a deployment stage resource exists in the API
func testAccQoveryDeploymentStageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("deployment stage not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("deployment_stage.id not found")
		}

		environmentID := rs.Primary.Attributes["environment_id"]
		if environmentID == "" {
			return fmt.Errorf("deployment_stage.environment_id not found")
		}

		_, err := qoveryServices.DeploymentStage.Get(context.TODO(), environmentID, rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

// Helper: check that a deployment stage resource has been destroyed
func testAccQoveryDeploymentStageDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("deployment stage not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("deployment_stage.id not found")
		}

		environmentID := rs.Primary.Attributes["environment_id"]
		if environmentID == "" {
			return fmt.Errorf("deployment_stage.environment_id not found")
		}

		_, err := qoveryServices.DeploymentStage.Get(context.TODO(), environmentID, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found deployment stage but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted deployment stage: %s", err.Error())
		}
		return nil
	}
}

// Config: default deployment stage
func testAccDeploymentStageDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_deployment_stage" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s-stage"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

// Config: updated deployment stage (name + description)
func testAccDeploymentStageUpdatedConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_deployment_stage" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s-stage-updated"
  description    = "updated description"
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName))
}

// Config: two stages with ordering (second is_after first)
func testAccDeploymentStageOrderingConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_deployment_stage" "first" {
  environment_id = qovery_environment.test.id
  name           = "%s-first"
}

resource "qovery_deployment_stage" "second" {
  environment_id = qovery_environment.test.id
  name           = "%s-second"
  is_after       = qovery_deployment_stage.first.id
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), generateTestName(testName))
}

// Config: three stages — third inserted after first (displacing second)
func testAccDeploymentStageReorderingConfig(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_deployment_stage" "first" {
  environment_id = qovery_environment.test.id
  name           = "%s-first"
}

resource "qovery_deployment_stage" "second" {
  environment_id = qovery_environment.test.id
  name           = "%s-second"
}

resource "qovery_deployment_stage" "third" {
  environment_id = qovery_environment.test.id
  name           = "%s-third"
  is_after       = qovery_deployment_stage.first.id
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), generateTestName(testName), generateTestName(testName))
}
