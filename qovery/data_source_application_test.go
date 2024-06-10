//go:build integration && !unit
// +build integration,!unit

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
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_application.test", "id", getTestApplicationID()),
					resource.TestCheckResourceAttr("data.qovery_application.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_application.test", "name", "test-http-server"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "git_repository.url", applicationRepositoryURL),
					resource.TestCheckResourceAttr("data.qovery_application.test", "git_repository.branch", applicationBranch),
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
					resource.TestCheckResourceAttr("data.qovery_application.test", "ports.0.internal_port", "8000"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "ports.0.external_port", "443"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "ports.0.publicly_accessible", "true"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "ports.0.protocol", "HTTP"),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_APPLICATION_VARIABLE",
						"value": "MY_TERRAFORM_APPLICATION_VARIABLE_VALUE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_application.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_APPLICATION_SECRET",
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_application.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckResourceAttr("data.qovery_application.test", "custom_domains.0.domain", "example.com"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "custom_domains.0.validation_domain", "zc4425337-z99aa979e-gtw.zc531a994.rustrocks.cloud"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "custom_domains.0.status", "VALIDATION_PENDING"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "external_host", "zc4425337-z99aa979e-gtw.zc531a994.rustrocks.cloud"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "internal_host", "app-z20501d1f"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "advanced_settings_json", "{\"build.timeout_max_sec\":1700,\"deployment.lifecycle.pre_stop_exec_command\":[]}"),
					resource.TestCheckResourceAttr("data.qovery_application.test", "auto_deploy", "true"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig(applicationID string) string {
	return fmt.Sprintf(`
data "qovery_application" "test" {
  id = "%s"
}
`, applicationID,
	)
}
