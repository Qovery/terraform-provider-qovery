//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_GitTokenDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing using the existing test git token
			{
				Config: testAccGitTokenDataSourceConfig(
					getTestQoverySandboxGitTokenID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "id", getTestQoverySandboxGitTokenID()),
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttrSet("data.qovery_git_token.test", "name"),
					resource.TestCheckResourceAttrSet("data.qovery_git_token.test", "type"),
				),
			},
		},
	})
}

func testAccGitTokenDataSourceConfig(gitTokenID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_git_token" "test" {
  id              = "%s"
  organization_id = "%s"
}
`, gitTokenID, organizationID)
}
