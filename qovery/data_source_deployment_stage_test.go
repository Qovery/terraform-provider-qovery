//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_DeploymentStageDataSource(t *testing.T) {
	t.Parallel()
	testName := "ds-deployment-stage"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a deployment stage first, then read it as a data source
			{
				Config: testAccDeploymentStageDataSourceConfig(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_deployment_stage.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttrSet("data.qovery_deployment_stage.test", "id"),
					resource.TestCheckResourceAttrSet("data.qovery_deployment_stage.test", "name"),
				),
			},
		},
	})
}

func testAccDeploymentStageDataSourceConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_deployment_stage" "source" {
  environment_id = "%s"
  name           = "%s"
}

data "qovery_deployment_stage" "test" {
  id             = qovery_deployment_stage.source.id
  environment_id = qovery_deployment_stage.source.environment_id
}
`, getTestEnvironmentID(), generateTestName(testName))
}
