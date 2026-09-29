//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// The Qovery API validates a git token against the git provider on every write, and CI has no
// real provider token available. The test therefore cannot create or update a git token: it checks
// the plan-time validation, then imports the pre-provisioned sandbox git token
// (TEST_QOVERY_SANDBOX_GIT_TOKEN_ID) without persisting it.
func TestAcc_GitToken(t *testing.T) {
	t.Parallel()
	const address = "qovery_git_token.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// A BITBUCKET token without a workspace fails at plan time
			{
				Config:      testAccGitTokenConfig("BITBUCKET"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`bitbucket_workspace\s+is\s+required\s+when\s+type\s+is\s+BITBUCKET`),
			},
			// Import records the remote values. The sandbox token has an empty description, which
			// reads as null, and no workspace. The API never returns the token value.
			{
				Config:        testAccGitTokenConfig("GITHUB"),
				ResourceName:  address,
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", getTestOrganizationID(), getTestQoverySandboxGitTokenID()),
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported git token, got %d", len(states))
					}
					attributes := states[0].Attributes
					expected := map[string]string{
						"id":              getTestQoverySandboxGitTokenID(),
						"organization_id": getTestOrganizationID(),
						"name":            "DO-NOT-DELETE-terraform-provider-tests",
						"type":            "GITHUB",
					}
					for key, value := range expected {
						if attributes[key] != value {
							return fmt.Errorf("imported %s is %q, expected %q", key, attributes[key], value)
						}
					}
					for _, key := range []string{"description", "bitbucket_workspace", "token"} {
						if value, ok := attributes[key]; ok {
							return fmt.Errorf("imported %s is %q, expected null", key, value)
						}
					}
					return nil
				},
			},
		},
	})
}

func testAccGitTokenConfig(tokenType string) string {
	return fmt.Sprintf(`
resource "qovery_git_token" "test" {
  organization_id = "%s"
  name            = "DO-NOT-DELETE-terraform-provider-tests"
  type            = "%s"
  token           = "not-a-real-token"
}
`, getTestOrganizationID(), tokenType)
}
