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

func TestAcc_ApiToken(t *testing.T) {
	t.Parallel()
	testName := "api-token"
	adminRoleID := getTestAdminRoleID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryApiTokenDestroy("qovery_api_token.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApiTokenDefaultConfig(
					testName,
					adminRoleID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApiTokenExists("qovery_api_token.test"),
					resource.TestCheckResourceAttr("qovery_api_token.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_api_token.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_api_token.test", "description", "this is a test api token"),
					resource.TestCheckResourceAttr("qovery_api_token.test", "role_id", adminRoleID),
					resource.TestCheckResourceAttrSet("qovery_api_token.test", "token"),
				),
			},
			// Update name forces a replacement (the API has no update endpoint)
			{
				Config: testAccApiTokenDefaultConfig(
					fmt.Sprintf("%s-updated", testName),
					adminRoleID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryApiTokenExists("qovery_api_token.test"),
					resource.TestCheckResourceAttr("qovery_api_token.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttrSet("qovery_api_token.test", "token"),
				),
			},
			// Import testing: the token value cannot be retrieved from the API, so it is ignored
			{
				ResourceName:            "qovery_api_token.test",
				ImportState:             true,
				ImportStateIdFunc:       getApiTokenImportStateId("qovery_api_token.test"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccApiTokenDefaultConfig(testName string, roleID string) string {
	return fmt.Sprintf(`
resource "qovery_api_token" "test" {
  organization_id = "%s"
  name            = "%s"
  description     = "this is a test api token"
  role_id         = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), roleID,
	)
}

// getTestAdminRoleID fetches the built-in admin role of the test organization.
// API tokens require a role_id and the built-in role ids differ per organization.
// The built-in role names are lowercase (e.g. "admin"), so the match is case-insensitive.
// It runs before resource.Test executes PreCheck (step configs are built eagerly),
// so the env precheck is repeated here to fail with the standard message when the
// environment is incomplete instead of an obscure API error.
func getTestAdminRoleID(t *testing.T) string {
	testAccPreCheck(t)
	roles, _, err := qoveryAPIClient.OrganizationMainCallsAPI.
		ListOrganizationAvailableRoles(context.TODO(), getTestOrganizationID()).
		Execute()
	if err != nil {
		t.Fatalf("failed to list organization available roles: %s", err)
	}
	for _, role := range roles.GetResults() {
		if strings.EqualFold(role.Name, "admin") {
			return role.Id
		}
	}
	t.Fatal("no admin role found in test organization")
	return ""
}

func getApiTokenImportStateId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("api token not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.ID), nil
	}
}

func testAccQoveryApiTokenExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("api token not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("api_token.id not found")
		}

		_, err := qoveryServices.ApiToken.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryApiTokenDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("api token not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("api_token.id not found")
		}

		_, err := qoveryServices.ApiToken.Get(context.TODO(), getTestOrganizationID(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found api token but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted api token: %s", err.Error())
		}
		return nil
	}
}
