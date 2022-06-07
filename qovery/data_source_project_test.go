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
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_project.test", "id", getTestProjectID()),
					resource.TestCheckResourceAttr("data.qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_project.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_project.test", "description", "Project used to run test for our terraform provider"),
					resource.TestCheckNoResourceAttr("data.qovery_project.test", "environment_variables.0.id"),
					resource.TestCheckNoResourceAttr("data.qovery_project.test", "built_in_environment_variables.0.id"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_project.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_PROJECT_SECRET",
					}),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(credentialsID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_project" "test" {
  id = "%s"
  organization_id = "%s"
}
`, credentialsID, organizationID,
	)
}
