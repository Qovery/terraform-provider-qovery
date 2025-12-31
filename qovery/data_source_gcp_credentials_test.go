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
	testName := "gcp-credentials-data-source"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create credentials first, then read via data source
			{
				Config: testAccGCPCredentialsDataSourceConfig(
					testName,
					getTestGCPCredentials(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.qovery_gcp_credentials.test", "id",
						"qovery_gcp_credentials.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.qovery_gcp_credentials.test", "organization_id",
						"qovery_gcp_credentials.test", "organization_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.qovery_gcp_credentials.test", "name",
						"qovery_gcp_credentials.test", "name",
					),
				),
			},
		},
	})
}

func testAccGCPCredentialsDataSourceConfig(testName string, gcpCredentials string) string {
	return fmt.Sprintf(`
resource "qovery_gcp_credentials" "test" {
  organization_id = "%s"
  name = "%s"
  gcp_credentials = "%s"
}

data "qovery_gcp_credentials" "test" {
  id              = qovery_gcp_credentials.test.id
  organization_id = qovery_gcp_credentials.test.organization_id
}
`, getTestOrganizationID(), generateTestName(testName), gcpCredentials,
	)
}
