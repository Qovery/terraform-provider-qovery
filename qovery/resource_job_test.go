//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
						EnvironmentVariables: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1"),
										"value": qovery.FromString(""),
									},
								},
							},
						},
						Secrets: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1"),
										"value": qovery.FromString(""),
									},
								},
							},
						},
						AdvancedSettingsJson: qovery.FromString("{\"deployment.termination_grace_period_seconds\":61}"),
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
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "",
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "internal_host"),
					resource.TestCheckResourceAttr("qovery_job.test", "advanced_settings_json", "{\"deployment.termination_grace_period_seconds\":61}"),
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
						EnvironmentVariables: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1"),
										"value": qovery.FromString("value1"),
									},
								},
							},
						},
						EnvironmentVariableAliases: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1_alias"),
										"value": qovery.FromString("key1"),
									},
								},
							},
						},
						EnvironmentVariableOverrides: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("environment_variable"),
										"value": qovery.FromString("override value"),
									},
								},
							},
						},
						Secrets: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1"),
										"value": qovery.FromString("secretvalue1"),
									},
								},
							},
						},
						SecretAliases: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1_alias"),
										"value": qovery.FromString("secretkey1"),
									},
								},
							},
						},
						SecretOverrides: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("environment_secret"),
										"value": qovery.FromString("override value"),
									},
								},
							},
						},
						AdvancedSettingsJson: qovery.FromString("{\"deployment.termination_grace_period_seconds\":61}"),
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
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "secretvalue1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secret_aliases.*", map[string]string{
						"key":   "secretkey1_alias",
						"value": "secretkey1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secret_overrides.*", map[string]string{
						"key":   "environment_secret",
						"value": "override value",
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "external_host"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "internal_host"),
					resource.TestCheckResourceAttr("qovery_job.test", "advanced_settings_json", "{\"deployment.termination_grace_period_seconds\":61}"),
				),
			},
			// Update variables
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
						EnvironmentVariables: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1"),
										"value": qovery.FromString("value1-updated"),
									},
								},
							},
						},
						EnvironmentVariableAliases: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1_alias_updated"),
										"value": qovery.FromString("key1"),
									},
								},
							},
						},
						EnvironmentVariableOverrides: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("environment_variable"),
										"value": qovery.FromString("override value updated"),
									},
								},
							},
						},
						Secrets: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1"),
										"value": qovery.FromString("secretvalue1-updated"),
									},
								},
							},
						},
						SecretAliases: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1_alias_updated"),
										"value": qovery.FromString("secretkey1"),
									},
								},
							},
						},
						SecretOverrides: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("environment_secret"),
										"value": qovery.FromString("override value updated"),
									},
								},
							},
						},
						AdvancedSettingsJson: qovery.FromString("{\"deployment.termination_grace_period_seconds\":61}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs("qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variable_aliases.*", map[string]string{
						"key":   "key1_alias_updated",
						"value": "key1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variable_overrides.*", map[string]string{
						"key":   "environment_variable",
						"value": "override value updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "secretvalue1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secret_aliases.*", map[string]string{
						"key":   "secretkey1_alias_updated",
						"value": "secretkey1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secret_overrides.*", map[string]string{
						"key":   "environment_secret",
						"value": "override value updated",
					}),
				),
			},
			// Delete variables
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
						EnvironmentVariables: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("key1"),
										"value": qovery.FromString("value1"),
									},
								},
							},
						},
						Secrets: types.Set{
							Elems: []attr.Value{
								types.Object{
									Attrs: map[string]attr.Value{
										"key":   qovery.FromString("secretkey1"),
										"value": qovery.FromString("secretvalue1"),
									},
								},
							},
						},
						AdvancedSettingsJson: qovery.FromString("{\"deployment.termination_grace_period_seconds\":61}"),
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchTypeSetElemNestedAttrs("qovery_job.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "environment_variable_aliases.0"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "environment_variable_overrides.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_job.test", "secrets.*", map[string]string{
						"key":   "secretkey1",
						"value": "secretvalue1",
					}),
					resource.TestCheckNoResourceAttr("qovery_job.test", "secret_aliases.0"),
					resource.TestCheckNoResourceAttr("qovery_job.test", "secret_overrides.0"),
				),
			},
			// Check Import
			{
				ResourceName:            "qovery_job.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secrets", "secret_aliases", "secret_overrides"},
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
		EnvironmentStr:       testAccEnvironmentDefaultConfigWithEnvironmentVariablesAndSecrets(testName, map[string]string{"environment_variable": "simple value"}, map[string]string{"environment_secret": "simple value"}),
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

  {{ if not .Job.EnvironmentVariableAliases.IsNull }}
  environment_variable_aliases = {{ .Job.EnvironmentVariableAliases.String }}
  {{ end }}

  {{ if not .Job.EnvironmentVariableOverrides.IsNull }}
  environment_variable_overrides = {{ .Job.EnvironmentVariableOverrides.String }}
  {{ end }}

  {{ if not .Job.Secrets.IsNull }}	
  secrets = {{ .Job.Secrets.String }}	
  {{ end }}

  {{ if not .Job.SecretAliases.IsNull }}
  secret_aliases = {{ .Job.SecretAliases.String }}
  {{ end }}

  {{ if not .Job.SecretOverrides.IsNull }}
  secret_overrides = {{ .Job.SecretOverrides.String }}
  {{ end }}

  healthchecks = {}

  {{ if not .Job.AdvancedSettingsJson.IsNull }}
  advanced_settings_json = jsonencode({{ .Job.AdvancedSettingsJson.Value }})
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
        {{ with .DockerFilePath }}	
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
			{{ with .GitTokenId }}
        	git_token_id = {{ .String }}
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
