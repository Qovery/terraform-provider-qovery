package qovery_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/qovery/terraform-provider-qovery/qovery"
)

func TestAcc_JobGitToken(t *testing.T) {
	t.Parallel()
	testName := "job-gittoken"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryJobDestroy("qovery_job.test"),
		Steps: []resource.TestStep{
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
							Docker: &qovery.Docker{
								GitRepository: qovery.GitRepository{
									Url:        qovery.FromString("https://github.com/Qovery/test_http_server.git"),
									Branch:     qovery.FromString("master"),
									RootPath:   qovery.FromString("/"),
									GitTokenId: qovery.FromString(getTestQoverySandboxGitTokenID()),
								},
								DockerFilePath: qovery.FromString("./Dockerfile"),
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
						EnvironmentVariables: generateVariableSet("key1", ""),
						Secrets:              generateVariableSet("secretkey1", ""),
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
					resource.TestCheckResourceAttr("qovery_job.test", "source.docker.dockerfile_path", "./Dockerfile"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.docker.git_repository.url", "https://github.com/Qovery/test_http_server.git"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.docker.git_repository.branch", "master"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.docker.git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("qovery_job.test", "source.docker.git_repository.git_token_id", getTestQoverySandboxGitTokenID()),
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
		},
	})
}
