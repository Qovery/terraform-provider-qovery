//go:build integration && !unit

package qovery_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_CustomRoleDataSource(t *testing.T) {
	t.Parallel()
	roleName := generateTestName("custom-role-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryCustomRoleDestroy("qovery_custom_role.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoleConfigNamed(roleName, "VIEWER") + `
data "qovery_custom_role" "test" {
  organization_id = qovery_custom_role.test.organization_id
  id              = qovery_custom_role.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.qovery_custom_role.test", "name", roleName),
					// The data source returns the FULL matrix (every project of the org),
					// so at minimum the declared project must be present.
					resource.TestCheckResourceAttrSet("data.qovery_custom_role.test", "project_permissions.#"),
				),
			},
		},
	})
}
