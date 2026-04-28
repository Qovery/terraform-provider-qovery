//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with `name` mutated to verify useStateUnlessNameChangesModifier
// invalidates cached built_in_environment_variables on a value-affecting
// attribute change. git_repository E2E coverage isn't reliable here — the
// Qovery API resolves and validates Dockerfile existence at the configured
// root_path, so toggling root_path between two valid-but-different values
// requires multiple Dockerfile-bearing subdirectories on the test repo;
// branch/url toggles likewise require alternate refs the test repo doesn't
// guarantee. The modifier's git_repository-change branch is covered by unit
// tests in plan_modifiers_test.go.
func TestAcc_Application_ApplyConsistencyOnNameChange(t *testing.T) {
	t.Parallel()
	testName := "application-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			{Config: testAccApplicationApplyConsistencyConfig(testName, generateTestName(testName))},
			{Config: testAccApplicationApplyConsistencyConfig(testName, generateTestName(testName)+"-renamed")},
		},
	})
}

func testAccApplicationApplyConsistencyConfig(testName, name string) string {
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
    url          = "%s"
    git_token_id = "%s"
  }
}
`,
		testAccEnvironmentDefaultConfig(testName),
		name,
		applicationRepositoryURL,
		getTestQoverySandboxGitTokenID(),
	)
}
