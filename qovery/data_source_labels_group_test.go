//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_LabelsGroupDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccLabelsGroupDataSourceConfig(
					getTestLabelsGroupID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "id", getTestLabelsGroupID()),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.0.key", "key1"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.0.value", "value1"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.0.propagate_to_cloud_provider", "false"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.1.key", "key2"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.1.value", "value2"),
					resource.TestCheckResourceAttr("data.qovery_labels_group.test", "labels.1.propagate_to_cloud_provider", "true"),
				),
			},
		},
	})
}

func testAccLabelsGroupDataSourceConfig(labelsGroupID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_labels_group" "test" {
  id = "%s"
  organization_id = "%s"
}
`, labelsGroupID, organizationID,
	)
}
