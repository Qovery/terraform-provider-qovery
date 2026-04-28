//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with the image tag mutated. A regression in
// useStateUnlessNameChangesModifier (state reused while built-in env vars
// actually changed — QOVERY_BUILD_ID embeds the tag) surfaces here as
// "Provider produced inconsistent result after apply".
func TestAcc_Container_ApplyConsistencyOnTagChange(t *testing.T) {
	t.Parallel()
	testName := "container-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			{Config: testAccContainerApplyConsistencyConfig(testName, "1.0.0")},
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
