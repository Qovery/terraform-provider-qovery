//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ContainerWithAutoscaling(t *testing.T) {
	t.Parallel()
	testName := "container-with-autoscaling"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create with a KEDA autoscaling block + scale-to-zero (min = 0).
			{
				Config: testAccContainerConfigWithAutoscalingJSON(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "0"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.polling_interval_seconds", "30"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.cooldown_period_seconds", "300"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "autoscaling.scalers.*", map[string]string{
						"scaler_type": "prometheus",
						"role":        "PRIMARY",
						"enabled":     "true",
					}),
				),
			},
			// Import.
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update: switch the scaler to config_yaml and add a SAFETY scaler.
			{
				Config: testAccContainerConfigWithAutoscalingYAML(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.#", "2"),
				),
			},
		},
	})
}

func TestAcc_ContainerScaleToZeroRequiresAutoscaling(t *testing.T) {
	t.Parallel()
	testName := "container-scale-to-zero-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerConfigMinZeroNoAutoscaling(testName),
				ExpectError: regexp.MustCompile(`scale-to-zero`),
			},
		},
	})
}

func TestAcc_ContainerAutoscalingConfigExclusivity(t *testing.T) {
	t.Parallel()
	testName := "container-config-exclusivity-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerConfigScalerBothConfigs(testName),
				ExpectError: regexp.MustCompile(`exactly one of config_json`),
			},
		},
	})
}

func TestAcc_ContainerAutoscalingMinMustBeLessThanMax(t *testing.T) {
	t.Parallel()
	testName := "container-autoscaling-min-max-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerConfigAutoscalingMinEqualsMax(testName),
				ExpectError: regexp.MustCompile(`strictly less than max_running_instances`),
			},
		},
	})
}

func TestAcc_ContainerAutoscalingRequiresEnabledScaler(t *testing.T) {
	t.Parallel()
	testName := "container-autoscaling-all-disabled-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerConfigAutoscalingAllDisabled(testName),
				ExpectError: regexp.MustCompile(`at least one scaler must be enabled`),
			},
		},
	})
}

func TestAcc_ContainerHpaToKedaTransitionRejected(t *testing.T) {
	t.Parallel()
	testName := "container-hpa-to-keda-invalid"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Step 1: create an HPA service (min != max, no autoscaling block).
			{
				Config: testAccContainerConfigHpa(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "5"),
				),
			},
			// Step 2: add a KEDA autoscaling block -> direct HPA->KEDA transition must be rejected at plan time.
			{
				Config:      testAccContainerConfigHpaToKeda(testName),
				ExpectError: regexp.MustCompile(`two-step`),
			},
		},
	})
}

func testAccContainerConfigAutoscalingMinEqualsMax(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
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
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigAutoscalingAllDisabled(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
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
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigHpa(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  min_running_instances = 2
  max_running_instances = 5
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigHpaToKeda(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
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
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigWithAutoscalingJSON(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
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
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigWithAutoscalingYAML(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  min_running_instances = 0
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_yaml = "query: up"
      },
      {
        scaler_type = "cron"
        role        = "SAFETY"
        config_json = jsonencode({ start = "0 0 * * *" })
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigMinZeroNoAutoscaling(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  min_running_instances = 0
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigScalerBothConfigs(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  healthchecks = {}
  autoscaling = {
    scalers = [
      {
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_json = jsonencode({ query = "up" })
        config_yaml = "query: up"
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}
