//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ContainerDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccContainerDataSourceConfig(
					getTestContainerID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_container.test", "id", getTestContainerID()),
					resource.TestCheckResourceAttr("data.qovery_container.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_container.test", "registry_id", getTestContainerRegistryID()),
					resource.TestCheckResourceAttr("data.qovery_container.test", "name", "test-container"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "image_name", containerImageName),
					resource.TestCheckResourceAttr("data.qovery_container.test", "tag", containerTag),
					resource.TestCheckResourceAttr("data.qovery_container.test", "cpu", "500"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "memory", "512"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "min_running_instances", "1"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "max_running_instances", "1"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "storage.0.id", "c176ee2a-de9a-418c-9856-39071772fcba"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "storage.0.type", "FAST_SSD"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "storage.0.size", "1"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "storage.0.mount_point", "/mnt/images"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.id", "5cedc809-8598-45c4-adeb-179403c01f80"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.name", "default-port"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.internal_port", "80"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.external_port", "443"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "ports.0.publicly_accessible", "true"),
					resource.TestMatchTypeSetElemNestedAttrs("data.qovery_container.test", "built_in_environment_variables.*", map[string]*regexp.Regexp{
						"key": regexp.MustCompile(`^QOVERY_`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_container.test", "environment_variables.*", map[string]string{
						"key":   "MY_TERRAFORM_CONTAINER_VARIABLE",
						"value": "MY_TERRAFORM_CONTAINER_VARIABLE_VALUE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.qovery_container.test", "secrets.*", map[string]string{
						"key": "MY_TERRAFORM_CONTAINER_SECRET",
					}),
					resource.TestCheckResourceAttr("data.qovery_container.test", "external_host", "zc4425337-z92544d94-gtw.zc531a994.rustrocks.cloud"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "internal_host", "container-za7d391bf"),
					resource.TestCheckResourceAttr("data.qovery_container.test", "state", "STOPPED"),
				),
			},
		},
	})
}

func testAccContainerDataSourceConfig(containerID string) string {
	return fmt.Sprintf(`
data "qovery_container" "test" {
  id = "%s"
}
`, containerID,
	)
}
