//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ApiTokenDataSource(t *testing.T) {
	t.Parallel()
	testName := "api-token-data-source"
	adminRoleID := getTestAdminRoleID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a token and read it back through the data source
			{
				Config: testAccApiTokenDataSourceConfig(
					testName,
					adminRoleID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.qovery_api_token.test", "id", "qovery_api_token.test", "id"),
					resource.TestCheckResourceAttr("data.qovery_api_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_api_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("data.qovery_api_token.test", "description", "this is a test api token"),
					resource.TestCheckResourceAttr("data.qovery_api_token.test", "role_id", adminRoleID),
					// The secret value is only returned at creation time: the data source never exposes it
					resource.TestCheckNoResourceAttr("data.qovery_api_token.test", "token"),
				),
			},
		},
	})
}

func testAccApiTokenDataSourceConfig(testName string, roleID string) string {
	return fmt.Sprintf(`
resource "qovery_api_token" "test" {
  organization_id = "%s"
  name            = "%s"
  description     = "this is a test api token"
  role_id         = "%s"
}

data "qovery_api_token" "test" {
  id              = qovery_api_token.test.id
  organization_id = qovery_api_token.test.organization_id
}
`, getTestOrganizationID(), generateTestName(testName), roleID,
	)
}
