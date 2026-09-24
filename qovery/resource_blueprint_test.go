//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// Helm blueprint: no cloud resources, runs on any test cluster
const testAccBlueprintVersion = "HELM/redis/8"

func TestAcc_Blueprint(t *testing.T) {
	t.Parallel()
	testName := "blueprint"
	blueprintName := generateTestName(testName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryBlueprintDestroy("qovery_blueprint.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccBlueprintConfig(testName, blueprintName, "256Mi", "first-password"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryBlueprintExists("qovery_blueprint.test"),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "name", blueprintName),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "blueprint", testAccBlueprintVersion),
					resource.TestMatchResourceAttr("qovery_blueprint.test", "tag", regexp.MustCompile(`^HELM/redis/8/.+`)),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "variables.memory_limit", "256Mi"),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "secret_variables.password", "first-password"),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "service_type", "HELM"),
					resource.TestCheckResourceAttrSet("qovery_blueprint.test", "service_id"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "service_id", "qovery_blueprint.test", "service_id"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "tag", "qovery_blueprint.test", "tag"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "blueprint", "qovery_blueprint.test", "blueprint"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "service_type", "qovery_blueprint.test", "service_type"),
					resource.TestCheckResourceAttr("data.qovery_blueprint.test", "variables.memory_limit", "256Mi"),
					resource.TestCheckTypeSetElemAttr("data.qovery_blueprint.test", "secret_variable_names.*", "password"),
				),
			},
			{
				Config: testAccBlueprintConfig(testName, blueprintName, "512Mi", "second-password"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryBlueprintExists("qovery_blueprint.test"),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "variables.memory_limit", "512Mi"),
					resource.TestMatchResourceAttr("qovery_blueprint.test", "tag", regexp.MustCompile(`^HELM/redis/8/.+`)),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "tag", "qovery_blueprint.test", "tag"),
					resource.TestCheckResourceAttr("qovery_blueprint.test", "secret_variables.password", "second-password"),
				),
			},
		},
	})
}

func testAccBlueprintConfig(testName string, blueprintName string, memoryLimit string, password string) string {
	return fmt.Sprintf(`
%s

resource "qovery_blueprint" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  blueprint      = "%s"

  variables = {
    memory_limit = "%s"
  }
  secret_variables = {
    password = "%s"
  }
}

data "qovery_blueprint" "test" {
  id = qovery_blueprint.test.id
}
`, testAccEnvironmentDefaultConfig(testName), blueprintName, testAccBlueprintVersion, memoryLimit, password)
}

func testAccQoveryBlueprintExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("blueprint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("blueprint.id not found")
		}
		_, err := qoveryServices.Blueprint.Get(context.TODO(), rs.Primary.ID)
		return err
	}
}

func testAccQoveryBlueprintDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("blueprint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("blueprint.id not found")
		}

		_, err := qoveryServices.Blueprint.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found blueprint but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted blueprint: %s", err.Error())
		}
		return nil
	}
}
