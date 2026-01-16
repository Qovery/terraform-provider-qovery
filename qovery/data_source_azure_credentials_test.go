//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_AzureCredentialsDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAzureCredentialsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_azure_credentials.test", "id", getTestAzureCredentialsID()),
					resource.TestCheckResourceAttr("data.qovery_azure_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttrSet("data.qovery_azure_credentials.test", "name"),
					resource.TestCheckResourceAttrSet("data.qovery_azure_credentials.test", "azure_subscription_id"),
					resource.TestCheckResourceAttrSet("data.qovery_azure_credentials.test", "azure_tenant_id"),
					resource.TestCheckResourceAttrSet("data.qovery_azure_credentials.test", "azure_application_id"),
					resource.TestCheckResourceAttrSet("data.qovery_azure_credentials.test", "azure_application_object_id"),
				),
			},
		},
	})
}

func testAccAzureCredentialsDataSourceConfig() string {
	return fmt.Sprintf(`
data "qovery_azure_credentials" "test" {
  id              = "%s"
  organization_id = "%s"
}
`, getTestAzureCredentialsID(), getTestOrganizationID(),
	)
}
