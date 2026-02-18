//go:build integration && !unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ClusterDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccClusterDataSourceConfig(
					getTestOrganizationID(),
					getTestClusterID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "id", getTestClusterID()),
					resource.TestCheckResourceAttr("data.qovery_cluster.test", "organization_id", getTestOrganizationID()),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig(organizationID string, clusterID string) string {
	return fmt.Sprintf(`
data "qovery_cluster" "test" {
  id = "%s"
  organization_id = "%s"
}
`, clusterID, organizationID,
	)
}
