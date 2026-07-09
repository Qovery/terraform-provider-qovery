//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_OrganizationMemberDataSource(t *testing.T) {
	t.Parallel()
	testName := "organization-member-data-source"
	adminRoleID := getTestAdminRoleID(t)
	email := fmt.Sprintf("%s@qovery-tf-acc-test.example.com", generateRandomName(testName))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationMemberDataSourceConfig(email, adminRoleID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_organization_member.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_organization_member.test", "email", email),
					resource.TestCheckResourceAttr("data.qovery_organization_member.test", "role_id", adminRoleID),
					resource.TestCheckResourceAttr("data.qovery_organization_member.test", "invitation_status", "PENDING"),
					resource.TestCheckResourceAttrSet("data.qovery_organization_member.test", "id"),
				),
			},
		},
	})
}

func testAccOrganizationMemberDataSourceConfig(email string, roleID string) string {
	return fmt.Sprintf(`
resource "qovery_organization_member" "test" {
  organization_id = "%s"
  email           = "%s"
  role_id         = "%s"
}

data "qovery_organization_member" "test" {
  organization_id = qovery_organization_member.test.organization_id
  email           = qovery_organization_member.test.email
}
`, getTestOrganizationID(), email, roleID,
	)
}
