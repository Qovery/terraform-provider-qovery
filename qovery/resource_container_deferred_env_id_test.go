//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAcc_Container_DeferredEnvironmentID_NoReplacement reproduces the QOV-1938
// scenario end-to-end: a container's environment_id flows through a resource
// whose output is "(known after apply)" at plan time, so the container's planned
// environment_id is unknown.
//
// The wiring uses terraform_data (built-in to Terraform 1.4+). Its `output`
// attribute equals `input` after apply, but becomes "(known after apply)" when
// the resource is being replaced (which `triggers_replace` forces).
//
// Step 1 creates the container with the wiring in place. Step 2 changes
// triggers_replace, forcing terraform_data to be replaced, which makes its
// `output` (and therefore container.environment_id) unknown at plan time. The
// assertion is that the container's resource ID is unchanged across the two
// steps — if RequiresReplaceIfKnownChange regresses to stock RequiresReplace,
// the container is destroyed and recreated and the ID changes.
func TestAcc_Container_DeferredEnvironmentID_NoReplacement(t *testing.T) {
	t.Parallel()
	testName := "container-deferred-env-id"
	containerName := generateTestName(testName)

	var containerIDStep1 string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDeferredEnvIDConfig(testName, containerName, "v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["qovery_container.test"]
						if !ok {
							return fmt.Errorf("qovery_container.test not in state")
						}
						containerIDStep1 = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccContainerDeferredEnvIDConfig(testName, containerName, "v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["qovery_container.test"]
						if !ok {
							return fmt.Errorf("qovery_container.test not in state after step 2")
						}
						if rs.Primary.ID != containerIDStep1 {
							return fmt.Errorf(
								"container was replaced: ID changed from %s to %s — RequiresReplaceIfKnownChange did not suppress replacement on deferred upstream",
								containerIDStep1, rs.Primary.ID,
							)
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccContainerDeferredEnvIDConfig wires environment_id through a terraform_data
// resource whose `output` becomes "(known after apply)" whenever triggers_replace
// changes. Changing triggerValue between steps forces the resource to be replaced,
// which makes terraform_data.env_id_holder.output unknown at plan time, which
// propagates to qovery_container.test.environment_id.
func testAccContainerDeferredEnvIDConfig(testName, containerName, triggerValue string) string {
	return fmt.Sprintf(`
%s

%s

resource "terraform_data" "env_id_holder" {
  input            = qovery_environment.test.id
  triggers_replace = ["%s"]
}

resource "qovery_container" "test" {
  environment_id = terraform_data.env_id_holder.output
  registry_id    = qovery_container_registry.test.id
  name           = "%s"
  image_name     = "%s"
  tag            = "%s"
  healthchecks   = {}
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccContainerRegistryDefaultConfig(testName),
		triggerValue,
		containerName,
		containerImageName,
		containerTag,
	)
}
