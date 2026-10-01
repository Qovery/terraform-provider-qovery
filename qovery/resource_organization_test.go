//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
)

// The test organization is a shared fixture: its name, plan and description are fixed.
const (
	testAccOrganizationName        = "Q Sandbox"
	testAccOrganizationPlan        = "ENTERPRISE"
	testAccOrganizationDescription = "Organization for team's test"
)

// Organizations cannot be created or deleted with terraform: the test imports the test organization,
// updates it, restores it, then removes it from the state without destroying it.
// Not parallel: it changes the description that TestAcc_OrganizationDataSource reads. Go runs the
// sequential tests of a package before it resumes the parallel ones.
func TestAcc_Organization(t *testing.T) {
	address := "qovery_organization.test"
	organizationID := getTestOrganizationID()
	updatedDescription := generateTestName("organization-description")
	var before qovery.Organization

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			before = testAccGetOrganization(t, organizationID)
			t.Cleanup(func() { testAccRestoreOrganizationDescription(t, organizationID, before) })
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Import changes nothing
			{
				Config: testAccOrganizationImportConfig(organizationID) + testAccOrganizationConfig(testAccOrganizationDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "id", organizationID),
					resource.TestCheckResourceAttr(address, "name", testAccOrganizationName),
					resource.TestCheckResourceAttr(address, "plan", testAccOrganizationPlan),
					resource.TestCheckResourceAttr(address, "description", testAccOrganizationDescription),
				),
			},
			// An update keeps the fields the provider does not manage, and succeeds although admin emails are set
			{
				Config: testAccOrganizationConfig(updatedDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "description", updatedDescription),
					testAccCheckOrganizationUnmanagedFieldsKept(organizationID, &before),
				),
			},
			// Restoring the description keeps them too
			{
				Config: testAccOrganizationConfig(testAccOrganizationDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "description", testAccOrganizationDescription),
					testAccCheckOrganizationUnmanagedFieldsKept(organizationID, &before),
				),
			},
			// Removing the description plans its removal, and the apply clears it
			{
				Config: testAccOrganizationConfigWithoutDescription(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(address, tfjsonpath.New("description"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(address, "description"),
					testAccCheckOrganizationDescription(organizationID, nil),
				),
			},
			// A description set outside Terraform shows up in the plan...
			{
				Config: testAccOrganizationConfigWithoutDescription(),
				Check: func(_ *terraform.State) error {
					return testAccSetOrganizationDescriptionOutOfBand(organizationID, before)
				},
				ExpectNonEmptyPlan: true,
			},
			// ... and the refresh stores it
			{
				RefreshState:       true,
				Check:              resource.TestCheckResourceAttr(address, "description", testAccOrganizationDescription),
				ExpectNonEmptyPlan: true,
			},
			// Declaring the description the organization holds plans nothing
			{
				Config: testAccOrganizationConfig(testAccOrganizationDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			// Import records the remote description
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete fails by design, so the organization leaves the state without being destroyed
			{
				Config: testAccOrganizationRemovedConfig(),
				Check: func(s *terraform.State) error {
					if _, ok := s.RootModule().Resources[address]; ok {
						return fmt.Errorf("%s is still in the state", address)
					}
					return nil
				},
			},
		},
	})
}

// TestAcc_OrganizationUpgradeFrom0x imports the test organization with the last 0.x release, then
// plans with this provider, which upgrades the 0.x state: the plan must be empty. Not parallel, for
// the reason given on TestAcc_Organization.
func TestAcc_OrganizationUpgradeFrom0x(t *testing.T) {
	organizationID := getTestOrganizationID()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccGetOrganization(t, organizationID)
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: testAccLastProvider0xFromRegistry,
				Config:            testAccOrganizationImportConfig(organizationID) + testAccOrganizationConfig(testAccOrganizationDescription),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccOrganizationConfig(testAccOrganizationDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			// Delete fails by design, so the organization leaves the state without being destroyed
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccOrganizationRemovedConfig(),
			},
		},
	})
}

// testAccGetOrganization reads the organization and fails the test if it is not the expected fixture,
// before any step can write to it.
func testAccGetOrganization(t *testing.T, organizationID string) qovery.Organization {
	t.Helper()
	orga, _, err := qoveryAPIClient.OrganizationMainCallsAPI.GetOrganization(context.Background(), organizationID).Execute()
	if err != nil {
		t.Fatalf("failed to read organization %s: %s", organizationID, err)
	}
	if orga.GetName() != testAccOrganizationName || string(orga.GetPlan()) != testAccOrganizationPlan || orga.GetDescription() != testAccOrganizationDescription {
		t.Fatalf("organization %s is not the expected fixture: got name %q, plan %q, description %q", organizationID, orga.GetName(), orga.GetPlan(), orga.GetDescription())
	}
	if len(orga.GetAdminEmails()) == 0 {
		t.Fatalf("organization %s has no admin emails: the test needs some to cover the update rejected when they are missing", organizationID)
	}
	return *orga
}

