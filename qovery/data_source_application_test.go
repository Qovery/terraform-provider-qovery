package qovery_test

import (
	"fmt"
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
					resource.TestCheckResourceAttr("data.qovery_application.test", "name", "MyTerraformApplication"),
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
						"value": "MY_TERRAFORM_APPLICATION_VALUE",
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
