//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAcc_Application_ApplyConsistencyOnGitRepositoryChange exercises the
// `git_repository` arm of the `useStateUnlessNameChangesModifier` invalidation
// list. Application built-in env vars (e.g. QOVERY_APPLICATION_*_GIT_*) embed
// values from the git_repository config — if the modifier reuses the cached
// list across a git_repository change, apply diverges from plan and the
// framework errors with "Provider produced inconsistent result after apply".
//
// We change `git_repository.root_path` because it's metadata-only at the
// qovery_application API layer (no repo-side validation required at save
// time, unlike branch which must exist on the remote), keeping the test
// stable independent of test-repo content.
//
// Companion: see resource_container/helm/job_apply_consistency_test.go.
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
