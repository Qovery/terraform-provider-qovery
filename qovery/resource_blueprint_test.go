//go:build integration && !unit

package qovery_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/pkg/errors"

	"github.com/qovery/terraform-provider-qovery/internal/domain/apierrors"
	"github.com/qovery/terraform-provider-qovery/internal/domain/blueprint"
)

// Helm blueprint: no cloud resources, runs on any test cluster
const testAccBlueprintVersion = "HELM/redis/8"

func TestAcc_Blueprint(t *testing.T) {
	t.Parallel()
	testName := "blueprint"
	// Chart suffixes k8s names off release name; full uuid overflows 63-char limit
	blueprintName := fmt.Sprintf("%s-bp-%s", testNamePrefix, testNameSuffix[:8])
	address := "qovery_blueprint.test"
	var blueprintID, createdServiceID string
	withoutVariables := testAccBlueprintConfig(testName, blueprintName, testAccBlueprintVariables{password: "second-password"})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryBlueprintDestroy(address),
		Steps: []resource.TestStep{
			// memory_limit is set to a value other than its catalog default (512Mi)
			{
				Config: testAccBlueprintConfig(testName, blueprintName, testAccBlueprintVariables{memoryLimit: "256Mi", password: "first-password"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryBlueprintExists(address),
					testAccCaptureResourceID(address, &blueprintID),
					resource.TestCheckResourceAttr(address, "name", blueprintName),
					resource.TestCheckResourceAttr(address, "blueprint", testAccBlueprintVersion),
					resource.TestMatchResourceAttr(address, "tag", regexp.MustCompile(`^HELM/redis/8/.+`)),
					resource.TestCheckResourceAttr(address, "icon_uri", "app://qovery-console/terraform"),
					resource.TestCheckResourceAttr(address, "variables.memory_limit", "256Mi"),
					resource.TestCheckResourceAttr(address, "secret_variables.password", "first-password"),
					resource.TestCheckResourceAttr(address, "service_type", "HELM"),
					resource.TestCheckResourceAttrSet(address, "service_id"),
					resource.TestCheckResourceAttrWith(address, "service_id", func(value string) error {
						createdServiceID = value
						return nil
					}),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "service_id", address, "service_id"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "tag", address, "tag"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "blueprint", address, "blueprint"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "service_type", address, "service_type"),
					resource.TestCheckResourceAttrPair("data.qovery_blueprint.test", "icon_uri", address, "icon_uri"),
					resource.TestCheckResourceAttr("data.qovery_blueprint.test", "variables.memory_limit", "256Mi"),
					resource.TestCheckTypeSetElemAttr("data.qovery_blueprint.test", "secret_variable_names.*", "password"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// Import records every value the API returns; secret values are not among them
			{
				ResourceName:            address,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_variables"},
			},
			// Removing variables plans the reset to the catalog default, which the apply writes
			{
				Config: withoutVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(address, tfjsonpath.New("variables"), knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(address, "variables.%"),
					resource.TestCheckResourceAttr(address, "secret_variables.password", "second-password"),
					// Updates converge the service in place: a new service_id would mean it was recreated
					resource.TestCheckResourceAttrWith(address, "service_id", func(value string) error {
						if value != createdServiceID {
							return fmt.Errorf("service_id changed from %s to %s: the update recreated the service", createdServiceID, value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("data.qovery_blueprint.test", "variables.memory_limit", "512Mi"),
				),
			},
			// A variable set from the Console shows in the plan, which reverts it...
			{
				Config: withoutVariables,
				Check: func(s *terraform.State) error {
					return testAccSetBlueprintVariableOutOfBand(s, address, "memory_limit", "768Mi")
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
				},
				ExpectNonEmptyPlan: true,
			},
			// ... the refresh stores it...
			{
				RefreshState:       true,
				Check:              resource.TestCheckResourceAttr(address, "variables.memory_limit", "768Mi"),
				ExpectNonEmptyPlan: true,
			},
			// ... and the corrective apply resets it to the catalog default.
			{
				Config: withoutVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(address, "variables.%"),
					resource.TestCheckResourceAttr("data.qovery_blueprint.test", "variables.memory_limit", "512Mi"),
				),
				ConfigPlanChecks: testAccEmptyPlanAfterApply,
			},
			// The API cannot change the icon of an existing blueprint
			{
				Config:      testAccBlueprintConfig(testName, blueprintName, testAccBlueprintVariables{password: "second-password", iconURI: "app://qovery-console/redis"}),
				ExpectError: regexp.MustCompile(`Cannot change icon_uri after creation`),
			},
		},
	})
}

type testAccBlueprintVariables struct {
	memoryLimit string
	password    string
	iconURI     string
}

func testAccBlueprintConfig(testName string, blueprintName string, v testAccBlueprintVariables) string {
	optional := ""
	if v.iconURI != "" {
		optional += fmt.Sprintf("  icon_uri = %q\n", v.iconURI)
	}
	if v.memoryLimit != "" {
		optional += fmt.Sprintf("  variables = {\n    memory_limit = %q\n  }\n", v.memoryLimit)
	}
	return fmt.Sprintf(`
%s

resource "qovery_blueprint" "test" {
  environment_id = qovery_environment.test.id
  name           = "%s"
  blueprint      = "%s"
%s
  secret_variables = {
    password = "%s"
  }
}

data "qovery_blueprint" "test" {
  id = qovery_blueprint.test.id
}
`, testAccEnvironmentDefaultConfig(testName), blueprintName, testAccBlueprintVersion, optional, v.password)
}

// testAccSetBlueprintVariableOutOfBand changes one variable the way the Qovery Console does: it
// patches only that variable, then applies the blueprint and waits for the deployment.
func testAccSetBlueprintVariableOutOfBand(s *terraform.State, resourceName string, name string, value string) error {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return fmt.Errorf("blueprint not found: %s", resourceName)
	}
	_, err := qoveryServices.Blueprint.Update(context.TODO(), rs.Primary.ID, blueprint.UpdateRequest{
		UpsertRequest: blueprint.UpsertRequest{
			Name:      rs.Primary.Attributes["name"],
			Tag:       rs.Primary.Attributes["tag"],
			IconURI:   rs.Primary.Attributes["icon_uri"],
			Variables: map[string]string{name: value},
		},
	})
	return err
}

func testAccQoveryBlueprintExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("blueprint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("blueprint.id not found")
		}
		_, err := qoveryServices.Blueprint.Get(context.TODO(), rs.Primary.ID)
		return err
	}
}

func testAccQoveryBlueprintDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("blueprint not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("blueprint.id not found")
		}

		_, err := qoveryServices.Blueprint.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("found blueprint but expected it to be deleted")
		}
		if !apierrors.IsErrNotFound(errors.Cause(err)) {
			return fmt.Errorf("unexpected error checking for deleted blueprint: %s", err.Error())
		}
		return nil
	}
}
