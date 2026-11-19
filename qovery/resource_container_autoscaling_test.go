//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_ContainerWithAutoscaling(t *testing.T) {
	t.Parallel()
	testName := "container-with-autoscaling"
	var scalerID, cronScalerID scalerIDTracker
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
					scalerID.record("qovery_container.test", "autoscaling.scalers.0.id"),
				),
			},
			// Import.
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Create stores "" when unset while Read returns "{}"; state is not refreshed
				// before import.
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
			// Update: switch the scaler to config_yaml and add a SAFETY scaler. The
			// first scaler is updated in place.
			{
				Config: testAccContainerConfigWithAutoscalingYAML(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.#", "2"),
					scalerID.unchanged("qovery_container.test", "autoscaling.scalers.0.id"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.1.scaler_type", "cron"),
					cronScalerID.record("qovery_container.test", "autoscaling.scalers.1.id"),
				),
			},
			// Reorder the scalers in the config: each one keeps its id.
			{
				Config: testAccContainerConfigWithAutoscalingScalers(testName, autoscalingCronJSONScaler, autoscalingPrometheusYAMLScaler),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.0.scaler_type", "cron"),
					cronScalerID.unchanged("qovery_container.test", "autoscaling.scalers.0.id"),
					scalerID.unchanged("qovery_container.test", "autoscaling.scalers.1.id"),
				),
			},
		},
	})
}

// Regression: the API ignores the name of an inline trigger authentication and
// generates one, which made the apply fail with "Provider produced inconsistent
// result after apply". The state keeps the configured name.
func TestAcc_ContainerAutoscalingTriggerAuthentication(t *testing.T) {
	t.Parallel()
	testName := "container-autoscaling-trigger-auth"
	var scalerID scalerIDTracker
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Create: the apply is consistent and the next plan is empty.
			{
				Config: testAccContainerConfigWithTriggerAuthentication(testName, "tf-acc-auth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "autoscaling.scalers.*", map[string]string{
						"scaler_type":                        "prometheus",
						"trigger_authentication.name":        "tf-acc-auth",
						"trigger_authentication.config_yaml": autoscalingTriggerAuthConfigYAML,
					}),
					scalerID.record("qovery_container.test", "autoscaling.scalers.0.id"),
				),
			},
			// Rename: the API keeps its generated name, the state takes the new one,
			// and the scaler is updated in place.
			{
				Config: testAccContainerConfigWithTriggerAuthentication(testName, "tf-acc-auth-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					scalerID.unchanged("qovery_container.test", "autoscaling.scalers.0.id"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_container.test", "autoscaling.scalers.*", map[string]string{
						"trigger_authentication.name": "tf-acc-auth-renamed",
					}),
				),
			},
			// Import: no prior state, so the state records the generated name.
			{
				ResourceName:      "qovery_container.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Create stores "" when unset while Read returns "{}"; state is not refreshed
				// before import.
				ImportStateVerifyIgnore: []string{"advanced_settings_json", "autoscaling.scalers.0.trigger_authentication.name"},
			},
		},
	})
}

// scalerIDTracker records the id of a scaler at one step and checks that a later
// step updated that scaler in place instead of recreating it.
type scalerIDTracker struct {
	id string
}

func (s *scalerIDTracker) record(name, key string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(name, key, func(value string) error {
		if value == "" {
			return fmt.Errorf("%s is empty", key)
		}
		s.id = value
		return nil
	})
}

func (s *scalerIDTracker) unchanged(name, key string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(name, key, func(value string) error {
		if value != s.id {
			return fmt.Errorf("%s changed from %q to %q: the scaler was recreated", key, s.id, value)
		}
		return nil
	})
}

// autoscalingTriggerAuthConfigYAML is a KEDA TriggerAuthentication spec that
// references no secret, shared by the container and application tests.
const autoscalingTriggerAuthConfigYAML = "podIdentity:\n  provider: none"

func testAccContainerConfigWithTriggerAuthentication(testName, triggerAuthName string) string {
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
        config_json = jsonencode({ query = "up" })
        trigger_authentication = {
          name        = "%s"
          config_yaml = %q
        }
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
		triggerAuthName, autoscalingTriggerAuthConfigYAML,
	)
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

// A service on HPA (min != max) can switch to KEDA in a single apply (QOV-2564).
func TestAcc_ContainerHpaToKedaTransitionAllowed(t *testing.T) {
	t.Parallel()
	testName := "container-hpa-to-keda-allowed"
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
			// Step 2: add the KEDA autoscaling block in the same apply -> must succeed.
			{
				Config: testAccContainerConfigHpaToKeda(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "2"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "5"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.#", "1"),
				),
			},
		},
	})
}

