//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAcc_Container_ApplyConsistencyOnTagChange exercises the path that
// caused "Provider produced inconsistent result after apply" errors against
// `built_in_environment_variables` when a Computed list's plan modifier
// preserved the prior state value across a change that the API would
// reflect into built-in env vars (e.g. QOVERY_BUILD_ID embeds the tag).
//
// The test applies the resource once, then re-applies with only the `tag`
// attribute changed. The terraform-plugin-testing framework auto-fails any
// step whose apply returns an error, so a regression in the plan modifier
// (state reused while built-in env vars actually changed) shows up as a
// concrete CI failure here.
//
// Companion files cover the same property for the other service resources:
//   - resource_application_apply_consistency_test.go (git_repository)
//   - resource_helm_apply_consistency_test.go        (source)
//   - resource_job_apply_consistency_test.go         (source.image)
func TestAcc_Container_ApplyConsistencyOnTagChange(t *testing.T) {
	t.Parallel()
	testName := "container-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Initial apply at tag "1.0.0".
			{Config: testAccContainerApplyConsistencyConfig(testName, "1.0.0")},
			// Change only the tag. If `built_in_environment_variables`'
			// plan modifier reuses state across this change, apply will
			// fail with "Provider produced inconsistent result after apply".
			{Config: testAccContainerApplyConsistencyConfig(testName, "1.0.1")},
		},
	})
}

func testAccContainerApplyConsistencyConfig(testName, tag string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccContainerRegistryDefaultConfig(testName),
		generateTestName(testName),
		containerImageName,
		tag,
	)
}
