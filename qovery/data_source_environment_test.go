package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_EnvironmentDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccEnvironmentDataSourceConfig(
					getTestEnvironmentID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_environment.test", "id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_environment.test", "project_id", getTestProjectID()),
					resource.TestCheckResourceAttr("data.qovery_environment.test", "cluster_id", getTestClusterID()),
					resource.TestCheckResourceAttr("data.qovery_environment.test", "name", "MyTerraformEnvironment"),
					resource.TestCheckResourceAttr("data.qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckNoResourceAttr("data.qovery_environment.test", "environment_variables.0"),
				),
			},
		},
	})
}

func testAccEnvironmentDataSourceConfig(environmentID string) string {
	return fmt.Sprintf(`
data "qovery_environment" "test" {
  id = "%s"
}
`, environmentID,
	)
}
