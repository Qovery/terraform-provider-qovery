//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The Qovery API validates a git token against the git provider on creation, and CI has no
// real provider token available. The test therefore reads the pre-provisioned sandbox git
// token (TEST_QOVERY_SANDBOX_GIT_TOKEN_ID) instead of creating one.
func TestAcc_GitTokenDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccGitTokenDataSourceConfig(
					getTestOrganizationID(),
					getTestQoverySandboxGitTokenID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "id", getTestQoverySandboxGitTokenID()),
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "name", "DO-NOT-DELETE-terraform-provider-tests"),
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("data.qovery_git_token.test", "description", ""),
					resource.TestCheckNoResourceAttr("data.qovery_git_token.test", "bitbucket_workspace"),
					// The API never returns the token value: the data source exposes it as null
					resource.TestCheckNoResourceAttr("data.qovery_git_token.test", "token"),
				),
			},
		},
	})
}

func testAccGitTokenDataSourceConfig(orgID string, gitTokenID string) string {
	return fmt.Sprintf(`
data "qovery_git_token" "test" {
  id              = "%s"
  organization_id = "%s"
}
`, gitTokenID, orgID,
	)
}
