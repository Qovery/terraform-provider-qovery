//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/internal/domain"
	"github.com/qovery/terraform-provider-qovery/internal/domain/advanced_settings"
)

// Covers build_settings next to the build.* advanced settings it replaces: the legacy keys keep working
// without the block, migrating to the block drops them, out-of-band changes show up as drift, and
// removing the block resets the build settings to their defaults.
func TestAcc_Job_BuildSettings(t *testing.T) {
	t.Parallel()
	testName := "job-build-settings"
	name := generateTestName(testName)
	var jobID string

	blockConfig := testAccJobBuildSettingsConfig(testName, name, `build_settings = {
    timeout_max_sec = 3600
  }`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy("qovery_job.test"),
		Steps: []resource.TestStep{
			// Legacy build.* key in advanced_settings_json, without the block
			{
				Config: testAccJobBuildSettingsConfig(testName, name, `advanced_settings_json = jsonencode({
    "build.timeout_max_sec" = 3000
  })`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccStoreResourceID("qovery_job.test", &jobID),
					resource.TestCheckNoResourceAttr("qovery_job.test", "build_settings.timeout_max_sec"),
					testAccCheckJobBuildTimeout(&jobID, 3000),
				),
			},
			// Migrating to the block: the build.* key advanced_settings_json kept from state is dropped
			{
				Config: blockConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_job.test", "build_settings.timeout_max_sec", "3600"),
					resource.TestCheckResourceAttr("qovery_job.test", "build_settings.cpu_max_in_milli", "4000"),
					resource.TestCheckResourceAttr("qovery_job.test", "advanced_settings_json", "{}"),
					testAccCheckJobBuildTimeout(&jobID, 3600),
				),
			},
			// A change made outside Terraform shows up on refresh
			{
				PreConfig: func() {
					err := advanced_settings.NewServiceAdvancedSettingsService(qoveryAPIClient.GetConfig()).
						UpdateServiceAdvancedSettings(domain.JOB, jobID, `{"build.timeout_max_sec":5000}`)
					if err != nil {
						t.Fatalf("failed to change the job build timeout out of band: %s", err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check:              resource.TestCheckResourceAttr("qovery_job.test", "build_settings.timeout_max_sec", "5000"),
			},
			// The next apply restores the configured value
			{
				Config: blockConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_job.test", "build_settings.timeout_max_sec", "3600"),
					testAccCheckJobBuildTimeout(&jobID, 3600),
				),
			},
			// Removing the block resets the build settings to their defaults
			{
				Config: testAccJobBuildSettingsConfig(testName, name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qovery_job.test", "build_settings.timeout_max_sec"),
					testAccCheckJobBuildTimeout(&jobID, 1800),
				),
			},
		},
	})
}

func testAccStoreResourceID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		*id = rs.Primary.ID
		return nil
	}
}

func testAccCheckJobBuildTimeout(jobID *string, expected int32) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, _, err := qoveryAPIClient.JobConfigurationAPI.GetJobAdvancedSettings(context.TODO(), *jobID).Execute()
		if err != nil {
			return fmt.Errorf("failed to read advanced settings of job %s: %w", *jobID, err)
		}
		if got := settings.GetBuildTimeoutMaxSec(); got != expected {
			return fmt.Errorf("expected build.timeout_max_sec %d on job %s, got %d", expected, *jobID, got)
		}
		return nil
	}
}

func testAccJobBuildSettingsConfig(testName, name, buildConfig string) string {
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

  %s
}
`,
		testAccEnvironmentDefaultConfig(testName),
		testAccContainerRegistryDefaultConfig(testName),
		name,
		jobImageName,
		jobImageTag,
		buildConfig,
	)
}
