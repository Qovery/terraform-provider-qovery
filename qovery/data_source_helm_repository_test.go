//go:build integration && !unit
// +build integration,!unit

package qovery_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_HelmRepositoryDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccHelmRepositoryDataSourceConfig(
					getTestOrganizationID(),
					getTestHelmRepositoryID(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "id", getTestHelmRepositoryID()),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "name", "Terraform Provider Tests"),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "kind", "OCI_DOCKER_HUB"),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "url", "oci://registry-1.docker.io"),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "description", "Helm Repository used for terraform tests."),
					resource.TestCheckResourceAttr("data.qovery_helm_repository.test", "skip_tls_verification", "false"),
				),
			},
		},
	})
}

func testAccHelmRepositoryDataSourceConfig(orgID string, helmID string) string {
	return fmt.Sprintf(`
data "qovery_helm_repository" "test" {
  id              = "%s"
  organization_id = "%s"
}
`, helmID, orgID,
	)
}
