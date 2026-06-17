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

func TestAcc_ApplicationAutoscalingMinMustBeLessThanMax(t *testing.T) {
	t.Parallel()
	testName := "application-autoscaling-min-max-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationConfigAutoscalingMinEqualsMax(testName),
				ExpectError: regexp.MustCompile(`strictly less than max_running_instances`),
			},
		},
	})
}

func TestAcc_ApplicationAutoscalingRequiresEnabledScaler(t *testing.T) {
	t.Parallel()
	testName := "application-autoscaling-all-disabled-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationConfigAutoscalingAllDisabled(testName),
				ExpectError: regexp.MustCompile(`at least one scaler must be enabled`),
			},
		},
	})
}

func TestAcc_ApplicationHpaToKedaTransitionRejected(t *testing.T) {
	t.Parallel()
	testName := "application-hpa-to-keda-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Step 1: create an HPA service (min != max, no autoscaling block).
			{
				Config: testAccApplicationConfigHpa(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "2"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "5"),
				),
			},
			// Step 2: add a KEDA autoscaling block -> direct HPA->KEDA transition must be rejected at plan time.
			{
				Config:      testAccApplicationConfigHpaToKeda(testName),
				ExpectError: regexp.MustCompile(`two-step`),
			},
		},
	})
}

// Regression: the documented two-step migration must be allowed. A service whose
// prior state has min == max (NONE mode, not HPA) can gain a KEDA autoscaling
// block in a single apply without being blocked by the HPA->KEDA guard.
func TestAcc_ApplicationEqualInstancesToKedaAllowed(t *testing.T) {
	t.Parallel()
	testName := "application-equal-to-keda-allowed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy("qovery_application.test"),
		Steps: []resource.TestStep{
			// Step 1: min == max (NONE mode, no autoscaling).
			{
				Config: testAccApplicationConfigEqualInstances(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "1"),
				),
			},
			// Step 2: add a KEDA autoscaling block with min < max -> must succeed.
			{
				Config: testAccApplicationConfigEqualInstancesToKeda(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApplicationExists("qovery_application.test"),
					resource.TestCheckResourceAttr("qovery_application.test", "min_running_instances", "0"),
					resource.TestCheckResourceAttr("qovery_application.test", "max_running_instances", "3"),
					resource.TestCheckResourceAttr("qovery_application.test", "autoscaling.scalers.#", "1"),
				),
			},
		},
	})
}

func testAccApplicationConfigEqualInstances(testName string) string {
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
  min_running_instances = 1
  max_running_instances = 1
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigEqualInstancesToKeda(testName string) string {
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
  max_running_instances = 3
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up" })
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigAutoscalingMinEqualsMax(testName string) string {
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
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up" })
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigAutoscalingAllDisabled(testName string) string {
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
  max_running_instances = 3
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up" })
        enabled     = false
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigHpa(testName string) string {
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
  min_running_instances = 2
  max_running_instances = 5
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
}

func testAccApplicationConfigHpaToKeda(testName string) string {
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
  min_running_instances = 2
  max_running_instances = 5
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up" })
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), applicationRepositoryURL, getTestQoverySandboxGitTokenID(),
	)
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
