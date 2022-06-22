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
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
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
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
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

func TestAcc_ProjectWithEnvironmentVariables(t *testing.T) {
	t.Parallel()
	testName := "project-with-environment-variables"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "project_key_1", value: "project_value_1", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
						"scope": "PROJECT",
					}),
				),
			},
			// Update environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "project_key_1", value: "project_value_1_updated", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1_updated",
						"scope": "PROJECT",
					}),
				),
			},
			// Add environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "project_key_1", value: "project_value_1", scope: "PROJECT"},
						{key: "project_key_2", value: "project_value_2", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
						"scope": "PROJECT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
						"scope": "PROJECT",
					}),
				),
			},
			// Remove environment variable
			{
				Config: testAccProjectDefaultConfigWithEnvironmentVariables(
					testName,
					[]environmentVariable{
						{key: "project_key_2", value: "project_value_2", scope: "PROJECT"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "environment_variables.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
						"scope": "PROJECT",
					}),
				),
			},
			//Check Import
			{
				ResourceName:      "qovery_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ProjectWithSecrets(t *testing.T) {
	t.Parallel()
	testName := "project-with-secrets"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccQoveryProjectDestroy("qovery_project.test"),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "project_key_1", value: "project_value_1"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
					}),
				),
			},
			// Update secrets
			{
				Config: testAccProjectDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "project_key_1", value: "project_value_1_updated"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1_updated",
					}),
				),
			},
			// Add secrets
			{
				Config: testAccProjectDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "project_key_1", value: "project_value_1"},
						{key: "project_key_2", value: "project_value_2"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "project_key_1",
						"value": "project_value_1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
					}),
				),
			},
			// Remove secrets
			{
				Config: testAccProjectDefaultConfigWithSecrets(
					testName,
					[]environmentSecret{
						{key: "project_key_2", value: "project_value_2"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccQoveryProjectExists("qovery_project.test"),
					resource.TestCheckResourceAttr("qovery_project.test", "organization_id", getTestOrganizationID()),
					resource.TestCheckResourceAttr("qovery_project.test", "name", generateTestName(testName)),
					resource.TestCheckResourceAttr("qovery_project.test", "description", ""),
					resource.TestCheckNoResourceAttr("qovery_project.test", "environment_variables.0"),
					resource.TestCheckNoResourceAttr("qovery_project.test", "built_in_environment_variables.0"),
					resource.TestCheckTypeSetElemNestedAttrs("qovery_project.test", "secrets.*", map[string]string{
						"key":   "project_key_2",
						"value": "project_value_2",
					}),
				),
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
					resource.TestCheckNoResourceAttr("qovery_project.test", "secrets.0"),
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

func testAccProjectDefaultConfigWithEnvironmentVariables(testName string, environmentVariables []environmentVariable) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  environment_variables = %s
}
`, getTestOrganizationID(), generateTestName(testName), convertEnvVarsToString(environmentVariables),
	)
}

func testAccProjectDefaultConfigWithSecrets(testName string, secrets []environmentSecret) string {
	return fmt.Sprintf(`
resource "qovery_project" "test" {
  organization_id = "%s"
  name = "%s"
  secrets = %s
}
`, getTestOrganizationID(), generateTestName(testName), convertEnvSecretsToString(secrets),
	)
}

type environmentVariable struct {
	key   string
	value string
	scope string
}

func convertEnvVarsToString(environmentVariables []environmentVariable) string {
	vars := make([]string, 0, len(environmentVariables))
	for _, e := range environmentVariables {
		if e.scope == "" {
			vars = append(vars, fmt.Sprintf(`{key: "%s", value: "%s"}`, e.key, e.value))
		} else {
			vars = append(vars, fmt.Sprintf(`{key: "%s", value: "%s", scope: "%s"}`, e.key, e.value, e.scope))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(vars, ","))
}

type environmentSecret struct {
	key   string
	value string
}

func convertEnvSecretsToString(secrets []environmentSecret) string {
	vars := make([]string, 0, len(secrets))
	for _, s := range secrets {
		vars = append(vars, fmt.Sprintf(`{key: "%s", value: "%s"}`, s.key, s.value))
	}
	return fmt.Sprintf("[%s]", strings.Join(vars, ","))
}
