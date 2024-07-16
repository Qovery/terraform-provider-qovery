//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_AnnotationsGroupDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAnnotationsGroupDataSourceConfig(
					getTestAnnotationsGroupID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "id", getTestAnnotationsGroupID()),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "annotations.key1", "value1"),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "annotations.key2", "value2"),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "scopes.0", "DEPLOYMENTS"),
					resource.TestCheckResourceAttr("data.qovery_annotations_group.test", "scopes.1", "SERVICES"),
				),
			},
		},
	})
}

func testAccAnnotationsGroupDataSourceConfig(annotationsGroupID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_annotations_group" "test" {
  id = "%s"
  organization_id = "%s"
}
`, annotationsGroupID, organizationID,
	)
}
