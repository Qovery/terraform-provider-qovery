//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ProjectDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccProjectDataSourceConfig(
					getTestProjectID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_project.test", "id", getTestProjectID()),
					resource.TestCheckResourceAttr("data.qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_project.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_project.test", "description", "Project used to run test for our terraform provider"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_PROJECT_VARIABLE",
						"value": "MY_TERRAFORM_PROJECT_VARIABLE_VALUE",
					}),
					resource.TestCheckNoResourceAttr("data.qovery_project.test", "built_in_environment_variables.0.id"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_project.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_PROJECT_SECRET",
					}),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
data "qovery_project" "test" {
  id = "%s"
}
`, projectID,
	)
}
