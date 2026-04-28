//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with git_repository mutated. A regression in
// useStateUnlessNameChangesModifier (state reused while built-in env vars
// actually changed) surfaces here as "Provider produced inconsistent result
// after apply". root_path is mutated because it requires no repo-side state.
func TestAcc_Application_ApplyConsistencyOnGitRepositoryChange(t *testing.T) {
	t.Parallel()
	testName := "application-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			{Config: testAccApplicationApplyConsistencyConfig(testName, "/")},
			{Config: testAccApplicationApplyConsistencyConfig(testName, "/cmd")},
		},
	})
}

func testAccApplicationApplyConsistencyConfig(testName, rootPath string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id  = qovery_environment.test.id
  name            = "%s"
  build_mode      = "DOCKER"
  dockerfile_path = "Dockerfile"
  cpu             = 500
  memory          = 512
  healthchecks    = {}

  git_repository = {
    url       = "%s"
    branch    = "%s"
    root_path = "%s"
  }
}
`,
		testAccEnvironmentDefaultConfig(testName),
		generateTestName(testName),
		applicationRepositoryURL,
		applicationBranch,
		rootPath,
	)
}
