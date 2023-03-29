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
					resource.TestCheckResourceAttr("data.qovery_job.test", "name", "test-job"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "auto_preview", "true"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "cpu", "500"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "memory", "512"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "max_duration_seconds", "23"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "max_nb_restart", "1"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "port", "5432"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.registry_id", getTestContainerRegistryID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.name", jobImageName),
					resource.TestCheckResourceAttr("data.qovery_job.test", "source.image.tag", jobImageTag),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_start"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_stop"),
					resource.TestCheckNoResourceAttr("data.qovery_job.test", "schedule.on_delete"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule", "*/2 * * * *"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.entrypoint", "/bin/sh -c"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.arguments.0", "timeout"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.arguments.1", "15s"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "schedule.cronjob.schedule.command.arguments.2", "yes"),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_job.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_CONTAINER_VARIABLE",
						"value": "MY_TERRAFORM_CONTAINER_VARIABLE_VALUE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_job.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_CONTAINER_SECRET",
					}),
					resource.TestCheckResourceAttr("data.qovery_job.test", "external_host", "zc4425337-z92544d94-gtw.zc531a994.rustrocks.cloud"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "internal_host", "job-za7d391bf"),
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
