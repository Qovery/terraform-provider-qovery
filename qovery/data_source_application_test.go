package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ApplicationDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccApplicationDataSourceConfig(
					getTestApplicationID(),
					getTestEnvironmentID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_application.test", "id", getTestApplicationID()),
					resource.TestCheckResourceAttr("data.qovery_application.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_application.test", "name", "http-server"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("data.qovery_application.test", "git_repository.branch", "main"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "git_repository.root_path", "/"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "build_mode", "DOCKER"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "dockerfile_path", "Dockerfile"),
					resource.TestCheckNoResourceAttr("data.qovery_application.test", "buildpack_language"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "cpu", "500"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "memory", "512"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "auto_preview", "false"),
					resource.TestCheckNoResourceAttr("data.qovery_application.test", "storage.0"),
					resource.TestCheckNoResourceAttr("data.qovery_application.test", "ports.0"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_APPLICATION_VARIABLE",
						"value": "MY_TERRAFORM_APPLICATION_VARIABLE_VALUE",
						"scope": "APPLICATION",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_ENVIRONMENT_VARIABLE",
						"value": "MY_TERRAFORM_ENVIRONMENT_VARIABLE_VALUE",
						"scope": "ENVIRONMENT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_PROJECT_VARIABLE",
						"value": "MY_TERRAFORM_PROJECT_VARIABLE_VALUE",
						"scope": "PROJECT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_APPLICATION_SECRET",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckResourceAttr("data.qovery_application.test", "state", "RUNNING"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig(credentialsID string, environmentID string) string {
	return fmt.Sprintf(`
data "qovery_application" "test" {
  id = "%s"
  environment_id = "%s"
}
`, credentialsID, environmentID,
	)
}
