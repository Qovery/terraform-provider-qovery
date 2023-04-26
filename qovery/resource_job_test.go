//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/qovery"
)

const (
	jobImageName          = "terraform-provider-tests-job"
	jobImageTag           = "1.0.0"
	jobScheduleCronString = "*/2 * * * *"
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
				Config: getJobConfigFromModel(
					testName,
					qovery.Job{
						Name:               qovery.FromString(generateTestName(testName)),
						AutoPreview:        qovery.FromBool(false),
						CPU:                qovery.FromInt32(500),
						Memory:             qovery.FromInt32(512),
						MaxDurationSeconds: qovery.FromUInt32(300),
						MaxNbRestart:       qovery.FromUInt32(0),
						Port:               qovery.FromInt32Pointer(nil),
						Source: &qovery.JobSource{
							Image: &qovery.Image{
								Name: qovery.FromString(jobImageName),
								Tag:  qovery.FromString(jobImageTag),
							},
						},
						Schedule: &qovery.JobSchedule{
							CronJob: &qovery.JobScheduleCron{
								Schedule: qovery.FromString(jobScheduleCronString),
								Command: qovery.ExecutionCommand{
									Entrypoint: qovery.FromString("test.sh"),
									Arguments:  []types.String{qovery.FromString("arg1"), qovery.FromString("arg2")},
								},
							},
						},
						EnvironmentVariables: types.Set{Null: true},
						Secrets:              types.Set{Null: true},
					},
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
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.entrypoint", "test.sh"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.arguments.0", "arg1"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.arguments.1", "arg2"),
					resource.TestMatchTypeSetElemNestedAttrs("qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "internal_host"),
				),
			},
			// Update name
			{
				Config: getJobConfigFromModel(
					testName,
					qovery.Job{
						Name:               qovery.FromString(generateTestName(testName) + "-updated"),
						AutoPreview:        qovery.FromBool(false),
						CPU:                qovery.FromInt32(500),
						Memory:             qovery.FromInt32(512),
						MaxDurationSeconds: qovery.FromUInt32(300),
						MaxNbRestart:       qovery.FromUInt32(0),
						Port:               qovery.FromInt32Pointer(nil),
						Source: &qovery.JobSource{
							Image: &qovery.Image{
								Name: qovery.FromString(jobImageName),
								Tag:  qovery.FromString(jobImageTag),
							},
						},
						Schedule: &qovery.JobSchedule{
							CronJob: &qovery.JobScheduleCron{
								Schedule: qovery.FromString(jobScheduleCronString),
								Command: qovery.ExecutionCommand{
									Entrypoint: qovery.FromString("test.sh"),
									Arguments:  []types.String{qovery.FromString("arg1"), qovery.FromString("arg2")},
								},
							},
						},
						EnvironmentVariables: types.Set{Null: true},
						Secrets:              types.Set{Null: true},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					testAccQoveryEnvironmentExists("qovery_environment.test"),
					testAccQoveryContainerRegistryExists("qovery_container_registry.test"),
					testAccQoveryJobExists("qovery_job.test"),
					resource.TestCheckResourceAttr("qovery_job.test", "name", fmt.Sprintf("%s-updated", generateTestName(testName))),
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
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.entrypoint", "test.sh"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.arguments.0", "arg1"),
					resource.TestCheckResourceAttr("qovery_job.test", "schedule.cronjob.command.arguments.1", "arg2"),
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

func getJobConfigFromModel(testName string, job qovery.Job) string {
	tmpl_model := struct {
		EnvironmentStr       string
		ContainerRegistryStr string
		Job                  qovery.Job
	}{
		EnvironmentStr:       testAccEnvironmentDefaultConfig(testName),
		ContainerRegistryStr: testAccContainerRegistryDefaultConfig(testName),
		Job:                  job,
	}

	tmpl, err := template.New("getJobConfigFromModel").Parse(`
{{ .EnvironmentStr }}

{{ .ContainerRegistryStr }}

resource "qovery_job" "test" {
  environment_id = qovery_environment.test.id
  name = {{ .Job.Name.String }}

  cpu = {{ .Job.CPU }}
  memory = {{ .Job.Memory }}
  max_duration_seconds = {{ .Job.MaxDurationSeconds }}
  max_nb_restart = {{ .Job.MaxNbRestart }}
  auto_preview = {{ .Job.AutoPreview }}

  {{ if not .Job.EnvironmentVariables.IsNull }}
  environment_variables = {{ .Job.EnvironmentVariables.String }}
  {{ end }}

  {{ if not .Job.Secrets.IsNull }}	
  secrets = {{ .Job.Secrets.String }}	
  {{ end }}

  {{ with .Job.Source }}	
  source = {
	{{ with .Image }}	
    image = {
      registry_id = qovery_container_registry.test.id
      name = {{ .Name.String }}
      tag = {{ .Tag.String }}
    }
    {{ end }}
	{{ with .Docker }}	
	docker = {
        {{ with .DockerDockerFilePath }}	
		dockerfile_path = {{ .String }}
        {{ end }}
		{{ with .GitRepository }}
		git_repository = {
			{{ with .Url }}
        	url = {{ .String }}
			{{ end }}
			{{ with .Branch }}
        	branch = {{ .String }}
			{{ end }}
			{{ with .RootPath }}	
        	root_path = {{ .String }}
			{{ end }}
        }
		{{ end }}
	}
    {{ end }}
  }
  {{ end }}

  {{ with .Job.Schedule }}	
  schedule = {
	{{ with .OnStart }}
    on_start = {
      {{ with .Entrypoint }}
	  entrypoint = {{ .String }}
	  {{ end }}
	  {{ with .Command.Arguments }}
        arguments = [{{ range $i, $a := . }}{{ if $i }}, {{ end }}{{ $a.String }}{{ end }}]
	  {{ end }}
    }
    {{ end }}
    {{ with .OnStop }}
    on_stop = {
      {{ with .Entrypoint }}
	  entrypoint = {{ .String }}
	  {{ end }}
	  {{ with .Command.Arguments }}
        arguments = [{{ range $i, $a := . }}{{ if $i }}, {{ end }}{{ $a.String }}{{ end }}]
	  {{ end }}
    }
    {{ end }}
    {{ with .OnDelete }}
    on_delete = {
      {{ with .Entrypoint }}
	  entrypoint = {{ .String }}
	  {{ end }}
	  {{ with .Command.Arguments }}
        arguments = [{{ range $i, $a := . }}{{ if $i }}, {{ end }}{{ $a.String }}{{ end }}]
	  {{ end }}
    }
    {{ end }}
	{{ with .CronJob }}	
    cronjob = {
      schedule = {{ .Schedule.String }}
      command = {
        {{ with .Command.Entrypoint }}
        entrypoint = {{ .String }}
        {{ end }}
		{{ with .Command.Arguments }}
        arguments = [{{ range $i, $a := . }}{{ if $i }}, {{ end }}{{ $a.String }}{{ end }}]
        {{ end }}
      }
    }
    {{ end }}
  }
  {{ end }}
}
`)

	var jobConfigStr bytes.Buffer
	err = tmpl.Execute(&jobConfigStr, tmpl_model)
	if err != nil {
		return ""
	}

	return jobConfigStr.String()
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
