//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_CustomRole(t *testing.T) {
	t.Parallel()
	roleName := generateTestName("custom-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryCustomRoleDestroy("qovery_custom_role.test"),
		Steps: []resource.TestStep{
			// Step 1: reserved name rejected at plan time (placed first so no state dangles)
			{
				Config:      testAccCustomRoleConfigNamed("admin", "MANAGER"),
				ExpectError: regexp.MustCompile(`reserved`),
			},
			// Step 2: create with a declared project permission (4 env types)
			{
				Config: testAccCustomRoleConfigNamed(roleName, "DEPLOYER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "name", roleName),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.#", "1"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.0.permissions.#", "4"),
				),
			},
			// Step 3: update a permission in place (PRODUCTION DEPLOYER -> MANAGER)
			{
				Config: testAccCustomRoleConfigNamed(roleName, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.#", "1"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "project_permissions.0.permissions.#", "4"),
				),
			},
			// Step 4: dropping the `description` attribute from config must apply cleanly.
			// The server persists an omitted description as "" (not null), so a plain
			// Optional attribute failed with "Provider produced inconsistent result after
			// apply" (null in config vs "" from the API). description is Optional+Computed
			// with UseStateForUnknown, so removing it keeps the prior value; a clean apply
			// plus the framework's empty post-apply plan is the regression guard.
			{
				Config: testAccCustomRoleConfigNoDescription(roleName, "MANAGER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryCustomRoleExists("qovery_custom_role.test"),
					resource.TestCheckResourceAttr("qovery_custom_role.test", "description", "acceptance test role"),
				),
			},
			// Step 5: adding an unrelated project must NOT produce a diff on the role
			// (THE perpetual-diff regression test: the server matrix now includes the new
			// project with default perms, which the Read must filter out).
			{
				Config:             testAccCustomRoleConfigWithExtraProject(roleName, "MANAGER"),
				Check:              testAccQoveryCustomRoleExists("qovery_custom_role.test"),
				ExpectNonEmptyPlan: false,
			},
			// Step 6: import keeps non-default entries (id format: "org_id,role_id")
			{
				ResourceName:      "qovery_custom_role.test",
				ImportState:       true,
				ImportStateIdFunc: testAccCustomRoleImportStateID("qovery_custom_role.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCustomRoleConfigNamed(roleName string, prodPermission string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"
  description     = "acceptance test role"

  project_permissions = [
    {
      project_id = "%s"
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "%s" },
      ]
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID(), prodPermission)
}

func testAccCustomRoleConfigNoDescription(roleName string, prodPermission string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test" {
  organization_id = "%s"
  name            = "%s"

  project_permissions = [
    {
      project_id = "%s"
      permissions = [
        { environment_type = "DEVELOPMENT", permission = "MANAGER" },
        { environment_type = "PREVIEW", permission = "MANAGER" },
        { environment_type = "STAGING", permission = "DEPLOYER" },
        { environment_type = "PRODUCTION", permission = "%s" },
      ]
    }
  ]
}
`, getTestOrganizationID(), roleName, getTestProjectID(), prodPermission)
}

func testAccCustomRoleConfigWithExtraProject(roleName string, prodPermission string) string {
	return testAccCustomRoleConfigNamed(roleName, prodPermission) + fmt.Sprintf(`
resource "qovery_project" "extra" {
  organization_id = "%s"
  name            = "%s-extra"
}
`, getTestOrganizationID(), roleName)
}

func testAccCustomRoleImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("custom role not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.ID), nil
	}
}

func testAccQoveryCustomRoleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("custom role not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("custom_role.id not found")
		}

		_, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryCustomRoleDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("custom role not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("custom_role.id not found")
		}

		_, err := qoveryServices.CustomRole.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found custom role but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted custom role: %s", err.Error())
		}
		return nil
	}
}
