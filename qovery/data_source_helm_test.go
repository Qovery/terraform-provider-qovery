//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_HelmDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccHelmDataSourceConfig(
					getTestHelmID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_helm.test", "id", getTestHelmID()),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "name", "test-helm"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "arguments.0", "--atomic"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "arguments.1", "--debug"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "arguments.2", "--wait"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "source.helm_repository.chart_version", "6.4.0"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "source.helm_repository.chart_name", "bitnamicharts/argo-workflows"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "values_override.set.test1", "value1"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "values_override.set_string.test2", "value2"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "auto_preview", "false"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "ports.myservice-p80.internal_port", "80"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "ports.myservice-p80.external_port", "443"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "ports.myservice-p80.protocol", "HTTP"),
					resource.TestCheckResourceAttr("data.qovery_helm.test", "ports.myservice-p80.service_name", "myservice"),
				),
			},
		},
	})
}

func testAccHelmDataSourceConfig(helmID string) string {
	return fmt.Sprintf(`
data "qovery_helm" "test" {
  id = "%s"
}
`, helmID,
	)
}
