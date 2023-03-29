//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

const (
	jobImageName = "terraform-provider-tests-container"
	jobImageTag  = "1.0.0"
)

func TestAcc_Job(t *testing.T) {
	t.Parallel()
	testName := "job"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy("qovery_job.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEnvironmentDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
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
			// Update name
			{
				Config: testAccJobDefaultConfigWithName(
					testName,
					fmt.Sprintf("%s-updated", testName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryJobExists("qovery_job.test"),
					resource.TestCheckResourceAttr("data.qovery_job.test", "id", getTestJobID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_job.test", "name", fmt.Sprintf("%s-updated", testName)),
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
			// Check Import
			{
				ResourceName:      "qovery_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccJobDefaultConfig(testName string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_job" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName),
	)
}

func testAccJobDefaultConfigWithName(testName string, name string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_job" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), name,
	)
}

func testAccQoveryJobExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("job not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("job.id not found")
		}

		_, err := qoveryServices.Job.Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryJobDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("environment not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("environment.id not found")
		}

		_, err := qoveryServices.Job.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found job but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted job: %s", err.Error())
		}
		return nil
	}
}
