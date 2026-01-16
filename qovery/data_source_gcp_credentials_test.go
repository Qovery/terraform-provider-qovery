//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_GcpCredentialsDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccGCPCredentialsDataSourceConfig(
					getTestGCPCredentialsID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_gcp_credentials.test", "id", getTestGCPCredentialsID()),
					resource.TestCheckResourceAttr("data.qovery_gcp_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_gcp_credentials.test", "name", "terraform-provider-test-gcp"),
				),
			},
		},
	})
}

func testAccGCPCredentialsDataSourceConfig(credentialsID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_gcp_credentials" "test" {
  id              = "%s"
  organization_id = "%s"
}
`, credentialsID, organizationID,
	)
}
