//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
					resource.TestCheckResourceAttr("data.qovery_organization.test", "name", "Q Sandbox"),
					resource.TestCheckResourceAttr("data.qovery_organization.test", "plan", "ENTERPRISE"),
					resource.TestCheckResourceAttr("data.qovery_organization.test", "description", "Organization for team's test"),
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
