//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

// serviceResources is the cpu and memory of a service with their limits in advanced_settings_json.
type serviceResources struct {
	cpu      int
	memory   int
	cpuLimit int
	ramLimit int
}

// resourceLimitsSteps walks a service through request and limit changes that each pass in one apply.
// q-core checks the service edit against the stored limits and the advanced settings update against
// the stored requests: before the fix, step 2 failed with "Ram limit (512) must be greater than ram
// request (1024)", and step 3 failed whatever the order of the two calls.
var resourceLimitsSteps = []serviceResources{
	// 1. Create.
	{cpu: 500, memory: 512, cpuLimit: 500, ramLimit: 512},
	// 2. Raise both requests and both limits.
	{cpu: 1000, memory: 1024, cpuLimit: 1000, ramLimit: 1024},
	// 3. Raise memory above its stored limit while lowering the cpu limit below the stored cpu.
	{cpu: 250, memory: 2048, cpuLimit: 500, ramLimit: 2048},
	// 4. Lower both requests and both limits.
	{cpu: 250, memory: 512, cpuLimit: 250, ramLimit: 512},
}

func TestAcc_ApplicationResourceLimits(t *testing.T) {
	t.Parallel()
	testName := "application-resource-limits"
	resourceName := "qovery_application.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccRequireClusterResourceOvercommit(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApplicationDestroy(resourceName),
		Steps: testAccResourceLimitsTestSteps(resourceName, testAccQoveryApplicationExists(resourceName), func(r serviceResources) string {
			return testAccApplicationDefaultConfigWithResourceLimits(testName, r)
		}),
	})
}

func TestAcc_ContainerResourceLimits(t *testing.T) {
	t.Parallel()
	testName := "container-resource-limits"
	resourceName := "qovery_container.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccRequireClusterResourceOvercommit(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryContainerDestroy(resourceName),
		Steps: testAccResourceLimitsTestSteps(resourceName, testAccQoveryContainerExists(resourceName), func(r serviceResources) string {
			return testAccContainerDefaultConfigWithResourceLimits(testName, r)
		}),
	})
}

// testAccResourceLimitsTestSteps applies each of resourceLimitsSteps, then checks the last one plans
// no change.
func testAccResourceLimitsTestSteps(resourceName string, exists resource.TestCheckFunc, config func(serviceResources) string) []resource.TestStep {
	steps := make([]resource.TestStep, 0, len(resourceLimitsSteps)+1)
	for _, r := range resourceLimitsSteps {
		steps = append(steps, resource.TestStep{
			Config: config(r),
			Check: resource.ComposeAggregateTestCheckFunc(
				exists,
				resource.TestCheckResourceAttr(resourceName, "cpu", strconv.Itoa(r.cpu)),
				resource.TestCheckResourceAttr(resourceName, "memory", strconv.Itoa(r.memory)),
				testAccCheckAdvancedSettingsJSON(resourceName, advanced_settings.ResourcesOverrideLimitCpuInMilliKey, int32(r.cpuLimit)),
				testAccCheckAdvancedSettingsJSON(resourceName, advanced_settings.ResourcesOverrideLimitRamInMibKey, int32(r.ramLimit)),
			),
		})
	}
	steps = append(steps, resource.TestStep{
		Config:             config(resourceLimitsSteps[len(resourceLimitsSteps)-1]),
		PlanOnly:           true,
		ExpectNonEmptyPlan: false,
	})
	return steps
}

// testAccRequireClusterResourceOvercommit fails the test unless the test cluster lets services set
// custom RAM and CPU limits: q-core rejects resources.override.limit.* otherwise.
func testAccRequireClusterResourceOvercommit(t *testing.T) {
	t.Helper()
	settings, res, err := qoveryAPIClient.ClustersAPI.GetClusterAdvancedSettings(context.TODO(), getTestOrganizationID(), getTestClusterID()).Execute()
	if err != nil {
		t.Fatalf("failed to read the advanced settings of test cluster %s: %s", getTestClusterID(), err)
	}
	if res != nil && res.StatusCode >= 400 {
		t.Fatalf("failed to read the advanced settings of test cluster %s: HTTP %d", getTestClusterID(), res.StatusCode)
	}
	for _, key := range []string{"allow_service_ram_overcommit", "allow_service_cpu_overcommit"} {
		if enabled, _ := settings.AdditionalProperties[key].(bool); !enabled {
			t.Fatalf("test cluster %s must have %s enabled in its advanced settings: q-core rejects resources.override.limit.* otherwise", getTestClusterID(), key)
		}
	}
}

func testAccResourceLimitsAdvancedSettingsJSON(r serviceResources) string {
	return fmt.Sprintf(`jsonencode({"%s": %d, "%s": %d})`,
		advanced_settings.ResourcesOverrideLimitCpuInMilliKey, r.cpuLimit,
		advanced_settings.ResourcesOverrideLimitRamInMibKey, r.ramLimit,
	)
}

func testAccApplicationDefaultConfigWithResourceLimits(testName string, r serviceResources) string {
	return fmt.Sprintf(`
%s

resource "qovery_application" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
  build_mode = "DOCKER"
  dockerfile_path = "Dockerfile"
  cpu = %d
  memory = %d
  git_repository = {
    url = "%s"
    git_token_id = "%s"
  }
  healthchecks = {}
  advanced_settings_json = %s
}
`, testAccEnvironmentDefaultConfig(testName), generateTestName(testName), r.cpu, r.memory, applicationRepositoryURL, getTestQoverySandboxGitTokenID(), testAccResourceLimitsAdvancedSettingsJSON(r),
	)
}

func testAccContainerDefaultConfigWithResourceLimits(testName string, r serviceResources) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_container" "test" {
  environment_id = qovery_environment.test.id
  registry_id = qovery_container_registry.test.id
  name = "%s"
  image_name = "%s"
  tag = "%s"
  cpu = %d
  memory = %d
  healthchecks = {}
  advanced_settings_json = %s
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName), containerImageName, containerTag, r.cpu, r.memory, testAccResourceLimitsAdvancedSettingsJSON(r),
	)
}
