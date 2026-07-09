//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

func TestAcc_OrganizationMember(t *testing.T) {
	// Deliberately not parallel: this test creates a custom role, and concurrent custom role
	// writes trigger a q-core race in the role/permission matrix (sporadic 500s in CI).
	testName := "organization-member"
	adminRoleID := getTestAdminRoleID(t)
	devopsRoleID := getTestOrganizationRoleID(t, "devops")
	email := fmt.Sprintf("%s@qovery-tf-acc-test.example.com", generateRandomName(testName))
	customRoleName := generateRandomName("member-custom-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryOrganizationMemberDestroy("qovery_organization_member.test", email),
		Steps: []resource.TestStep{
			// Create and Read testing (pending invitation)
			{
				Config: testAccOrganizationMemberConfig(email, adminRoleID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryOrganizationMemberExists("qovery_organization_member.test", email),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "email", email),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "role_id", adminRoleID),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "invitation_status", "PENDING"),
					resource.TestCheckResourceAttrSet("qovery_organization_member.test", "id"),
					resource.TestCheckNoResourceAttr("qovery_organization_member.test", "user_id"),
				),
			},
			// Update role on a pending invitation (delete + re-invite, id changes)
			{
				Config: testAccOrganizationMemberConfig(email, devopsRoleID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryOrganizationMemberExists("qovery_organization_member.test", email),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "email", email),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "role_id", devopsRoleID),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "invitation_status", "PENDING"),
				),
			},
			// Update role to a custom role (headline use case: role_id = qovery_custom_role.x.id)
			{
				Config: testAccOrganizationMemberConfigWithCustomRole(email, customRoleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryOrganizationMemberExists("qovery_organization_member.test", email),
					resource.TestCheckResourceAttrPair("qovery_organization_member.test", "role_id", "qovery_custom_role.test_member", "id"),
					resource.TestCheckResourceAttr("qovery_organization_member.test", "invitation_status", "PENDING"),
				),
			},
			// Import testing by organization_id,email
			{
				ResourceName:      "qovery_organization_member.test",
				ImportState:       true,
				ImportStateIdFunc: getOrganizationMemberImportStateId("qovery_organization_member.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganizationMemberConfig(email string, roleID string) string {
	return fmt.Sprintf(`
resource "qovery_organization_member" "test" {
  organization_id = "%s"
  email           = "%s"
  role_id         = "%s"
}
`, getTestOrganizationID(), email, roleID,
	)
}

func testAccOrganizationMemberConfigWithCustomRole(email string, customRoleName string) string {
	return fmt.Sprintf(`
resource "qovery_custom_role" "test_member" {
  organization_id = "%s"
  name            = "%s"
  description     = "custom role for organization member acceptance test"
}

resource "qovery_organization_member" "test" {
  organization_id = "%s"
  email           = "%s"
  role_id         = qovery_custom_role.test_member.id
}
`, getTestOrganizationID(), customRoleName, getTestOrganizationID(), email,
	)
}

// getTestOrganizationRoleID fetches a built-in role of the test organization by name.
// Built-in role names are lowercase (admin/devops/billing), so the match is case-insensitive.
// It runs before resource.Test executes PreCheck (step configs are built eagerly), so the
// env precheck is repeated here.
func getTestOrganizationRoleID(t *testing.T, roleName string) string {
	testAccPreCheck(t)
	roles, _, err := qoveryAPIClient.OrganizationMainCallsAPI.
		ListOrganizationAvailableRoles(context.TODO(), getTestOrganizationID()).
		Execute()
	if err != nil {
		t.Fatalf("failed to list organization available roles: %s", err)
	}
	for _, role := range roles.GetResults() {
		if strings.EqualFold(role.Name, roleName) {
			return role.Id
		}
	}
	t.Fatalf("no %q role found in test organization", roleName)
	return ""
}

func getOrganizationMemberImportStateId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("organization member not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["email"]), nil
	}
}

func testAccQoveryOrganizationMemberExists(resourceName string, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("organization member not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("organization_member.id not found")
		}

		_, err := qoveryServices.OrganizationMember.Get(context.TODO(), getTestOrganizationID(), email)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryOrganizationMemberDestroy(resourceName string, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("organization member not found: %s", resourceName)
		}

		_, err := qoveryServices.OrganizationMember.Get(context.TODO(), getTestOrganizationID(), email)
		if err == nil {
			return fmt.Errorf("found organization member but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted organization member: %s", err.Error())
		}
		return nil
	}
}
