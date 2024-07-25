//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_JobDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccJobDataSourceConfig(
					getTestJobID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_job.test", "id", getTestJobID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "name", "cron-job"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "icon_uri", "app://qovery-console/cron-job"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "auto_preview", "true"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "cpu", "100"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "memory", "256"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "max_duration_seconds", "300"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "max_nb_restart", "0"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "port"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.registry_id", getTestContainerRegistryID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.name", jobImageName),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.tag", jobImageTag),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_start"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_stop"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_delete"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule", "*/2 * * * *"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.entrypoint"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.arguments"),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "internal_host"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "advanced_settings_json", "{\"deployment.termination_grace_period_seconds\":61}"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "auto_deploy", "true"),
				),
			},
		},
	})
}

func testAccJobDataSourceConfig(jobID string) string {
	return fmt.Sprintf(`
data "qovery_job" "test" {
  id = "%s"
}
`, jobID,
	)
}
