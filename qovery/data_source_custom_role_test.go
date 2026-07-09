//go:build integration && !unit

package qovery_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Deliberately NOT t.Parallel(): see TestAcc_CustomRole — this test also creates/deletes a
// custom role, which races q-core's unlocked project_role_permission matrix maintenance and
// 500s concurrently-running project-creating tests.
func TestAcc_CustomRoleDataSource(t *testing.T) {
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
					// so the returned project_permissions set must be non-empty (not "0").
					resource.TestMatchResourceAttr("data.qovery_custom_role.test", "project_permissions.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
		},
	})
}
