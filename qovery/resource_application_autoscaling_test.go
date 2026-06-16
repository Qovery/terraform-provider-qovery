//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ApplicationWithAutoscaling(t *testing.T) {
	t.Parallel()
	testName := "application-with-autoscaling"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Create with a KEDA autoscaling block + scale-to-zero (min = 0).
			{
				Config: testAccApplicationConfigWithAutoscaling(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "0"),
					resource.TestCheckResourceAttr("qovery_application.test", "autoscaling.scalers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_application.test", "autoscaling.scalers.*", map[string]string{
						"scaler_type": "prometheus",
						"role":        "PRIMARY",
					}),
				),
			},
			// Import.
			{
				ResourceName:      "qovery_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ApplicationScaleToZeroRequiresAutoscaling(t *testing.T) {
	t.Parallel()
	testName := "application-scale-to-zero-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationConfigMinZeroNoAutoscaling(testName),
				ExpectError: regexp.MustCompile(`scale-to-zero`),
			},
		},
	})
}

func testAccApplicationConfigWithAutoscaling(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
    git_token_id = "%s"
  }
  min_running_instances = 0
  healthchecks = {}
  autoscaling = {
    polling_interval_seconds = 30
    cooldown_period_seconds  = 300
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up", threshold = "1" })
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigMinZeroNoAutoscaling(testName string) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  git_repository = {
    url = "%s"
    git_token_id = "%s"
  }
  min_running_instances = 0
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}
