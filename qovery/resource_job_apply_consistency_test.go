//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with `name` mutated to verify useStateUnlessNameChangesModifier
// invalidates cached built_in_environment_variables on a value-affecting
// attribute change. source.image tag E2E coverage is blocked by the
// single-tag test ECR fixture; that branch is covered by unit tests in
// plan_modifiers_test.go.
func TestAcc_Job_ApplyConsistencyOnNameChange(t *testing.T) {
	t.Parallel()
	testName := "job-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy("qovery_job.test"),
		Steps: []resource.TestStep{
			{Config: testAccJobApplyConsistencyConfig(testName, generateTestName(testName))},
			{Config: testAccJobApplyConsistencyConfig(testName, generateTestName(testName)+"-renamed")},
		},
	})
}

func testAccJobApplyConsistencyConfig(testName, name string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_job" "test" {
  environment_id       = qovery_environment.test.id
  name                 = "%s"
  cpu                  = 500
  memory               = 512
  max_duration_seconds = 300
  max_nb_restart       = 0
  auto_preview         = false
  healthchecks         = {}

  source = {
    image = {
      registry_id = qovery_container_registry.test.id
      name        = "%s"
      tag         = "%s"
    }
  }

  schedule = {
    cronjob = {
      schedule = "*/2 * * * *"
      command = {
        entrypoint = "test.sh"
      }
    }
  }
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccContainerRegistryDefaultConfig(testName),
		name,
		jobImageName,
		jobImageTag,
	)
}
