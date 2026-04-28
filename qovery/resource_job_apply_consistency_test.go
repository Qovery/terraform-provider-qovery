//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with the source.image tag mutated. A regression in
// useStateUnlessNameChangesModifier (state reused while built-in env vars
// actually changed — QOVERY_BUILD_ID embeds the tag) surfaces here as
// "Provider produced inconsistent result after apply".
func TestAcc_Job_ApplyConsistencyOnImageTagChange(t *testing.T) {
	t.Parallel()
	testName := "job-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy("qovery_job.test"),
		Steps: []resource.TestStep{
			{Config: testAccJobApplyConsistencyConfig(testName, "1.0.0")},
			{Config: testAccJobApplyConsistencyConfig(testName, "1.0.1")},
		},
	})
}

func testAccJobApplyConsistencyConfig(testName, tag string) string {
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
		generateTestName(testName),
		jobImageName,
		tag,
	)
}
