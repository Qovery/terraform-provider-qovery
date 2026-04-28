//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Re-applies with `name` mutated to verify useStateUnlessNameChangesModifier
// invalidates cached built_in_environment_variables on a value-affecting
// attribute change. Tag/image_name/registry_id E2E coverage is blocked by
// the single-tag test ECR fixture; those branches are covered by unit tests
// in plan_modifiers_test.go.
func TestAcc_Container_ApplyConsistencyOnNameChange(t *testing.T) {
	t.Parallel()
	testName := "container-apply-consistency"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			{Config: testAccContainerApplyConsistencyConfig(testName, generateTestName(testName))},
			{Config: testAccContainerApplyConsistencyConfig(testName, generateTestName(testName)+"-renamed")},
		},
	})
}

func testAccContainerApplyConsistencyConfig(testName, name string) string {
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
		name,
		containerImageName,
		containerTag,
	)
}
