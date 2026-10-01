//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_OrganizationDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccOrganizationDataSourceConfig(
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_organization.test", "id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_organization.test", "name", testAccOrganizationName),
					resource.TestCheckResourceAttr("data.qovery_organization.test", "plan", testAccOrganizationPlan),
					resource.TestCheckResourceAttr("data.qovery_organization.test", "description", testAccOrganizationDescription),
				),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig(organizationID string) string {
	return fmt.Sprintf(`
data "qovery_organization" "test" {
  id = "%s"
}
`, organizationID)
}
