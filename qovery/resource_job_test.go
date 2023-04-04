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
	jobImageName                     = "terraform-provider-tests-job"
	jobImageTag                      = "1.0.0"
	jobScheduleCronString            = "*/2 * * * *"
	jobScheduleCronCommandEntrypoint = ""
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
				Config: testAccJobDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryJobExists("qovery_job.test"),
					resource.TestCheckResourceAttr("qovery_job.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_job.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_job.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_job.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_job.test", "max_duration_seconds", "300"),
					resource.TestCheckResourceAttr("qovery_job.test", "max_nb_restart", "0"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "port"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.image.name", jobImageName),
					resource.TestCheckResourceAttr("qovery_job.test", "source.image.tag", jobImageTag),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_start"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_stop"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_delete"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.schedule", "*/2 * * * *"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.cronjob.schedule.command.entrypoint"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.cronjob.schedule.command.arguments"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "internal_host"),
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
					resource.TestCheckResourceAttr("qovery_job.test", "name", fmt.Sprintf("%s-updated", testName)),
					resource.TestCheckResourceAttr("qovery_job.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("qovery_job.test", "cpu", "500"),
					resource.TestCheckResourceAttr("qovery_job.test", "memory", "512"),
					resource.TestCheckResourceAttr("qovery_job.test", "max_duration_seconds", "300"),
					resource.TestCheckResourceAttr("qovery_job.test", "max_nb_restart", "0"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "port"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.image.name", jobImageName),
					resource.TestCheckResourceAttr("qovery_job.test", "source.image.tag", jobImageTag),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_start"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_stop"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.on_delete"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.schedule", "*/2 * * * *"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.cronjob.schedule.command.entrypoint"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "schedule.cronjob.schedule.command.arguments"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "internal_host"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_job.test",
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

  source = {
    image = {
      registry_id = qovery_container_registry.test.id
      name = "%s"
      tag = "%s"
    }
  }

  schedule = {
    cronjob = {
      schedule = "%s"
        command = {
          entrypoint = "%s"
          arguments = []
        }
      }
    }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), generateTestName(testName),
		jobImageName, jobImageTag, jobScheduleCronString, jobScheduleCronCommandEntrypoint)
}

func testAccJobDefaultConfigWithName(testName string, name string) string {
	return fmt.Sprintf(`
%s

%s

resource "qovery_job" "test" {
  environment_id = qovery_environment.test.id
  name = "%s"

  source = {
    image = {
      registry_id = qovery_container_registry.test.id
      name = "%s"
      tag = "%s"
    }
  }

  schedule = {
    cronjob = {
      schedule = "%s"
        command = {
          entrypoint = "%s"
          arguments = []
        }
      }
    }
}
`, testAccEnvironmentDefaultConfig(testName), testAccContainerRegistryDefaultConfig(testName), name,
		jobImageName, jobImageTag, jobScheduleCronString, jobScheduleCronCommandEntrypoint)
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
			return fmt.Errorf("job not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("job.id not found")
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