// testAccRestoreOrganizationDescription puts back the description of before when a failed step left another one.
func testAccRestoreOrganizationDescription(t *testing.T, organizationID string, before qovery.Organization) {
	t.Helper()
	orga, err := qoveryServices.Organization.Get(context.Background(), organizationID)
	if err != nil {
		t.Errorf("failed to read organization %s to restore its description: %s", organizationID, err)
		return
	}
	if orga.Description != nil && *orga.Description == before.GetDescription() {
		return
	}
	if _, err := qoveryServices.Organization.Update(context.Background(), organizationID, organization.UpdateRequest{
		Name:        before.GetName(),
		Description: before.Description.Get(),
	}); err != nil {
		t.Errorf("failed to restore the description of organization %s: %s", organizationID, err)
	}
}

// testAccSetOrganizationDescriptionOutOfBand puts back the description of before through the API,
// bypassing Terraform. The edit resends the fields of the organization, which the API replaces.
func testAccSetOrganizationDescriptionOutOfBand(organizationID string, before qovery.Organization) error {
	return testAccEditServiceOutOfBand("/organization/"+organizationID, nil, func(orga map[string]any) {
		orga["description"] = before.GetDescription()
	})
}

// testAccCheckOrganizationDescription checks the description the API holds, nil meaning none.
func testAccCheckOrganizationDescription(organizationID string, expected *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		orga, _, err := qoveryAPIClient.OrganizationMainCallsAPI.GetOrganization(context.Background(), organizationID).Execute()
		if err != nil {
			return fmt.Errorf("failed to read organization %s: %w", organizationID, err)
		}
		got := orga.Description.Get()
		if (got == nil) != (expected == nil) || (got != nil && *got != *expected) {
			return fmt.Errorf("organization %s has description %v, expected %v", organizationID, got, expected)
		}
		return nil
	}
}

// testAccCheckOrganizationUnmanagedFieldsKept checks that the fields the provider does not manage still have the values of before.
func testAccCheckOrganizationUnmanagedFieldsKept(organizationID string, before *qovery.Organization) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		after, _, err := qoveryAPIClient.OrganizationMainCallsAPI.GetOrganization(context.Background(), organizationID).Execute()
		if err != nil {
			return fmt.Errorf("failed to read organization %s: %w", organizationID, err)
		}
		fields := []struct {
			name          string
			before, after string
		}{
			{name: "logo_url", before: before.GetLogoUrl(), after: after.GetLogoUrl()},
			{name: "website_url", before: before.GetWebsiteUrl(), after: after.GetWebsiteUrl()},
			{name: "icon_url", before: before.GetIconUrl(), after: after.GetIconUrl()},
			{name: "repository", before: before.GetRepository(), after: after.GetRepository()},
		}
		for _, field := range fields {
			if field.before != field.after {
				return fmt.Errorf("organization %s changed from %q to %q", field.name, field.before, field.after)
			}
		}
		if !slices.Equal(before.GetAdminEmails(), after.GetAdminEmails()) {
			return fmt.Errorf("organization admin_emails changed from %d to %d entries", len(before.GetAdminEmails()), len(after.GetAdminEmails()))
		}
		return nil
	}
}

func testAccOrganizationImportConfig(organizationID string) string {
	return fmt.Sprintf(`
import {
  to = qovery_organization.test
  id = "%s"
}
`, organizationID)
}

func testAccOrganizationConfig(description string) string {
	return fmt.Sprintf(`
resource "qovery_organization" "test" {
  name        = "%s"
  plan        = "%s"
  description = "%s"
}
`, testAccOrganizationName, testAccOrganizationPlan, description)
}

func testAccOrganizationConfigWithoutDescription() string {
	return fmt.Sprintf(`
resource "qovery_organization" "test" {
  name = "%s"
  plan = "%s"
}
`, testAccOrganizationName, testAccOrganizationPlan)
}

func testAccOrganizationRemovedConfig() string {
	return `
removed {
  from = qovery_organization.test

  lifecycle {
    destroy = false
  }
}
`
}
