package qovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
)

// NOTE: skipped because organization creation is not allowed with terraform
func TestAcc_Organization(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "organization"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryOrganizationDestroy("qovery_organization.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationConfig(
					testName,
					"FREE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryOrganizationExists("qovery_organization.test"),
					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
				),
			},
			// Update name
			{
				Config: testAccOrganizationConfig(
					fmt.Sprintf("%s-updated", testName),
					"FREE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
				),
			},
			// Add description
			{
				Config: testAccOrganizationConfigWithDescription(
					fmt.Sprintf("%s-updated", testName),
					"FREE",
					"this is my description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateTestName(fmt.Sprintf("%s-updated", testName))),
					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
					resource.TestCheckResourceAttr("qovery_organization.test", "description", "this is my description"),
				),
			},
		},
	})
}

// NOTE: skipped because organization creation is not allowed with terraform
func TestAcc_OrganizationImport(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	testName := "organization-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryOrganizationDestroy("qovery_organization.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationConfig(
					testName,
					"FREE",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryOrganizationExists("qovery_organization.test"),
					resource.TestCheckResourceAttr("qovery_organization.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_organization.test", "plan", "FREE"),
					resource.TestCheckNoResourceAttr("qovery_organization.test", "description"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryOrganizationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("organization not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("organization.id not found")
		}

		_, err := qoveryServices.OrganizationService.Get(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccQoveryOrganizationDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("organization not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("organization.id not found")
		}

		_, err := qoveryServices.OrganizationService.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			// TODO: handle orga delete properly
			// return fmt.Errorf("found organization but expected it to have been deleted")
			return nil
		}
		if !apierrors.IsErrNotFound(err) {
			return fmt.Errorf("unexpected error checking for deleted organization: %s", err.Error())
		}
		return nil
	}
}

func testAccOrganizationConfig(testName string, plan string) string {
	return fmt.Sprintf(`
resource "qovery_organization" "test" {
 name = "%s"
 plan = "%s"
}
`, generateTestName(testName), plan,
	)
}

func testAccOrganizationConfigWithDescription(testName string, plan string, description string) string {
	return fmt.Sprintf(`
resource "qovery_organization" "test" {
 name = "%s"
 plan = "%s"
 description = "%s"
}
`, generateTestName(testName), plan, description)
}