// A service whose prior state has min == max (NONE mode) can gain a KEDA
// autoscaling block in a single apply.
func TestAcc_ContainerEqualInstancesToKedaAllowed(t *testing.T) {
	t.Parallel()
	testName := "container-equal-to-keda-allowed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy("qovery_container.test"),
		Steps: []resource.TestStep{
			// Step 1: min == max (NONE mode, no autoscaling).
			{
				Config: testAccContainerConfigEqualInstances(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "1"),
				),
			},
			// Step 2: add a KEDA autoscaling block with min < max -> must succeed.
			{
				Config: testAccContainerConfigEqualInstancesToKeda(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists("qovery_container.test"),
					resource.TestCheckResourceAttr("qovery_container.test", "min_running_instances", "0"),
					resource.TestCheckResourceAttr("qovery_container.test", "max_running_instances", "3"),
					resource.TestCheckResourceAttr("qovery_container.test", "autoscaling.scalers.#", "1"),
				),
			},
		},
	})
}

func testAccContainerConfigEqualInstances(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  min_running_instances = 1
  max_running_instances = 1
  healthchecks = {}
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
}

func testAccContainerConfigEqualInstancesToKeda(testName string) string {
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
      },
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
	)
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

const autoscalingPrometheusYAMLScaler = `{
        scaler_type = "prometheus"
        role        = "PRIMARY"
        config_yaml = "query: up"
      }`

const autoscalingCronJSONScaler = `{
        scaler_type = "cron"
        role        = "SAFETY"
        config_json = jsonencode({ start = "0 0 * * *" })
      }`

func testAccContainerConfigWithAutoscalingYAML(testName string) string {
	return testAccContainerConfigWithAutoscalingScalers(testName, autoscalingPrometheusYAMLScaler, autoscalingCronJSONScaler)
}

func testAccContainerConfigWithAutoscalingScalers(testName, firstScaler, secondScaler string) string {
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
      %s,
      %s,
    ]
  }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag,
		firstScaler, secondScaler,
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

const autoscalingAdvancedYAML = "horizontalPodAutoscalerConfig:\n  behavior:\n    scaleDown:\n      stabilizationWindowSeconds: 60\n"

// The KEDA ScaledObject settings: set on create, kept on import, a change made
// outside Terraform shows on refresh, and removing them from the config clears them.
func TestAcc_ContainerAutoscalingScaledObjectSettings(t *testing.T) {
	t.Parallel()
	testName := "container-autoscaling-scaled-object-settings"
	address := "qovery_container.test"
	settings := `
    initial_cooldown_period_seconds = 120
    idle_replica_count              = 0
    advanced_yaml                   = ` + fmt.Sprintf("%q", autoscalingAdvancedYAML)
	var containerID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(address),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfigWithScaledObjectSettings(testName, 1, settings),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(address),
					testAccCaptureResourceID(address, &containerID),
					resource.TestCheckResourceAttr(address, "autoscaling.initial_cooldown_period_seconds", "120"),
					resource.TestCheckResourceAttr(address, "autoscaling.idle_replica_count", "0"),
					resource.TestCheckResourceAttr(address, "autoscaling.advanced_yaml", autoscalingAdvancedYAML),
				),
			},
			{
				ResourceName:            address,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"advanced_settings_json"},
			},
			// A change made outside Terraform shows as a difference on refresh.
			{
				PreConfig: func() {
					err := testAccEditServiceOutOfBand("/container/"+containerID, func(container map[string]any) {
						if registry, ok := container["registry"].(map[string]any); ok {
							container["registry_id"] = registry["id"]
						}
					}, func(container map[string]any) {
						container["autoscaling"].(map[string]any)["initial_cooldown_period_seconds"] = 60
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config:             testAccContainerConfigWithScaledObjectSettings(testName, 1, settings),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Removing the settings from the config clears them.
			{
				Config: testAccContainerConfigWithScaledObjectSettings(testName, 1, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryContainerExists(address),
					resource.TestCheckNoResourceAttr(address, "autoscaling.initial_cooldown_period_seconds"),
					resource.TestCheckNoResourceAttr(address, "autoscaling.idle_replica_count"),
					resource.TestCheckNoResourceAttr(address, "autoscaling.advanced_yaml"),
				),
			},
		},
	})
}

func TestAcc_ContainerAutoscalingScaledObjectSettingsInvalid(t *testing.T) {
	t.Parallel()
	testName := "container-autoscaling-scaled-object-settings-invalid"
	testCases := []struct {
		TestName            string
		MinRunningInstances int
		Settings            string
		ExpectError         *regexp.Regexp
	}{
		{"idle_replica_count_not_below_min", 0, "idle_replica_count = 0", regexp.MustCompile(`strictly less than\s+min_running_instances`)},
		{"idle_replica_count_not_zero", 2, "idle_replica_count = 1", regexp.MustCompile(`value must be one of`)},
		{"initial_cooldown_out_of_range", 1, "initial_cooldown_period_seconds = 3601", regexp.MustCompile(`value must be between 0\s+and 3600`)},
		{"blank_advanced_yaml", 1, `advanced_yaml = " "`, regexp.MustCompile(`must not be blank`)},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccContainerConfigWithScaledObjectSettings(testName, tc.MinRunningInstances, tc.Settings),
						ExpectError: tc.ExpectError,
					},
				},
			})
		})
	}
}

func testAccContainerConfigWithScaledObjectSettings(testName string, minRunningInstances int, settings string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  min_running_instances = %d
  max_running_instances = 5
  healthchecks = {}
  autoscaling = {
    %s
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
		minRunningInstances, settings,
	)
}
