package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_AWSCredentialsDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAwsCredentialsDataSourceConfig(
					getTestAWSCredentialsID(),
					getTestOrganizationID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_aws_credentials.test", "id", getTestAWSCredentialsID()),
					resource.TestCheckResourceAttr("data.qovery_aws_credentials.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_aws_credentials.test", "name", "bbenamira"),
				),
			},
		},
	})
}

func testAccAwsCredentialsDataSourceConfig(credentialsID string, organizationID string) string {
	return fmt.Sprintf(`
data "qovery_aws_credentials" "test" {
  id = "%s"
  organization_id = "%s"
}
`, credentialsID, organizationID,
	)
}
