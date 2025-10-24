//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ScalewayCredentialsDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAwsCredentialsDataSourceConfig(
					getTestScalewayCredentialsID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_scaleway_credentials.test", "id", getTestScalewayCredentialsID()),
					resource.TestCheckResourceAttr("data.qovery_scaleway_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_scaleway_credentials.test", "name", "terraform-provider-test-scaleway"),
				),
			},
		},
	})
}

func testAccAwsCredentialsDataSourceConfig(credentialsID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_scaleway_credentials" "test" {
  id = "%s"
  organization_id = "%s"
}
`, credentialsID, organizationID,
	)
}
