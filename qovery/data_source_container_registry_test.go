//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ContainerRegistryDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccContainerRegistryDataSourceConfig(
					getTestContainerRegistryID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_container_registry.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_container_registry.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_container_registry.test", "kind", "ECR"),
					resource.TestCheckResourceAttr("data.qovery_container_registry.test", "url", "https://default.com"),
					resource.TestCheckResourceAttr("data.qovery_container_registry.test", "description", "Container Registry used to run test for our terraform provider"),
				),
			},
		},
	})
}

func testAccContainerRegistryDataSourceConfig(containerRegistryID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_container_registry" "test" {
  id = "%s"
  organization_id = "%s"
}
`, containerRegistryID, organizationID,
	)
}
