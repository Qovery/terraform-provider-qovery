//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAcc_Job_ApplyConsistencyOnImageTagChange is the regression test for
// the customer-reported "Provider produced inconsistent result after apply"
// error against `built_in_environment_variables[N].value`, where N held an
// entry like QOVERY_BUILD_ID containing the prior image tag.
//
// The bug pattern: the original `UseStateUnlessNameChanges` modifier only
// invalidated the cached env-var list when `name` or `ports` changed. A tag
// change left the list reused from state, and the API then returned a list
// containing the new tag — apply diverged from plan, framework errored.
//
// This test applies once at tag "1.0.0", then re-applies at "1.0.1". A
// regression here surfaces as a concrete apply-time failure.
//
// Companion: see resource_container/helm/application_apply_consistency_test.go.
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
