//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// NOTE: This test requires a pre-existing terraform service in the test environment.
// If TEST_TERRAFORM_SERVICE_ID is not set, the test will be skipped.
func TestAcc_TerraformServiceDataSource(t *testing.T) {
	t.Parallel()

	// Create a terraform service first, then read it as a data source
	testName := "ds-terraform-service"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a terraform service first, then read it as a data source
			{
				Config: testAccTerraformServiceDataSourceConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_terraform_service.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttrSet("data.qovery_terraform_service.test", "id"),
					resource.TestCheckResourceAttrSet("data.qovery_terraform_service.test", "name"),
				),
			},
		},
	})
}

func testAccTerraformServiceDataSourceConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_terraform_service" "source" {
  environment_id = "%s"
  name           = "%s"
  auto_deploy    = false

  source {
    git_repository {
      url       = "https://github.com/Qovery/terraform-provider-qovery.git"
      branch    = "main"
      root_path = "/"
    }
  }

  backend {
    kubernetes {}
  }

  engine = "TERRAFORM"
  engine_version {
    version = "1.5.7"
  }

  resources {
    cpu    = 500
    memory = 512
  }
}

data "qovery_terraform_service" "test" {
  id             = qovery_terraform_service.source.id
  environment_id = qovery_terraform_service.source.environment_id
}
`, getTestEnvironmentID(), generateTestName(testName))
}
