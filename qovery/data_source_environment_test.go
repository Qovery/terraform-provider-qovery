//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"regexp"
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
					resource.TestCheckResourceAttr("data.qovery_environment.test", "name", "tests"),
					resource.TestCheckResourceAttr("data.qovery_environment.test", "mode", "DEVELOPMENT"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_environment.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_ENVIRONMENT_VARIABLE",
						"value": "MY_TERRAFORM_ENVIRONMENT_VARIABLE_VALUE",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_environment.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_environment.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_ENVIRONMENT_SECRET",
					}),
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
