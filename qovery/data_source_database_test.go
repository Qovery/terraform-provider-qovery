package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_DatabaseDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDatabaseDataSourceConfig(
					getTestDatabaseID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_database.test", "id", getTestDatabaseID()),
					resource.TestCheckResourceAttr("data.qovery_database.test", "environment_id", getTestEnvironmentID()),
					resource.TestCheckResourceAttr("data.qovery_database.test", "name", "redis"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "type", "REDIS"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "version", "6"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "mode", "CONTAINER"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "accessibility", "PRIVATE"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "cpu", "250"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "memory", "256"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "storage", "10"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "state", "RUNNING"),
					resource.TestCheckResourceAttr("data.qovery_database.test", "port", "6379"),
					resource.TestCheckResourceAttrSet("data.qovery_database.test", "external_host"),
					resource.TestCheckResourceAttrSet("data.qovery_database.test", "internal_host"),
					resource.TestCheckResourceAttrSet("data.qovery_database.test", "login"),
					resource.TestCheckResourceAttrSet("data.qovery_database.test", "password"),
				),
			},
		},
	})
}

func testAccDatabaseDataSourceConfig(databaseID string) string {
	return fmt.Sprintf(`
data "qovery_database" "test" {
  id = "%s"
}
`, databaseID,
	)
}
