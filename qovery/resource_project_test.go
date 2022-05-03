package qovery_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/qovery/terraform-provider-qovery/client/apierrors"
)

func TestAcc_Project(t *testing.T) {
	t.Parallel()
	testName := "project"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
				),
			},
			// Add description
			{
				Config: testAccProjectDefaultConfigWithDescription(
					testName,
					"this is a description",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", "this is a description"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
				),
			},
			// Add environment variables
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVars(
					testName,
					map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
				),
			},
			// Update environment variables
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVars(
					testName,
					map[string]string{
						"key1": "value1-updated",
						"key2": "value2-updated",
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key1",
						"value": "value1-updated",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "key2",
						"value": "value2-updated",
					}),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ProjectImport(t *testing.T) {
	t.Parallel()
	testName := "project-import"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfig(
					testName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
				),
			},
			// Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccQoveryProjectExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("project not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project.id not found")
		}

		_, apiErr := apiClient.GetProject(context.TODO(), rs.Primary.ID)
		if apiErr != nil {
			return apiErr
		}
		return nil
	}
}

func testAccQoveryProjectDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("project not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project.id not found")
		}

		_, apiErr := apiClient.GetProject(context.TODO(), rs.Primary.ID)
		if apiErr == nil {
			return fmt.Errorf("found project but expected it to be deleted")
		}
		if !apierrors.IsNotFound(apiErr) {
			return fmt.Errorf("unexpected error checking for deleted project: %s", apiErr.Summary())
		}
		return nil
	}
}

func testAccProjectDefaultConfig(testName string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
}
`, getTestOrganizationID(), generateTestName(testName),
	)
}

func testAccProjectDefaultConfigWithDescription(testName string, description string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  description = "%s"
}
`, getTestOrganizationID(), generateTestName(testName), description,
	)
}

func testAccProjectDefaultConfigWithEnvironmentVars(testName string, environmentVariables map[string]string) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  environment_variables = %s
}
`, getTestOrganizationID(), generateTestName(testName), convertEnvVarsToString(environmentVariables),
	)
}

func convertEnvVarsToString(environmentVariables map[string]string) string {
	vars := make([]string, 0, len(environmentVariables))
	for key, value := range environmentVariables {
		vars = append(vars, fmt.Sprintf(`{key: "%s", value: "%s"}`, key, value))
	}
	return fmt.Sprintf("[%s]", strings.Join(vars, ","))
}
